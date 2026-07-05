package mission

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
	packet.ArtifactsAgree = finalReconciliationArtifactsAgree(packet, r)
	if packet.ArtifactsAgree {
		packet.Status = "ready"
	}
	return packet
}

func finalReconciliationArtifactsAgree(packet MissionFinalReconciliationPacket, r Record) bool {
	if r.Status != "done" || packet.CommandStatus != r.Status || !packet.FinalResponseAllowed {
		return false
	}
	if r.Evidence.AtlasRecommendation == nil {
		return false
	}
	if r.Evidence.AtlasRecommendation.Status != "completed" ||
		r.Evidence.AtlasRecommendation.TotalNodes == 0 ||
		r.Evidence.AtlasRecommendation.CompletedNodes != r.Evidence.AtlasRecommendation.TotalNodes ||
		r.Evidence.AtlasRecommendation.ReadyNodes != 0 {
		return false
	}
	if r.Evidence.FoundryRollup != nil &&
		(r.Evidence.FoundryRollup.CompletedNodes != r.Evidence.AtlasRecommendation.CompletedNodes ||
			r.Evidence.FoundryRollup.TotalNodes != r.Evidence.AtlasRecommendation.TotalNodes) {
		return false
	}
	return !packet.PromotionClaimed && packet.RSIRemainsDenied && !packet.ClaimsAuthorityAdvance
}
