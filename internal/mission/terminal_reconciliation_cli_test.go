package mission

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCreateSliceCheckpointAppendsEvidenceBoundS01WithoutLifecycleMutation(t *testing.T) {
	home := t.TempDir()
	store := NewStore(home)
	contract, err := store.StartObjective(
		"Coordinate one bounded implementation workgraph",
		ObjectiveStartOptions{CorrelationID: "cross-platform-baseline-test"},
	)
	if err != nil {
		t.Fatal(err)
	}
	record, err := store.Load(contract.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	digest := writeSliceCheckpointEvidence(t, store, record, "S01", nil)
	before, err := store.Load(record.MissionID)
	if err != nil {
		t.Fatal(err)
	}

	bundle, err := CreateSliceCheckpoint(store, record.MissionID, SliceCheckpointOptions{
		Slice: "S01", EvidenceDigest: digest,
	})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.CheckpointCount != 1 || bundle.LatestCheckpoint == nil {
		t.Fatalf("missing S01 checkpoint: %+v", bundle)
	}
	if bundle.LatestCheckpoint.Result != "slice_pass:S01:"+digest {
		t.Fatalf("wrong evidence binding: %+v", bundle.LatestCheckpoint)
	}
	if bundle.ExecutesWork || bundle.ApprovesWork || bundle.MutatesRepositories {
		t.Fatalf("slice checkpoint widened authority: %+v", bundle)
	}

	after, err := store.Load(record.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	before.UpdatedAtUTC = ""
	after.UpdatedAtUTC = ""
	before.Checkpoints = nil
	after.Checkpoints = nil
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("slice checkpoint changed Mission lifecycle:\nbefore=%+v\nafter=%+v", before, after)
	}
}

func TestCreateSliceCheckpointReplayConflictAndOrder(t *testing.T) {
	newMission := func(t *testing.T) (Store, Record) {
		t.Helper()
		store := NewStore(t.TempDir())
		contract, err := store.StartObjective(
			"Coordinate one bounded implementation workgraph",
			ObjectiveStartOptions{CorrelationID: "slice-order-test"},
		)
		if err != nil {
			t.Fatal(err)
		}
		record, err := store.Load(contract.MissionID)
		if err != nil {
			t.Fatal(err)
		}
		return store, record
	}

	t.Run("exact replay is idempotent", func(t *testing.T) {
		store, record := newMission(t)
		digest := writeSliceCheckpointEvidence(t, store, record, "S01", nil)
		options := SliceCheckpointOptions{Slice: "S01", EvidenceDigest: digest}
		first, err := CreateSliceCheckpoint(store, record.MissionID, options)
		if err != nil {
			t.Fatal(err)
		}
		replay, err := CreateSliceCheckpoint(store, record.MissionID, options)
		if err != nil {
			t.Fatal(err)
		}
		if replay.CheckpointCount != first.CheckpointCount {
			t.Fatalf("exact replay appended a checkpoint: first=%d replay=%d", first.CheckpointCount, replay.CheckpointCount)
		}
	})

	t.Run("same slice with another digest conflicts", func(t *testing.T) {
		store, record := newMission(t)
		first := writeSliceCheckpointEvidence(t, store, record, "S01", nil)
		second := writeSliceCheckpointEvidence(t, store, record, "S01", func(document map[string]any) {
			document["producer_note"] = "distinct reviewed evidence"
		})
		if _, err := CreateSliceCheckpoint(store, record.MissionID, SliceCheckpointOptions{Slice: "S01", EvidenceDigest: first}); err != nil {
			t.Fatal(err)
		}
		_, err := CreateSliceCheckpoint(store, record.MissionID, SliceCheckpointOptions{Slice: "S01", EvidenceDigest: second})
		if err == nil || !strings.Contains(err.Error(), "slice S01 already checkpointed with a different evidence digest") {
			t.Fatalf("conflicting replay was not rejected: %v", err)
		}
	})

	t.Run("S02 requires S01", func(t *testing.T) {
		store, record := newMission(t)
		digest := writeSliceCheckpointEvidence(t, store, record, "S02", nil)
		_, err := CreateSliceCheckpoint(store, record.MissionID, SliceCheckpointOptions{Slice: "S02", EvidenceDigest: digest})
		if err == nil || !strings.Contains(err.Error(), "slice checkpoint is out of order") {
			t.Fatalf("out-of-order S02 was not rejected: %v", err)
		}
	})

	t.Run("S02 follows S01 but S03 cannot skip S02", func(t *testing.T) {
		store, record := newMission(t)
		s01 := writeSliceCheckpointEvidence(t, store, record, "S01", nil)
		s02 := writeSliceCheckpointEvidence(t, store, record, "S02", nil)
		s03 := writeSliceCheckpointEvidence(t, store, record, "S03", nil)
		if _, err := CreateSliceCheckpoint(store, record.MissionID, SliceCheckpointOptions{Slice: "S01", EvidenceDigest: s01}); err != nil {
			t.Fatal(err)
		}
		if _, err := CreateSliceCheckpoint(store, record.MissionID, SliceCheckpointOptions{Slice: "S03", EvidenceDigest: s03}); err == nil || !strings.Contains(err.Error(), "slice checkpoint is out of order") {
			t.Fatalf("skipped S02 was not rejected: %v", err)
		}
		bundle, err := CreateSliceCheckpoint(store, record.MissionID, SliceCheckpointOptions{Slice: "S02", EvidenceDigest: s02})
		if err != nil {
			t.Fatal(err)
		}
		if bundle.CheckpointCount != 2 || bundle.LatestCheckpoint.Result != "slice_pass:S02:"+s02 {
			t.Fatalf("S02 checkpoint mismatch: %+v", bundle)
		}
	})
}

func writeSliceCheckpointEvidence(
	t *testing.T,
	store Store,
	record Record,
	slice string,
	mutate func(map[string]any),
) string {
	t.Helper()
	document := map[string]any{
		"schema":         "ao.architecture.development-baseline-slice-evidence.v1",
		"correlation_id": record.CorrelationID,
		"mission_ref":    record.MissionID,
		"slice":          slice,
		"result":         "pass",
		"authority": map[string]any{
			"safe_to_execute":          false,
			"executes_work":            false,
			"approves_work":            false,
			"mutates_repositories":     false,
			"provider_calls":           false,
			"credential_use":           false,
			"release":                  false,
			"publication":              false,
			"deployment":               false,
			"promotion":                false,
			"compatibility_activation": false,
			"external_beta":            false,
			"rsi":                      false,
		},
	}
	if mutate != nil {
		mutate(document)
	}
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	contentRef, digest, err := store.retainArtifact(body)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Update(record.MissionID, func(candidate *Record) error {
		candidate.ArtifactRefs = append(candidate.ArtifactRefs, ArtifactRef{
			Schema: ArtifactRefSchema, Ref: "slice-evidence.json", ContentRef: contentRef,
			Digest: digest, Kind: "correlation-evidence",
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return digest
}

func TestOperatorReadbackSchemasDeclareTerminalProjectionFields(t *testing.T) {
	for _, name := range []string{"command-status-v0.1.schema.json", "dashboard-readback-v0.1.schema.json"} {
		body, err := os.ReadFile(filepath.Join("..", "..", "docs", "contracts", name))
		if err != nil {
			t.Fatal(err)
		}
		var schema struct {
			Properties map[string]struct {
				Type string `json:"type"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(body, &schema); err != nil {
			t.Fatal(err)
		}
		for field, want := range map[string]string{
			"source_record_status":          "string",
			"terminal_projection_status":    "string",
			"terminal_projection_read_only": "boolean",
			"effective_operator_status":     "string",
		} {
			if got := schema.Properties[field].Type; got != want {
				t.Errorf("%s property %s type = %q, want %q", name, field, got, want)
			}
		}
	}
}

func TestCheckpointCreateAppendsOneIdempotentReadOnlyCheckpoint(t *testing.T) {
	home := t.TempDir()
	store := NewStore(home)
	record, err := store.Start("checkpoint the current operator state")
	if err != nil {
		t.Fatal(err)
	}
	before := record
	for attempt := 0; attempt < 2; attempt++ {
		var stdout, stderr bytes.Buffer
		if code := Run([]string{"--home", home, "checkpoint", "create", "--mission", record.MissionID, "--json"}, &stdout, &stderr); code != 0 {
			t.Fatalf("checkpoint create attempt %d: code=%d stderr=%s", attempt+1, code, stderr.String())
		}
		var bundle MissionCheckpointBundle
		if err := json.Unmarshal(stdout.Bytes(), &bundle); err != nil {
			t.Fatal(err)
		}
		if bundle.CheckpointCount != 1 || bundle.LatestCheckpoint == nil || bundle.LatestCheckpoint.Result != "checkpoint_created" ||
			bundle.ExecutesWork || bundle.ApprovesWork || bundle.MutatesRepositories {
			t.Fatalf("unexpected checkpoint bundle: %+v", bundle)
		}
	}
	after, err := store.Load(record.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Checkpoints) != 1 || len(after.Steps) != len(before.Steps) ||
		after.Status != before.Status || after.CurrentRoute != before.CurrentRoute ||
		after.CurrentPhase != before.CurrentPhase || after.ExactNextAction != before.ExactNextAction {
		t.Fatalf("checkpoint create advanced Mission state: before=%+v after=%+v", before, after)
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"--home", home, "checkpoint", "create", "--json"}, &stdout, &stderr); code == 0 ||
		!strings.Contains(stderr.String(), "requires --mission") {
		t.Fatalf("missing Mission identity code=%d stderr=%q", code, stderr.String())
	}
}

func TestCheckpointCreateDoesNotChangePauseSemantics(t *testing.T) {
	store := NewStore(t.TempDir())
	record, err := store.Start("pause without creating a checkpoint")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Pause(store, record.MissionID); err != nil {
		t.Fatal(err)
	}
	after, err := store.Load(record.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Checkpoints) != 0 {
		t.Fatalf("pause created checkpoints: %+v", after.Checkpoints)
	}
}

func TestGenericMissionViewsProjectValidatedTerminalState(t *testing.T) {
	home := t.TempDir()
	store := NewStore(home)
	record, err := store.Start("run governed pool external beta")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Update(record.MissionID, func(candidate *Record) error {
		candidate.CurrentPhase = "lifecycle-canary"
		candidate.ExactNextAction = "continue lifecycle canary"
		candidate.GoalLease = &GoalLease{Schema: GoalLeaseSchema, MaxIterations: 1, CheckpointPolicy: "after_each_node_or_timed_interval"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	statePath := writeTerminalProjectionState(t, store, record.MissionID, func(*TerminalIndexImportReadback) {})

	for _, test := range []struct {
		name string
		args []string
	}{
		{name: "status", args: []string{"status", "--mission", record.MissionID, "--terminal-state", statePath, "--json"}},
		{name: "inspect", args: []string{"mission", "inspect", "--mission", record.MissionID, "--terminal-state", statePath, "--json"}},
		{name: "dashboard", args: []string{"mission", "dashboard", "--mission", record.MissionID, "--terminal-state", statePath, "--json"}},
		{name: "command", args: []string{"command", "status", "--mission", record.MissionID, "--terminal-state", statePath, "--json"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			args := append([]string{"--home", home}, test.args...)
			if code := Run(args, &stdout, &stderr); code != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			text := stdout.String()
			for _, want := range []string{
				`"status": "done"`, `"current_phase": "reconciled"`, `"exact_next_action": "none"`,
				`"source_record_status": "active"`, `"terminal_projection_status": "done"`,
				`"terminal_projection_read_only": true`, `"effective_operator_status": "done"`,
			} {
				if !strings.Contains(text, want) {
					t.Fatalf("view does not project %s: %s", want, text)
				}
			}
			if !strings.Contains(text, `"completed": 7`) && !strings.Contains(text, `"completed_nodes": 7`) {
				t.Fatalf("view does not project terminal counts: %s", text)
			}
		})
	}

	persisted, err := store.Load(record.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Status != "active" || persisted.CurrentPhase != "lifecycle-canary" {
		t.Fatalf("read-only terminal projection mutated Mission: %+v", persisted)
	}
	persistedJSON, err := json.Marshal(persisted)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(persistedJSON), "terminal_projection_status") {
		t.Fatalf("persisted Mission contains read-only projection fields: %s", persistedJSON)
	}
	projected, err := projectRecordWithTerminalState(persisted, statePath)
	if err != nil {
		t.Fatal(err)
	}
	if projected.Reconciliation == nil || projected.Reconciliation.Status != "reconciled" ||
		projected.Reconciliation.AtlasReadyNodes != 0 || projected.Reconciliation.ExactNextAction != "none" {
		t.Fatalf("route reconciliation contradicts terminal state: %+v", projected.Reconciliation)
	}
	if projected.GoalLease == nil || projected.GoalLease.MaxIterations != 1 ||
		projected.GoalLease.CheckpointPolicy != "after_each_node_or_timed_interval" {
		t.Fatalf("terminal projection replaced the Mission lease policy: %+v", projected.GoalLease)
	}
}

func TestTerminalProjectionDistinguishesSourceAndEffectiveStatuses(t *testing.T) {
	tests := []struct {
		name           string
		sourceStatus   string
		edit           func(*TerminalIndexImportReadback)
		terminalStatus string
		effective      string
	}{
		{
			name: "active source plus nonterminal projection", sourceStatus: "active",
			edit: func(state *TerminalIndexImportReadback) {
				state.Status = "reconciled_fail_closed"
				state.Counts.Completed = 6
				state.Counts.Ready = 1
				state.CompletionObserved = false
				state.ReadinessPassed = false
				state.ReturnGateStatus = "early_return_denied"
				state.FinalResponseAllowed = false
				state.ExactNextAction = "continue node 7"
			},
			terminalStatus: "active", effective: "active",
		},
		{
			name: "done source plus done projection", sourceStatus: "done",
			edit: func(*TerminalIndexImportReadback) {}, terminalStatus: "done", effective: "done",
		},
		{
			name: "done source plus nonterminal projection", sourceStatus: "done",
			edit: func(state *TerminalIndexImportReadback) {
				state.Status = "reconciled_fail_closed"
				state.Counts.Completed = 6
				state.Counts.Ready = 1
				state.CompletionObserved = false
				state.ReadinessPassed = false
				state.ReturnGateStatus = "early_return_denied"
				state.FinalResponseAllowed = false
				state.ExactNextAction = "continue node 7"
			},
			terminalStatus: "active", effective: "done",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			store := NewStore(home)
			record, err := store.Start(test.name)
			if err != nil {
				t.Fatal(err)
			}
			if test.sourceStatus == "done" {
				record, err = store.Update(record.MissionID, func(candidate *Record) error {
					candidate.Status = "done"
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			statePath := writeTerminalProjectionState(t, store, record.MissionID, test.edit)
			var stdout, stderr bytes.Buffer
			if code := Run([]string{"--home", home, "status", "--mission", record.MissionID, "--terminal-state", statePath, "--json"}, &stdout, &stderr); code != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			for _, want := range []string{
				`"source_record_status": "` + test.sourceStatus + `"`,
				`"terminal_projection_status": "` + test.terminalStatus + `"`,
				`"terminal_projection_read_only": true`,
				`"effective_operator_status": "` + test.effective + `"`,
			} {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("projection missing %s: %s", want, stdout.String())
				}
			}
		})
	}
}

func TestDurableMissionRejectsTerminalProjectionFields(t *testing.T) {
	for _, operation := range []string{"save", "load"} {
		t.Run(operation, func(t *testing.T) {
			store := NewStore(t.TempDir())
			record, err := store.Start("keep terminal projection out of durable state")
			if err != nil {
				t.Fatal(err)
			}
			record.SourceRecordStatus = "active"
			record.TerminalProjectionStatus = "done"
			record.TerminalProjectionReadOnly = true
			record.EffectiveOperatorStatus = "done"
			if operation == "save" {
				if err := store.Save(record); err == nil || !strings.Contains(err.Error(), "projection") {
					t.Fatalf("Save error = %v, want projection rejection", err)
				}
				return
			}
			body, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(store.path(record.MissionID), append(body, '\n'), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := store.Load(record.MissionID); err == nil || !strings.Contains(err.Error(), "projection") {
				t.Fatalf("Load error = %v, want projection rejection", err)
			}
		})
	}
}

func TestGenericMissionViewsRejectInvalidTerminalState(t *testing.T) {
	tests := []struct {
		name string
		edit func(*TerminalIndexImportReadback)
		want string
	}{
		{name: "wrong mission", edit: func(state *TerminalIndexImportReadback) { state.MissionID = "mission-other" }, want: "mission identity"},
		{name: "stale", edit: func(state *TerminalIndexImportReadback) { state.GeneratedAtUTC = "2020-01-01T00:00:00Z" }, want: "stale"},
		{name: "unsafe", edit: func(state *TerminalIndexImportReadback) { state.ExecutesWork = true }, want: "safety"},
		{name: "contradictory", edit: func(state *TerminalIndexImportReadback) { state.Counts.Ready = 1 }, want: "contradictory"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			store := NewStore(home)
			record, err := store.Start("reject bad terminal projection")
			if err != nil {
				t.Fatal(err)
			}
			statePath := writeTerminalProjectionState(t, store, record.MissionID, test.edit)
			var stdout, stderr bytes.Buffer
			code := Run([]string{"--home", home, "status", "--mission", record.MissionID, "--terminal-state", statePath, "--json"}, &stdout, &stderr)
			if code == 0 || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("code=%d error=%q want %q", code, stderr.String(), test.want)
			}
		})
	}

	home := t.TempDir()
	store := NewStore(home)
	record, err := store.Start("reject altered terminal projection")
	if err != nil {
		t.Fatal(err)
	}
	statePath := writeTerminalProjectionState(t, store, record.MissionID, func(*TerminalIndexImportReadback) {})
	body, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.Replace(body, []byte(`"completed": 7`), []byte(`"completed": 6`), 1)
	if err := os.WriteFile(statePath, body, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"--home", home, "status", "--mission", record.MissionID, "--terminal-state", statePath, "--json"}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "digest mismatch") {
		t.Fatalf("altered state code=%d error=%q", code, stderr.String())
	}
}

func writeTerminalProjectionState(t *testing.T, store Store, missionID string, edit func(*TerminalIndexImportReadback)) string {
	t.Helper()
	record, err := store.Load(missionID)
	if err != nil {
		t.Fatal(err)
	}
	state := TerminalIndexImportReadback{
		Schema: TerminalIndexImportSchema, Surface: "import", MissionID: missionID,
		IndexDigest:    "sha256:afbf4c71026c5214495eb90ccfc18eb023c9613285e17a0a14c2a022c0e00101",
		GeneratedAtUTC: record.UpdatedAtUTC, Status: "reconciled",
		Counts:             TerminalIndexCounts{Total: 7, Minimum: 7, Completed: 7},
		Lease:              TerminalIndexLease{TargetMinutes: 120, MaximumMinutes: 360, ElapsedMinutes: 115, Status: "within_window"},
		CompletionObserved: true, TimingCompliant: true, CanonicalEvidenceAgreement: true,
		ReadinessPassed: true, ReturnGateStatus: "final_response_allowed", FinalResponseAllowed: true,
		ConflictCodes: []string{}, ExactNextAction: "none", ReadOnly: true,
	}
	edit(&state)
	signTerminalIndexImport(&state)
	path := filepath.Join(t.TempDir(), "terminal-state.json")
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
