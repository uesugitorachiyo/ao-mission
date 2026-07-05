package mission

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissionLifecycleAndSnapshot(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	r, err := s.Start("build a long-running atlas workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	if r.CurrentRoute != "ao-atlas" {
		t.Fatalf("route=%s", r.CurrentRoute)
	}
	r, err = Continue(s, r.MissionID, ContinueOptions{UntilDone: true, MaxIterations: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Steps) != 10 {
		t.Fatalf("steps=%d", len(r.Steps))
	}
	if r.GoalLease == nil || r.GoalLease.MinNodes != 10 || r.GoalLease.ReturnOnlyWhen == "" {
		t.Fatalf("missing long-run lease: %+v", r.GoalLease)
	}
	if len(r.Checkpoints) != len(r.Steps) {
		t.Fatalf("checkpoints=%d steps=%d", len(r.Checkpoints), len(r.Steps))
	}
	snap := Snapshot(r)
	if snap.SafeToExecute {
		t.Fatal("snapshot must not be safe to execute")
	}
	if snap.ExecutesWork || snap.ApprovesWork || snap.ProviderCalls {
		t.Fatal("authority boundary widened")
	}
}

func TestContinuePersistsEventLoopDecision(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("build a long-running atlas workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Continue(s, rec.MissionID, ContinueOptions{UntilDone: true, MaxIterations: 3}); err != nil {
		t.Fatal(err)
	}
	decision, err := s.LoadEventLoopDecision(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Schema != EventLoopDecisionSchema || decision.MissionID != rec.MissionID {
		t.Fatalf("bad event loop decision: %+v", decision)
	}
	if decision.Status != "handoff_required" || decision.Route != "ao-atlas" || decision.Iteration != 3 {
		t.Fatalf("unexpected event loop decision: %+v", decision)
	}
	if decision.ExecutesWork || decision.ApprovesWork || decision.MutatesRepositories {
		t.Fatalf("event loop widened authority: %+v", decision)
	}
}

func TestContinueUntilDoneDoesNotStopAfterOneHandoff(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("supervise long-running atlas workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	continued, err := Continue(s, rec.MissionID, ContinueOptions{UntilDone: true, MaxIterations: 4})
	if err != nil {
		t.Fatal(err)
	}
	if len(continued.Steps) != 4 {
		t.Fatalf("until-done stopped early after %d steps", len(continued.Steps))
	}
	if continued.ReturnGate == nil || continued.ReturnGate.FinalResponseAllowed {
		t.Fatalf("premature final return should be denied: %+v", continued.ReturnGate)
	}
}

func TestContinueWritesCheckpointBundleAndDoctorSupervisorHealth(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("supervise checkpointed atlas workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	continued, err := Continue(s, rec.MissionID, ContinueOptions{UntilDone: true, MaxIterations: 3})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := s.LoadCheckpointBundle(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Schema != CheckpointBundleSchema || bundle.CheckpointCount != 3 || bundle.ExecutesWork || bundle.ApprovesWork || bundle.MutatesRepositories {
		t.Fatalf("bad checkpoint bundle: %+v", bundle)
	}
	if bundle.ReturnGate == nil || bundle.ReturnGate.FinalResponseAllowed {
		t.Fatalf("checkpoint bundle should carry early-return denial: %+v", bundle.ReturnGate)
	}
	doctor := BuildMissionDoctorReadback(s)
	if doctor.Schema != "ao.mission.doctor-readback.v0.1" || doctor.LeaseCount != 1 || doctor.FreshCheckpoints != 1 || doctor.EarlyReturnRisks != 1 {
		t.Fatalf("doctor missing supervisor health: %+v continued=%+v", doctor, continued)
	}
	for _, want := range []string{"lease_health_checked", "checkpoint_freshness_checked", "stale_route_reconciliation_checked", "early_return_risk_checked"} {
		if !stringSliceContains(doctor.Checks, want) {
			t.Fatalf("doctor missing check %q: %+v", want, doctor)
		}
	}
}

func TestSchedulerFailsClosedWhenCronMissing(t *testing.T) {
	old := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", old) })
	os.Setenv("PATH", t.TempDir())
	rb := ScheduleReadback("mission-x", "1m", true)
	if rb.Status != "blocked" {
		t.Fatalf("status=%s", rb.Status)
	}
	if !strings.Contains(rb.Reason, "missing") {
		t.Fatalf("reason=%s", rb.Reason)
	}
}

func TestSchedulerReadbackImportRecordsWakeupOnlyEvidence(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("schedule long-running workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "scheduler-readback.json")
	body := `{"schema":"ao.mission.scheduler-readback.v0.1","mission_id":"` + rec.MissionID + `","status":"ready","scheduler":"codex-cron","event_loop":true,"reason":"fixture wakeup only","generated_at_utc":"2026-07-03T00:00:00Z"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "scheduler-readback", path); err != nil {
		t.Fatal(err)
	}
	updated, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Evidence.SchedulerReadback == nil || updated.Evidence.SchedulerReadback.Scheduler != "codex-cron" {
		t.Fatalf("scheduler evidence missing: %+v", updated.Evidence.SchedulerReadback)
	}
	if updated.Evidence.SchedulerReadback.ExecutesWork || updated.CurrentPhase != "scheduler_readback_recorded" {
		t.Fatalf("scheduler import widened authority or wrong phase: %+v", updated)
	}
}

func TestSchedulerReadbackImportRejectsExecutionAuthority(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("schedule long-running workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "scheduler-readback.json")
	body := `{"schema":"ao.mission.scheduler-readback.v0.1","mission_id":"` + rec.MissionID + `","status":"ready","scheduler":"codex-cron","event_loop":true,"executes_work":true,"reason":"unsafe fixture","generated_at_utc":"2026-07-03T00:00:00Z"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "scheduler-readback", path); err == nil || !strings.Contains(err.Error(), "executes_work") {
		t.Fatalf("expected scheduler authority rejection, got %v", err)
	}
	updated, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Evidence.SchedulerReadback != nil {
		t.Fatalf("unsafe scheduler readback was recorded: %+v", updated.Evidence.SchedulerReadback)
	}
}

func TestRouterSendsConcreteBatchSupervisionToAtlasNotBlueprint(t *testing.T) {
	decision := DecideRoute("mission-route", "supervise twenty bounded implementation and evidence nodes", nil)
	if decision.Route != "ao-atlas" {
		t.Fatalf("concrete long-run batch should route to Atlas, got %+v", decision)
	}
	underspecified := DecideRoute("mission-route", "figure out", nil)
	if underspecified.Route != "ao-blueprint" {
		t.Fatalf("underspecified objective should still route to Blueprint, got %+v", underspecified)
	}
}

func TestTelegramIntentOnly(t *testing.T) {
	rb := HandleTelegramCommand(TelegramCommand{ChatID: "1001", Command: "/continue", Role: "admin"}, map[string]string{"1001": "admin"})
	if rb.Status != "intent_recorded" || rb.MutationAuthority {
		t.Fatalf("bad readback: %+v", rb)
	}
	denied := HandleTelegramCommand(TelegramCommand{ChatID: "9", Command: "/continue", Role: "admin"}, map[string]string{})
	if denied.Status != "denied" {
		t.Fatalf("want denied")
	}
}
func TestA2AAgentCardIntentOnly(t *testing.T) {
	card := AgentCard()
	if card.MutationAuthority {
		t.Fatal("a2a must not grant mutation authority")
	}
	task := A2ATaskFor("mission.start")
	if task.MutationAuthority || task.Status != "intent_recorded" {
		t.Fatalf("bad task %+v", task)
	}
}
func TestCLIStartStatusNext(t *testing.T) {
	dir := t.TempDir()
	old := os.Getenv("AO_MISSION_HOME")
	t.Cleanup(func() { os.Setenv("AO_MISSION_HOME", old) })
	os.Setenv("AO_MISSION_HOME", dir)
	var out, errb bytes.Buffer
	if code := Run([]string{"init"}, &out, &errb); code != 0 {
		t.Fatalf("init: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"start", "small objective"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"next", "--mission", rec.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("next: %s", errb.String())
	}
	var d RouteDecision
	if err := json.Unmarshal(out.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d.SafeToExecute {
		t.Fatal("next must not be safe to execute")
	}
	if _, err := os.Stat(filepath.Join(dir, "missions", rec.MissionID+".json")); err != nil {
		t.Fatal(err)
	}
}
func TestPublicSafeTextRejectsSecrets(t *testing.T) {
	if ValidatePublicSafeText("tok"+"en: abc") == nil {
		t.Fatal("expected token rejection")
	}
	if err := ValidatePublicSafeText("safe fixture with redacted token example"); err != nil {
		t.Fatal(err)
	}
}

func TestGlobalHomeAndFinalRollup(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "atlas workgraph objective"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "final", "rollup", "--mission", rec.MissionID}, &out, &errb); code != 0 {
		t.Fatalf("rollup: %s", errb.String())
	}
	var rollup FinalRollup
	if err := json.Unmarshal(out.Bytes(), &rollup); err != nil {
		t.Fatal(err)
	}
	if rollup.SafeToExecute || rollup.ExecutesWork {
		t.Fatal("final rollup widened authority")
	}
}

func TestValidateContractAndImports(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("build atlas workgraph")
	if err != nil {
		t.Fatal(err)
	}
	authPath := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(authPath, []byte(`{"schema":"ao.blueprint.build-authorization.v0.1","project_id":"demo","status":"ready","next_allowed_action":"ao-atlas"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateContractFile(filepath.Join("..", "..", "examples", "valid", "mission-record.json")); err != nil {
		t.Fatal(err)
	}
	rb, err := ImportArtifact(s, rec.MissionID, "blueprint-authorization", authPath)
	if err != nil {
		t.Fatal(err)
	}
	if rb.SafeToExecute || rb.Kind != "blueprint-authorization" {
		t.Fatalf("bad readback: %+v", rb)
	}
	updated, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.ArtifactRefs) != 1 {
		t.Fatalf("artifact refs=%d", len(updated.ArtifactRefs))
	}
}

func TestContractValidationRejectsSchemaTypeMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad-record.json")
	body := `{"schema":"ao.mission.record.v0.1","mission_id":42,"objective_digest":"sha256:abc","status":"active","created_at_utc":"2026-07-03T00:00:00Z","current_route":"ao-blueprint"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := ValidateContractFile(path)
	if err == nil {
		t.Fatal("expected schema type mismatch")
	}
	if result.Status != "blocked" || !strings.Contains(strings.Join(result.Blockers, ";"), "mission_id") {
		t.Fatalf("unexpected validation result: %+v", result)
	}
}

func TestSupervisorV03ContractsValidate(t *testing.T) {
	dir := t.TempDir()
	fixtures := map[string]any{
		"goal-lease.json": GoalLease{
			Schema:           GoalLeaseSchema,
			MinNodes:         10,
			MinMinutes:       120,
			MaxMinutes:       180,
			MaxIterations:    20,
			ReturnOnlyWhen:   defaultReturnOnlyWhen,
			CheckpointPolicy: defaultCheckpointPolicy,
		},
		"checkpoint.json": MissionCheckpoint{
			Schema:          MissionCheckpointSchema,
			MissionID:       "mission-contract",
			Sequence:        1,
			Iteration:       1,
			Route:           "ao-atlas",
			Phase:           "handoff_required",
			Result:          "handoff_required",
			ExactNextAction: "send objective to AO Atlas for workgraph sequencing",
			ResumeCommand:   "ao-mission continue --mission mission-contract --until-done --max-iterations 10",
			GeneratedAtUTC:  "2026-07-04T00:00:00Z",
		},
		"return-gate.json": ReturnGate{
			Schema:               ReturnGateSchema,
			MissionID:            "mission-contract",
			Status:               "early_return_denied",
			FinalResponseAllowed: false,
			Reason:               "lease minimum unmet",
			CompletedNodes:       1,
			MinNodes:             10,
			ReadyNodesRemaining:  1,
			HardBlocker:          false,
			ExactNextAction:      "continue mission",
			GeneratedAtUTC:       "2026-07-04T00:00:00Z",
		},
		"route-reconciliation.json": RouteReconciliation{
			Schema:                RouteReconciliationSchema,
			MissionID:             "mission-contract",
			Status:                "ready",
			CurrentRoute:          "ao-atlas",
			LatestRoute:           "ao-atlas",
			AtlasReadyNodes:       1,
			CommandReadbackBound:  false,
			PromoterReadbackBound: false,
			ExactNextAction:       "continue from latest exact next action",
			GeneratedAtUTC:        "2026-07-04T00:00:00Z",
		},
		"checkpoint-bundle.json": MissionCheckpointBundle{
			Schema:              CheckpointBundleSchema,
			MissionID:           "mission-contract",
			Status:              "ready",
			CheckpointCount:     1,
			ResumePrompt:        "ao-mission continue --mission mission-contract --until-done --max-iterations 10",
			SafeToExecute:       false,
			ExecutesWork:        false,
			ApprovesWork:        false,
			MutatesRepositories: false,
			GeneratedAtUTC:      "2026-07-04T00:00:00Z",
		},
	}
	for name, fixture := range fixtures {
		path := filepath.Join(dir, name)
		writeJSONForTest(t, path, fixture)
		if result, err := ValidateContractFile(path); err != nil {
			t.Fatalf("%s failed contract validation: %v %+v", name, err, result)
		}
	}
}

func TestMissionListInspectCommandStatusAndArtifactManifest(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "first objective"}, &out, &errb); code != 0 {
		t.Fatalf("start first: %s", errb.String())
	}
	var first Record
	if err := json.Unmarshal(out.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "start", "second objective"}, &out, &errb); code != 0 {
		t.Fatalf("start second: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "list", "--json"}, &out, &errb); code != 0 {
		t.Fatalf("mission list: %s", errb.String())
	}
	var list []Record
	if err := json.Unmarshal(out.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("mission list len=%d", len(list))
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "list", "--route", first.CurrentRoute, "--status", first.Status, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("mission list filters: %s", errb.String())
	}
	var filtered []Record
	if err := json.Unmarshal(out.Bytes(), &filtered); err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 2 {
		t.Fatalf("filtered mission list len=%d", len(filtered))
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "list", "--route", "complete", "--json"}, &out, &errb); code != 0 {
		t.Fatalf("mission list empty filter: %s", errb.String())
	}
	filtered = nil
	if err := json.Unmarshal(out.Bytes(), &filtered); err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 0 {
		t.Fatalf("empty filtered mission list len=%d", len(filtered))
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "inspect", "--mission", first.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("mission inspect: %s", errb.String())
	}
	var inspected Record
	if err := json.Unmarshal(out.Bytes(), &inspected); err != nil {
		t.Fatal(err)
	}
	if inspected.MissionID != first.MissionID {
		t.Fatalf("inspect mission=%s", inspected.MissionID)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "command", "status", "--mission", first.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("command status: %s", errb.String())
	}
	var status CommandStatus
	if err := json.Unmarshal(out.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.MissionID != first.MissionID || status.ExecutesWork || status.ApprovesWork || status.MutatesRepositories {
		t.Fatalf("bad command status: %+v", status)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "artifacts", "manifest", "--mission", first.MissionID}, &out, &errb); code != 0 {
		t.Fatalf("artifact manifest: %s", errb.String())
	}
	var manifest ArtifactManifest
	if err := json.Unmarshal(out.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.ManifestDigest == "" || manifest.Signature == "" || manifest.ExecutesWork || manifest.ApprovesWork {
		t.Fatalf("bad artifact manifest: %+v", manifest)
	}
}

func TestImportAtlasWorkgraphCountsAndFoundryFinalRollupCompletion(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("build atlas workgraph")
	if err != nil {
		t.Fatal(err)
	}
	workgraphPath := filepath.Join(dir, "workgraph.json")
	workgraph := `{"schema":"ao.atlas.workgraph.v0.1","nodes":[{"node_id":"n1","status":"ready"},{"node_id":"n2","status":"blocked"},{"node_id":"n3","status":"completed"}]}`
	if err := os.WriteFile(workgraphPath, []byte(workgraph), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "atlas-workgraph", workgraphPath); err != nil {
		t.Fatal(err)
	}
	updated, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Evidence.AtlasWorkgraph == nil || updated.Evidence.AtlasWorkgraph.Total != 3 || updated.Evidence.AtlasWorkgraph.Ready != 1 {
		t.Fatalf("bad workgraph counts: %+v", updated.Evidence.AtlasWorkgraph)
	}
	rollupPath := filepath.Join(dir, "foundry-final-rollup.json")
	rollup := `{"schema":"ao.foundry.final-rollup.v0.1","status":"completed","completed_nodes":3,"total_nodes":3}`
	if err := os.WriteFile(rollupPath, []byte(rollup), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "foundry-final-rollup", rollupPath); err != nil {
		t.Fatal(err)
	}
	done, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != "done" || done.CurrentPhase != "complete" {
		t.Fatalf("mission not complete: %+v", done)
	}
	if done.Evidence.FoundryRollup == nil || done.Evidence.FoundryRollup.CompletedNodes != 3 {
		t.Fatalf("bad rollup evidence: %+v", done.Evidence.FoundryRollup)
	}
}

func TestImportAtlasRecommendationReadbackCompletesMission(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("import completed Atlas recommendation wave")
	if err != nil {
		t.Fatal(err)
	}
	readbackPath := filepath.Join(dir, "recommendation-readback.json")
	readback := `{
		"schema":"ao.atlas.recommendation-readback.v0.1",
		"status":"completed",
		"total_nodes":40,
		"completed_nodes":40,
		"ready_nodes":0,
		"checkpoint_count":40,
		"elapsed_minutes":491,
		"min_minutes_met":true,
		"lease_time_status":"minimum_minutes_met",
		"return_gate_status":"final_response_allowed",
		"final_response_allowed":true,
		"safe_to_execute":false,
		"executes_work":false,
		"approves_work":false,
		"mutates_repositories":false,
		"provider_calls":false,
		"release_or_publish":false,
		"credential_use":false,
		"direct_main_mutation":false,
		"concurrent_mutation":false,
		"exact_next_action":"Finalize AO Atlas long-run wave with Promoter, Command, and public-safety readbacks."
	}`
	if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
		t.Fatal(err)
	}
	importReadback, err := ImportArtifact(s, rec.MissionID, "atlas-recommendation-readback", readbackPath)
	if err != nil {
		t.Fatal(err)
	}
	if importReadback.SafeToExecute || importReadback.ExecutesWork || importReadback.ApprovesWork {
		t.Fatalf("import widened authority: %+v", importReadback)
	}
	done, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != "done" || done.CurrentRoute != "complete" || done.CurrentPhase != "complete" {
		t.Fatalf("completed Atlas readback should close mission: %+v", done)
	}
	if done.Evidence.AtlasRecommendation == nil || done.Evidence.AtlasRecommendation.CompletedNodes != 40 {
		t.Fatalf("Atlas recommendation evidence missing: %+v", done.Evidence.AtlasRecommendation)
	}
	if done.ReturnGate == nil || !done.ReturnGate.FinalResponseAllowed || done.ReturnGate.ReadyNodesRemaining != 0 {
		t.Fatalf("terminal return gate not allowed: %+v", done.ReturnGate)
	}
}

func TestCLIImportsAtlasRecommendationReadback(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "import completed Atlas recommendation wave"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	readbackPath := filepath.Join(dir, "recommendation-readback.json")
	readback := `{"schema":"ao.atlas.recommendation-readback.v0.1","status":"completed","total_nodes":40,"completed_nodes":40,"ready_nodes":0,"checkpoint_count":40,"elapsed_minutes":491,"min_minutes_met":true,"lease_time_status":"minimum_minutes_met","return_gate_status":"final_response_allowed","final_response_allowed":true,"safe_to_execute":false,"executes_work":false,"approves_work":false,"mutates_repositories":false,"provider_calls":false,"release_or_publish":false,"credential_use":false,"direct_main_mutation":false,"concurrent_mutation":false,"exact_next_action":"Finalize AO Atlas long-run wave."}`
	if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "import", "atlas-recommendation-readback", "--mission", rec.MissionID, "--path", readbackPath}, &out, &errb); code != 0 {
		t.Fatalf("import: %s", errb.String())
	}
	var importReadback ImportReadback
	if err := json.Unmarshal(out.Bytes(), &importReadback); err != nil {
		t.Fatal(err)
	}
	if importReadback.Kind != "atlas-recommendation-readback" || importReadback.ExecutesWork || importReadback.ApprovesWork {
		t.Fatalf("bad import readback: %+v", importReadback)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "inspect", "--mission", rec.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("inspect: %s", errb.String())
	}
	var done Record
	if err := json.Unmarshal(out.Bytes(), &done); err != nil {
		t.Fatal(err)
	}
	if done.Status != "done" || done.Evidence.AtlasRecommendation == nil || done.Evidence.AtlasRecommendation.TotalNodes != 40 {
		t.Fatalf("CLI import did not complete mission: %+v", done)
	}
}

func TestImportAtlasRecommendationReadbackDeniesFinalWhenLeaseMinimumUnmet(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("import short Atlas recommendation wave")
	if err != nil {
		t.Fatal(err)
	}
	readbackPath := filepath.Join(dir, "recommendation-readback-short.json")
	readback := `{
		"schema":"ao.atlas.recommendation-readback.v0.1",
		"status":"completed",
		"total_nodes":40,
		"completed_nodes":40,
		"ready_nodes":0,
		"checkpoint_count":40,
		"elapsed_minutes":22,
		"min_minutes_met":false,
		"lease_time_status":"minimum_minutes_unmet",
		"return_gate_status":"blocked_minimum_minutes_unmet",
		"final_response_allowed":false,
		"safe_to_execute":false,
		"executes_work":false,
		"approves_work":false,
		"mutates_repositories":false,
		"provider_calls":false,
		"release_or_publish":false,
		"credential_use":false,
		"direct_main_mutation":false,
		"concurrent_mutation":false,
		"exact_next_action":"Continue AO Atlas wave until the minimum lease duration is met."
	}`
	if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "atlas-recommendation-readback", readbackPath); err != nil {
		t.Fatal(err)
	}
	continued, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if continued.Status == "done" || continued.CurrentRoute != "ao-atlas" {
		t.Fatalf("short Atlas readback should keep Mission routed to Atlas: %+v", continued)
	}
	if continued.ReturnGate == nil || continued.ReturnGate.FinalResponseAllowed {
		t.Fatalf("short Atlas readback should deny final response: %+v", continued.ReturnGate)
	}
	if !strings.Contains(continued.ReturnGate.Reason, "blocked_minimum_minutes_unmet") {
		t.Fatalf("return gate should carry Atlas lease blocker, got %+v", continued.ReturnGate)
	}
	if !strings.Contains(continued.ReturnGate.ExactNextAction, "minimum lease") {
		t.Fatalf("return gate should preserve Atlas lease continuation action, got %+v", continued.ReturnGate)
	}
}

func TestImportAtlasRecommendationReadbackRejectsAuthorityAdvanceClaim(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("reject unsafe Atlas recommendation readback")
	if err != nil {
		t.Fatal(err)
	}
	readbackPath := filepath.Join(dir, "recommendation-readback-authority.json")
	readback := `{"schema":"ao.atlas.recommendation-readback.v0.1","status":"completed","total_nodes":40,"completed_nodes":40,"ready_nodes":0,"checkpoint_count":40,"elapsed_minutes":491,"min_minutes_met":true,"lease_time_status":"minimum_minutes_met","return_gate_status":"final_response_allowed","final_response_allowed":true,"claims_authority_advance":true,"safe_to_execute":false,"executes_work":false,"approves_work":false,"mutates_repositories":false,"provider_calls":false,"release_or_publish":false,"credential_use":false,"direct_main_mutation":false,"concurrent_mutation":false}`
	if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
		t.Fatal(err)
	}
	err = func() error {
		_, importErr := ImportArtifact(s, rec.MissionID, "atlas-recommendation-readback", readbackPath)
		return importErr
	}()
	if err == nil || !strings.Contains(err.Error(), "claims_authority_advance") {
		t.Fatalf("expected authority-advance rejection, got %v", err)
	}
	unchanged, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.Evidence.AtlasRecommendation != nil || unchanged.Status == "done" {
		t.Fatalf("unsafe Atlas recommendation readback was recorded: %+v", unchanged)
	}
}

func TestImportAtlasRecommendationReadbackDeniedAndBlockedBecomeExactBlockers(t *testing.T) {
	for _, tc := range []struct {
		status string
		reason string
	}{
		{status: "denied", reason: "missing stop-gate evidence for node-17"},
		{status: "blocked", reason: "Command readback disagreement"},
	} {
		t.Run(tc.status, func(t *testing.T) {
			dir := t.TempDir()
			s := NewStore(dir)
			rec, err := s.Start("import terminal Atlas blocker")
			if err != nil {
				t.Fatal(err)
			}
			readbackPath := filepath.Join(dir, "recommendation-readback-"+tc.status+".json")
			readback := `{"schema":"ao.atlas.recommendation-readback.v0.1","status":"` + tc.status + `","total_nodes":40,"completed_nodes":39,"ready_nodes":0,"checkpoint_count":39,"elapsed_minutes":141,"min_minutes_met":true,"lease_time_status":"minimum_minutes_met","return_gate_status":"` + tc.status + `","final_response_allowed":true,"blocker":"` + tc.reason + `","safe_to_execute":false,"executes_work":false,"approves_work":false,"mutates_repositories":false,"provider_calls":false,"release_or_publish":false,"credential_use":false,"direct_main_mutation":false,"concurrent_mutation":false,"claims_authority_advance":false,"rsi_remains_denied":true,"exact_next_action":"Repair exact blocker through AO Atlas before final response."}`
			if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := ImportArtifact(s, rec.MissionID, "atlas-recommendation-readback", readbackPath); err != nil {
				t.Fatal(err)
			}
			blocked, err := s.Load(rec.MissionID)
			if err != nil {
				t.Fatal(err)
			}
			if blocked.Status != "blocked" || blocked.CurrentRoute != "ao-atlas" {
				t.Fatalf("terminal Atlas %s should block Mission: %+v", tc.status, blocked)
			}
			if !strings.Contains(strings.Join(blocked.Blockers, ";"), tc.reason) {
				t.Fatalf("missing exact blocker %q in %+v", tc.reason, blocked.Blockers)
			}
			if blocked.ReturnGate == nil || !blocked.ReturnGate.HardBlocker || !strings.Contains(blocked.ReturnGate.Reason, "terminal hard blocker") {
				t.Fatalf("terminal Atlas blocker should set hard-blocker gate: %+v", blocked.ReturnGate)
			}
		})
	}
}

func TestPromotedFoundryRollupClosesMission(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("build atlas workgraph for promoted rollup")
	if err != nil {
		t.Fatal(err)
	}
	rollupPath := filepath.Join(dir, "foundry-final-rollup.json")
	rollup := `{"schema":"ao.foundry.final-rollup.v0.1","status":"promoted","completed_nodes":2,"total_nodes":2}`
	if err := os.WriteFile(rollupPath, []byte(rollup), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "foundry-final-rollup", rollupPath); err != nil {
		t.Fatal(err)
	}
	done, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != "done" || done.CurrentRoute != "complete" || done.Evidence.FoundryRollup.Status != "promoted" {
		t.Fatalf("promoted rollup should close mission: %+v", done)
	}
}

func TestDeniedFoundryRollupBlocksMissionWithExactNextAction(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("build atlas workgraph for denied rollup")
	if err != nil {
		t.Fatal(err)
	}
	rollupPath := filepath.Join(dir, "foundry-final-rollup.json")
	rollup := `{"schema":"ao.foundry.final-rollup.v0.1","status":"denied","completed_nodes":1,"total_nodes":2}`
	if err := os.WriteFile(rollupPath, []byte(rollup), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "foundry-final-rollup", rollupPath); err != nil {
		t.Fatal(err)
	}
	blocked, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if blocked.Status != "blocked" || !strings.Contains(blocked.ExactNextAction, "Foundry rollup denied") {
		t.Fatalf("denied rollup should block with exact next action: %+v", blocked)
	}
}

func TestFoundryTerminalStateBindingFixtureCoversClosureAndBlockers(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "examples", "valid", "foundry-terminal-state-binding.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Schema string `json:"schema"`
		States []struct {
			Status          string `json:"status"`
			CompletedNodes  int    `json:"completed_nodes"`
			TotalNodes      int    `json:"total_nodes"`
			ExpectedMission string `json:"expected_mission_status"`
			ExpectedRoute   string `json:"expected_route"`
			ExpectedPhase   string `json:"expected_phase"`
			ExactBlocker    string `json:"exact_blocker,omitempty"`
		} `json:"states"`
	}
	if err := json.Unmarshal(body, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Schema != "ao.foundry.terminal-state-binding.v0.1" || len(fixture.States) != 4 {
		t.Fatalf("unexpected fixture coverage: %+v", fixture)
	}
	for _, state := range fixture.States {
		t.Run(state.Status, func(t *testing.T) {
			dir := t.TempDir()
			s := NewStore(dir)
			rec, err := s.Start("bind Foundry terminal state " + state.Status)
			if err != nil {
				t.Fatal(err)
			}
			rollupPath := filepath.Join(dir, "foundry-final-rollup.json")
			rollup := map[string]any{
				"schema":          "ao.foundry.final-rollup.v0.1",
				"status":          state.Status,
				"completed_nodes": state.CompletedNodes,
				"total_nodes":     state.TotalNodes,
				"safe_to_execute": false,
				"executes_work":   false,
				"approves_work":   false,
			}
			rollupBody, err := json.Marshal(rollup)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(rollupPath, rollupBody, 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := ImportArtifact(s, rec.MissionID, "foundry-final-rollup", rollupPath); err != nil {
				t.Fatal(err)
			}
			got, err := s.Load(rec.MissionID)
			if err != nil {
				t.Fatal(err)
			}
			if got.Status != state.ExpectedMission || got.CurrentRoute != state.ExpectedRoute || got.CurrentPhase != state.ExpectedPhase {
				t.Fatalf("terminal state mismatch: got status=%s route=%s phase=%s want status=%s route=%s phase=%s", got.Status, got.CurrentRoute, got.CurrentPhase, state.ExpectedMission, state.ExpectedRoute, state.ExpectedPhase)
			}
			if state.ExactBlocker != "" && !strings.Contains(strings.Join(got.Blockers, ";"), state.ExactBlocker) {
				t.Fatalf("missing exact blocker %q in %+v", state.ExactBlocker, got.Blockers)
			}
		})
	}
}

func TestCommandCompactTimelineFixtureCoversAtlasAndReconciliationEvents(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "examples", "valid", "command-compact-timeline-readback.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Schema              string         `json:"schema"`
		Status              string         `json:"status"`
		Compact             bool           `json:"compact"`
		IncludesKinds       []string       `json:"includes_event_kinds"`
		RecentEvents        []MissionEvent `json:"recent_events"`
		SafeToExecute       bool           `json:"safe_to_execute"`
		ExecutesWork        bool           `json:"executes_work"`
		ApprovesWork        bool           `json:"approves_work"`
		MutatesRepositories bool           `json:"mutates_repositories"`
		RSIRemainsDenied    bool           `json:"rsi_remains_denied"`
		ExactNextAction     string         `json:"exact_next_action"`
	}
	if err := json.Unmarshal(body, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Schema != "ao.command.compact-timeline-readback.v0.1" || fixture.Status != "ready" || !fixture.Compact {
		t.Fatalf("unexpected command timeline fixture header: %+v", fixture)
	}
	for _, want := range []string{"atlas_recommendation", "final_reconciliation"} {
		foundKind := false
		for _, kind := range fixture.IncludesKinds {
			if kind == want {
				foundKind = true
			}
		}
		if !foundKind {
			t.Fatalf("fixture missing included kind %q: %+v", want, fixture.IncludesKinds)
		}
		foundEvent := false
		for _, event := range fixture.RecentEvents {
			if event.Kind == want {
				foundEvent = true
				if event.Status == "" || event.Summary == "" {
					t.Fatalf("event %q missing status or summary: %+v", want, event)
				}
			}
		}
		if !foundEvent {
			t.Fatalf("fixture missing recent event %q: %+v", want, fixture.RecentEvents)
		}
	}
	if fixture.SafeToExecute || fixture.ExecutesWork || fixture.ApprovesWork || fixture.MutatesRepositories {
		t.Fatalf("command timeline fixture widened authority: %+v", fixture)
	}
	if !fixture.RSIRemainsDenied || !strings.Contains(fixture.ExactNextAction, "continue node 19") {
		t.Fatalf("fixture should preserve RSI denial and exact next action: %+v", fixture)
	}
}

func TestEventIndexSearchesSupervisorEvidence(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("index supervisor checkpoint and rollup evidence")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Continue(s, rec.MissionID, ContinueOptions{UntilDone: true, MaxIterations: 2}); err != nil {
		t.Fatal(err)
	}
	index, err := BuildMissionEventIndex(s)
	if err != nil {
		t.Fatal(err)
	}
	preTerminalResults := SearchMissionEvents(index, MissionEventSearchFilters{MissionID: rec.MissionID, Query: "early_return_denied"})
	if preTerminalResults.TotalMatches == 0 {
		t.Fatalf("event index did not expose early-return risk before terminal rollup: %+v", index)
	}
	rollupPath := filepath.Join(dir, "foundry-final-rollup.json")
	rollup := `{"schema":"ao.foundry.final-rollup.v0.1","status":"blocked","completed_nodes":1,"total_nodes":2}`
	if err := os.WriteFile(rollupPath, []byte(rollup), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "foundry-final-rollup", rollupPath); err != nil {
		t.Fatal(err)
	}
	index, err = BuildMissionEventIndex(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{"checkpoint", "foundry rollup", "blocked"} {
		results := SearchMissionEvents(index, MissionEventSearchFilters{MissionID: rec.MissionID, Query: query})
		if results.TotalMatches == 0 {
			t.Fatalf("event index did not find %q: %+v", query, index)
		}
		if results.ExecutesWork || results.ApprovesWork || results.MutatesRepositories {
			t.Fatalf("event search widened authority: %+v", results)
		}
	}
}

func TestEventIndexSearchesAtlasRecommendationReadbackEvidence(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("index Atlas recommendation readback evidence")
	if err != nil {
		t.Fatal(err)
	}
	readbackPath := filepath.Join(dir, "recommendation-readback-short.json")
	readback := `{"schema":"ao.atlas.recommendation-readback.v0.1","status":"completed","total_nodes":40,"completed_nodes":40,"ready_nodes":0,"checkpoint_count":40,"elapsed_minutes":22,"min_minutes_met":false,"lease_time_status":"minimum_minutes_unmet","return_gate_status":"blocked_minimum_minutes_unmet","final_response_allowed":false,"safe_to_execute":false,"executes_work":false,"approves_work":false,"mutates_repositories":false,"provider_calls":false,"release_or_publish":false,"credential_use":false,"direct_main_mutation":false,"concurrent_mutation":false,"claims_authority_advance":false,"exact_next_action":"Continue AO Atlas wave until the minimum lease duration is met."}`
	if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "atlas-recommendation-readback", readbackPath); err != nil {
		t.Fatal(err)
	}
	index, err := BuildMissionEventIndex(s)
	if err != nil {
		t.Fatal(err)
	}
	results := SearchMissionEvents(index, MissionEventSearchFilters{MissionID: rec.MissionID, Kind: "atlas_recommendation", Query: "blocked_minimum_minutes_unmet"})
	if results.TotalMatches != 1 {
		t.Fatalf("expected one Atlas recommendation event, got %+v", results)
	}
	event := results.Events[0]
	if event.Status != "completed" || !strings.Contains(event.Summary, "elapsed_minutes=22") || !strings.Contains(event.Summary, "ready_nodes=0") {
		t.Fatalf("Atlas recommendation event missing terminal details: %+v", event)
	}
	if results.ExecutesWork || results.ApprovesWork || results.MutatesRepositories {
		t.Fatalf("event search widened authority: %+v", results)
	}
}

func TestEventIndexSearchesFinalReconciliationEvidence(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("index final reconciliation evidence")
	if err != nil {
		t.Fatal(err)
	}
	readbackPath := filepath.Join(dir, "recommendation-readback.json")
	readback := `{"schema":"ao.atlas.recommendation-readback.v0.1","status":"completed","total_nodes":40,"completed_nodes":40,"ready_nodes":0,"checkpoint_count":40,"elapsed_minutes":491,"min_minutes_met":true,"lease_time_status":"minimum_minutes_met","return_gate_status":"final_response_allowed","final_response_allowed":true,"safe_to_execute":false,"executes_work":false,"approves_work":false,"mutates_repositories":false,"provider_calls":false,"release_or_publish":false,"credential_use":false,"direct_main_mutation":false,"concurrent_mutation":false,"claims_authority_advance":false}`
	if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "atlas-recommendation-readback", readbackPath); err != nil {
		t.Fatal(err)
	}
	index, err := BuildMissionEventIndex(s)
	if err != nil {
		t.Fatal(err)
	}
	results := SearchMissionEvents(index, MissionEventSearchFilters{MissionID: rec.MissionID, Kind: "final_reconciliation", Query: "artifacts_agree=true"})
	if results.TotalMatches != 1 {
		t.Fatalf("expected one final reconciliation event, got %+v", results)
	}
	if results.Events[0].Status != "ready" || !strings.Contains(results.Events[0].Summary, "rsi_remains_denied=true") {
		t.Fatalf("bad final reconciliation event: %+v", results.Events[0])
	}
}

func TestCommandStatusIncludesAtlasRecommendationSummary(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("command status over Atlas recommendation readback")
	if err != nil {
		t.Fatal(err)
	}
	readbackPath := filepath.Join(dir, "recommendation-readback.json")
	readback := `{"schema":"ao.atlas.recommendation-readback.v0.1","status":"completed","total_nodes":40,"completed_nodes":40,"ready_nodes":0,"checkpoint_count":40,"elapsed_minutes":491,"min_minutes_met":true,"lease_time_status":"minimum_minutes_met","return_gate_status":"final_response_allowed","final_response_allowed":true,"safe_to_execute":false,"executes_work":false,"approves_work":false,"mutates_repositories":false,"provider_calls":false,"release_or_publish":false,"credential_use":false,"direct_main_mutation":false,"concurrent_mutation":false,"claims_authority_advance":false}`
	if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "atlas-recommendation-readback", readbackPath); err != nil {
		t.Fatal(err)
	}
	done, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	status := BuildCommandStatus(done)
	if status.AtlasRecommendation == nil {
		t.Fatalf("command status missing Atlas recommendation summary: %+v", status)
	}
	if status.AtlasRecommendation.CompletedNodes != 40 || status.AtlasRecommendation.ReadyNodes != 0 || status.AtlasRecommendation.ReturnGateStatus != "final_response_allowed" || !status.AtlasRecommendation.FinalResponseAllowed {
		t.Fatalf("bad Atlas recommendation summary: %+v", status.AtlasRecommendation)
	}
	if status.ExecutesWork || status.ApprovesWork || status.MutatesRepositories {
		t.Fatalf("command status widened authority: %+v", status)
	}
}

func TestCLICommandStatusTextIncludesAtlasRecommendationSummary(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "command status text Atlas recommendation"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	readbackPath := filepath.Join(dir, "recommendation-readback.json")
	readback := `{"schema":"ao.atlas.recommendation-readback.v0.1","status":"completed","total_nodes":40,"completed_nodes":40,"ready_nodes":0,"checkpoint_count":40,"elapsed_minutes":491,"min_minutes_met":true,"lease_time_status":"minimum_minutes_met","return_gate_status":"final_response_allowed","final_response_allowed":true,"safe_to_execute":false,"executes_work":false,"approves_work":false,"mutates_repositories":false,"provider_calls":false,"release_or_publish":false,"credential_use":false,"direct_main_mutation":false,"concurrent_mutation":false,"claims_authority_advance":false}`
	if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "import", "atlas-recommendation-readback", "--mission", rec.MissionID, "--path", readbackPath}, &out, &errb); code != 0 {
		t.Fatalf("import: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "command", "status", "--mission", rec.MissionID}, &out, &errb); code != 0 {
		t.Fatalf("command status: %s", errb.String())
	}
	text := out.String()
	for _, want := range []string{"atlas_recommendation=completed", "completed_nodes=40", "ready_nodes=0", "final_response_allowed=true"} {
		if !strings.Contains(text, want) {
			t.Fatalf("command status text missing %q: %s", want, text)
		}
	}
}

func TestFinalReconciliationPacketAgreesAcrossAtlasFoundryAndCommand(t *testing.T) {
	rec := Record{
		Schema:          RecordSchema,
		MissionID:       "mission-reconcile",
		Status:          "done",
		CurrentRoute:    "complete",
		CurrentPhase:    "complete",
		ExactNextAction: "mission complete; read final rollup and recommended next tasks",
		Evidence: EvidenceSummary{
			AtlasRecommendation: &AtlasRecommendationReadbackCounts{
				Status:               "completed",
				TotalNodes:           40,
				CompletedNodes:       40,
				ReadyNodes:           0,
				CheckpointCount:      40,
				MinMinutesMet:        true,
				LeaseTimeStatus:      "minimum_minutes_met",
				ReturnGateStatus:     "final_response_allowed",
				FinalResponseAllowed: true,
			},
			FoundryRollup: &FoundryRollupCounts{Status: "completed", CompletedNodes: 40, TotalNodes: 40},
		},
	}
	packet := BuildFinalReconciliationPacket(rec)
	if packet.Status != "ready" || !packet.ArtifactsAgree || !packet.FinalResponseAllowed {
		t.Fatalf("reconciliation should agree for completed evidence: %+v", packet)
	}
	if packet.AtlasRecommendationStatus != "completed" || packet.CommandStatus != "done" || packet.FoundryStatus != "completed" {
		t.Fatalf("packet missing status summary: %+v", packet)
	}
	if packet.PromotionClaimed || !packet.RSIRemainsDenied || packet.ClaimsAuthorityAdvance {
		t.Fatalf("packet widened promotion or RSI boundary: %+v", packet)
	}
}

func TestFinalReconciliationPacketReportsFoundryAtlasMismatch(t *testing.T) {
	rec := Record{
		Schema:       RecordSchema,
		MissionID:    "mission-reconcile-mismatch",
		Status:       "done",
		CurrentRoute: "complete",
		CurrentPhase: "complete",
		Evidence: EvidenceSummary{
			AtlasRecommendation: &AtlasRecommendationReadbackCounts{
				Status:               "completed",
				TotalNodes:           40,
				CompletedNodes:       40,
				ReadyNodes:           0,
				CheckpointCount:      40,
				MinMinutesMet:        true,
				LeaseTimeStatus:      "minimum_minutes_met",
				ReturnGateStatus:     "final_response_allowed",
				FinalResponseAllowed: true,
			},
			FoundryRollup: &FoundryRollupCounts{Status: "completed", CompletedNodes: 39, TotalNodes: 40},
		},
	}
	packet := BuildFinalReconciliationPacket(rec)
	if packet.Status != "blocked" || packet.ArtifactsAgree {
		t.Fatalf("reconciliation mismatch should block: %+v", packet)
	}
	if !strings.Contains(packet.Blocker, "Foundry completed_nodes=39") || !strings.Contains(packet.Blocker, "Atlas completed_nodes=40") {
		t.Fatalf("packet missing exact mismatch blocker: %+v", packet)
	}
}

func TestCLIFinalReconcileEmitsPacket(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "cli final reconciliation"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	readbackPath := filepath.Join(dir, "recommendation-readback.json")
	readback := `{"schema":"ao.atlas.recommendation-readback.v0.1","status":"completed","total_nodes":40,"completed_nodes":40,"ready_nodes":0,"checkpoint_count":40,"elapsed_minutes":491,"min_minutes_met":true,"lease_time_status":"minimum_minutes_met","return_gate_status":"final_response_allowed","final_response_allowed":true,"safe_to_execute":false,"executes_work":false,"approves_work":false,"mutates_repositories":false,"provider_calls":false,"release_or_publish":false,"credential_use":false,"direct_main_mutation":false,"concurrent_mutation":false,"claims_authority_advance":false}`
	if err := os.WriteFile(readbackPath, []byte(readback), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "import", "atlas-recommendation-readback", "--mission", rec.MissionID, "--path", readbackPath}, &out, &errb); code != 0 {
		t.Fatalf("import: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "final", "reconcile", "--mission", rec.MissionID}, &out, &errb); code != 0 {
		t.Fatalf("final reconcile: %s", errb.String())
	}
	var packet MissionFinalReconciliationPacket
	if err := json.Unmarshal(out.Bytes(), &packet); err != nil {
		t.Fatal(err)
	}
	if packet.Schema != "ao.mission.final-reconciliation-packet.v0.1" || !packet.ArtifactsAgree || !packet.FinalResponseAllowed {
		t.Fatalf("bad reconciliation packet: %+v", packet)
	}
	if packet.ExecutesWork || packet.ApprovesWork || packet.MutatesRepositories {
		t.Fatalf("reconciliation packet widened authority: %+v", packet)
	}
}

func TestFeatureDepthRecommendationsReturnAtLeastTenActionableTasks(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("long-running atlas workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	rec.Evidence.AtlasWorkgraph = &NodeCounts{Total: 12, Ready: 3, Blocked: 2, Completed: 7}
	rollup := BuildFinalRollup(rec)
	if len(rollup.FeatureDepthRecommendations) < 10 {
		t.Fatalf("recommendations too shallow: %d", len(rollup.FeatureDepthRecommendations))
	}
	for _, item := range rollup.FeatureDepthRecommendations {
		if item.ID == "" || item.Task == "" || item.Owner == "" || item.ExactNextAction == "" {
			t.Fatalf("recommendation is not actionable: %+v", item)
		}
	}
}

func TestFinalRollupDeniesFinalResponseWhenReadyNodesRemain(t *testing.T) {
	rec := Record{
		Schema:          RecordSchema,
		MissionID:       "mission-ready-remains",
		Status:          "active",
		CurrentRoute:    "ao-foundry",
		CurrentPhase:    "atlas_workgraph_ready",
		ExactNextAction: "send first safe Atlas node to AO Foundry",
		Evidence: EvidenceSummary{
			AtlasWorkgraph: &NodeCounts{Total: 4, Ready: 2, Blocked: 0, Completed: 2},
		},
	}
	rollup := BuildFinalRollup(rec)
	if rollup.FinalResponseAllowed || rollup.ReturnGateStatus != "early_return_denied" {
		t.Fatalf("final response should be denied while ready nodes remain: %+v", rollup)
	}
	if !strings.Contains(rollup.ExactNextAction, "continue") {
		t.Fatalf("rollup did not preserve executable next action: %+v", rollup)
	}
}

func TestTelegramCommandFixtureMatrix(t *testing.T) {
	allowlist := map[string]string{"1001": "admin", "1002": "user"}
	matrix, err := LoadTelegramCommandMatrix(filepath.Join("..", "..", "examples", "valid", "telegram-command-matrix.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range matrix.Commands {
		chat := "1002"
		if tc.Role == "admin" {
			chat = "1001"
		}
		rb := HandleTelegramCommand(TelegramCommand{ChatID: chat, Command: tc.Command, Role: tc.Role}, allowlist)
		if rb.Status != tc.ExpectedStatus || rb.MutationAuthority {
			t.Fatalf("%s/%s: %+v", tc.Command, tc.Role, rb)
		}
	}
}

func TestTelegramInvalidCommandMatrix(t *testing.T) {
	allowlist := map[string]string{"1001": "admin", "1002": "user"}
	matrix, err := LoadTelegramCommandMatrix(filepath.Join("..", "..", "examples", "invalid", "telegram-command-matrix-denied.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range matrix.Commands {
		chat := "1002"
		if tc.Role == "admin" {
			chat = "1001"
		}
		if tc.Role == "none" {
			chat = "9999"
		}
		rb := HandleTelegramCommand(TelegramCommand{ChatID: chat, Command: tc.Command, Role: tc.Role}, allowlist)
		if rb.Status != tc.ExpectedStatus || rb.MutationAuthority {
			t.Fatalf("%s/%s: %+v", tc.Command, tc.Role, rb)
		}
	}
}

func TestA2AJSONRPCHandlerIntentOnly(t *testing.T) {
	server := httptest.NewServer(A2AHandler())
	defer server.Close()
	req := strings.NewReader(`{"jsonrpc":"2.0","id":"req-1","method":"mission.status","params":{"mission_id":"mission-demo"}}`)
	resp, err := http.Post(server.URL+"/", "application/json", req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var rpc A2AJSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpc); err != nil {
		t.Fatal(err)
	}
	if rpc.JSONRPC != "2.0" || rpc.ID != "req-1" || rpc.Result.MutationAuthority {
		t.Fatalf("bad json-rpc response: %+v", rpc)
	}
	if rpc.Result.Method != "mission.status" || rpc.Result.Status != "intent_recorded" {
		t.Fatalf("bad a2a task: %+v", rpc.Result)
	}
}

func TestA2AJSONRPCHandlerValidatesMethodParams(t *testing.T) {
	server := httptest.NewServer(A2AHandler())
	defer server.Close()
	req := strings.NewReader(`{"jsonrpc":"2.0","id":"req-2","method":"mission.continue","params":{}}`)
	resp, err := http.Post(server.URL+"/", "application/json", req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var rpc A2AJSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpc); err != nil {
		t.Fatal(err)
	}
	if rpc.Result.Status != "invalid" || rpc.Result.MutationAuthority {
		t.Fatalf("expected invalid intent-only response: %+v", rpc)
	}

	req = strings.NewReader(`{"jsonrpc":"2.0","id":"req-3","method":"mission.start","params":{"objective":"build mission gateway readback"}}`)
	resp, err = http.Post(server.URL+"/", "application/json", req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&rpc); err != nil {
		t.Fatal(err)
	}
	if rpc.Result.Status != "intent_recorded" || rpc.Result.Method != "mission.start" || rpc.Result.MutationAuthority {
		t.Fatalf("expected valid intent-only response: %+v", rpc)
	}
}

func TestA2AInvalidJSONRPCExamples(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "examples", "invalid", "a2a-jsonrpc-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no invalid A2A examples")
	}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var req struct {
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatal(err)
		}
		task := A2ATaskForParams(req.Method, req.Params)
		if task.Status != "invalid" || task.MutationAuthority {
			t.Fatalf("%s expected invalid intent-only task, got %+v", path, task)
		}
	}
}

func TestContractSchemasDeclareRequiredProperties(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "docs", "contracts", "*.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no contract schemas found")
	}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var schema map[string]any
		if err := json.Unmarshal(body, &schema); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		props, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatalf("%s missing properties", path)
		}
		for _, item := range schema["required"].([]any) {
			field := item.(string)
			if _, ok := props[field]; !ok {
				t.Fatalf("%s missing property for required field %s", path, field)
			}
		}
	}
}

func TestGatewayRunbookDocumentsIntentOnlyReferences(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "docs", "gateway-readback-runbook.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{"Hermes-style", "Telegram", "A2A", "intent/readback only", "mutation_authority=false"} {
		if !strings.Contains(text, want) {
			t.Fatalf("runbook missing %q", want)
		}
	}
	if err := ValidatePublicSafeText(text); err != nil {
		t.Fatal(err)
	}
}

func TestTelegramConfigReadback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "telegram.json")
	body := `{"schema":"ao.mission.telegram-config.v0.1","token_env":"AO_MISSION_TELEGRAM_REDACTED","allowed_chats":{"1001":"admin","1002":"user"}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadTelegramConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	rb := TelegramConfigReadback(cfg)
	if rb.AllowedChatCount != 2 || rb.MutationAuthority {
		t.Fatalf("bad gateway readback: %+v", rb)
	}
}

func TestA2AHTTPHandler(t *testing.T) {
	server := httptest.NewServer(A2AHandler())
	defer server.Close()
	resp, err := http.Get(server.URL + "/.well-known/agent-card.json")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var card A2AAgentCard
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		t.Fatal(err)
	}
	if card.MutationAuthority {
		t.Fatal("agent card must not grant mutation authority")
	}
}

func TestRouteDecisionHistoryPersistsAcrossNextAndImports(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "build atlas workgraph mission"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	if len(rec.RouteHistory) != 1 || rec.RouteHistory[0].Route != "ao-atlas" {
		t.Fatalf("start did not persist initial route history: %+v", rec.RouteHistory)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "next", "--mission", rec.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("next: %s", errb.String())
	}
	updated, err := NewStore(dir).Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.RouteHistory) != 2 || updated.RouteHistory[1].Route != "ao-atlas" {
		t.Fatalf("next did not append route history: %+v", updated.RouteHistory)
	}
	workgraphPath := filepath.Join(dir, "workgraph.json")
	if err := os.WriteFile(workgraphPath, []byte(`{"schema":"ao.atlas.workgraph.v0.1","nodes":[{"status":"ready"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(NewStore(dir), rec.MissionID, "atlas-workgraph", workgraphPath); err != nil {
		t.Fatal(err)
	}
	updated, err = NewStore(dir).Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	last := updated.RouteHistory[len(updated.RouteHistory)-1]
	if last.Route != "ao-foundry" || last.ExactNextAction != "send first safe Atlas node to AO Foundry" {
		t.Fatalf("atlas import did not append Foundry route history: %+v", updated.RouteHistory)
	}
}

func TestTelegramReplayFixtureProducesIntentOnlyReadback(t *testing.T) {
	readback, err := ReplayTelegramCommandMatrix(
		filepath.Join("..", "..", "examples", "valid", "telegram-command-matrix.json"),
		map[string]string{"1001": "admin", "1002": "user"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.telegram-replay-readback.v0.1" || readback.Status != "ready" {
		t.Fatalf("bad replay readback: %+v", readback)
	}
	if readback.Total != len(readback.Results) || readback.Denied != 2 || readback.Invalid != 0 {
		t.Fatalf("unexpected replay counts: %+v", readback)
	}
	if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("telegram replay widened authority: %+v", readback)
	}
}

func TestA2AHTTPFixtureReplayProducesIntentOnlyReadback(t *testing.T) {
	readback, err := ReplayA2AHTTPFixture(filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"))
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.a2a-http-replay-readback.v0.1" || readback.Status != "ready" {
		t.Fatalf("bad A2A HTTP replay readback: %+v", readback)
	}
	if readback.Total != 3 || readback.Invalid != 1 {
		t.Fatalf("unexpected A2A replay counts: %+v", readback)
	}
	if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("A2A replay widened authority: %+v", readback)
	}
	for _, result := range readback.Results {
		if result.RequestID == "" || result.ResponseID != result.RequestID {
			t.Fatalf("A2A HTTP replay did not bind request/response pair IDs: %+v", result)
		}
	}
}

func TestA2AHTTPFixtureReplayRecordsRequestResponsePairIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a2a-http-mismatch.json")
	if err := os.WriteFile(path, []byte(`{
  "schema": "ao.mission.a2a-http-integration-fixture.v0.1",
  "requests": [
    {
      "jsonrpc": "2.0",
      "id": "req-status",
      "method": "mission.status",
      "params": {"mission_id": "mission-demo"},
      "expected_status": "intent_recorded"
    }
  ],
  "mutation_authority": false
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	readback, err := ReplayA2AHTTPFixture(path)
	if err != nil {
		t.Fatal(err)
	}
	if readback.Status != "ready" || len(readback.Results) != 1 || readback.Results[0].ResponseID != "req-status" {
		t.Fatalf("A2A HTTP replay should validate matching request/response IDs: %+v", readback)
	}
}

func TestGatewayIntentLedgerPersistsTelegramAndA2AReplayWithoutAuthority(t *testing.T) {
	telegram, err := ReplayTelegramUpdates(filepath.Join("..", "..", "examples", "valid", "telegram-update-replay.json"), map[string]string{"1001": "admin", "1002": "user"})
	if err != nil {
		t.Fatal(err)
	}
	a2a, err := ReplayA2AHTTPFixture(filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"))
	if err != nil {
		t.Fatal(err)
	}
	ledger := BuildGatewayIntentLedger("mission-demo", telegram, a2a)
	if ledger.Schema != "ao.mission.gateway-intent-ledger.v0.1" || ledger.Status != "ready" {
		t.Fatalf("bad ledger: %+v", ledger)
	}
	if ledger.Total != len(ledger.Intents) || ledger.IntentRecorded != 4 || ledger.Denied != 1 || ledger.Invalid != 2 {
		t.Fatalf("bad ledger counts: %+v", ledger)
	}
	if ledger.MutationAuthority || ledger.ExecutesWork || ledger.ApprovesWork {
		t.Fatalf("ledger widened authority: %+v", ledger)
	}
	for _, intent := range ledger.Intents {
		if intent.MissionID != "mission-demo" || intent.MutationAuthority || intent.ExecutesWork || intent.ApprovesWork {
			t.Fatalf("unsafe intent record: %+v", intent)
		}
	}
}

func TestGatewayLedgerCommandWritesReplayLedger(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "gateway-ledger.json")
	var out, errb bytes.Buffer
	code := Run([]string{
		"gateway", "ledger",
		"--mission", "mission-demo",
		"--telegram-updates", filepath.Join("..", "..", "examples", "valid", "telegram-update-replay.json"),
		"--telegram-config", filepath.Join("..", "..", "examples", "valid", "telegram-config.json"),
		"--a2a-http", filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"),
		"--out", outPath,
	}, &out, &errb)
	if code != 0 {
		t.Fatalf("gateway ledger failed: %s", errb.String())
	}
	if !strings.Contains(out.String(), "gateway_intent_ledger="+outPath) {
		t.Fatalf("missing ledger path output: %s", out.String())
	}
	var ledger GatewayIntentLedger
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &ledger); err != nil {
		t.Fatal(err)
	}
	if ledger.Total != 7 || ledger.MutationAuthority || ledger.ExecutesWork || ledger.ApprovesWork {
		t.Fatalf("bad written ledger: %+v", ledger)
	}
}

func TestA2ATaskLifecycleFixtureRecordsCancellationAsIntentOnly(t *testing.T) {
	readback, err := ReplayA2ATaskLifecycle(filepath.Join("..", "..", "examples", "valid", "a2a-task-lifecycle.json"))
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.a2a-task-lifecycle-readback.v0.1" || readback.Status != "ready" {
		t.Fatalf("bad lifecycle readback: %+v", readback)
	}
	if readback.Total != 3 || readback.Cancelled != 1 || readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("lifecycle widened authority or bad counts: %+v", readback)
	}
}

func TestA2ATaskLifecycleFixtureRecordsResumeAndCancelEdges(t *testing.T) {
	readback, err := ReplayA2ATaskLifecycle(filepath.Join("..", "..", "examples", "valid", "a2a-task-lifecycle-edges.json"))
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.a2a-task-lifecycle-readback.v0.1" || readback.Status != "ready" {
		t.Fatalf("bad lifecycle edge readback: %+v", readback)
	}
	if readback.Total != 5 || readback.CancelRequested != 1 || readback.Cancelled != 1 || readback.ResumeRequested != 1 || readback.Resumed != 1 {
		t.Fatalf("bad lifecycle edge counts: %+v", readback)
	}
	if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("lifecycle edge replay widened authority: %+v", readback)
	}
}

func TestSchedulerReadbackImportClassifiesFreshness(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("schedule long-running workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "scheduler-readback.json")
	body := `{"schema":"ao.mission.scheduler-readback.v0.1","mission_id":"` + rec.MissionID + `","status":"ready","scheduler":"codex-cron","event_loop":true,"reason":"fixture wakeup only","generated_at_utc":"2026-07-01T00:00:00Z"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "scheduler-readback", path); err != nil {
		t.Fatal(err)
	}
	updated, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Evidence.SchedulerReadback == nil || updated.Evidence.SchedulerReadback.FreshnessStatus != "stale" {
		t.Fatalf("scheduler freshness was not classified as stale: %+v", updated.Evidence.SchedulerReadback)
	}
	snap := Snapshot(updated)
	if snap.EvidenceFreshnessStatus != "stale" {
		t.Fatalf("snapshot freshness=%s", snap.EvidenceFreshnessStatus)
	}
}

func TestSchedulerRecoveryReadbackImportRecordsMissedWakeupEvidence(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("schedule long-running workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("..", "..", "examples", "valid", "scheduler-recovery-readback.json")
	if _, err := ImportArtifact(s, rec.MissionID, "scheduler-recovery-readback", path); err != nil {
		t.Fatal(err)
	}
	updated, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Evidence.SchedulerRecovery == nil || updated.Evidence.SchedulerRecovery.MissedWakeups != 2 {
		t.Fatalf("scheduler recovery evidence missing: %+v", updated.Evidence)
	}
	if updated.CurrentPhase != "scheduler_recovery_recorded" || !strings.Contains(updated.ExactNextAction, "continue") {
		t.Fatalf("scheduler recovery did not set continuation action: %+v", updated)
	}
}

func TestMissionEvidenceImportRejectsAuthorityDrift(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("reject evidence authority drift")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		kind string
		body string
		want string
	}{
		{
			name: "scheduler recovery safe to execute",
			kind: "scheduler-recovery-readback",
			body: `{"schema":"ao.mission.scheduler-recovery-readback.v0.1","mission_id":"mission-demo","safe_to_execute":true,"executes_work":false}`,
			want: "safe_to_execute",
		},
		{
			name: "scheduler recovery schedules work",
			kind: "scheduler-recovery-readback",
			body: `{"schema":"ao.mission.scheduler-recovery-readback.v0.1","mission_id":"mission-demo","schedules_work":true,"executes_work":false}`,
			want: "schedules_work",
		},
		{
			name: "ledger compaction repository mutation",
			kind: "ledger-compaction-readback",
			body: `{"schema":"ao.mission.ledger-compaction-readback.v0.1","mission_id":"mission-demo","mutates_repositories":true,"executes_work":false}`,
			want: "mutates_repositories",
		},
		{
			name: "scheduler readback provider call",
			kind: "scheduler-readback",
			body: `{"schema":"ao.mission.scheduler-readback.v0.1","mission_id":"mission-demo","provider_calls":true,"executes_work":false}`,
			want: "provider_calls",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, strings.ReplaceAll(tc.name, " ", "-")+".json")
			if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := ImportArtifact(s, rec.MissionID, tc.kind, path); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %s rejection, got %v", tc.want, err)
			}
		})
	}
}

func TestArtifactManifestIncludesImportedRecoveryAndCompactionReadbacks(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("manifest recovery and compaction evidence")
	if err != nil {
		t.Fatal(err)
	}
	recoveryPath := filepath.Join("..", "..", "examples", "valid", "scheduler-recovery-readback.json")
	compactionPath := filepath.Join("..", "..", "examples", "valid", "ledger-compaction-readback.json")
	if _, err := ImportArtifact(s, rec.MissionID, "scheduler-recovery-readback", recoveryPath); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(s, rec.MissionID, "ledger-compaction-readback", compactionPath); err != nil {
		t.Fatal(err)
	}
	updated, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	manifest := BuildArtifactManifest(updated)
	kinds := map[string]bool{}
	for _, ref := range manifest.ArtifactRefs {
		kinds[ref.Kind] = true
	}
	if !kinds["scheduler-recovery-readback"] || !kinds["ledger-compaction-readback"] {
		t.Fatalf("manifest missing recovery or compaction refs: %+v", manifest.ArtifactRefs)
	}
	if manifest.ExecutesWork || manifest.ApprovesWork || manifest.SafeToExecute {
		t.Fatalf("manifest widened authority: %+v", manifest)
	}
}

func TestOperatorNextActionsDocsAreConcreteAndPublicSafe(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "operator-next-actions.md")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		"ao-mission start",
		"ao-mission next --mission",
		"ao-mission continue --mission",
		"ao-mission artifacts manifest --mission",
		"Move to AO Atlas",
		"Move to AO Foundry",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("operator docs missing %q", want)
		}
	}
	if err := ValidatePublicSafeText(text); err != nil {
		t.Fatal(err)
	}
}

func TestArtifactManifestCommandWritesOutFile(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "mission artifact manifest output"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, "artifact-manifest.json")
	out.Reset()
	if code := Run([]string{"--home", dir, "artifacts", "manifest", "--mission", rec.MissionID, "--out", manifestPath}, &out, &errb); code != 0 {
		t.Fatalf("artifact manifest --out: %s", errb.String())
	}
	if !strings.Contains(out.String(), "artifact_manifest="+manifestPath) {
		t.Fatalf("expected output path summary, got %s", out.String())
	}
	var manifest ArtifactManifest
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.MissionID != rec.MissionID || manifest.ManifestDigest == "" || manifest.ExecutesWork || manifest.ApprovesWork {
		t.Fatalf("bad written manifest: %+v", manifest)
	}
}

func TestArtifactManifestRepairCommandRecomputesDigests(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, "route.json")
	if err := os.WriteFile(artifactPath, []byte(`{"schema":"ao.mission.route-decision.v0.1"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := FinalizeArtifactManifest(ArtifactManifest{
		MissionID: "mission-demo",
		ArtifactRefs: []ArtifactRef{{
			Schema: ArtifactRefSchema,
			Ref:    artifactPath,
			Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Kind:   "route_readback",
		}},
	})
	manifestPath := filepath.Join(dir, "artifact-manifest.json")
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	repairedPath := filepath.Join(dir, "artifact-manifest.repaired.json")
	var out, errb bytes.Buffer
	if code := Run([]string{"artifacts", "repair-manifest", "--path", manifestPath, "--out", repairedPath}, &out, &errb); code != 0 {
		t.Fatalf("repair-manifest: %s", errb.String())
	}
	result, err := ValidateArtifactManifestFile(repairedPath)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "passed" || result.ArtifactCount != 1 || result.ExecutesWork || result.ApprovesWork {
		t.Fatalf("bad repaired manifest validation: %+v", result)
	}
}

func TestMissionHistoryCommandExportsRouteHistory(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "history atlas workgraph mission"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "next", "--mission", rec.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("next: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "history", "--mission", rec.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("mission history: %s", errb.String())
	}
	var history []RouteDecision
	if err := json.Unmarshal(out.Bytes(), &history); err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].Route != "ao-atlas" || history[1].SafeToExecute {
		t.Fatalf("bad route history: %+v", history)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "history", "--mission", rec.MissionID}, &out, &errb); code != 0 {
		t.Fatalf("mission history text: %s", errb.String())
	}
	if !strings.Contains(out.String(), "route=ao-atlas") || strings.Contains(out.String(), "safe_to_execute=true") {
		t.Fatalf("bad history text: %s", out.String())
	}
}

func TestSchedulerReplayFixtureClassifiesFreshness(t *testing.T) {
	readback, err := ReplaySchedulerReadbacks(filepath.Join("..", "..", "examples", "valid", "scheduler-readback-replay.json"))
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.scheduler-replay-readback.v0.1" || readback.Status != "ready" {
		t.Fatalf("bad scheduler replay readback: %+v", readback)
	}
	if readback.Total != 3 || readback.Fresh != 1 || readback.Stale != 1 || readback.Unknown != 1 {
		t.Fatalf("bad scheduler freshness counts: %+v", readback)
	}
	if readback.EvaluatedAtUTC != "2026-07-03T12:00:00Z" {
		t.Fatalf("scheduler replay did not preserve fixture evaluation time: %+v", readback)
	}
	if readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("scheduler replay widened authority: %+v", readback)
	}
}

func TestSchedulerRecoveryRecommendsImmediateContinuationForMissedWakeups(t *testing.T) {
	replay, err := ReplaySchedulerReadbacks(filepath.Join("..", "..", "examples", "valid", "scheduler-readback-replay.json"))
	if err != nil {
		t.Fatal(err)
	}
	recovery := BuildSchedulerRecoveryReadback("mission-demo", replay)
	if recovery.Schema != "ao.mission.scheduler-recovery-readback.v0.1" || recovery.Status != "attention_required" {
		t.Fatalf("bad scheduler recovery: %+v", recovery)
	}
	if recovery.MissedWakeups != 2 || recovery.RecoveryMode != "immediate_continue_recommended" {
		t.Fatalf("bad scheduler recovery counts: %+v", recovery)
	}
	if !strings.Contains(recovery.ExactNextAction, "ao-mission continue --mission mission-demo --until-done") {
		t.Fatalf("bad scheduler recovery next action: %+v", recovery)
	}
	if recovery.ExecutesWork || recovery.ApprovesWork {
		t.Fatalf("scheduler recovery widened authority: %+v", recovery)
	}
}

func TestSchedulerAlertSummaryHighlightsStaleReadbacks(t *testing.T) {
	replay, err := ReplaySchedulerReadbacks(filepath.Join("..", "..", "examples", "valid", "scheduler-readback-replay.json"))
	if err != nil {
		t.Fatal(err)
	}
	alerts := BuildSchedulerAlertSummary(replay)
	if alerts.Schema != "ao.mission.scheduler-alert-summary.v0.1" || alerts.Status != "attention_required" {
		t.Fatalf("bad scheduler alert summary: %+v", alerts)
	}
	if alerts.Stale != 1 || alerts.Unknown != 1 || len(alerts.Alerts) != 2 {
		t.Fatalf("bad scheduler alerts: %+v", alerts)
	}
	if alerts.ExecutesWork || alerts.ApprovesWork {
		t.Fatalf("scheduler alerts widened authority: %+v", alerts)
	}
}

func TestMissionLedgerCompactionTrimsHistoryAndRecordsEvidence(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("build a long-running atlas workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	rec, err = s.Update(rec.MissionID, func(r *Record) error {
		for i := 0; i < 5; i++ {
			r.Steps = append(r.Steps, ContinuationStep{
				Schema:          StepSchema,
				MissionID:       r.MissionID,
				Iteration:       i + 1,
				Route:           "ao-atlas",
				Result:          "handoff_required",
				ExactNextAction: "continue through Atlas",
				GeneratedAtUTC:  "2026-07-03T00:00:00Z",
			})
			AppendRouteHistory(r, RouteDecision{
				Schema:          RouteSchema,
				MissionID:       r.MissionID,
				Route:           "ao-atlas",
				Reason:          "test route",
				SafeToRequest:   true,
				SafeToExecute:   false,
				SafeToPromote:   false,
				ExactNextAction: "continue through Atlas",
				GeneratedAtUTC:  "2026-07-03T00:00:00Z",
			})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	readback, err := CompactMissionLedger(s, rec.MissionID, LedgerCompactionOptions{KeepRouteHistory: 2, KeepSteps: 3})
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.ledger-compaction-readback.v0.1" || readback.Status != "compacted" {
		t.Fatalf("bad compaction readback: %+v", readback)
	}
	if readback.RouteHistoryBefore <= readback.RouteHistoryAfter || readback.RouteHistoryAfter != 2 || readback.StepsAfter != 3 {
		t.Fatalf("bad compaction counts: %+v", readback)
	}
	if readback.ExecutesWork || readback.ApprovesWork || readback.MutatesRepositories {
		t.Fatalf("compaction widened authority: %+v", readback)
	}
	updated, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.RouteHistory) != 2 || len(updated.Steps) != 3 {
		t.Fatalf("record was not compacted: route_history=%d steps=%d", len(updated.RouteHistory), len(updated.Steps))
	}
	if updated.Evidence.LedgerCompaction == nil || updated.Evidence.LedgerCompaction.RouteHistoryAfter != 2 {
		t.Fatalf("compaction evidence missing: %+v", updated.Evidence)
	}
}

func TestMissionTimelineCompactionEmitsDigestBoundReadback(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "timeline compaction"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	s := NewStore(dir)
	_, err := s.Update(rec.MissionID, func(r *Record) error {
		for i := 0; i < 4; i++ {
			AppendRouteHistory(r, RouteDecision{Schema: RouteSchema, MissionID: r.MissionID, Route: "ao-atlas", SafeToRequest: true, SafeToExecute: false, SafeToPromote: false, GeneratedAtUTC: "2026-07-03T00:00:00Z"})
			r.Steps = append(r.Steps, ContinuationStep{Schema: StepSchema, MissionID: r.MissionID, Iteration: i + 1, Route: "ao-atlas", Result: "recorded", ExactNextAction: "continue", GeneratedAtUTC: "2026-07-03T00:00:00Z"})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "compact", "--mission", rec.MissionID, "--keep-route-history", "2", "--keep-steps", "2", "--timeline"}, &out, &errb); code != 0 {
		t.Fatalf("mission compact timeline: %s", errb.String())
	}
	var readback map[string]any
	if err := json.Unmarshal(out.Bytes(), &readback); err != nil {
		t.Fatal(err)
	}
	if readback["schema"] != "ao.mission.timeline-compaction-readback.v0.1" || readback["status"] != "compacted" {
		t.Fatalf("bad timeline compaction readback: %#v", readback)
	}
	if digest, _ := readback["timeline_digest"].(string); !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("timeline compaction missing digest: %#v", readback)
	}
	if readback["executes_work"] != false || readback["approves_work"] != false || readback["mutates_repositories"] != false {
		t.Fatalf("timeline compaction widened authority: %#v", readback)
	}
}

func TestMissionCompactCLIEmitsReadback(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("build a long-running atlas workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(rec.MissionID, func(r *Record) error {
		for i := 0; i < 4; i++ {
			r.Steps = append(r.Steps, ContinuationStep{Schema: StepSchema, MissionID: r.MissionID, Iteration: i + 1, Route: "ao-atlas", Result: "handoff_required", GeneratedAtUTC: "2026-07-03T00:00:00Z"})
			AppendRouteHistory(r, RouteDecision{Schema: RouteSchema, MissionID: r.MissionID, Route: "ao-atlas", SafeToRequest: true, SafeToExecute: false, SafeToPromote: false, GeneratedAtUTC: "2026-07-03T00:00:00Z"})
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "mission", "compact", "--mission", rec.MissionID, "--keep-route-history", "2", "--keep-steps", "2"}, &out, &errb); code != 0 {
		t.Fatalf("mission compact: %s", errb.String())
	}
	var readback LedgerCompactionReadback
	if err := json.Unmarshal(out.Bytes(), &readback); err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.ledger-compaction-readback.v0.1" || readback.RouteHistoryAfter != 2 || readback.StepsAfter != 2 {
		t.Fatalf("bad compact CLI readback: %+v", readback)
	}
}

func TestMissionCompactDryRunDoesNotMutateRecord(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("build a long-running atlas workgraph mission")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(rec.MissionID, func(r *Record) error {
		for i := 0; i < 4; i++ {
			r.Steps = append(r.Steps, ContinuationStep{Schema: StepSchema, MissionID: r.MissionID, Iteration: i + 1, Route: "ao-atlas", Result: "handoff_required", GeneratedAtUTC: "2026-07-03T00:00:00Z"})
			AppendRouteHistory(r, RouteDecision{Schema: RouteSchema, MissionID: r.MissionID, Route: "ao-atlas", SafeToRequest: true, SafeToExecute: false, SafeToPromote: false, GeneratedAtUTC: "2026-07-03T00:00:00Z"})
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	readback, err := CompactMissionLedger(s, rec.MissionID, LedgerCompactionOptions{KeepRouteHistory: 2, KeepSteps: 2, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if readback.Status != "dry_run" || readback.RouteHistoryAfter != 2 || readback.StepsAfter != 2 {
		t.Fatalf("bad dry-run compaction readback: %+v", readback)
	}
	updated, err := s.Load(rec.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.RouteHistory) != 5 || len(updated.Steps) != 4 || updated.Evidence.LedgerCompaction != nil {
		t.Fatalf("dry-run compaction mutated record: route_history=%d steps=%d evidence=%+v", len(updated.RouteHistory), len(updated.Steps), updated.Evidence)
	}
}

func TestScheduleRecoverCLIEmitsImmediateContinuationReadback(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run([]string{"schedule", "recover", "--mission", "mission-demo", "--fixture", filepath.Join("..", "..", "examples", "valid", "scheduler-readback-replay.json")}, &out, &errb); code != 0 {
		t.Fatalf("schedule recover: %s", errb.String())
	}
	var readback SchedulerRecoveryReadback
	if err := json.Unmarshal(out.Bytes(), &readback); err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.scheduler-recovery-readback.v0.1" || readback.RecoveryMode != "immediate_continue_recommended" {
		t.Fatalf("bad schedule recovery CLI readback: %+v", readback)
	}
	if readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("schedule recovery widened authority: %+v", readback)
	}
}

func TestArtifactManifestSelfValidationRejectsTampering(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, "route.json")
	if err := os.WriteFile(artifactPath, []byte(`{"schema":"ao.mission.route-decision.v0.1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	digest := digestBytesForTest(t, artifactPath)
	manifest := ArtifactManifest{
		Schema:    "ao.mission.artifact-manifest.v0.1",
		MissionID: "mission-demo",
		ArtifactRefs: []ArtifactRef{{
			Schema: "ao.mission.artifact-ref.v0.1",
			Ref:    artifactPath,
			Digest: digest,
			Kind:   "route_readback",
		}},
		SafeToExecute: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	}
	manifest = FinalizeArtifactManifest(manifest)
	manifestPath := filepath.Join(dir, "artifact-manifest.json")
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := ValidateArtifactManifestFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "passed" || result.ManifestDigest != manifest.ManifestDigest || result.ArtifactCount != 1 {
		t.Fatalf("bad manifest validation: %+v", result)
	}

	if err := os.WriteFile(artifactPath, []byte(`{"schema":"tampered"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err = ValidateArtifactManifestFile(manifestPath)
	if err == nil || result.Status != "failed" || !strings.Contains(err.Error(), "artifact digest mismatch") {
		t.Fatalf("expected tamper failure, result=%+v err=%v", result, err)
	}
}

func TestArtifactManifestSelfValidationRejectsInvalidDigestFormat(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, "route.json")
	if err := os.WriteFile(artifactPath, []byte(`{"schema":"ao.mission.route-decision.v0.1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := FinalizeArtifactManifest(ArtifactManifest{
		MissionID: "mission-demo",
		ArtifactRefs: []ArtifactRef{{
			Schema: "ao.mission.artifact-ref.v0.1",
			Ref:    artifactPath,
			Digest: "not-a-sha256-digest",
			Kind:   "route_readback",
		}},
	})
	manifestPath := filepath.Join(dir, "artifact-manifest.json")
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := ValidateArtifactManifestFile(manifestPath)
	if err == nil || result.Status != "failed" || !strings.Contains(err.Error(), "digest must start with sha256:") {
		t.Fatalf("expected digest format failure, result=%+v err=%v", result, err)
	}
}

func TestArtifactManifestValidationNormalizesTextLineEndings(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, "route.json")
	lfBody := []byte("{\n  \"schema\": \"ao.mission.route-decision.v0.1\"\n}\n")
	if err := os.WriteFile(artifactPath, []byte("{\r\n  \"schema\": \"ao.mission.route-decision.v0.1\"\r\n}\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(lfBody)
	manifest := FinalizeArtifactManifest(ArtifactManifest{
		MissionID: "mission-demo",
		ArtifactRefs: []ArtifactRef{{
			Schema: "ao.mission.artifact-ref.v0.1",
			Ref:    artifactPath,
			Digest: "sha256:" + hex.EncodeToString(sum[:]),
			Kind:   "route_readback",
		}},
	})
	manifestPath := filepath.Join(dir, "artifact-manifest.json")
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := ValidateArtifactManifestFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "passed" || result.ArtifactCount != 1 {
		t.Fatalf("bad normalized manifest validation: %+v", result)
	}
}

func TestArtifactManifestFixtureBindsRecoveryAndCompactionEvidence(t *testing.T) {
	result, err := ValidateArtifactManifestFile(filepath.Join("..", "..", "examples", "valid", "artifact-manifest-recovery-compaction.json"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "passed" || result.ArtifactCount != 2 || result.ExecutesWork || result.ApprovesWork {
		t.Fatalf("bad recovery/compaction manifest validation: %+v", result)
	}
}

func TestTelegramUpdateReplayFixtureProducesIntentOnlyReadback(t *testing.T) {
	readback, err := ReplayTelegramUpdates(filepath.Join("..", "..", "examples", "valid", "telegram-update-replay.json"), map[string]string{"1001": "admin", "1002": "user"})
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.telegram-update-replay-readback.v0.1" || readback.Status != "ready" {
		t.Fatalf("bad telegram update replay: %+v", readback)
	}
	if readback.Total != 4 || readback.IntentRecorded != 2 || readback.Denied != 1 || readback.Invalid != 1 {
		t.Fatalf("bad telegram update counts: %+v", readback)
	}
	if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("telegram update replay widened authority: %+v", readback)
	}
}

func TestTelegramWebhookReplayFixtureMatchesUpdateReplayBoundary(t *testing.T) {
	readback, err := ReplayTelegramWebhookFixture(filepath.Join("..", "..", "examples", "valid", "telegram-webhook-replay.json"), map[string]string{"1001": "admin", "1002": "user"})
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.telegram-webhook-replay-readback.v0.1" || readback.Gateway != "telegram_webhook" || readback.Status != "ready" {
		t.Fatalf("bad webhook replay readback: %+v", readback)
	}
	if readback.Total != 5 || readback.IntentRecorded != 2 || readback.Denied != 2 || readback.Invalid != 1 {
		t.Fatalf("bad webhook replay counts: %+v", readback)
	}
	if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("telegram webhook replay widened authority: %+v", readback)
	}
}

func TestTelegramWebhookReplayTracksDuplicateUpdates(t *testing.T) {
	readback, err := ReplayTelegramWebhookFixture(filepath.Join("..", "..", "examples", "valid", "telegram-webhook-duplicate-replay.json"), map[string]string{"1001": "admin"})
	if err != nil {
		t.Fatal(err)
	}
	if readback.Total != 3 || readback.Duplicates != 1 || readback.IntentRecorded != 2 {
		t.Fatalf("bad duplicate webhook accounting: %+v", readback)
	}
	if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("telegram webhook duplicate replay widened authority: %+v", readback)
	}
}

func TestTelegramWebhookReplayClassifiesFreshness(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "telegram-webhook-freshness.json")
	if err := os.WriteFile(fixturePath, []byte(`{
  "schema": "ao.mission.telegram-webhook-fixture.v0.1",
  "updates": [
    {"update_id": 3001, "chat_id": "1001", "text": "/status", "expected_status": "intent_recorded", "generated_at_utc": "2099-01-01T00:00:00Z"},
    {"update_id": 3002, "chat_id": "1001", "text": "/next", "expected_status": "intent_recorded", "generated_at_utc": "2000-01-01T00:00:00Z"},
    {"update_id": 3003, "chat_id": "1001", "text": "/where", "expected_status": "intent_recorded"}
  ]
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	readback, err := ReplayTelegramWebhookFixture(fixturePath, map[string]string{"1001": "admin"})
	if err != nil {
		t.Fatal(err)
	}
	if readback.Fresh != 1 || readback.Stale != 1 || readback.UnknownFreshness != 1 || readback.FreshnessStatus != "stale" {
		t.Fatalf("telegram webhook freshness was not classified: %+v", readback)
	}
	if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("telegram webhook freshness widened authority: %+v", readback)
	}
}

func TestTelegramWebhookReplayCLIEmitsIntentOnlyReadback(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run([]string{
		"telegram", "webhook-replay",
		"--fixture", filepath.Join("..", "..", "examples", "valid", "telegram-webhook-replay.json"),
		"--config", filepath.Join("..", "..", "examples", "valid", "telegram-config.json"),
	}, &out, &errb); code != 0 {
		t.Fatalf("telegram webhook replay: %s", errb.String())
	}
	var readback GatewayReplayReadback
	if err := json.Unmarshal(out.Bytes(), &readback); err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.telegram-webhook-replay-readback.v0.1" || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("bad webhook CLI readback: %+v", readback)
	}
}

func TestA2AServeOnceEmitsReplayableFixtureServerReadback(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run([]string{"a2a", "serve", "--http", "--once"}, &out, &errb); code != 0 {
		t.Fatalf("a2a serve once: %s", errb.String())
	}
	var readback map[string]any
	if err := json.Unmarshal(out.Bytes(), &readback); err != nil {
		t.Fatal(err)
	}
	if readback["schema"] != "ao.mission.a2a-fixture-server-readback.v0.1" || readback["status"] != "ready" {
		t.Fatalf("bad A2A fixture server readback: %#v", readback)
	}
	if readback["agent_card_path"] != "/.well-known/agent-card.json" || readback["jsonrpc_path"] != "/" {
		t.Fatalf("A2A fixture server readback missing replay paths: %#v", readback)
	}
	if readback["mutation_authority"] != false || readback["executes_work"] != false || readback["approves_work"] != false {
		t.Fatalf("A2A fixture server widened authority: %#v", readback)
	}
}

func TestA2AAgentCardIncludesProtocolMetadata(t *testing.T) {
	card := AgentCard()
	if card.ProtocolVersion == "" || card.Endpoint == "" || card.Description == "" {
		t.Fatalf("agent card missing protocol metadata: %+v", card)
	}
	for _, want := range []string{"streaming=false", "push_notifications=false", "mutation_authority=false"} {
		if !stringSliceContains(card.Capabilities, want) {
			t.Fatalf("agent card missing capability %q: %+v", want, card.Capabilities)
		}
	}
	if !card.CapabilitiesDetail["state_transition_history"] || card.CapabilitiesDetail["streaming"] || card.CapabilitiesDetail["push_notifications"] {
		t.Fatalf("agent card capabilities detail must expose readback-only A2A capabilities: %+v", card.CapabilitiesDetail)
	}
	if len(card.Skills) < 3 {
		t.Fatalf("agent card should expose mission readback skills: %+v", card.Skills)
	}
	if card.Skills[0].ID == "" || len(card.Skills[0].Tags) == 0 {
		t.Fatalf("agent card skill metadata incomplete: %+v", card.Skills[0])
	}
	if card.MutationAuthority {
		t.Fatal("agent card must remain intent/readback only")
	}
}

func TestA2ALifecycleTracksArtifactAndCancelReadbacks(t *testing.T) {
	readback, err := ReplayA2ATaskLifecycle(filepath.Join("..", "..", "examples", "valid", "a2a-task-lifecycle-artifacts.json"))
	if err != nil {
		t.Fatal(err)
	}
	if readback.Total != 4 || readback.CancelRequested != 1 || readback.Cancelled != 1 || readback.ArtifactReadbacks != 2 {
		t.Fatalf("bad A2A artifact/cancel readback counts: %+v", readback)
	}
	if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("A2A lifecycle artifact readbacks widened authority: %+v", readback)
	}
}

func TestA2ATaskCarriesArtifactRefsWithoutMutationAuthority(t *testing.T) {
	task := A2ATaskForParams("mission.artifacts", map[string]any{"mission_id": "mission-demo"})
	if task.Status != "intent_recorded" {
		t.Fatalf("bad artifact task status: %+v", task)
	}
	if len(task.ArtifactRefs) != 1 || task.ArtifactRefs[0].Kind != "mission_artifact_readback" {
		t.Fatalf("artifact task should carry readback artifact refs: %+v", task.ArtifactRefs)
	}
	if task.MutationAuthority {
		t.Fatalf("artifact task widened authority: %+v", task)
	}
}

func TestGatewayReplaySuiteCombinesTelegramAndA2AWithoutAuthority(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "gateway-replay-suite.json")
	var out, errb bytes.Buffer
	code := Run([]string{
		"gateway", "replay-suite",
		"--telegram-config", filepath.Join("..", "..", "examples", "valid", "telegram-config.json"),
		"--telegram-webhook", filepath.Join("..", "..", "examples", "valid", "telegram-webhook-replay.json"),
		"--telegram-updates", filepath.Join("..", "..", "examples", "valid", "telegram-update-replay.json"),
		"--a2a-http", filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"),
		"--a2a-lifecycle", filepath.Join("..", "..", "examples", "valid", "a2a-task-lifecycle-artifacts.json"),
		"--out", outPath,
	}, &out, &errb)
	if code != 0 {
		t.Fatalf("gateway replay-suite failed: %s", errb.String())
	}
	var suite map[string]any
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &suite); err != nil {
		t.Fatal(err)
	}
	if suite["schema"] != "ao.mission.gateway-replay-suite-readback.v0.1" || suite["status"] != "ready" {
		t.Fatalf("bad gateway replay suite: %#v", suite)
	}
	if suite["telegram_replays"] != float64(2) || suite["a2a_replays"] != float64(2) {
		t.Fatalf("gateway replay suite should bind two Telegram and two A2A replays: %#v", suite)
	}
	for _, key := range []string{"mutation_authority", "executes_work", "approves_work"} {
		if suite[key] != false {
			t.Fatalf("gateway replay suite widened %s: %#v", key, suite)
		}
	}
}

func TestA2ACompatibilityCommandValidatesAgentCardTasksAndArtifacts(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "a2a-compatibility.json")
	var out, errb bytes.Buffer
	code := Run([]string{
		"a2a", "compatibility",
		"--agent-card", filepath.Join("..", "..", "examples", "valid", "a2a-agent-card.json"),
		"--http", filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"),
		"--lifecycle", filepath.Join("..", "..", "examples", "valid", "a2a-task-lifecycle-artifacts.json"),
		"--out", outPath,
	}, &out, &errb)
	if code != 0 {
		t.Fatalf("a2a compatibility failed: %s", errb.String())
	}
	var readback map[string]any
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &readback); err != nil {
		t.Fatal(err)
	}
	if readback["schema"] != "ao.mission.a2a-compatibility-readback.v0.1" || readback["status"] != "ready" {
		t.Fatalf("bad A2A compatibility readback: %#v", readback)
	}
	if readback["agent_card_skills"] != float64(3) || readback["artifact_readbacks"] != float64(2) {
		t.Fatalf("A2A compatibility should bind skills and artifacts: %#v", readback)
	}
	if readback["mutation_authority"] != false || readback["executes_work"] != false {
		t.Fatalf("A2A compatibility widened authority: %#v", readback)
	}
}

func TestGovernanceSnapshotDiffCommandReportsRouteChangeWithoutAuthority(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("small mission")
	if err != nil {
		t.Fatal(err)
	}
	before := Snapshot(rec)
	rec.CurrentRoute = "ao-atlas"
	rec.CurrentPhase = "atlas_import"
	after := Snapshot(rec)
	beforePath := filepath.Join(dir, "before.json")
	afterPath := filepath.Join(dir, "after.json")
	writeJSONForTest(t, beforePath, before)
	writeJSONForTest(t, afterPath, after)
	var out, errb bytes.Buffer
	if code := Run([]string{"governance", "diff", "--before", beforePath, "--after", afterPath}, &out, &errb); code != 0 {
		t.Fatalf("governance diff: %s", errb.String())
	}
	var diff map[string]any
	if err := json.Unmarshal(out.Bytes(), &diff); err != nil {
		t.Fatal(err)
	}
	if diff["schema"] != "ao.mission.governance-snapshot-diff.v0.1" || diff["changed_fields"] != float64(2) {
		t.Fatalf("bad snapshot diff: %#v", diff)
	}
	if diff["safe_to_execute"] != false || diff["executes_work"] != false || diff["approves_work"] != false {
		t.Fatalf("snapshot diff widened authority: %#v", diff)
	}
}

func TestMissionArchiveExportWritesDigestBoundPublicSafeBundle(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "archive mission evidence"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	archivePath := filepath.Join(dir, "mission-archive.json")
	if code := Run([]string{"--home", dir, "mission", "archive", "--mission", rec.MissionID, "--out", archivePath}, &out, &errb); code != 0 {
		t.Fatalf("mission archive: %s", errb.String())
	}
	var archive map[string]any
	body, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &archive); err != nil {
		t.Fatal(err)
	}
	if archive["schema"] != "ao.mission.archive.v0.1" || archive["mission_id"] != rec.MissionID {
		t.Fatalf("bad archive: %#v", archive)
	}
	if digest, _ := archive["archive_digest"].(string); !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("archive digest missing: %#v", archive)
	}
	if archive["safe_to_execute"] != false || archive["executes_work"] != false || archive["approves_work"] != false {
		t.Fatalf("archive widened authority: %#v", archive)
	}
}

func TestMissionArchiveValidateAndImportRoundTripWithoutAuthority(t *testing.T) {
	exportHome := t.TempDir()
	importHome := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", exportHome, "start", "archive import round trip"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(exportHome, "mission-archive.json")
	out.Reset()
	if code := Run([]string{"--home", exportHome, "mission", "archive", "--mission", rec.MissionID, "--out", archivePath}, &out, &errb); code != 0 {
		t.Fatalf("mission archive: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"mission", "validate-archive", "--path", archivePath}, &out, &errb); code != 0 {
		t.Fatalf("mission validate-archive: %s", errb.String())
	}
	var validation map[string]any
	if err := json.Unmarshal(out.Bytes(), &validation); err != nil {
		t.Fatal(err)
	}
	if validation["schema"] != "ao.mission.archive-validation.v0.1" || validation["status"] != "ready" || validation["mission_id"] != rec.MissionID {
		t.Fatalf("bad archive validation: %#v", validation)
	}
	if validation["safe_to_execute"] != false || validation["executes_work"] != false || validation["approves_work"] != false {
		t.Fatalf("archive validation widened authority: %#v", validation)
	}
	out.Reset()
	if code := Run([]string{"--home", importHome, "mission", "import-archive", "--path", archivePath}, &out, &errb); code != 0 {
		t.Fatalf("mission import-archive: %s", errb.String())
	}
	var imported map[string]any
	if err := json.Unmarshal(out.Bytes(), &imported); err != nil {
		t.Fatal(err)
	}
	if imported["schema"] != "ao.mission.archive-import-readback.v0.1" || imported["status"] != "ready" || imported["mission_id"] != rec.MissionID {
		t.Fatalf("bad archive import readback: %#v", imported)
	}
	out.Reset()
	if code := Run([]string{"--home", importHome, "mission", "inspect", "--mission", rec.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("inspect imported mission: %s", errb.String())
	}
	var roundTrip Record
	if err := json.Unmarshal(out.Bytes(), &roundTrip); err != nil {
		t.Fatal(err)
	}
	if roundTrip.MissionID != rec.MissionID || roundTrip.Objective != rec.Objective {
		t.Fatalf("archive import did not restore record: %#v", roundTrip)
	}
}

func TestTelegramRoleMatrixExportListsAllowlistedRolesWithoutAuthority(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "telegram-role-matrix.json")
	var out, errb bytes.Buffer
	if code := Run([]string{
		"telegram", "role-matrix",
		"--config", filepath.Join("..", "..", "examples", "valid", "telegram-config.json"),
		"--out", outPath,
	}, &out, &errb); code != 0 {
		t.Fatalf("telegram role-matrix: %s", errb.String())
	}
	var matrix map[string]any
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix["schema"] != "ao.mission.telegram-role-matrix-readback.v0.1" || matrix["status"] != "ready" {
		t.Fatalf("bad Telegram role matrix: %#v", matrix)
	}
	if matrix["chat_count"] != float64(2) || matrix["admin_count"] != float64(1) || matrix["user_count"] != float64(1) {
		t.Fatalf("bad Telegram role counts: %#v", matrix)
	}
	if matrix["mutation_authority"] != false || matrix["executes_work"] != false || matrix["approves_work"] != false {
		t.Fatalf("Telegram role matrix widened authority: %#v", matrix)
	}
}

func TestA2AStreamingDeniedReadbackRejectsStreamingAgentCard(t *testing.T) {
	var out, errb bytes.Buffer
	code := Run([]string{
		"a2a", "streaming-denial",
		"--agent-card", filepath.Join("..", "..", "examples", "invalid", "a2a-agent-card-streaming.json"),
	}, &out, &errb)
	if code != 0 {
		t.Fatalf("a2a streaming-denial should emit denial readback, got stderr=%s", errb.String())
	}
	var readback map[string]any
	if err := json.Unmarshal(out.Bytes(), &readback); err != nil {
		t.Fatal(err)
	}
	if readback["schema"] != "ao.mission.a2a-streaming-denial-readback.v0.1" || readback["status"] != "denied" {
		t.Fatalf("bad A2A streaming denial: %#v", readback)
	}
	if readback["streaming_requested"] != true || readback["mutation_authority"] != false || readback["executes_work"] != false {
		t.Fatalf("A2A streaming denial missing boundary flags: %#v", readback)
	}
}

func TestA2AStreamingDeniedReadbackRejectsSSEFixture(t *testing.T) {
	var out, errb bytes.Buffer
	code := Run([]string{
		"a2a", "streaming-denial",
		"--agent-card", filepath.Join("..", "..", "examples", "invalid", "a2a-agent-card-streaming-sse.json"),
	}, &out, &errb)
	if code != 0 {
		t.Fatalf("a2a streaming-denial should emit SSE denial readback, got stderr=%s", errb.String())
	}
	var readback map[string]any
	if err := json.Unmarshal(out.Bytes(), &readback); err != nil {
		t.Fatal(err)
	}
	if readback["schema"] != "ao.mission.a2a-streaming-denial-readback.v0.1" || readback["status"] != "denied" {
		t.Fatalf("bad A2A SSE denial: %#v", readback)
	}
	if readback["sse_requested"] != true || readback["streaming_requested"] != true || readback["denied_capability"] != "streaming_or_push" {
		t.Fatalf("A2A SSE denial missing requested capability flags: %#v", readback)
	}
	if readback["mutation_authority"] != false || readback["executes_work"] != false || readback["approves_work"] != false {
		t.Fatalf("A2A SSE denial widened authority: %#v", readback)
	}
}

func TestMissionEventIndexSearchAndCLIReadback(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "event search atlas mission"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "next", "--mission", rec.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("next: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "events", "index", "--out", filepath.Join(dir, "mission-event-index.json")}, &out, &errb); code != 0 {
		t.Fatalf("mission events index: %s", errb.String())
	}
	var index MissionEventIndex
	if err := json.Unmarshal(out.Bytes(), &index); err != nil {
		t.Fatal(err)
	}
	if index.Schema != "ao.mission.event-index.v0.2" || index.IndexVersion != "v0.2" || index.TotalEvents < 2 || index.ExecutesWork || index.ApprovesWork || index.MutatesRepositories {
		t.Fatalf("bad event index: %+v", index)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "events", "search", "--mission", rec.MissionID, "--kind", "route_decision", "--query", "AO Atlas", "--json"}, &out, &errb); code != 0 {
		t.Fatalf("mission events search: %s", errb.String())
	}
	var results MissionEventSearchReadback
	if err := json.Unmarshal(out.Bytes(), &results); err != nil {
		t.Fatal(err)
	}
	if results.Schema != "ao.mission.event-search-readback.v0.1" || results.TotalMatches == 0 {
		t.Fatalf("bad event search results: %+v", results)
	}
	if results.ExecutesWork || results.ApprovesWork || results.MutatesRepositories {
		t.Fatalf("event search widened authority: %+v", results)
	}
}

func TestDoctorCommandReportsLocalStoreHealthWithoutAuthority(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "doctor mission loop"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "doctor", "--json"}, &out, &errb); code != 0 {
		t.Fatalf("doctor: %s", errb.String())
	}
	var readback MissionDoctorReadback
	if err := json.Unmarshal(out.Bytes(), &readback); err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.doctor-readback.v0.1" || readback.Status != "ready" || readback.MissionCount != 1 {
		t.Fatalf("bad doctor readback: %+v", readback)
	}
	if readback.SafeToExecute || readback.ExecutesWork || readback.ApprovesWork || readback.MutatesRepositories {
		t.Fatalf("doctor widened authority: %+v", readback)
	}
}

func TestSchedulerRecoveryDoesNotRecommendContinuationWhenReplayFresh(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scheduler-fresh-replay.json")
	if err := os.WriteFile(path, []byte(`{
  "schema": "ao.mission.scheduler-replay-fixture.v0.1",
  "evaluated_at_utc": "2026-07-03T12:00:00Z",
  "readbacks": [
    {
      "schema": "ao.mission.scheduler-readback.v0.1",
      "mission_id": "mission-demo",
      "status": "ready",
      "scheduler": "codex-cron",
      "event_loop": true,
      "generated_at_utc": "2026-07-03T11:55:00Z",
      "executes_work": false,
      "approves_work": false
    }
  ]
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	replay, err := ReplaySchedulerReadbacks(path)
	if err != nil {
		t.Fatal(err)
	}
	recovery := BuildSchedulerRecoveryReadback("mission-demo", replay)
	if recovery.Status != "ready" || recovery.RecoveryMode != "none_required" || strings.Contains(recovery.ExactNextAction, "continue --mission") {
		t.Fatalf("fresh scheduler replay should not request recovery continuation: %+v", recovery)
	}
	if recovery.ExecutesWork || recovery.ApprovesWork {
		t.Fatalf("fresh scheduler recovery widened authority: %+v", recovery)
	}
}

func TestTelegramCommandReplayMatrixCoversAllAllowedCommandsAndDeniedRoles(t *testing.T) {
	readback, err := ReplayTelegramCommandMatrix(
		filepath.Join("..", "..", "examples", "valid", "telegram-command-matrix.json"),
		map[string]string{"1001": "admin", "1002": "user"},
	)
	if err != nil {
		t.Fatal(err)
	}
	covered := map[string]bool{}
	for _, result := range readback.Results {
		covered[result.Command] = true
	}
	for command := range allowedTelegramCommands {
		if !covered[command] {
			t.Fatalf("telegram replay matrix missing allowed command %s", command)
		}
	}
	if readback.Total < len(allowedTelegramCommands)+2 || readback.Denied < 2 {
		t.Fatalf("telegram replay matrix should cover allowed and denied role cases: %+v", readback)
	}
	if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
		t.Fatalf("telegram replay matrix widened authority: %+v", readback)
	}
}

func TestMissionEventIndexCarriesVersionAndDigest(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	rec, err := s.Start("event index digest proof")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Continue(s, rec.MissionID, ContinueOptions{UntilDone: true, MaxIterations: 1}); err != nil {
		t.Fatal(err)
	}
	index, err := BuildMissionEventIndex(s)
	if err != nil {
		t.Fatal(err)
	}
	if index.Schema != "ao.mission.event-index.v0.2" || index.IndexVersion != "v0.2" {
		t.Fatalf("event index should be versioned as v0.2: %+v", index)
	}
	if !strings.HasPrefix(index.IndexDigest, "sha256:") || index.IndexDigest == "sha256:" {
		t.Fatalf("event index missing digest: %+v", index)
	}
	if !strings.HasPrefix(index.SourceDigest, "sha256:") || index.SourceDigest == "sha256:" {
		t.Fatalf("event index missing source digest: %+v", index)
	}
	if err := ValidateMissionEventIndexDigest(index); err != nil {
		t.Fatalf("event index digest did not validate: %v", err)
	}
	index.Events[0].Summary = "tampered"
	if err := ValidateMissionEventIndexDigest(index); err == nil {
		t.Fatal("tampered event index digest should fail validation")
	}
}

func TestMissionReadinessBundleVerifierBindsLocalRepoSummaries(t *testing.T) {
	dir := t.TempDir()
	missionReady := filepath.Join(dir, "mission-readiness.txt")
	atlasReady := filepath.Join(dir, "atlas-readiness.txt")
	if err := os.WriteFile(missionReady, []byte("AO Mission production readiness: 100/100 status=ready\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(atlasReady, []byte("status=ready\nscore=100/100\nsummary=target/production-readiness/summary.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	readback, err := BuildMissionReadinessBundleReadback([]MissionReadinessBundleInput{
		{Repo: "ao-mission", Path: missionReady},
		{Repo: "ao-atlas", Path: atlasReady},
	})
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.readiness-bundle-readback.v0.1" || readback.Status != "ready" || readback.ReadyRepos != 2 {
		t.Fatalf("bad readiness bundle: %+v", readback)
	}
	if len(readback.Repos) != 2 || !strings.HasPrefix(readback.Repos[0].SHA256, "sha256:") {
		t.Fatalf("readiness bundle missing repo digest evidence: %+v", readback)
	}
	if readback.SafeToExecute || readback.ExecutesWork || readback.ApprovesWork || readback.MutatesRepositories {
		t.Fatalf("readiness bundle widened authority: %+v", readback)
	}
	outPath := filepath.Join(dir, "readiness-bundle.json")
	var out, errb bytes.Buffer
	if code := Run([]string{
		"mission", "readiness-bundle",
		"--repo", "ao-mission=" + missionReady,
		"--repo", "ao-atlas=" + atlasReady,
		"--out", outPath,
	}, &out, &errb); code != 0 {
		t.Fatalf("mission readiness-bundle CLI: %s", errb.String())
	}
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"schema": "ao.mission.readiness-bundle-readback.v0.1"`) {
		t.Fatalf("readiness bundle CLI did not write expected readback: %s", string(body))
	}
}

func TestGatewayReplayBundleBindsSchedulerTelegramAndA2A(t *testing.T) {
	readback, err := BuildGatewayReplayBundleReadback(GatewayReplayBundleInputs{
		TelegramConfigPath:  filepath.Join("..", "..", "examples", "valid", "telegram-config.json"),
		TelegramMatrixPath:  filepath.Join("..", "..", "examples", "valid", "telegram-command-matrix.json"),
		TelegramUpdatesPath: filepath.Join("..", "..", "examples", "valid", "telegram-update-replay.json"),
		TelegramWebhookPath: filepath.Join("..", "..", "examples", "valid", "telegram-webhook-replay.json"),
		A2AHTTPPath:         filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"),
		A2ALifecyclePath:    filepath.Join("..", "..", "examples", "valid", "a2a-task-lifecycle-artifacts.json"),
		SchedulerPath:       filepath.Join("..", "..", "examples", "valid", "scheduler-readback-replay.json"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if readback.Schema != "ao.mission.gateway-replay-bundle-readback.v0.1" || readback.Status != "ready" {
		t.Fatalf("bad replay bundle: %+v", readback)
	}
	if readback.TelegramReadbacks != 3 || readback.A2AReadbacks != 2 || readback.SchedulerReadbacks != 1 {
		t.Fatalf("replay bundle missing expected readback families: %+v", readback)
	}
	if readback.SafeToExecute || readback.ExecutesWork || readback.ApprovesWork || readback.MutatesRepositories {
		t.Fatalf("replay bundle widened authority: %+v", readback)
	}
	outPath := filepath.Join(t.TempDir(), "gateway-replay-bundle.json")
	var out, errb bytes.Buffer
	if code := Run([]string{
		"gateway", "replay-bundle",
		"--telegram-config", filepath.Join("..", "..", "examples", "valid", "telegram-config.json"),
		"--telegram-matrix", filepath.Join("..", "..", "examples", "valid", "telegram-command-matrix.json"),
		"--telegram-updates", filepath.Join("..", "..", "examples", "valid", "telegram-update-replay.json"),
		"--a2a-http", filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"),
		"--scheduler", filepath.Join("..", "..", "examples", "valid", "scheduler-readback-replay.json"),
		"--out", outPath,
	}, &out, &errb); code != 0 {
		t.Fatalf("gateway replay-bundle CLI: %s", errb.String())
	}
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"schema": "ao.mission.gateway-replay-bundle-readback.v0.1"`) {
		t.Fatalf("replay bundle CLI did not write expected readback: %s", string(body))
	}
}

func TestMissionDashboardReadbackSummarizesCompactOperatorLoop(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", dir, "start", "dashboard readback atlas loop"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "next", "--mission", rec.MissionID, "--json"}, &out, &errb); code != 0 {
		t.Fatalf("next: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"--home", dir, "mission", "dashboard", "--mission", rec.MissionID, "--compact", "--json"}, &out, &errb); code != 0 {
		t.Fatalf("mission dashboard: %s", errb.String())
	}
	var dashboard MissionDashboardReadback
	if err := json.Unmarshal(out.Bytes(), &dashboard); err != nil {
		t.Fatal(err)
	}
	if dashboard.Schema != "ao.mission.dashboard-readback.v0.1" || dashboard.MissionID != rec.MissionID || dashboard.EventCount == 0 {
		t.Fatalf("bad dashboard readback: %+v", dashboard)
	}
	if !dashboard.Compact || dashboard.LatestRoute == "" || dashboard.EventIndexDigest == "" {
		t.Fatalf("dashboard missing compact route/index evidence: %+v", dashboard)
	}
	if dashboard.SafeToExecute || dashboard.ExecutesWork || dashboard.ApprovesWork || dashboard.MutatesRepositories {
		t.Fatalf("dashboard widened authority: %+v", dashboard)
	}
}

func TestGatewayReadinessRollupCombinesReadbacksWithoutAuthority(t *testing.T) {
	dir := t.TempDir()
	suitePath := filepath.Join(dir, "suite.json")
	compatPath := filepath.Join(dir, "compat.json")
	archiveValidationPath := filepath.Join(dir, "archive-validation.json")
	diffPath := filepath.Join(dir, "snapshot-diff.json")
	rollupPath := filepath.Join(dir, "gateway-readiness-rollup.json")
	var out, errb bytes.Buffer
	if code := Run([]string{
		"gateway", "replay-suite",
		"--telegram-config", filepath.Join("..", "..", "examples", "valid", "telegram-config.json"),
		"--telegram-updates", filepath.Join("..", "..", "examples", "valid", "telegram-update-replay.json"),
		"--a2a-http", filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"),
		"--out", suitePath,
	}, &out, &errb); code != 0 {
		t.Fatalf("gateway replay-suite: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{
		"a2a", "compatibility",
		"--agent-card", filepath.Join("..", "..", "examples", "valid", "a2a-agent-card.json"),
		"--http", filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"),
		"--lifecycle", filepath.Join("..", "..", "examples", "valid", "a2a-task-lifecycle-artifacts.json"),
		"--out", compatPath,
	}, &out, &errb); code != 0 {
		t.Fatalf("a2a compatibility: %s", errb.String())
	}
	exportHome := filepath.Join(dir, "export-home")
	out.Reset()
	if code := Run([]string{"--home", exportHome, "start", "gateway readiness rollup"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(dir, "archive.json")
	out.Reset()
	if code := Run([]string{"--home", exportHome, "mission", "archive", "--mission", rec.MissionID, "--out", archivePath}, &out, &errb); code != 0 {
		t.Fatalf("archive: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{"mission", "validate-archive", "--path", archivePath, "--out", archiveValidationPath}, &out, &errb); code != 0 {
		t.Fatalf("validate archive: %s", errb.String())
	}
	beforePath := filepath.Join(dir, "before.json")
	afterPath := filepath.Join(dir, "after.json")
	before := Snapshot(rec)
	rec.CurrentRoute = "ao-atlas"
	after := Snapshot(rec)
	writeJSONForTest(t, beforePath, before)
	writeJSONForTest(t, afterPath, after)
	out.Reset()
	if code := Run([]string{"governance", "diff", "--before", beforePath, "--after", afterPath, "--out", diffPath}, &out, &errb); code != 0 {
		t.Fatalf("governance diff: %s", errb.String())
	}
	out.Reset()
	if code := Run([]string{
		"gateway", "readiness-rollup",
		"--mission", rec.MissionID,
		"--suite", suitePath,
		"--a2a-compatibility", compatPath,
		"--archive-validation", archiveValidationPath,
		"--snapshot-diff", diffPath,
		"--out", rollupPath,
	}, &out, &errb); code != 0 {
		t.Fatalf("gateway readiness-rollup: %s", errb.String())
	}
	var rollup map[string]any
	body, err := os.ReadFile(rollupPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &rollup); err != nil {
		t.Fatal(err)
	}
	if rollup["schema"] != "ao.mission.gateway-readiness-rollup.v0.1" || rollup["status"] != "ready" {
		t.Fatalf("bad gateway readiness rollup: %#v", rollup)
	}
	if rollup["mission_id"] != rec.MissionID {
		t.Fatalf("gateway readiness rollup missing mission_id: %#v", rollup)
	}
	if rollup["readback_count"] != float64(4) || rollup["safe_to_execute"] != false || rollup["executes_work"] != false {
		t.Fatalf("gateway readiness rollup missing no-authority evidence: %#v", rollup)
	}
}

func TestGatewayReadinessRollupCarriesCorrelationID(t *testing.T) {
	dir := t.TempDir()
	readbackPath := filepath.Join(dir, "gateway-readback.json")
	outPath := filepath.Join(dir, "gateway-readiness-rollup.json")
	if err := os.WriteFile(readbackPath, []byte(`{
  "schema": "ao.mission.gateway-replay-suite-readback.v0.1",
  "status": "ready",
  "safe_to_execute": false,
  "executes_work": false,
  "approves_work": false,
  "mutation_authority": false
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := Run([]string{
		"gateway", "readiness-rollup",
		"--mission", "mission-demo",
		"--suite", readbackPath,
		"--correlation-id", "corr-gateway-001",
		"--out", outPath,
	}, &out, &errb)
	if code != 0 {
		t.Fatalf("gateway readiness-rollup with correlation failed: %s", errb.String())
	}
	var rollup map[string]any
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &rollup); err != nil {
		t.Fatal(err)
	}
	if rollup["correlation_id"] != "corr-gateway-001" {
		t.Fatalf("rollup missing correlation_id: %#v", rollup)
	}
	if rollup["mission_id"] != "mission-demo" {
		t.Fatalf("rollup missing mission_id: %#v", rollup)
	}
	if rollup["safe_to_execute"] != false || rollup["executes_work"] != false || rollup["approves_work"] != false {
		t.Fatalf("correlated rollup widened authority: %#v", rollup)
	}
}

func TestGatewayReadinessRollupDerivesCorrelationIDFromTelegramReplaySuite(t *testing.T) {
	dir := t.TempDir()
	suitePath := filepath.Join(dir, "gateway-replay-suite.json")
	rollupPath := filepath.Join(dir, "gateway-readiness-rollup.json")
	telegramPath := filepath.Join("..", "..", "examples", "valid", "telegram-update-replay.json")
	var out, errb bytes.Buffer
	if code := Run([]string{
		"gateway", "replay-suite",
		"--telegram-config", filepath.Join("..", "..", "examples", "valid", "telegram-config.json"),
		"--telegram-updates", telegramPath,
		"--out", suitePath,
	}, &out, &errb); code != 0 {
		t.Fatalf("gateway replay-suite: %s", errb.String())
	}
	var suite map[string]any
	body, err := os.ReadFile(suitePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &suite); err != nil {
		t.Fatal(err)
	}
	if suite["correlation_id"] != "corr-telegram-replay-001" {
		t.Fatalf("replay suite did not carry Telegram correlation ID: %#v", suite)
	}
	out.Reset()
	if code := Run([]string{
		"gateway", "readiness-rollup",
		"--mission", "mission-demo",
		"--suite", suitePath,
		"--out", rollupPath,
	}, &out, &errb); code != 0 {
		t.Fatalf("gateway readiness-rollup: %s", errb.String())
	}
	var rollup map[string]any
	body, err = os.ReadFile(rollupPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &rollup); err != nil {
		t.Fatal(err)
	}
	if rollup["correlation_id"] != "corr-telegram-replay-001" {
		t.Fatalf("rollup did not derive correlation ID from replay suite: %#v", rollup)
	}
	if rollup["safe_to_execute"] != false || rollup["executes_work"] != false || rollup["approves_work"] != false {
		t.Fatalf("derived correlation rollup widened authority: %#v", rollup)
	}
}

func TestA2ACancellationReplayRequiresRequestAndCancelWithoutAuthority(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "a2a-cancellation-replay.json")
	var out, errb bytes.Buffer
	code := Run([]string{
		"a2a", "cancellation-replay",
		"--lifecycle", filepath.Join("..", "..", "examples", "valid", "a2a-task-lifecycle.json"),
		"--out", outPath,
	}, &out, &errb)
	if code != 0 {
		t.Fatalf("a2a cancellation-replay failed: %s", errb.String())
	}
	var replay map[string]any
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &replay); err != nil {
		t.Fatal(err)
	}
	if replay["schema"] != "ao.mission.a2a-cancellation-replay-readback.v0.1" || replay["status"] != "ready" {
		t.Fatalf("bad cancellation replay: %#v", replay)
	}
	if replay["cancel_requested"] != float64(1) || replay["cancelled"] != float64(1) {
		t.Fatalf("cancellation replay missing request/cancel counts: %#v", replay)
	}
	if replay["mutation_authority"] != false || replay["executes_work"] != false || replay["approves_work"] != false {
		t.Fatalf("cancellation replay widened authority: %#v", replay)
	}
}

func TestMissionV02ReadbackContractFixturesValidate(t *testing.T) {
	for _, path := range []string{
		filepath.Join("..", "..", "examples", "valid", "mission-event-index-v0.2.json"),
		filepath.Join("..", "..", "examples", "valid", "mission-readiness-bundle-readback.json"),
		filepath.Join("..", "..", "examples", "valid", "gateway-replay-bundle-readback.json"),
		filepath.Join("..", "..", "examples", "valid", "mission-dashboard-readback.json"),
		filepath.Join("..", "..", "examples", "valid", "mission-verification-bundle-readback.json"),
	} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			result, err := ValidateContractFile(path)
			if err != nil {
				t.Fatalf("valid fixture rejected: result=%+v err=%v", result, err)
			}
			if result.Status != "ready" || result.Executes || result.Approves || result.Mutates {
				t.Fatalf("bad contract validation result: %+v", result)
			}
		})
	}
}

func TestMissionV02ReadbackContractFixturesRejectAuthorityDrift(t *testing.T) {
	for _, path := range []string{
		filepath.Join("..", "..", "examples", "invalid", "mission-readiness-bundle-authority.json"),
		filepath.Join("..", "..", "examples", "invalid", "gateway-replay-bundle-authority.json"),
		filepath.Join("..", "..", "examples", "invalid", "mission-dashboard-authority.json"),
		filepath.Join("..", "..", "examples", "invalid", "mission-verification-bundle-authority.json"),
	} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			result, err := ValidateContractFile(path)
			if err == nil || result.Status != "blocked" {
				t.Fatalf("invalid fixture accepted: result=%+v err=%v", result, err)
			}
		})
	}
}

func TestMissionVerificationBundleBindsReadbacksAndRejectsAuthorityDrift(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(filepath.Join(dir, "home"))
	rec, err := s.Start("verification bundle operator handoff")
	if err != nil {
		t.Fatal(err)
	}
	readyPath := filepath.Join(dir, "readiness.txt")
	if err := os.WriteFile(readyPath, []byte("status=ready\nscore=100/100\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	readiness, err := BuildMissionReadinessBundleReadback([]MissionReadinessBundleInput{{Repo: "ao-mission", Path: readyPath}})
	if err != nil {
		t.Fatal(err)
	}
	readinessPath := filepath.Join(dir, "readiness-bundle.json")
	writeJSONForTest(t, readinessPath, readiness)
	replay, err := BuildGatewayReplayBundleReadback(GatewayReplayBundleInputs{
		TelegramConfigPath: filepath.Join("..", "..", "examples", "valid", "telegram-config.json"),
		TelegramMatrixPath: filepath.Join("..", "..", "examples", "valid", "telegram-command-matrix.json"),
		A2AHTTPPath:        filepath.Join("..", "..", "examples", "valid", "a2a-http-integration.json"),
		SchedulerPath:      filepath.Join("..", "..", "examples", "valid", "scheduler-readback-replay.json"),
	})
	if err != nil {
		t.Fatal(err)
	}
	replayPath := filepath.Join(dir, "gateway-replay-bundle.json")
	writeJSONForTest(t, replayPath, replay)
	bundle, err := BuildMissionVerificationBundleReadback(s, rec.MissionID, MissionVerificationBundleOptions{
		ReadinessBundlePath:     readinessPath,
		GatewayReplayBundlePath: replayPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Schema != "ao.mission.verification-bundle-readback.v0.1" || bundle.Status != "ready" || bundle.ComponentCount < 5 {
		t.Fatalf("bad verification bundle: %+v", bundle)
	}
	if !strings.HasPrefix(bundle.BundleDigest, "sha256:") {
		t.Fatalf("verification bundle missing digest: %+v", bundle)
	}
	if bundle.SafeToExecute || bundle.ExecutesWork || bundle.ApprovesWork || bundle.MutatesRepositories {
		t.Fatalf("verification bundle widened authority: %+v", bundle)
	}

	var unsafe map[string]any
	body, err := os.ReadFile(readinessPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &unsafe); err != nil {
		t.Fatal(err)
	}
	unsafe["safe_to_execute"] = true
	unsafePath := filepath.Join(dir, "unsafe-readiness-bundle.json")
	writeJSONForTest(t, unsafePath, unsafe)
	if _, err := BuildMissionVerificationBundleReadback(s, rec.MissionID, MissionVerificationBundleOptions{ReadinessBundlePath: unsafePath}); err == nil {
		t.Fatal("verification bundle accepted authority drift")
	}
}

func TestMissionVerificationBundleCLIWritesDigestManifest(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	var out, errb bytes.Buffer
	if code := Run([]string{"--home", home, "start", "verification bundle cli"}, &out, &errb); code != 0 {
		t.Fatalf("start: %s", errb.String())
	}
	var rec Record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	readyPath := filepath.Join(dir, "readiness.txt")
	if err := os.WriteFile(readyPath, []byte("AO Mission production readiness: 100/100 status=ready\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	readiness, err := BuildMissionReadinessBundleReadback([]MissionReadinessBundleInput{{Repo: "ao-mission", Path: readyPath}})
	if err != nil {
		t.Fatal(err)
	}
	readinessPath := filepath.Join(dir, "readiness-bundle.json")
	writeJSONForTest(t, readinessPath, readiness)
	outPath := filepath.Join(dir, "verification-bundle.json")
	out.Reset()
	if code := Run([]string{
		"--home", home,
		"mission", "verification-bundle",
		"--mission", rec.MissionID,
		"--readiness-bundle", readinessPath,
		"--out", outPath,
	}, &out, &errb); code != 0 {
		t.Fatalf("mission verification-bundle: %s", errb.String())
	}
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var bundle MissionVerificationBundleReadback
	if err := json.Unmarshal(body, &bundle); err != nil {
		t.Fatal(err)
	}
	if bundle.MissionID != rec.MissionID || bundle.ComponentCount == 0 || !strings.HasPrefix(bundle.BundleDigest, "sha256:") {
		t.Fatalf("bad CLI verification bundle: %+v", bundle)
	}
}

func digestBytesForTest(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func writeJSONForTest(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
