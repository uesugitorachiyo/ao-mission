package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

type ImportReadback struct {
	Schema          string      `json:"schema"`
	MissionID       string      `json:"mission_id"`
	Kind            string      `json:"kind"`
	Status          string      `json:"status"`
	Artifact        ArtifactRef `json:"artifact"`
	ExactNextAction string      `json:"exact_next_action"`
	SafeToExecute   bool        `json:"safe_to_execute"`
	ExecutesWork    bool        `json:"executes_work"`
	ApprovesWork    bool        `json:"approves_work"`
	GeneratedAtUTC  string      `json:"generated_at_utc"`
}

func ImportArtifact(s Store, missionID, kind, path string) (ImportReadback, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return ImportReadback{}, err
	}
	if err := ValidatePublicSafeText(string(body)); err != nil {
		return ImportReadback{}, err
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return ImportReadback{}, err
	}
	if kind == "scheduler-readback" && boolFromAny(doc["executes_work"]) {
		return ImportReadback{}, fmt.Errorf("scheduler-readback executes_work must be false")
	}
	ref := ArtifactRef{Schema: ArtifactRefSchema, Ref: path, Digest: digestBytes(body), Kind: kind}
	r, err := s.Update(missionID, func(rec *Record) error {
		rec.ArtifactRefs = append(rec.ArtifactRefs, ref)
		switch kind {
		case "blueprint-authorization":
			rec.CurrentRoute = "ao-atlas"
			rec.CurrentPhase = "blueprint_authorized"
			rec.ExactNextAction = "send authorized Blueprint pack to AO Atlas"
		case "atlas-workgraph":
			counts := countWorkgraphNodes(doc)
			rec.Evidence.AtlasWorkgraph = &counts
			rec.CurrentRoute = "ao-foundry"
			rec.CurrentPhase = "atlas_workgraph_ready"
			rec.ExactNextAction = "send first safe Atlas node to AO Foundry"
		case "foundry-run-link":
			rec.CurrentPhase = "foundry_run_link_recorded"
			rec.ExactNextAction = "read next Atlas dependency-unblocked node or final rollup"
		case "foundry-final-rollup":
			rollup := parseFoundryRollupCounts(doc)
			rec.Evidence.FoundryRollup = &rollup
			if rollup.Status == "completed" && rollup.TotalNodes > 0 && rollup.CompletedNodes == rollup.TotalNodes {
				rec.Status = "done"
				rec.CurrentRoute = "complete"
				rec.CurrentPhase = "complete"
				rec.ExactNextAction = "mission complete; read final rollup and recommended next tasks"
			} else {
				rec.CurrentPhase = "foundry_final_rollup_recorded"
				rec.ExactNextAction = "review final rollup blockers before continuing"
			}
		case "scheduler-readback":
			rec.Evidence.SchedulerReadback = &SchedulerEvidenceCounts{
				Status:       stringFromAny(doc["status"]),
				Scheduler:    stringFromAny(doc["scheduler"]),
				EventLoop:    boolFromAny(doc["event_loop"]),
				ExecutesWork: false,
			}
			rec.CurrentPhase = "scheduler_readback_recorded"
			rec.ExactNextAction = "scheduler wakeup readback recorded; continue mission through AO Mission event loop"
		default:
			return fmt.Errorf("unsupported import kind %q", kind)
		}
		return nil
	})
	if err != nil {
		return ImportReadback{}, err
	}
	return ImportReadback{
		Schema:          "ao.mission.import-readback.v0.1",
		MissionID:       r.MissionID,
		Kind:            kind,
		Status:          "recorded",
		Artifact:        ref,
		ExactNextAction: r.ExactNextAction,
		SafeToExecute:   false,
		ExecutesWork:    false,
		ApprovesWork:    false,
		GeneratedAtUTC:  now(nil),
	}, nil
}

func digestBytes(body []byte) string {
	sum := sha256.Sum256(body)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func countWorkgraphNodes(doc map[string]any) NodeCounts {
	var counts NodeCounts
	nodes, _ := doc["nodes"].([]any)
	for _, node := range nodes {
		counts.Total++
		obj, _ := node.(map[string]any)
		status, _ := obj["status"].(string)
		switch status {
		case "ready":
			counts.Ready++
		case "blocked":
			counts.Blocked++
		case "completed", "complete", "done":
			counts.Completed++
		case "failed", "fail":
			counts.Failed++
		}
	}
	return counts
}

func parseFoundryRollupCounts(doc map[string]any) FoundryRollupCounts {
	status, _ := doc["status"].(string)
	return FoundryRollupCounts{
		Status:         status,
		CompletedNodes: intFromAny(doc["completed_nodes"]),
		TotalNodes:     intFromAny(doc["total_nodes"]),
	}
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func stringFromAny(v any) string {
	s, _ := v.(string)
	return s
}

func boolFromAny(v any) bool {
	b, _ := v.(bool)
	return b
}
