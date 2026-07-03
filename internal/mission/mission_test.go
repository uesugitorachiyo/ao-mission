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
