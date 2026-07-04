package mission

import (
	"fmt"
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
		Schema:              "ao.mission.event-index.v0.1",
		Status:              "ready",
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
	return index, nil
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
	readback.Checks = append(readback.Checks, "mission_records_readable", "mission_event_index_readable", "authority_flags_false")
	return readback
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
