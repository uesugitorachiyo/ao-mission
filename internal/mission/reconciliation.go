package mission

import "fmt"

func BuildFinalReconciliationPacket(r Record) MissionFinalReconciliationPacket {
	command := BuildCommandStatus(r)
	gate := EvaluateReturnGate(r)
	packet := MissionFinalReconciliationPacket{
		Schema:                 "ao.mission.final-reconciliation-packet.v0.1",
		MissionID:              r.MissionID,
		Status:                 "blocked",
		MissionStatus:          r.Status,
		CommandStatus:          command.Status,
		CompletedNodes:         gate.CompletedNodes,
		ReadyNodes:             gate.ReadyNodesRemaining,
		FinalResponseAllowed:   gate.FinalResponseAllowed,
		ReturnGateStatus:       gate.Status,
		PromotionClaimed:       false,
		RSIRemainsDenied:       true,
		ClaimsAuthorityAdvance: false,
		SafeToExecute:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		MutatesRepositories:    false,
		GeneratedAtUTC:         now(nil),
	}
	if r.Evidence.AtlasRecommendation != nil {
		packet.AtlasRecommendationStatus = r.Evidence.AtlasRecommendation.Status
		packet.CompletedNodes = r.Evidence.AtlasRecommendation.CompletedNodes
		packet.TotalNodes = r.Evidence.AtlasRecommendation.TotalNodes
		packet.ReadyNodes = r.Evidence.AtlasRecommendation.ReadyNodes
		packet.ReturnGateStatus = r.Evidence.AtlasRecommendation.ReturnGateStatus
	}
	if r.Evidence.FoundryRollup != nil {
		packet.FoundryStatus = r.Evidence.FoundryRollup.Status
		if packet.TotalNodes == 0 {
			packet.TotalNodes = r.Evidence.FoundryRollup.TotalNodes
		}
	}
	packet.Blocker = finalReconciliationBlocker(packet, r)
	packet.ArtifactsAgree = packet.Blocker == ""
	if packet.ArtifactsAgree {
		packet.Status = "ready"
	}
	return packet
}

func finalReconciliationBlocker(packet MissionFinalReconciliationPacket, r Record) string {
	if r.Status != "done" || packet.CommandStatus != r.Status || !packet.FinalResponseAllowed {
		return fmt.Sprintf("Mission status=%s Command status=%s final_response_allowed=%t", r.Status, packet.CommandStatus, packet.FinalResponseAllowed)
	}
	if r.Evidence.AtlasRecommendation == nil {
		return "Atlas recommendation evidence missing"
	}
	if r.Evidence.AtlasRecommendation.Status != "completed" ||
		r.Evidence.AtlasRecommendation.TotalNodes == 0 ||
		r.Evidence.AtlasRecommendation.CompletedNodes != r.Evidence.AtlasRecommendation.TotalNodes ||
		r.Evidence.AtlasRecommendation.ReadyNodes != 0 {
		return fmt.Sprintf("Atlas recommendation incomplete status=%s completed_nodes=%d total_nodes=%d ready_nodes=%d", r.Evidence.AtlasRecommendation.Status, r.Evidence.AtlasRecommendation.CompletedNodes, r.Evidence.AtlasRecommendation.TotalNodes, r.Evidence.AtlasRecommendation.ReadyNodes)
	}
	if r.Evidence.FoundryRollup != nil &&
		(r.Evidence.FoundryRollup.CompletedNodes != r.Evidence.AtlasRecommendation.CompletedNodes ||
			r.Evidence.FoundryRollup.TotalNodes != r.Evidence.AtlasRecommendation.TotalNodes) {
		return fmt.Sprintf("Foundry completed_nodes=%d total_nodes=%d disagrees with Atlas completed_nodes=%d total_nodes=%d", r.Evidence.FoundryRollup.CompletedNodes, r.Evidence.FoundryRollup.TotalNodes, r.Evidence.AtlasRecommendation.CompletedNodes, r.Evidence.AtlasRecommendation.TotalNodes)
	}
	if packet.PromotionClaimed || !packet.RSIRemainsDenied || packet.ClaimsAuthorityAdvance {
		return fmt.Sprintf("promotion boundary mismatch promotion_claimed=%t rsi_remains_denied=%t claims_authority_advance=%t", packet.PromotionClaimed, packet.RSIRemainsDenied, packet.ClaimsAuthorityAdvance)
	}
	return ""
}
