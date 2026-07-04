package mission

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type MissionEventSearchFilters struct {
	MissionID string
	Kind      string
	Query     string
}

func BuildMissionEventIndex(s Store) (MissionEventIndex, error) {
	records, err := s.List()
	if err != nil {
		return MissionEventIndex{}, err
	}
	index := MissionEventIndex{
		Schema:              "ao.mission.event-index.v0.2",
		Status:              "ready",
		IndexVersion:        "v0.2",
		Root:                s.Root,
		MissionCount:        len(records),
		Events:              []MissionEvent{},
		SafeToExecute:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		GeneratedAtUTC:      now(s.Clock),
	}
	for _, record := range records {
		index.Events = append(index.Events, missionEventsForRecord(record)...)
	}
	index.TotalEvents = len(index.Events)
	sourceBody, err := json.Marshal(records)
	if err != nil {
		return MissionEventIndex{}, err
	}
	index.SourceDigest = digestBytes(sourceBody)
	index.IndexDigest, err = digestMissionEventIndex(index)
	if err != nil {
		return MissionEventIndex{}, err
	}
	return index, nil
}

func ValidateMissionEventIndexDigest(index MissionEventIndex) error {
	if index.Schema != "ao.mission.event-index.v0.2" {
		return fmt.Errorf("mission event index schema must be ao.mission.event-index.v0.2")
	}
	if index.IndexVersion != "v0.2" {
		return fmt.Errorf("mission event index version must be v0.2")
	}
	if !strings.HasPrefix(index.IndexDigest, "sha256:") {
		return fmt.Errorf("mission event index digest must start with sha256:")
	}
	expected, err := digestMissionEventIndex(index)
	if err != nil {
		return err
	}
	if index.IndexDigest != expected {
		return fmt.Errorf("mission event index digest mismatch")
	}
	return nil
}

func digestMissionEventIndex(index MissionEventIndex) (string, error) {
	copy := index
	copy.IndexDigest = ""
	body, err := json.Marshal(copy)
	if err != nil {
		return "", err
	}
	return digestBytes(body), nil
}

func SearchMissionEvents(index MissionEventIndex, filters MissionEventSearchFilters) MissionEventSearchReadback {
	query := strings.ToLower(strings.TrimSpace(filters.Query))
	missionID := strings.TrimSpace(filters.MissionID)
	kind := strings.TrimSpace(filters.Kind)
	readback := MissionEventSearchReadback{
		Schema:              "ao.mission.event-search-readback.v0.1",
		Status:              "ready",
		Query:               filters.Query,
		MissionID:           missionID,
		Kind:                kind,
		Events:              []MissionEvent{},
		SafeToExecute:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		GeneratedAtUTC:      now(nil),
	}
	for _, event := range index.Events {
		if missionID != "" && event.MissionID != missionID {
			continue
		}
		if kind != "" && event.Kind != kind {
			continue
		}
		if query != "" && !missionEventMatchesQuery(event, query) {
			continue
		}
		readback.Events = append(readback.Events, event)
	}
	readback.TotalMatches = len(readback.Events)
	return readback
}

func BuildMissionDoctorReadback(s Store) MissionDoctorReadback {
	readback := MissionDoctorReadback{
		Schema:              "ao.mission.doctor-readback.v0.1",
		Status:              "ready",
		Root:                s.Root,
		Checks:              []string{},
		Blockers:            []string{},
		SafeToExecute:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		GeneratedAtUTC:      now(s.Clock),
	}
	if err := s.Init(); err != nil {
		readback.Status = "blocked"
		readback.Blockers = append(readback.Blockers, "mission store init failed: "+err.Error())
		return readback
	}
	readback.Checks = append(readback.Checks, "mission_store_initialized")
	index, err := BuildMissionEventIndex(s)
	if err != nil {
		readback.Status = "blocked"
		readback.Blockers = append(readback.Blockers, "mission event index failed: "+err.Error())
		return readback
	}
	readback.MissionCount = index.MissionCount
	readback.EventCount = index.TotalEvents
	records, err := s.List()
	if err != nil {
		readback.Status = "blocked"
		readback.Blockers = append(readback.Blockers, "mission records reload failed: "+err.Error())
		return readback
	}
	for _, record := range records {
		if record.GoalLease != nil {
			readback.LeaseCount++
		}
		if len(record.Checkpoints) > 0 {
			readback.FreshCheckpoints++
		}
		if record.ReturnGate != nil && !record.ReturnGate.FinalResponseAllowed {
			readback.EarlyReturnRisks++
		}
		if record.Reconciliation != nil && record.Reconciliation.Status == "stale_route_detected" {
			readback.StaleRoutes++
			readback.Status = "blocked"
			readback.Blockers = append(readback.Blockers, "stale route reconciliation for "+record.MissionID)
		}
	}
	readback.Checks = append(readback.Checks,
		"mission_records_readable",
		"mission_event_index_readable",
		"authority_flags_false",
		"lease_health_checked",
		"checkpoint_freshness_checked",
		"stale_route_reconciliation_checked",
		"early_return_risk_checked",
	)
	return readback
}

func BuildMissionReadinessBundleReadback(inputs []MissionReadinessBundleInput) (MissionReadinessBundleReadback, error) {
	readback := MissionReadinessBundleReadback{
		Schema:              "ao.mission.readiness-bundle-readback.v0.1",
		Status:              "ready",
		Repos:               []MissionReadinessRepoReadback{},
		SafeToExecute:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		ExactNextAction:     "review blocked repo readiness summaries before PR lifecycle work",
		GeneratedAtUTC:      now(nil),
	}
	for _, input := range inputs {
		repo := strings.TrimSpace(input.Repo)
		path := strings.TrimSpace(input.Path)
		if repo == "" || path == "" {
			return MissionReadinessBundleReadback{}, fmt.Errorf("readiness bundle inputs require repo and path")
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return MissionReadinessBundleReadback{}, err
		}
		text := string(body)
		if err := ValidatePublicSafeText(text); err != nil {
			return MissionReadinessBundleReadback{}, err
		}
		repoStatus := "blocked"
		if strings.Contains(text, "status=ready") && (strings.Contains(text, "100/100") || strings.Contains(text, "score=100/100")) {
			repoStatus = "ready"
			readback.ReadyRepos++
		} else {
			readback.BlockedRepos++
			readback.Status = "blocked"
		}
		readback.Repos = append(readback.Repos, MissionReadinessRepoReadback{
			Repo:   repo,
			Path:   path,
			Status: repoStatus,
			Score:  readinessScoreLabel(text),
			SHA256: digestBytes(body),
		})
	}
	readback.RepoCount = len(readback.Repos)
	if readback.RepoCount == 0 {
		readback.Status = "blocked"
		readback.ExactNextAction = "provide at least one local readiness summary"
	}
	if readback.Status == "ready" {
		readback.ExactNextAction = "readiness bundle verified locally; remote PR lifecycle remains operator-controlled"
	}
	return readback, nil
}

func readinessScoreLabel(text string) string {
	if strings.Contains(text, "score=100/100") {
		return "100/100"
	}
	if strings.Contains(text, "100/100") {
		return "100/100"
	}
	return ""
}

func BuildMissionDashboardReadback(s Store, missionID string, compact bool) (MissionDashboardReadback, error) {
	record, err := s.Load(missionID)
	if err != nil {
		return MissionDashboardReadback{}, err
	}
	index, err := BuildMissionEventIndex(s)
	if err != nil {
		return MissionDashboardReadback{}, err
	}
	events := []MissionEvent{}
	for _, event := range index.Events {
		if event.MissionID == record.MissionID {
			events = append(events, event)
		}
	}
	if compact && len(events) > 5 {
		events = events[len(events)-5:]
	}
	latestRoute := record.CurrentRoute
	if n := len(record.RouteHistory); n > 0 && strings.TrimSpace(record.RouteHistory[n-1].Route) != "" {
		latestRoute = record.RouteHistory[n-1].Route
	}
	return MissionDashboardReadback{
		Schema:              "ao.mission.dashboard-readback.v0.1",
		Status:              "ready",
		MissionID:           record.MissionID,
		MissionStatus:       record.Status,
		CurrentPhase:        record.CurrentPhase,
		CurrentRoute:        record.CurrentRoute,
		LatestRoute:         latestRoute,
		EventCount:          len(events),
		EventIndexDigest:    index.IndexDigest,
		Compact:             compact,
		RecentEvents:        events,
		SafeToExecute:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		ExactNextAction:     record.ExactNextAction,
		GeneratedAtUTC:      now(s.Clock),
	}, nil
}

func missionEventsForRecord(record Record) []MissionEvent {
	events := []MissionEvent{{
		Schema:         "ao.mission.event.v0.1",
		MissionID:      record.MissionID,
		Kind:           "mission_record",
		Sequence:       0,
		Status:         record.Status,
		Route:          record.CurrentRoute,
		Phase:          record.CurrentPhase,
		Summary:        fmt.Sprintf("mission %s status %s route %s phase %s next %s", record.MissionID, record.Status, record.CurrentRoute, record.CurrentPhase, record.ExactNextAction),
		GeneratedAtUTC: record.UpdatedAtUTC,
	}}
	for i, route := range record.RouteHistory {
		events = append(events, MissionEvent{
			Schema:         "ao.mission.event.v0.1",
			MissionID:      record.MissionID,
			Kind:           "route_decision",
			Sequence:       i + 1,
			Route:          route.Route,
			Summary:        strings.TrimSpace(route.Reason + " " + route.ExactNextAction),
			GeneratedAtUTC: route.GeneratedAtUTC,
		})
	}
	for i, step := range record.Steps {
		events = append(events, MissionEvent{
			Schema:         "ao.mission.event.v0.1",
			MissionID:      record.MissionID,
			Kind:           "continuation_step",
			Sequence:       i + 1,
			Status:         step.Result,
			Route:          step.Route,
			Summary:        step.ExactNextAction,
			GeneratedAtUTC: step.GeneratedAtUTC,
		})
	}
	for i, ref := range record.ArtifactRefs {
		events = append(events, MissionEvent{
			Schema:       "ao.mission.event.v0.1",
			MissionID:    record.MissionID,
			Kind:         "artifact_ref",
			Sequence:     i + 1,
			ArtifactKind: ref.Kind,
			Summary:      strings.TrimSpace(ref.Kind + " " + ref.Ref + " " + ref.Digest),
		})
	}
	if record.GoalLease != nil {
		events = append(events, MissionEvent{
			Schema:         "ao.mission.event.v0.1",
			MissionID:      record.MissionID,
			Kind:           "goal_lease",
			Sequence:       len(events) + 1,
			Status:         "ready",
			Summary:        fmt.Sprintf("lease min_nodes=%d min_minutes=%d max_minutes=%d return_only_when=%s checkpoint_policy=%s", record.GoalLease.MinNodes, record.GoalLease.MinMinutes, record.GoalLease.MaxMinutes, record.GoalLease.ReturnOnlyWhen, record.GoalLease.CheckpointPolicy),
			GeneratedAtUTC: record.GoalLease.UpdatedAtUTC,
		})
	}
	for i, checkpoint := range record.Checkpoints {
		events = append(events, MissionEvent{
			Schema:         "ao.mission.event.v0.1",
			MissionID:      record.MissionID,
			Kind:           "checkpoint",
			Sequence:       i + 1,
			Status:         checkpoint.Result,
			Route:          checkpoint.Route,
			Phase:          checkpoint.Phase,
			Summary:        checkpoint.ResumeCommand + " " + checkpoint.ExactNextAction,
			GeneratedAtUTC: checkpoint.GeneratedAtUTC,
		})
	}
	if record.Evidence.FoundryRollup != nil {
		events = append(events, MissionEvent{
			Schema:         "ao.mission.event.v0.1",
			MissionID:      record.MissionID,
			Kind:           "foundry_rollup",
			Sequence:       len(events) + 1,
			Status:         record.Evidence.FoundryRollup.Status,
			Summary:        fmt.Sprintf("foundry rollup status %s completed_nodes=%d total_nodes=%d", record.Evidence.FoundryRollup.Status, record.Evidence.FoundryRollup.CompletedNodes, record.Evidence.FoundryRollup.TotalNodes),
			GeneratedAtUTC: record.UpdatedAtUTC,
		})
	}
	if record.ReturnGate != nil {
		events = append(events, MissionEvent{
			Schema:         "ao.mission.event.v0.1",
			MissionID:      record.MissionID,
			Kind:           "return_gate",
			Sequence:       len(events) + 1,
			Status:         record.ReturnGate.Status,
			Summary:        record.ReturnGate.Reason + " " + record.ReturnGate.ExactNextAction,
			GeneratedAtUTC: record.ReturnGate.GeneratedAtUTC,
		})
	}
	if record.Reconciliation != nil {
		events = append(events, MissionEvent{
			Schema:         "ao.mission.event.v0.1",
			MissionID:      record.MissionID,
			Kind:           "route_reconciliation",
			Sequence:       len(events) + 1,
			Status:         record.Reconciliation.Status,
			Route:          record.Reconciliation.CurrentRoute,
			Summary:        record.Reconciliation.ExactNextAction,
			GeneratedAtUTC: record.Reconciliation.GeneratedAtUTC,
		})
	}
	return events
}

func missionEventMatchesQuery(event MissionEvent, query string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		event.MissionID,
		event.Kind,
		event.Status,
		event.Route,
		event.Phase,
		event.ArtifactKind,
		event.Summary,
	}, " "))
	return strings.Contains(haystack, query)
}
