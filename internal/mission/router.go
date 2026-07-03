package mission

import "strings"

func DecideRoute(missionID, objective string, artifacts []ArtifactRef) RouteDecision {
	lower := strings.ToLower(objective)
	words := len(strings.Fields(objective))
	decision := RouteDecision{Schema: RouteSchema, MissionID: missionID, SafeToRequest: true, SafeToExecute: false, SafeToPromote: false, GeneratedAtUTC: now(nil)}
	switch {
	case words < 4 || strings.Contains(lower, "figure out") || strings.Contains(lower, "not sure"):
		decision.Route = "ao-blueprint"
		decision.Reason = "objective is underspecified"
		decision.ExactNextAction = "send objective to AO Blueprint for requirements and authorization"
	case strings.Contains(lower, "workgraph") || strings.Contains(lower, "long-running") || strings.Contains(lower, "oversized") || strings.Contains(lower, "mutation-class") || strings.Contains(lower, "atlas"):
		decision.Route = "ao-atlas"
		decision.Reason = "objective requires workgraph, context, or long-running task management"
		decision.ExactNextAction = "send authorized pack to AO Atlas"
	case strings.Contains(lower, "ready node") || strings.Contains(lower, "foundry import"):
		decision.Route = "ao-foundry"
		decision.Reason = "ready workgraph node is present"
		decision.ExactNextAction = "send first safe node to AO Foundry"
	default:
		decision.Route = "ao-blueprint"
		decision.Reason = "default conservative requirements front door"
		decision.ExactNextAction = "send objective to AO Blueprint for requirements and authorization"
	}
	return decision
}

func NextAction(r Record) RouteDecision { return DecideRoute(r.MissionID, r.Objective, r.ArtifactRefs) }

func AppendRouteHistory(r *Record, decision RouteDecision) {
	if decision.Schema == "" {
		decision.Schema = RouteSchema
	}
	if decision.MissionID == "" {
		decision.MissionID = r.MissionID
	}
	if decision.GeneratedAtUTC == "" {
		decision.GeneratedAtUTC = now(nil)
	}
	r.RouteHistory = append(r.RouteHistory, decision)
}
