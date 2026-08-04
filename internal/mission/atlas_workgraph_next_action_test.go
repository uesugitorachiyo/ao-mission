package mission

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtlasWorkgraphImportBindsFirstReadyNodeID(t *testing.T) {
	store := NewStore(t.TempDir())
	record, err := store.Start("bind the first ready Atlas node")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Root, "workgraph.json")
	body := `{
  "contract_version": "ao.atlas.workgraph.v0.1",
  "nodes": [
    {"id":"completed-node","status":"completed"},
    {"id":"month2-gap-closure-reconciliation","status":"ready"},
    {"node_id":"later-ready-node","status":"ready"}
  ]
}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	readback, err := ImportArtifact(store, record.MissionID, "atlas-workgraph", path)
	if err != nil {
		t.Fatal(err)
	}
	if readback.ExactNextAction != "month2-gap-closure-reconciliation" {
		t.Fatalf("import readback next action = %q", readback.ExactNextAction)
	}
	updated, err := store.Load(record.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ExactNextAction != "month2-gap-closure-reconciliation" {
		t.Fatalf("durable next action = %q", updated.ExactNextAction)
	}
	last := updated.RouteHistory[len(updated.RouteHistory)-1]
	if last.ExactNextAction != updated.ExactNextAction {
		t.Fatalf("route history next action = %q, durable = %q", last.ExactNextAction, updated.ExactNextAction)
	}
}

func TestAtlasWorkgraphImportRejectsArtifactPathDigestDrift(t *testing.T) {
	store := NewStore(t.TempDir())
	record, err := store.Start("reject Atlas workgraph artifact-path drift")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Root, "workgraph.json")
	first := `{"contract_version":"ao.atlas.workgraph.v0.1","nodes":[{"id":"first-ready","status":"ready"}]}`
	if err := os.WriteFile(path, []byte(first), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(store, record.MissionID, "atlas-workgraph", path); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(store, record.MissionID, "atlas-workgraph", path); err != nil {
		t.Fatalf("exact reimport error = %v", err)
	}
	before, err := store.Load(record.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.ArtifactRefs) != 1 {
		t.Fatalf("artifact refs after exact reimport = %d, want 1", len(before.ArtifactRefs))
	}

	second := `{"contract_version":"ao.atlas.workgraph.v0.1","nodes":[{"id":"second-ready","status":"ready"}]}`
	if err := os.WriteFile(path, []byte(second), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportArtifact(store, record.MissionID, "atlas-workgraph", path); err == nil || !strings.Contains(err.Error(), "artifact path already bound to a different digest") {
		t.Fatalf("drift import error = %v", err)
	}
	after, err := store.Load(record.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.ArtifactRefs) != 1 || after.ArtifactRefs[0] != before.ArtifactRefs[0] {
		t.Fatalf("drift import mutated artifact refs: before=%+v after=%+v", before.ArtifactRefs, after.ArtifactRefs)
	}
	if after.ExactNextAction != "first-ready" {
		t.Fatalf("drift import changed next action to %q", after.ExactNextAction)
	}
}

func TestAtlasWorkgraphImportSkipsReadyNodesWithIncompleteDependencies(t *testing.T) {
	store := NewStore(t.TempDir())
	record, err := store.Start("bind only a dependency-ready Atlas node")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Root, "workgraph.json")
	body := `{
  "contract_version": "ao.atlas.workgraph.v0.1",
  "nodes": [
    {"id":"blocked-prerequisite","status":"blocked"},
    {"id":"must-not-run","status":"ready","dependencies":["blocked-prerequisite"]},
    {"id":"independent-ready","status":"ready","dependencies":[]}
  ]
}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	readback, err := ImportArtifact(store, record.MissionID, "atlas-workgraph", path)
	if err != nil {
		t.Fatal(err)
	}
	if readback.ExactNextAction != "independent-ready" {
		t.Fatalf("import readback next action = %q, want independent-ready", readback.ExactNextAction)
	}
}

func TestAtlasWorkgraphImportRejectsReadyNodesWithoutAnExecutableDependencyPath(t *testing.T) {
	store := NewStore(t.TempDir())
	record, err := store.Start("reject a non-executable Atlas workgraph")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Root, "workgraph.json")
	body := `{
  "contract_version": "ao.atlas.workgraph.v0.1",
  "nodes": [
    {"id":"blocked-prerequisite","status":"blocked"},
    {"id":"must-not-run","status":"ready","dependencies":["blocked-prerequisite"]}
  ]
}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err = ImportArtifact(store, record.MissionID, "atlas-workgraph", path)
	if err == nil || !strings.Contains(err.Error(), "no dependency-ready node") {
		t.Fatalf("error = %v, want dependency-ready rejection", err)
	}
	updated, loadErr := store.Load(record.MissionID)
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if len(updated.ArtifactRefs) != 0 || updated.Evidence.AtlasWorkgraph != nil {
		t.Fatalf("rejected workgraph mutated durable state: %+v", updated)
	}
}

func TestAtlasWorkgraphImportBindsLegacyNodeIDAlias(t *testing.T) {
	store := NewStore(t.TempDir())
	record, err := store.Start("bind a legacy Atlas node_id")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Root, "workgraph.json")
	if err := os.WriteFile(path, []byte(`{"schema":"ao.atlas.workgraph.v0.1","nodes":[{"node_id":"legacy-ready-node","status":"ready"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	readback, err := ImportArtifact(store, record.MissionID, "atlas-workgraph", path)
	if err != nil {
		t.Fatal(err)
	}
	if readback.ExactNextAction != "legacy-ready-node" {
		t.Fatalf("legacy node_id next action = %q", readback.ExactNextAction)
	}
}

func TestAtlasWorkgraphImportRejectsAmbiguousOrUnsafeReadyNodeIDs(t *testing.T) {
	for _, tc := range []struct {
		name string
		node string
		want string
	}{
		{name: "conflicting aliases", node: `{"id":"node-one","node_id":"node-two","status":"ready"}`, want: "conflicting id and node_id"},
		{name: "unsafe identifier", node: `{"id":"node one","status":"ready"}`, want: "bounded ASCII identifier"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := NewStore(t.TempDir())
			record, err := store.Start("reject an unsafe Atlas node identity")
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(store.Root, "workgraph.json")
			if err := os.WriteFile(path, []byte(`{"contract_version":"ao.atlas.workgraph.v0.1","nodes":[`+tc.node+`]}`), 0o600); err != nil {
				t.Fatal(err)
			}

			_, err = ImportArtifact(store, record.MissionID, "atlas-workgraph", path)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
			unchanged, loadErr := store.Load(record.MissionID)
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			if len(unchanged.ArtifactRefs) != 0 || unchanged.Evidence.AtlasWorkgraph != nil {
				t.Fatalf("rejected import mutated durable state: %+v", unchanged)
			}
		})
	}
}

func TestAtlasWorkgraphImportRetainsAnonymousLegacyFallback(t *testing.T) {
	store := NewStore(t.TempDir())
	record, err := store.Start("retain anonymous workgraph compatibility")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Root, "workgraph.json")
	if err := os.WriteFile(path, []byte(`{"schema":"ao.atlas.workgraph.v0.1","nodes":[{"status":"ready"},{"id":"later-node","status":"ready"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	readback, err := ImportArtifact(store, record.MissionID, "atlas-workgraph", path)
	if err != nil {
		t.Fatal(err)
	}
	if readback.ExactNextAction != "send first safe Atlas node to AO Foundry" {
		t.Fatalf("legacy fallback = %q", readback.ExactNextAction)
	}
}
