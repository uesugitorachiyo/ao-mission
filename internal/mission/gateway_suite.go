package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func BuildGatewayReplaySuite(readbacks []GatewayReplayReadback, lifecycle *A2ATaskLifecycleReadback, refs []string) GatewayReplaySuiteReadback {
	suite := GatewayReplaySuiteReadback{
		Schema:            "ao.mission.gateway-replay-suite-readback.v0.1",
		Status:            "ready",
		ReplayRefs:        append([]string(nil), refs...),
		Replays:           append([]GatewayReplayReadback(nil), readbacks...),
		A2ALifecycle:      lifecycle,
		MutationAuthority: false,
		ExecutesWork:      false,
		ApprovesWork:      false,
		GeneratedAtUTC:    now(nil),
	}
	for _, readback := range readbacks {
		if strings.HasPrefix(readback.Gateway, "telegram") {
			suite.TelegramReplays++
		}
		if readback.Gateway == "a2a" {
			suite.A2AReplays++
		}
		if readback.Status == "blocked" {
			suite.Status = "blocked"
		}
		suite.Total += readback.Total
		suite.IntentRecorded += readback.IntentRecorded
		suite.Denied += readback.Denied
		suite.Invalid += readback.Invalid
		if readback.MutationAuthority || readback.ExecutesWork || readback.ApprovesWork {
			suite.Status = "blocked"
			suite.MutationAuthority = suite.MutationAuthority || readback.MutationAuthority
			suite.ExecutesWork = suite.ExecutesWork || readback.ExecutesWork
			suite.ApprovesWork = suite.ApprovesWork || readback.ApprovesWork
		}
	}
	if lifecycle != nil {
		suite.A2AReplays++
		suite.Total += lifecycle.Total
		suite.IntentRecorded += lifecycle.IntentRecorded + lifecycle.ResumeRequested + lifecycle.Resumed + lifecycle.CancelRequested + lifecycle.Cancelled
		suite.ArtifactReadbacks += lifecycle.ArtifactReadbacks
		if lifecycle.Status == "blocked" || lifecycle.MutationAuthority || lifecycle.ExecutesWork || lifecycle.ApprovesWork {
			suite.Status = "blocked"
			suite.MutationAuthority = suite.MutationAuthority || lifecycle.MutationAuthority
			suite.ExecutesWork = suite.ExecutesWork || lifecycle.ExecutesWork
			suite.ApprovesWork = suite.ApprovesWork || lifecycle.ApprovesWork
		}
	}
	return suite
}

func BuildA2ACompatibilityReadback(agentCardPath, httpFixturePath, lifecyclePath string) (A2ACompatibilityReadback, error) {
	var card A2AAgentCard
	body, err := os.ReadFile(agentCardPath)
	if err != nil {
		return A2ACompatibilityReadback{}, err
	}
	if err := ValidatePublicSafeText(string(body)); err != nil {
		return A2ACompatibilityReadback{}, err
	}
	if err := json.Unmarshal(body, &card); err != nil {
		return A2ACompatibilityReadback{}, err
	}
	if card.Schema != A2AAgentCardSchema {
		return A2ACompatibilityReadback{}, fmt.Errorf("agent card schema must be %s", A2AAgentCardSchema)
	}
	if card.MutationAuthority || card.CapabilitiesDetail["streaming"] || card.CapabilitiesDetail["push_notifications"] {
		return A2ACompatibilityReadback{}, fmt.Errorf("agent card must remain readback-only")
	}
	if len(card.Skills) == 0 || !card.CapabilitiesDetail["artifact_readbacks"] {
		return A2ACompatibilityReadback{}, fmt.Errorf("agent card must expose readback skills and artifact readbacks")
	}
	httpReadback, err := ReplayA2AHTTPFixture(httpFixturePath)
	if err != nil {
		return A2ACompatibilityReadback{}, err
	}
	lifecycle, err := ReplayA2ATaskLifecycle(lifecyclePath)
	if err != nil {
		return A2ACompatibilityReadback{}, err
	}
	status := "ready"
	if httpReadback.Status != "ready" || lifecycle.Status != "ready" {
		status = "blocked"
	}
	return A2ACompatibilityReadback{
		Schema:            "ao.mission.a2a-compatibility-readback.v0.1",
		Status:            status,
		ProtocolVersion:   card.ProtocolVersion,
		AgentCardSkills:   len(card.Skills),
		Methods:           len(card.Methods),
		HTTPRequests:      httpReadback.Total,
		LifecycleTasks:    lifecycle.Total,
		ArtifactReadbacks: lifecycle.ArtifactReadbacks,
		MutationAuthority: false,
		ExecutesWork:      false,
		ApprovesWork:      false,
		GeneratedAtUTC:    now(nil),
	}, nil
}

func DiffGovernanceSnapshots(before, after GovernanceSnapshot) GovernanceSnapshotDiff {
	fields := []string{}
	if before.MissionID != after.MissionID {
		fields = append(fields, "mission_id")
	}
	if before.CurrentRoute != after.CurrentRoute {
		fields = append(fields, "current_route")
	}
	if before.CurrentPhase != after.CurrentPhase {
		fields = append(fields, "current_phase")
	}
	if before.ExactNextAction != after.ExactNextAction {
		fields = append(fields, "exact_next_action")
	}
	if before.EvidenceFreshnessStatus != after.EvidenceFreshnessStatus {
		fields = append(fields, "evidence_freshness_status")
	}
	status := "unchanged"
	if len(fields) > 0 {
		status = "changed"
	}
	return GovernanceSnapshotDiff{
		Schema:         "ao.mission.governance-snapshot-diff.v0.1",
		Status:         status,
		MissionID:      after.MissionID,
		ChangedFields:  len(fields),
		Fields:         fields,
		SafeToExecute:  false,
		ExecutesWork:   false,
		ApprovesWork:   false,
		GeneratedAtUTC: now(nil),
	}
}

func LoadGovernanceSnapshot(path string) (GovernanceSnapshot, error) {
	var snapshot GovernanceSnapshot
	body, err := os.ReadFile(path)
	if err != nil {
		return snapshot, err
	}
	if err := ValidatePublicSafeText(string(body)); err != nil {
		return snapshot, err
	}
	return snapshot, json.Unmarshal(body, &snapshot)
}

func BuildMissionArchive(record Record) (MissionArchive, error) {
	archive := MissionArchive{
		Schema:         "ao.mission.archive.v0.1",
		MissionID:      record.MissionID,
		Record:         record,
		Snapshot:       Snapshot(record),
		FinalRollup:    BuildFinalRollup(record),
		ArtifactCount:  len(record.ArtifactRefs),
		SafeToExecute:  false,
		ExecutesWork:   false,
		ApprovesWork:   false,
		GeneratedAtUTC: now(nil),
	}
	body, err := json.Marshal(archive)
	if err != nil {
		return MissionArchive{}, err
	}
	sum := sha256.Sum256(body)
	archive.ArchiveDigest = "sha256:" + hex.EncodeToString(sum[:])
	return archive, nil
}
