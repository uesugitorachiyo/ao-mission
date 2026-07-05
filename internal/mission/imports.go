package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
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
	if isMissionEvidenceReadback(kind) {
		for _, field := range []string{"safe_to_execute", "schedules_work", "executes_work", "approves_work", "mutates_repositories", "provider_calls", "release_or_publish", "credential_use", "direct_main_mutation", "concurrent_mutation", "claims_authority_advance"} {
			if boolFromAny(doc[field]) {
				return ImportReadback{}, fmt.Errorf("%s %s must be false", kind, field)
			}
		}
	}
	ref := ArtifactRef{Schema: ArtifactRefSchema, Ref: path, Digest: digestBytes(body), Kind: kind}
	r, err := s.Update(missionID, func(rec *Record) error {
		rec.ArtifactRefs = append(rec.ArtifactRefs, ref)
		switch kind {
		case "blueprint-authorization":
			rec.CurrentRoute = "ao-atlas"
			rec.CurrentPhase = "blueprint_authorized"
			rec.ExactNextAction = "send authorized Blueprint pack to AO Atlas"
			AppendRouteHistory(rec, routeFromRecord(*rec, "Blueprint authorization imported"))
		case "atlas-workgraph":
			counts := countWorkgraphNodes(doc)
			rec.Evidence.AtlasWorkgraph = &counts
			rec.CurrentRoute = "ao-foundry"
			rec.CurrentPhase = "atlas_workgraph_ready"
			rec.ExactNextAction = "send first safe Atlas node to AO Foundry"
			AppendRouteHistory(rec, routeFromRecord(*rec, "Atlas workgraph imported"))
			gate := EvaluateReturnGate(*rec)
			rec.ReturnGate = &gate
			reconciliation := BuildRouteReconciliation(*rec)
			rec.Reconciliation = &reconciliation
		case "atlas-recommendation-readback":
			readback := parseAtlasRecommendationReadbackCounts(doc)
			rec.Evidence.AtlasRecommendation = &readback
			rec.Evidence.AtlasWorkgraph = &NodeCounts{
				Total:     readback.TotalNodes,
				Ready:     readback.ReadyNodes,
				Completed: readback.CompletedNodes,
			}
			rec.ExactNextAction = readback.ExactNextAction
			switch {
			case atlasRecommendationReadbackClosesMission(readback):
				rec.Status = "done"
				rec.CurrentRoute = "complete"
				rec.CurrentPhase = "complete"
				rec.ExactNextAction = "mission complete; read final rollup and recommended next tasks"
			case atlasRecommendationReadbackTerminalBlocker(readback):
				rec.Status = "blocked"
				rec.CurrentRoute = "ao-atlas"
				rec.CurrentPhase = "atlas_recommendation_" + readback.Status
				blocker := atlasRecommendationBlocker(readback)
				rec.Blockers = appendMissingString(rec.Blockers, blocker)
				rec.ExactNextAction = "Atlas recommendation readback " + readback.Status + ": " + blocker
			default:
				rec.CurrentRoute = "ao-atlas"
				rec.CurrentPhase = "atlas_recommendation_readback_recorded"
				if rec.ExactNextAction == "" {
					rec.ExactNextAction = "continue AO Atlas recommendation wave from latest ready node"
				}
			}
			AppendRouteHistory(rec, routeFromRecord(*rec, "Atlas recommendation readback imported"))
			gate := EvaluateReturnGate(*rec)
			rec.ReturnGate = &gate
			reconciliation := BuildRouteReconciliation(*rec)
			rec.Reconciliation = &reconciliation
		case "atlas-final-synthesis-readback":
			readback := parseAtlasFinalSynthesisReadbackCounts(doc)
			if err := validateAtlasFinalSynthesisReadback(readback); err != nil {
				return err
			}
			rec.Evidence.AtlasFinalSynthesis = &readback
			rec.Evidence.AtlasRecommendation = atlasRecommendationFromFinalSynthesis(readback)
			rec.Evidence.AtlasWorkgraph = &NodeCounts{
				Total:     readback.TotalNodes,
				Ready:     readback.ReadyNodes,
				Blocked:   readback.BlockedNodes,
				Completed: readback.CompletedNodes,
			}
			rec.ExactNextAction = readback.ExactNextAction
			switch {
			case atlasFinalSynthesisClosesMission(readback):
				rec.Status = "done"
				rec.CurrentRoute = "complete"
				rec.CurrentPhase = "complete"
				rec.ExactNextAction = "mission complete; read final rollup and recommended next tasks"
			case readback.Status == "blocked" || readback.Status == "denied":
				rec.Status = "blocked"
				rec.CurrentRoute = "ao-atlas"
				rec.CurrentPhase = "atlas_final_synthesis_" + readback.Status
				blocker := atlasFinalSynthesisBlocker(readback)
				rec.Blockers = appendMissingString(rec.Blockers, blocker)
				rec.ExactNextAction = "Atlas final synthesis readback " + readback.Status + ": " + blocker
			default:
				rec.CurrentRoute = "ao-atlas"
				rec.CurrentPhase = "atlas_final_synthesis_readback_recorded"
				if rec.ExactNextAction == "" {
					rec.ExactNextAction = "continue AO Atlas final synthesis reconciliation from latest exact next action"
				}
			}
			AppendRouteHistory(rec, routeFromRecord(*rec, "Atlas final synthesis readback imported"))
			gate := EvaluateReturnGate(*rec)
			rec.ReturnGate = &gate
			reconciliation := BuildRouteReconciliation(*rec)
			rec.Reconciliation = &reconciliation
		case "foundry-run-link":
			rec.CurrentPhase = "foundry_run_link_recorded"
			rec.ExactNextAction = "read next Atlas dependency-unblocked node or final rollup"
			AppendRouteHistory(rec, routeFromRecord(*rec, "Foundry run-link imported"))
			gate := EvaluateReturnGate(*rec)
			rec.ReturnGate = &gate
			reconciliation := BuildRouteReconciliation(*rec)
			rec.Reconciliation = &reconciliation
		case "foundry-final-rollup":
			rollup := parseFoundryRollupCounts(doc)
			rec.Evidence.FoundryRollup = &rollup
			switch normalizeFoundryRollupStatus(rollup.Status) {
			case "completed", "promoted":
				if !foundryRollupClosesMission(rollup) {
					rec.CurrentPhase = "foundry_final_rollup_recorded"
					rec.ExactNextAction = "review final rollup node counts before closure"
					break
				}
				rec.Status = "done"
				rec.CurrentRoute = "complete"
				rec.CurrentPhase = "complete"
				rec.ExactNextAction = "mission complete; read final rollup and recommended next tasks"
			case "denied":
				rec.Status = "blocked"
				rec.CurrentRoute = "ao-atlas"
				rec.CurrentPhase = "foundry_rollup_denied"
				rec.ExactNextAction = "Foundry rollup denied; generate repair/repack support node through AO Atlas"
				rec.Blockers = appendMissingString(rec.Blockers, "foundry final rollup status denied")
			case "blocked":
				rec.Status = "blocked"
				rec.CurrentRoute = "ao-atlas"
				rec.CurrentPhase = "foundry_rollup_blocked"
				rec.ExactNextAction = "Foundry rollup blocked; resolve exact blocker before continuing"
				rec.Blockers = appendMissingString(rec.Blockers, "foundry final rollup status blocked")
			default:
				rec.CurrentPhase = "foundry_final_rollup_recorded"
				rec.ExactNextAction = "review final rollup blockers before continuing"
			}
			AppendRouteHistory(rec, routeFromRecord(*rec, "Foundry final rollup imported"))
			gate := EvaluateReturnGate(*rec)
			rec.ReturnGate = &gate
			reconciliation := BuildRouteReconciliation(*rec)
			rec.Reconciliation = &reconciliation
		case "scheduler-readback":
			rec.Evidence.SchedulerReadback = &SchedulerEvidenceCounts{
				Status:          stringFromAny(doc["status"]),
				Scheduler:       stringFromAny(doc["scheduler"]),
				EventLoop:       boolFromAny(doc["event_loop"]),
				FreshnessStatus: classifyFreshness(stringFromAny(doc["generated_at_utc"])),
				ExecutesWork:    false,
			}
			rec.CurrentPhase = "scheduler_readback_recorded"
			rec.ExactNextAction = "scheduler wakeup readback recorded; continue mission through AO Mission event loop"
			AppendRouteHistory(rec, routeFromRecord(*rec, "Scheduler readback imported"))
		case "scheduler-recovery-readback":
			rec.Evidence.SchedulerRecovery = &SchedulerRecoveryCounts{
				Status:        stringFromAny(doc["status"]),
				RecoveryMode:  stringFromAny(doc["recovery_mode"]),
				MissedWakeups: intFromAny(doc["missed_wakeups"]),
				ExecutesWork:  false,
			}
			rec.CurrentPhase = "scheduler_recovery_recorded"
			rec.ExactNextAction = stringFromAny(doc["exact_next_action"])
			if rec.ExactNextAction == "" {
				rec.ExactNextAction = "scheduler recovery readback recorded; continue mission through AO Mission event loop"
			}
			AppendRouteHistory(rec, routeFromRecord(*rec, "Scheduler recovery readback imported"))
		case "ledger-compaction-readback":
			rec.Evidence.LedgerCompaction = &LedgerCompactionCounts{
				RouteHistoryBefore: intFromAny(doc["route_history_before"]),
				RouteHistoryAfter:  intFromAny(doc["route_history_after"]),
				StepsBefore:        intFromAny(doc["steps_before"]),
				StepsAfter:         intFromAny(doc["steps_after"]),
			}
			rec.CurrentPhase = "ledger_compaction_recorded"
			rec.ExactNextAction = "ledger compaction readback recorded; continue from retained route and step evidence"
			AppendRouteHistory(rec, routeFromRecord(*rec, "Ledger compaction readback imported"))
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

func isMissionEvidenceReadback(kind string) bool {
	switch kind {
	case "atlas-recommendation-readback", "atlas-final-synthesis-readback", "scheduler-readback", "scheduler-recovery-readback", "ledger-compaction-readback":
		return true
	default:
		return false
	}
}

func routeFromRecord(rec Record, reason string) RouteDecision {
	return RouteDecision{
		Schema:          RouteSchema,
		MissionID:       rec.MissionID,
		Route:           rec.CurrentRoute,
		Reason:          reason,
		SafeToRequest:   true,
		SafeToExecute:   false,
		SafeToPromote:   false,
		ExactNextAction: rec.ExactNextAction,
		GeneratedAtUTC:  now(nil),
	}
}

func classifyFreshness(generatedAt string) string {
	return classifyFreshnessAt(generatedAt, time.Now().UTC())
}

func classifyFreshnessAt(generatedAt string, evaluatedAt time.Time) string {
	if generatedAt == "" {
		return "unknown"
	}
	stamp, err := time.Parse(time.RFC3339, generatedAt)
	if err != nil {
		return "unknown"
	}
	age := evaluatedAt.Sub(stamp)
	if age > 24*time.Hour {
		return "stale"
	}
	return "fresh"
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
		Status:         normalizeFoundryRollupStatus(status),
		CompletedNodes: intFromAny(doc["completed_nodes"]),
		TotalNodes:     intFromAny(doc["total_nodes"]),
	}
}

func parseAtlasRecommendationReadbackCounts(doc map[string]any) AtlasRecommendationReadbackCounts {
	return AtlasRecommendationReadbackCounts{
		Status:               stringFromAny(doc["status"]),
		TotalNodes:           intFromAny(doc["total_nodes"]),
		CompletedNodes:       intFromAny(doc["completed_nodes"]),
		ReadyNodes:           intFromAny(doc["ready_nodes"]),
		CheckpointCount:      intFromAny(doc["checkpoint_count"]),
		ElapsedMinutes:       intFromAny(doc["elapsed_minutes"]),
		MinMinutesMet:        boolFromAny(doc["min_minutes_met"]),
		LeaseTimeStatus:      stringFromAny(doc["lease_time_status"]),
		ReturnGateStatus:     stringFromAny(doc["return_gate_status"]),
		FinalResponseAllowed: boolFromAny(doc["final_response_allowed"]),
		Blocker:              stringFromAny(doc["blocker"]),
		RSIRemainsDenied:     boolFromAny(doc["rsi_remains_denied"]),
		ExactNextAction:      stringFromAny(doc["exact_next_action"]),
	}
}

func parseAtlasFinalSynthesisReadbackCounts(doc map[string]any) AtlasFinalSynthesisReadbackCounts {
	return AtlasFinalSynthesisReadbackCounts{
		ContractVersion:      stringFromAny(doc["contract_version"]),
		Status:               stringFromAny(doc["status"]),
		TotalNodes:           intFromAny(doc["total_nodes"]),
		CompletedNodes:       intFromAny(doc["completed_nodes"]),
		ReadyNodes:           intFromAny(doc["ready_nodes"]),
		BlockedNodes:         intFromAny(doc["blocked_nodes"]),
		MinimumNodes:         intFromAny(doc["minimum_nodes"]),
		ReturnGateStatus:     stringFromAny(doc["return_gate_status"]),
		FinalResponseAllowed: boolFromAny(doc["final_response_allowed"]),
		FinalResponseReason:  stringFromAny(doc["final_response_reason"]),
		AtlasWorkgraphStatus: stringFromAny(doc["atlas_workgraph_status"]),
		FoundryRollup:        stringFromAny(doc["foundry_rollup"]),
		PromoterStatus:       stringFromAny(doc["promoter_status"]),
		CommandReadback:      stringFromAny(doc["command_readback"]),
		EventSearchBound:     boolFromAny(doc["event_search_bound"]),
		BranchCleanupBound:   boolFromAny(doc["branch_cleanup_bound"]),
		RSIRemainsDenied:     boolFromAny(doc["rsi_remains_denied"]),
		ExactNextAction:      stringFromAny(doc["exact_next_action"]),
	}
}

func validateAtlasFinalSynthesisReadback(readback AtlasFinalSynthesisReadbackCounts) error {
	switch {
	case readback.ContractVersion != "ao.atlas.ao-mission-final-synthesis-readback.v0.1":
		return fmt.Errorf("contract_version must be ao.atlas.ao-mission-final-synthesis-readback.v0.1")
	case readback.TotalNodes != readback.CompletedNodes+readback.ReadyNodes+readback.BlockedNodes:
		return fmt.Errorf("total_nodes must equal completed_nodes plus ready_nodes plus blocked_nodes")
	case readback.FinalResponseAllowed && readback.ReadyNodes > 0:
		return fmt.Errorf("final response cannot be allowed while ready nodes remain")
	case readback.FinalResponseAllowed && readback.BlockedNodes > 0:
		return fmt.Errorf("final response cannot be allowed while blocked nodes remain")
	case readback.FinalResponseAllowed && readback.CompletedNodes < readback.MinimumNodes:
		return fmt.Errorf("final response requires completed_nodes to meet minimum_nodes")
	case readback.FinalResponseAllowed && readback.ReturnGateStatus != "final_response_allowed":
		return fmt.Errorf("final response requires return_gate_status final_response_allowed")
	case readback.FinalResponseAllowed && readback.Status != "completed":
		return fmt.Errorf("final response requires completed status")
	case readback.FinalResponseAllowed && readback.AtlasWorkgraphStatus != "completed":
		return fmt.Errorf("final response requires completed Atlas workgraph status")
	case readback.FinalResponseAllowed && readback.CommandReadback != "ready":
		return fmt.Errorf("final response requires ready command_readback")
	case readback.FinalResponseAllowed && !readback.EventSearchBound:
		return fmt.Errorf("final response requires event search binding")
	case readback.FinalResponseAllowed && !readback.BranchCleanupBound:
		return fmt.Errorf("final response requires branch cleanup binding")
	case !readback.RSIRemainsDenied:
		return fmt.Errorf("rsi_remains_denied must be true")
	default:
		return nil
	}
}

func atlasRecommendationFromFinalSynthesis(readback AtlasFinalSynthesisReadbackCounts) *AtlasRecommendationReadbackCounts {
	leaseStatus := "minimum_minutes_unmet"
	minMinutesMet := false
	checkpoints := 0
	if readback.FinalResponseAllowed {
		leaseStatus = "minimum_minutes_met"
		minMinutesMet = true
		checkpoints = readback.TotalNodes
	}
	return &AtlasRecommendationReadbackCounts{
		Status:               readback.Status,
		TotalNodes:           readback.TotalNodes,
		CompletedNodes:       readback.CompletedNodes,
		ReadyNodes:           readback.ReadyNodes,
		CheckpointCount:      checkpoints,
		MinMinutesMet:        minMinutesMet,
		LeaseTimeStatus:      leaseStatus,
		ReturnGateStatus:     readback.ReturnGateStatus,
		FinalResponseAllowed: readback.FinalResponseAllowed,
		Blocker:              atlasFinalSynthesisBlocker(readback),
		RSIRemainsDenied:     readback.RSIRemainsDenied,
		ExactNextAction:      readback.ExactNextAction,
	}
}

func atlasRecommendationReadbackClosesMission(readback AtlasRecommendationReadbackCounts) bool {
	return readback.Status == "completed" &&
		readback.TotalNodes > 0 &&
		readback.CompletedNodes == readback.TotalNodes &&
		readback.ReadyNodes == 0 &&
		readback.CheckpointCount >= readback.TotalNodes &&
		readback.MinMinutesMet &&
		readback.LeaseTimeStatus == "minimum_minutes_met" &&
		readback.ReturnGateStatus == "final_response_allowed" &&
		readback.FinalResponseAllowed
}

func atlasFinalSynthesisClosesMission(readback AtlasFinalSynthesisReadbackCounts) bool {
	return readback.Status == "completed" &&
		readback.TotalNodes > 0 &&
		readback.CompletedNodes == readback.TotalNodes &&
		readback.ReadyNodes == 0 &&
		readback.BlockedNodes == 0 &&
		readback.CompletedNodes >= readback.MinimumNodes &&
		readback.ReturnGateStatus == "final_response_allowed" &&
		readback.FinalResponseAllowed &&
		readback.AtlasWorkgraphStatus == "completed" &&
		readback.CommandReadback == "ready" &&
		readback.PromoterStatus != "" &&
		readback.EventSearchBound &&
		readback.BranchCleanupBound &&
		readback.RSIRemainsDenied
}

func atlasFinalSynthesisBlocker(readback AtlasFinalSynthesisReadbackCounts) string {
	switch {
	case readback.FinalResponseAllowed && readback.ReadyNodes > 0:
		return "final response cannot be allowed while ready nodes remain"
	case readback.FinalResponseAllowed && readback.BlockedNodes > 0:
		return "final response cannot be allowed while blocked nodes remain"
	case readback.CompletedNodes < readback.MinimumNodes:
		return "minimum nodes unmet"
	case readback.ReturnGateStatus != "" && readback.ReturnGateStatus != "final_response_allowed":
		return "return gate status " + readback.ReturnGateStatus
	case readback.CommandReadback != "" && readback.CommandReadback != "ready":
		return "command readback status " + readback.CommandReadback
	case !readback.RSIRemainsDenied:
		return "RSI denial evidence missing"
	default:
		return ""
	}
}

func atlasRecommendationReadbackTerminalBlocker(readback AtlasRecommendationReadbackCounts) bool {
	return readback.Status == "denied" || readback.Status == "blocked"
}

func atlasRecommendationBlocker(readback AtlasRecommendationReadbackCounts) string {
	if readback.Blocker != "" {
		return readback.Blocker
	}
	if readback.ReturnGateStatus != "" {
		return "return gate status " + readback.ReturnGateStatus
	}
	return "terminal Atlas recommendation status " + readback.Status
}

func appendMissingString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
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
