package mission

import (
	"fmt"
	"strings"
)

const (
	GoalLeaseSchema              = "ao.mission.goal-lease.v0.3"
	MissionCheckpointSchema      = "ao.mission.checkpoint.v0.3"
	ReturnGateSchema             = "ao.mission.return-gate.v0.3"
	RouteReconciliationSchema    = "ao.mission.route-reconciliation.v0.3"
	CheckpointBundleSchema       = "ao.mission.checkpoint-resume-bundle.v0.3"
	defaultMinNodes              = 10
	defaultMinMinutes            = 120
	defaultMaxMinutes            = 180
	defaultReturnOnlyWhen        = "mission_done_or_true_hard_blocker_or_no_ready_work_and_no_exact_next_action"
	defaultCheckpointPolicy      = "after_each_node_or_timed_interval"
	minFeatureDepthTaskMinutes   = 6
	minFeatureDepthEvidenceItems = 3
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
		CorrelationID:   r.CorrelationID,
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
		CorrelationID:       r.CorrelationID,
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
	case r.Evidence.AtlasRecommendation != nil && !r.Evidence.AtlasRecommendation.FinalResponseAllowed:
		gate.Status = "early_return_denied"
		gate.FinalResponseAllowed = false
		gate.Reason = fmt.Sprintf("Atlas recommendation readback return gate blocked: %s lease_time_status=%s elapsed_minutes=%d", r.Evidence.AtlasRecommendation.ReturnGateStatus, r.Evidence.AtlasRecommendation.LeaseTimeStatus, r.Evidence.AtlasRecommendation.ElapsedMinutes)
		if strings.TrimSpace(r.Evidence.AtlasRecommendation.ExactNextAction) != "" {
			gate.ExactNextAction = r.Evidence.AtlasRecommendation.ExactNextAction
		}
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
	completed := 0
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
	missionID := strings.TrimSpace(r.MissionID)
	if missionID == "" {
		missionID = "<mission-id>"
	}
	continueCommand := fmt.Sprintf("ao-mission continue --mission %s --until-done --max-iterations 10", missionID)
	eventIndexCommand := fmt.Sprintf("ao-mission mission events index --out tmp/%s-event-index.json", missionID)
	seeds := []FeatureDepthRecommendation{
		featureDepthRecommendation("mission-continue-loop", "ao-mission", "Continue the governed supervisor loop until lease minimums, ready work, or terminal blockers are resolved.", "run "+continueCommand, "Final response remains denied while lease minimums, ready nodes, or exact next action remain.", continueCommand, []string{"goal lease", "return gate", "checkpoint bundle"}),
		featureDepthRecommendation("atlas-workgraph-refresh", "ao-atlas", "Refresh the Atlas workgraph from current Mission readiness, event-index digest, and final rollup evidence.", "import the current Mission record and emit the next bounded Atlas workgraph node", "Atlas workgraph refresh records total, completed, ready, and blocked node counts.", "atlas workgraph import --mission "+missionID, []string{"mission record", "event index digest", "final rollup digest"}),
		featureDepthRecommendation("foundry-single-node-import", "ao-foundry", "Consume exactly one safe ready Atlas node and return a run-link or terminal rollup.", "run Foundry on the first ready node only and record run-link evidence", "Exactly one executable mutation node is active and it produces run-link or terminal evidence.", "foundry run --mission "+missionID+" --one-node", []string{"Foundry import", "run-link", "rollback record"}),
		featureDepthRecommendation("checkpoint-resume-bundle", "ao-mission", "Persist and verify a checkpoint/resume bundle after each completed node or timed interval.", "inspect the latest checkpoint bundle before any final response", "Checkpoint bundle includes latest route, phase, exact next action, and return gate.", "ao-mission checkpoint inspect --mission "+missionID, []string{"checkpoint bundle", "resume command", "return gate"}),
		featureDepthRecommendation("route-reconciliation", "ao-mission", "Reconcile current route, Atlas readiness, Foundry rollup, Promoter verdict, and Command readback.", "run mission events search for route, rollup, promoter, command, CI, and blocker evidence", "Route reconciliation binds the newest route and blocks stale final status.", eventIndexCommand, []string{"route history", "Foundry rollup", "Command readback"}),
		featureDepthRecommendation("event-index-search-binding", "ao-mission", "Index and search route, node, PR, CI, rollup, and blocker evidence for the active mission.", "record event search readbacks for every terminal node before closure", "Event search returns mission-bound evidence for node, PR, CI, rollup, and blocker kinds.", eventIndexCommand, []string{"event index", "search readback", "artifact refs"}),
		featureDepthRecommendation("command-compact-readback", "ao-command", "Emit a compact long-run mission timeline readback for operator scanning.", "bind Mission dashboard and event-index output into AO Command readback", "Command readback summarizes status, ready work, blockers, and exact next action.", "ao command mission status --mission "+missionID+" --compact", []string{"Command status", "mission timeline", "exact next action"}),
		featureDepthRecommendation("promoter-summary", "ao-promoter", "Record promotion or no-promotion readiness for the exact terminal rollup status.", "keep broad RSI denied unless governed evidence separately proves it", "Promoter summary records no-promotion or exact missing evidence without broad RSI claims.", "ao-promoter summarize --mission "+missionID, []string{"Promoter verdict", "denied surfaces", "RSI denied"}),
		featureDepthRecommendation("sentinel-wording-scan", "ao-sentinel", "Scan changed public docs and readbacks for unsafe or stale public-safety wording.", "record Sentinel/public-safety wording evidence before rollup closure", "Sentinel scan passes without broad capability or RSI wording.", "ao-sentinel scan docs/evidence docs/operator-next-actions.md", []string{"public docs", "readbacks", "wording scan"}),
		featureDepthRecommendation("operator-runbook", "ao-mission", "Keep the operator runbook explicit about Mission, Atlas, Blueprint, Foundry, and long-run requests.", "update docs/operator-next-actions.md with 2-3 hour Mission usage", "Runbook explains when to use Mission, Atlas, Blueprint, Foundry, and how to request long runs.", "ao-mission docs verify --operator-runbook", []string{"operator docs", "route guidance", "long-run prompt"}),
		featureDepthRecommendation("production-readiness", "ao-mission", "Run tests, vet, build, production-readiness, and public-safety wording scans over changed artifacts.", "record verification command output in the mission evidence root", "Verification evidence includes local tests, build, readiness, and wording scan status.", "scripts/production-readiness.sh", []string{"test output", "build output", "readiness output"}),
		featureDepthRecommendation("final-response-gate", "ao-mission", "Prove final response is denied while ready nodes or exact next actions remain.", "add or run a final-response denial regression before node closure", "Regression fails if ready nodes can still allow final response.", "go test ./internal/mission -run TestFinalRollupDeniesFinalResponseWhenReadyNodesRemain -count=1", []string{"return gate", "ready node fixture", "regression output"}),
		featureDepthRecommendation("duration-ledger-review", "ao-mission", "Compare the latest session duration metadata with the lease target without reading session contents.", "record duration-ledger readback and exact continuation reason when target time is unmet", "Duration ledger reports elapsed minutes and never inspects prompt or credential content.", "ao-mission duration readback --mission "+missionID, []string{"duration ledger", "metadata-only readback", "lease target"}),
		featureDepthRecommendation("atlas-prompt-packet", "ao-mission", "Generate an Atlas continuation prompt packet from current readiness, final rollup, and event-index evidence.", "refresh Atlas prompt packet and deny final response if ready work remains", "Prompt packet includes event-index digest, final-rollup digest, and ready-node denial wording.", "ao-mission final atlas-prompt --mission "+missionID+" --event-index tmp/event-index.json --out tmp/atlas-prompt.json", []string{"Atlas prompt packet", "event-index digest", "final-rollup digest"}),
		featureDepthRecommendation("stale-rollup-normalization", "ao-mission", "Normalize completed, promoted, denied, and blocked terminal rollup statuses before route decisions.", "import a terminal Foundry rollup fixture and verify exact Mission route/readback state", "Terminal status normalization records done or blocker status without preserving stale denial blindly.", "go test ./internal/mission -run TestImportFoundryFinalRollup -count=1", []string{"Foundry rollup", "normalized status", "route readback"}),
		featureDepthRecommendation("branch-lifecycle-evidence", "ao-mission", "Record local and remote codex branch cleanup evidence after each merged node.", "fetch with prune and record branch cleanup in node evidence", "Branch cleanup evidence shows no current node branch remains after merge.", "git fetch --prune && git branch --list 'codex/*'", []string{"local branches", "remote branches", "merge PR"}),
		featureDepthRecommendation("ci-run-link-evidence", "ao-mission", "Bind PR and CI run links to the node evidence root before completing the node.", "record PR number, CI job results, and merge head for the active node", "CI evidence contains passing jobs and merge head for the completed node.", "gh pr checks --watch --interval 10", []string{"PR link", "CI checks", "merge head"}),
		featureDepthRecommendation("rollback-record-audit", "ao-mission", "Audit every active node for a rollback record before implementation evidence is accepted.", "reject node closure if rollback evidence is missing", "Each node evidence root contains rollback record and verification output.", "find docs/evidence -name rollback-record.json -print", []string{"rollback record", "node gate", "verification"}),
		featureDepthRecommendation("blueprint-routing-audit", "ao-atlas", "Route directly through Atlas or Foundry unless a node genuinely needs new requirements or authorization from Blueprint.", "record route decision explaining why Blueprint was not used for implementation nodes", "Route decision denies Blueprint for ordinary bounded implementation work.", "atlas route decide --mission "+missionID, []string{"route decision", "Blueprint denial", "Foundry import"}),
		featureDepthRecommendation("final-synthesis-readback", "ao-mission", "Regenerate final synthesis readback and verify it blocks final status until minimum nodes and ready-work gates clear.", "run final synthesis and compare completed, ready, blocked, and final-response fields", "Final synthesis carries exact next action and at least the minimum Feature Depth recommendations.", "ao-mission final synthesize --mission "+missionID+" --evidence-root docs/evidence/ao-mission-doubled-wave-v01", []string{"final synthesis", "workgraph", "Feature Depth recommendations"}),
	}
	for len(seeds) < min {
		n := len(seeds) + 1
		seeds = append(seeds, featureDepthRecommendation(
			fmt.Sprintf("feature-depth-continuation-%02d", n),
			"ao-atlas",
			fmt.Sprintf("Create and execute continuation node %02d with gate, rollback, implementation evidence, verification, PR/CI readback, and checkpoint bundle.", n),
			"materialize the next Atlas-owned bounded node only after the active mutation node closes",
			fmt.Sprintf("Continuation node %02d cannot close until evidence, verification, and branch cleanup are recorded.", n),
			continueCommand,
			[]string{"node gate", "implementation evidence", "verification output"},
		))
	}
	return seeds
}

func featureDepthRecommendation(id, owner, task, exactNextAction, gate, command string, evidence []string) FeatureDepthRecommendation {
	return FeatureDepthRecommendation{
		ID:                  id,
		Owner:               owner,
		Task:                task,
		Gate:                gate,
		EvidenceRequired:    append([]string(nil), evidence...),
		EstimatedMinutes:    minFeatureDepthTaskMinutes,
		ContinuationCommand: command,
		ExactNextAction:     exactNextAction,
		StopCondition:       defaultReturnOnlyWhen,
	}
}

func ValidateFeatureDepthRecommendations(items []FeatureDepthRecommendation, min int) error {
	if min <= 0 {
		min = defaultMinNodes
	}
	if len(items) < min {
		return fmt.Errorf("feature depth recommendations too shallow: got %d want at least %d", len(items), min)
	}
	seen := map[string]bool{}
	totalMinutes := 0
	for i, item := range items {
		label := item.ID
		if strings.TrimSpace(label) == "" {
			label = fmt.Sprintf("index %d", i)
		}
		if seen[item.ID] {
			return fmt.Errorf("duplicate feature depth recommendation id %q", item.ID)
		}
		seen[item.ID] = true
		if strings.TrimSpace(item.ID) == "" ||
			strings.TrimSpace(item.Owner) == "" ||
			strings.TrimSpace(item.Task) == "" ||
			strings.TrimSpace(item.Gate) == "" ||
			strings.TrimSpace(item.ContinuationCommand) == "" ||
			strings.TrimSpace(item.ExactNextAction) == "" ||
			strings.TrimSpace(item.StopCondition) == "" {
			return fmt.Errorf("feature depth recommendation %s is missing concrete fields", label)
		}
		if item.EstimatedMinutes < minFeatureDepthTaskMinutes {
			return fmt.Errorf("feature depth recommendation %s estimated_minutes=%d below minimum %d", label, item.EstimatedMinutes, minFeatureDepthTaskMinutes)
		}
		if len(item.EvidenceRequired) < minFeatureDepthEvidenceItems {
			return fmt.Errorf("feature depth recommendation %s evidence_required too shallow: got %d want at least %d", label, len(item.EvidenceRequired), minFeatureDepthEvidenceItems)
		}
		for _, evidence := range item.EvidenceRequired {
			if strings.TrimSpace(evidence) == "" {
				return fmt.Errorf("feature depth recommendation %s has empty evidence requirement", label)
			}
		}
		totalMinutes += item.EstimatedMinutes
	}
	requiredMinutes := min * minFeatureDepthTaskMinutes
	if totalMinutes < requiredMinutes {
		return fmt.Errorf("feature depth recommendations under budget: got %d minutes want at least %d", totalMinutes, requiredMinutes)
	}
	return nil
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
	commandBound := artifactKindBound(r, "command-readback") || artifactKindBound(r, "command-status") || atlasFinalSynthesisCommandBound(r)
	promoterBound := artifactKindBound(r, "promoter-readback") || artifactKindBound(r, "promoter-verdict") || atlasFinalSynthesisPromoterBound(r)
	return RouteReconciliation{
		Schema:                RouteReconciliationSchema,
		MissionID:             r.MissionID,
		CorrelationID:         r.CorrelationID,
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

func atlasFinalSynthesisCommandBound(r Record) bool {
	return r.Evidence.AtlasFinalSynthesis != nil && r.Evidence.AtlasFinalSynthesis.CommandReadback == "ready"
}

func atlasFinalSynthesisPromoterBound(r Record) bool {
	return r.Evidence.AtlasFinalSynthesis != nil && strings.TrimSpace(r.Evidence.AtlasFinalSynthesis.PromoterStatus) != ""
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
