package mission

import (
	"bytes"
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
	if len(r.Steps) != 1 {
		t.Fatalf("steps=%d", len(r.Steps))
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
	if decision.Status != "handoff_required" || decision.Route != "ao-atlas" || decision.Iteration != 1 {
		t.Fatalf("unexpected event loop decision: %+v", decision)
	}
	if decision.ExecutesWork || decision.ApprovesWork || decision.MutatesRepositories {
		t.Fatalf("event loop widened authority: %+v", decision)
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
