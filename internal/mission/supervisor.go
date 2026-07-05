package mission

import (
	"fmt"
	"strings"
)

const (
	GoalLeaseSchema           = "ao.mission.goal-lease.v0.3"
	MissionCheckpointSchema   = "ao.mission.checkpoint.v0.3"
	ReturnGateSchema          = "ao.mission.return-gate.v0.3"
	RouteReconciliationSchema = "ao.mission.route-reconciliation.v0.3"
	CheckpointBundleSchema    = "ao.mission.checkpoint-resume-bundle.v0.3"
	defaultMinNodes           = 10
	defaultMinMinutes         = 120
	defaultMaxMinutes         = 180
	defaultReturnOnlyWhen     = "mission_done_or_true_hard_blocker_or_no_ready_work_and_no_exact_next_action"
	defaultCheckpointPolicy   = "after_each_node_or_timed_interval"
)

func ensureGoalLease(r *Record, opts ContinueOptions) GoalLease {
	stamp := now(nil)
	minNodes := opts.MinNodes
	if minNodes <= 0 {
		minNodes = defaultMinNodes
	}
	minMinutes := opts.MinMinutes
	if minMinutes <= 0 {
		minMinutes = defaultMinMinutes
	}
	maxMinutes := opts.MaxMinutes
	if maxMinutes <= 0 {
		maxMinutes = defaultMaxMinutes
	}
	maxIterations := opts.MaxIterations
	if maxIterations <= 0 {
		maxIterations = minNodes
	}
	returnOnlyWhen := strings.TrimSpace(opts.ReturnOnlyWhen)
	if returnOnlyWhen == "" {
		returnOnlyWhen = defaultReturnOnlyWhen
	}
	checkpointPolicy := strings.TrimSpace(opts.CheckpointPolicy)
	if checkpointPolicy == "" {
		checkpointPolicy = defaultCheckpointPolicy
	}
	if r.GoalLease == nil {
		r.GoalLease = &GoalLease{
			Schema:           GoalLeaseSchema,
			MinNodes:         minNodes,
			MinMinutes:       minMinutes,
			MaxMinutes:       maxMinutes,
			MaxIterations:    maxIterations,
			ReturnOnlyWhen:   returnOnlyWhen,
			CheckpointPolicy: checkpointPolicy,
			CreatedAtUTC:     stamp,
			UpdatedAtUTC:     stamp,
		}
		return *r.GoalLease
	}
	r.GoalLease.Schema = GoalLeaseSchema
	if r.GoalLease.MinNodes <= 0 {
		r.GoalLease.MinNodes = minNodes
	}
	if r.GoalLease.MinMinutes <= 0 {
		r.GoalLease.MinMinutes = minMinutes
	}
	if r.GoalLease.MaxMinutes <= 0 {
		r.GoalLease.MaxMinutes = maxMinutes
	}
	if r.GoalLease.MaxIterations <= 0 || opts.MaxIterations > r.GoalLease.MaxIterations {
		r.GoalLease.MaxIterations = maxIterations
	}
	if strings.TrimSpace(r.GoalLease.ReturnOnlyWhen) == "" {
		r.GoalLease.ReturnOnlyWhen = returnOnlyWhen
	}
	if strings.TrimSpace(r.GoalLease.CheckpointPolicy) == "" {
		r.GoalLease.CheckpointPolicy = checkpointPolicy
	}
	r.GoalLease.UpdatedAtUTC = stamp
	return *r.GoalLease
}

func appendMissionCheckpoint(r *Record, step ContinuationStep) MissionCheckpoint {
	checkpoint := MissionCheckpoint{
		Schema:          MissionCheckpointSchema,
		MissionID:       r.MissionID,
		Sequence:        len(r.Checkpoints) + 1,
		Iteration:       step.Iteration,
		Route:           step.Route,
		Phase:           r.CurrentPhase,
		Result:          step.Result,
		ExactNextAction: step.ExactNextAction,
		ResumeCommand:   fmt.Sprintf("ao-mission continue --mission %s --until-done --max-iterations 10", r.MissionID),
		GeneratedAtUTC:  step.GeneratedAtUTC,
	}
	r.Checkpoints = append(r.Checkpoints, checkpoint)
	return checkpoint
}

func BuildCheckpointBundle(r Record) MissionCheckpointBundle {
	var latest *MissionCheckpoint
	if n := len(r.Checkpoints); n > 0 {
		cp := r.Checkpoints[n-1]
		latest = &cp
	}
	gate := EvaluateReturnGate(r)
	return MissionCheckpointBundle{
		Schema:              CheckpointBundleSchema,
		MissionID:           r.MissionID,
		Status:              "ready",
		CheckpointCount:     len(r.Checkpoints),
		LatestCheckpoint:    latest,
		ReturnGate:          &gate,
		ResumePrompt:        fmt.Sprintf("ao-mission continue --mission %s --until-done --max-iterations 10", r.MissionID),
		SafeToExecute:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		GeneratedAtUTC:      now(nil),
	}
}

func EvaluateReturnGate(r Record) ReturnGate {
	minNodes := defaultMinNodes
	if r.GoalLease != nil && r.GoalLease.MinNodes > 0 {
		minNodes = r.GoalLease.MinNodes
	}
	completedNodes := completedEvidenceNodes(r)
	readyNodes := readyNodesRemaining(r)
	hardBlocker := hardBlockerExists(r)
	gate := ReturnGate{
		Schema:               ReturnGateSchema,
		MissionID:            r.MissionID,
		Status:               "return_allowed",
		FinalResponseAllowed: true,
		Reason:               "mission has no ready work, no unmet lease minimum, and no exact next action",
		CompletedNodes:       completedNodes,
		MinNodes:             minNodes,
		ReadyNodesRemaining:  readyNodes,
		HardBlocker:          hardBlocker,
		ExactNextAction:      r.ExactNextAction,
		Blockers:             append([]string{}, r.Blockers...),
		GeneratedAtUTC:       now(nil),
	}
	switch {
	case r.Status == "done":
		gate.Reason = "mission status is done"
	case hardBlocker:
		gate.Reason = "mission has a terminal hard blocker for operator review"
	case completedNodes < minNodes:
		gate.Status = "early_return_denied"
		gate.FinalResponseAllowed = false
		gate.Reason = fmt.Sprintf("lease minimum unmet: completed_nodes=%d min_nodes=%d", completedNodes, minNodes)
	case readyNodes > 0:
		gate.Status = "early_return_denied"
		gate.FinalResponseAllowed = false
		gate.Reason = fmt.Sprintf("ready Atlas nodes remain: %d", readyNodes)
	case strings.TrimSpace(r.ExactNextAction) != "" && r.Status != "done":
		gate.Status = "early_return_denied"
		gate.FinalResponseAllowed = false
		gate.Reason = "exact next action remains"
	}
	if !gate.FinalResponseAllowed && !strings.HasPrefix(gate.ExactNextAction, "continue") {
		gate.ExactNextAction = "continue mission: " + strings.TrimSpace(r.ExactNextAction)
	}
	if strings.TrimSpace(gate.ExactNextAction) == "" {
		gate.ExactNextAction = "read final rollup and preserve denied authority boundaries"
	}
	return gate
}

func completedEvidenceNodes(r Record) int {
	completed := len(r.Steps)
	if r.Evidence.AtlasWorkgraph != nil && r.Evidence.AtlasWorkgraph.Completed > completed {
		completed = r.Evidence.AtlasWorkgraph.Completed
	}
	if r.Evidence.AtlasRecommendation != nil && r.Evidence.AtlasRecommendation.CompletedNodes > completed {
		completed = r.Evidence.AtlasRecommendation.CompletedNodes
	}
	if r.Evidence.FoundryRollup != nil && r.Evidence.FoundryRollup.CompletedNodes > completed {
		completed = r.Evidence.FoundryRollup.CompletedNodes
	}
	return completed
}

func readyNodesRemaining(r Record) int {
	if r.Evidence.AtlasRecommendation != nil {
		return r.Evidence.AtlasRecommendation.ReadyNodes
	}
	if r.Evidence.AtlasWorkgraph == nil {
		return 0
	}
	return r.Evidence.AtlasWorkgraph.Ready
}

func hardBlockerExists(r Record) bool {
	if r.Status == "blocked" || len(r.Blockers) > 0 {
		return true
	}
	if r.Evidence.FoundryRollup == nil {
		return false
	}
	switch normalizeFoundryRollupStatus(r.Evidence.FoundryRollup.Status) {
	case "blocked", "denied":
		return true
	default:
		return false
	}
}

func BuildFeatureDepthRecommendations(r Record, min int) []FeatureDepthRecommendation {
	if min <= 0 {
		min = defaultMinNodes
	}
	seeds := []FeatureDepthRecommendation{
		{ID: "mission-continue-loop", Owner: "ao-mission", Task: "Continue the governed supervisor loop until lease minimums or terminal blockers are reached.", ExactNextAction: "run ao-mission continue --mission " + r.MissionID + " --until-done --max-iterations 10"},
		{ID: "atlas-workgraph-refresh", Owner: "ao-atlas", Task: "Refresh the Atlas workgraph from current Mission readiness and event-index evidence.", ExactNextAction: "import the current Mission record and emit the next bounded Atlas workgraph node"},
		{ID: "foundry-single-node-import", Owner: "ao-foundry", Task: "Consume exactly one safe ready Atlas node and return a run-link or terminal rollup.", ExactNextAction: "run Foundry on the first ready node only and record run-link evidence"},
		{ID: "checkpoint-resume-bundle", Owner: "ao-mission", Task: "Persist a checkpoint/resume bundle after each node or timed interval.", ExactNextAction: "inspect the latest checkpoint bundle before any final response"},
		{ID: "route-reconciliation", Owner: "ao-mission", Task: "Reconcile current route, Atlas readiness, Foundry rollup, Promoter verdict, and Command readback.", ExactNextAction: "run mission events search for route, rollup, promoter, command, CI, and blocker evidence"},
		{ID: "command-compact-readback", Owner: "ao-command", Task: "Emit a compact long-run mission timeline readback for operator scanning.", ExactNextAction: "bind Mission dashboard and event-index output into AO Command readback"},
		{ID: "promoter-summary", Owner: "ao-promoter", Task: "Record promotion or no-promotion readiness for the exact terminal rollup status.", ExactNextAction: "keep broad RSI denied unless governed evidence separately proves it"},
		{ID: "sentinel-wording-scan", Owner: "ao-sentinel", Task: "Scan changed public docs and readbacks for unsafe or stale public-safety wording.", ExactNextAction: "record Sentinel/public-safety wording evidence before rollup closure"},
		{ID: "operator-runbook", Owner: "ao-mission", Task: "Keep the operator runbook explicit about Mission, Atlas, Blueprint, Foundry, and long-run requests.", ExactNextAction: "update docs/operator-next-actions.md with 2-3 hour Mission usage"},
		{ID: "production-readiness", Owner: "ao-mission", Task: "Run tests, vet, build, production-readiness, and public-safety wording scans over changed artifacts.", ExactNextAction: "record verification command output in the mission evidence root"},
	}
	for len(seeds) < min {
		n := len(seeds) + 1
		seeds = append(seeds, FeatureDepthRecommendation{
			ID:              fmt.Sprintf("feature-depth-wave-%02d", n),
			Owner:           "ao-atlas",
			Task:            fmt.Sprintf("Generate bounded continuation node %02d from current readiness, blockers, and route evidence.", n),
			ExactNextAction: "materialize the next Atlas-owned bounded node only after the active mutation node closes",
		})
	}
	return seeds
}

func BuildRouteReconciliation(r Record) RouteReconciliation {
	latestRoute := r.CurrentRoute
	if n := len(r.RouteHistory); n > 0 && strings.TrimSpace(r.RouteHistory[n-1].Route) != "" {
		latestRoute = r.RouteHistory[n-1].Route
	}
	foundryStatus := ""
	if r.Evidence.FoundryRollup != nil {
		foundryStatus = normalizeFoundryRollupStatus(r.Evidence.FoundryRollup.Status)
	}
	status := "ready"
	next := "route and readback evidence reconciled; continue from latest exact next action"
	if latestRoute != r.CurrentRoute {
		status = "stale_route_detected"
		next = "refresh route decision before final response"
	}
	if r.Status == "done" && r.CurrentRoute != "complete" {
		status = "stale_route_detected"
		next = "reconcile terminal mission route with final rollup"
	}
	commandBound := artifactKindBound(r, "command-readback") || artifactKindBound(r, "command-status")
	promoterBound := artifactKindBound(r, "promoter-readback") || artifactKindBound(r, "promoter-verdict")
	return RouteReconciliation{
		Schema:                RouteReconciliationSchema,
		MissionID:             r.MissionID,
		Status:                status,
		CurrentRoute:          r.CurrentRoute,
		LatestRoute:           latestRoute,
		FoundryTerminalStatus: foundryStatus,
		AtlasReadyNodes:       readyNodesRemaining(r),
		CommandReadbackBound:  commandBound,
		PromoterReadbackBound: promoterBound,
		ExactNextAction:       next,
		GeneratedAtUTC:        now(nil),
	}
}

func artifactKindBound(r Record, kind string) bool {
	for _, ref := range r.ArtifactRefs {
		if strings.EqualFold(ref.Kind, kind) {
			return true
		}
	}
	return false
}

func normalizeFoundryRollupStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "complete", "completed", "done":
		return "completed"
	case "promote", "promoted", "promotion_ready":
		return "promoted"
	case "deny", "denied":
		return "denied"
	case "block", "blocked":
		return "blocked"
	default:
		return strings.ToLower(strings.TrimSpace(status))
	}
}

func foundryRollupClosesMission(rollup FoundryRollupCounts) bool {
	switch normalizeFoundryRollupStatus(rollup.Status) {
	case "completed", "promoted":
		return rollup.TotalNodes > 0 && rollup.CompletedNodes == rollup.TotalNodes
	default:
		return false
	}
}
