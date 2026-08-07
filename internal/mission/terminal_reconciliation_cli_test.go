package mission

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
			for _, want := range []string{`"status": "done"`, `"current_phase": "reconciled"`, `"exact_next_action": "none"`} {
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
