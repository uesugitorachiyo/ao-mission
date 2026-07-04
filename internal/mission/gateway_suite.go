package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
		if suite.CorrelationID == "" && strings.TrimSpace(readback.CorrelationID) != "" {
			suite.CorrelationID = strings.TrimSpace(readback.CorrelationID)
		}
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

func BuildA2AStreamingDenialReadback(agentCardPath string) (A2AStreamingDenialReadback, error) {
	var card A2AAgentCard
	body, err := os.ReadFile(agentCardPath)
	if err != nil {
		return A2AStreamingDenialReadback{}, err
	}
	if err := ValidatePublicSafeText(string(body)); err != nil {
		return A2AStreamingDenialReadback{}, err
	}
	if err := json.Unmarshal(body, &card); err != nil {
		return A2AStreamingDenialReadback{}, err
	}
	streaming := card.CapabilitiesDetail["streaming"]
	push := card.CapabilitiesDetail["push_notifications"]
	denied := "none"
	status := "ready"
	if streaming {
		denied = "streaming"
		status = "denied"
	}
	if push {
		if denied == "none" {
			denied = "push_notifications"
		} else {
			denied += ",push_notifications"
		}
		status = "denied"
	}
	return A2AStreamingDenialReadback{
		Schema:             "ao.mission.a2a-streaming-denial-readback.v0.1",
		Status:             status,
		StreamingRequested: streaming,
		PushRequested:      push,
		DeniedCapability:   denied,
		MutationAuthority:  false,
		ExecutesWork:       false,
		ApprovesWork:       false,
		ExactNextAction:    "keep A2A gateway in readback-only non-streaming mode",
		GeneratedAtUTC:     now(nil),
	}, nil
}

func BuildA2ACancellationReplayReadback(lifecyclePath string) (A2ACancellationReplayReadback, error) {
	lifecycle, err := ReplayA2ATaskLifecycle(lifecyclePath)
	if err != nil {
		return A2ACancellationReplayReadback{}, err
	}
	status := "ready"
	if lifecycle.CancelRequested == 0 || lifecycle.Cancelled == 0 || lifecycle.MutationAuthority || lifecycle.ExecutesWork || lifecycle.ApprovesWork {
		status = "blocked"
	}
	return A2ACancellationReplayReadback{
		Schema:            "ao.mission.a2a-cancellation-replay-readback.v0.1",
		Status:            status,
		Total:             lifecycle.Total,
		CancelRequested:   lifecycle.CancelRequested,
		Cancelled:         lifecycle.Cancelled,
		MutationAuthority: false,
		ExecutesWork:      false,
		ApprovesWork:      false,
		ExactNextAction:   "record A2A cancellation as readback only; route any continuation through AO Mission gates",
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

func BuildTelegramRoleMatrix(cfg TelegramConfig) TelegramRoleMatrixReadback {
	entries := make([]TelegramRoleEntry, 0, len(cfg.AllowedChats))
	for chatID, role := range cfg.AllowedChats {
		entries = append(entries, TelegramRoleEntry{ChatID: chatID, Role: role})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ChatID < entries[j].ChatID })
	readback := TelegramRoleMatrixReadback{
		Schema:            "ao.mission.telegram-role-matrix-readback.v0.1",
		Status:            "ready",
		ChatCount:         len(entries),
		Roles:             entries,
		MutationAuthority: false,
		ExecutesWork:      false,
		ApprovesWork:      false,
		GeneratedAtUTC:    now(nil),
	}
	for _, entry := range entries {
		switch entry.Role {
		case "admin":
			readback.AdminCount++
		case "user":
			readback.UserCount++
		}
	}
	return readback
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

func LoadMissionArchive(path string) (MissionArchive, error) {
	var archive MissionArchive
	body, err := os.ReadFile(path)
	if err != nil {
		return archive, err
	}
	if err := ValidatePublicSafeText(string(body)); err != nil {
		return archive, err
	}
	if err := json.Unmarshal(body, &archive); err != nil {
		return archive, err
	}
	if archive.Schema != "ao.mission.archive.v0.1" {
		return archive, fmt.Errorf("mission archive schema must be ao.mission.archive.v0.1")
	}
	if archive.SafeToExecute || archive.ExecutesWork || archive.ApprovesWork {
		return archive, fmt.Errorf("mission archive must not claim execution or approval authority")
	}
	return archive, nil
}

func ValidateMissionArchive(path string) (MissionArchiveValidation, error) {
	archive, err := LoadMissionArchive(path)
	if err != nil {
		return MissionArchiveValidation{}, err
	}
	expected := archive.ArchiveDigest
	archive.ArchiveDigest = ""
	body, err := json.Marshal(archive)
	if err != nil {
		return MissionArchiveValidation{}, err
	}
	sum := sha256.Sum256(body)
	actual := "sha256:" + hex.EncodeToString(sum[:])
	if expected == "" || actual != expected {
		return MissionArchiveValidation{}, fmt.Errorf("mission archive digest mismatch")
	}
	return MissionArchiveValidation{
		Schema:         "ao.mission.archive-validation.v0.1",
		Status:         "ready",
		MissionID:      archive.MissionID,
		ArchiveDigest:  expected,
		ArtifactCount:  archive.ArtifactCount,
		SafeToExecute:  false,
		ExecutesWork:   false,
		ApprovesWork:   false,
		GeneratedAtUTC: now(nil),
	}, nil
}

func ImportMissionArchive(store Store, path string) (MissionArchiveImportReadback, error) {
	validation, err := ValidateMissionArchive(path)
	if err != nil {
		return MissionArchiveImportReadback{}, err
	}
	archive, err := LoadMissionArchive(path)
	if err != nil {
		return MissionArchiveImportReadback{}, err
	}
	if archive.Record.Schema != RecordSchema || archive.Record.MissionID != archive.MissionID {
		return MissionArchiveImportReadback{}, fmt.Errorf("mission archive record does not match archive mission_id")
	}
	if err := store.Save(archive.Record); err != nil {
		return MissionArchiveImportReadback{}, err
	}
	return MissionArchiveImportReadback{
		Schema:         "ao.mission.archive-import-readback.v0.1",
		Status:         "ready",
		MissionID:      archive.MissionID,
		ArchiveDigest:  validation.ArchiveDigest,
		SafeToExecute:  false,
		ExecutesWork:   false,
		ApprovesWork:   false,
		GeneratedAtUTC: now(nil),
	}, nil
}

func BuildGatewayReadinessRollup(paths ...string) (GatewayReadinessRollup, error) {
	return BuildGatewayReadinessRollupWithCorrelation("", paths...)
}

func BuildGatewayReadinessRollupWithCorrelation(correlationID string, paths ...string) (GatewayReadinessRollup, error) {
	return BuildGatewayReadinessRollupWithMissionAndCorrelation("", correlationID, paths...)
}

func BuildGatewayReadinessRollupWithMissionAndCorrelation(missionID, correlationID string, paths ...string) (GatewayReadinessRollup, error) {
	rollup := GatewayReadinessRollup{
		Schema:              "ao.mission.gateway-readiness-rollup.v0.1",
		MissionID:           strings.TrimSpace(missionID),
		Status:              "ready",
		CorrelationID:       strings.TrimSpace(correlationID),
		ReadbackRefs:        []string{},
		SafeToExecute:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		ExactNextAction:     "route ready gateway readbacks through AO Mission, Atlas, Foundry, and Command gates",
		GeneratedAtUTC:      now(nil),
	}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return GatewayReadinessRollup{}, err
		}
		if err := ValidatePublicSafeText(string(body)); err != nil {
			return GatewayReadinessRollup{}, err
		}
		var packet map[string]any
		if err := json.Unmarshal(body, &packet); err != nil {
			return GatewayReadinessRollup{}, err
		}
		if rollup.CorrelationID == "" {
			if value, ok := packet["correlation_id"].(string); ok {
				rollup.CorrelationID = strings.TrimSpace(value)
			}
		}
		rollup.ReadbackCount++
		rollup.ReadbackRefs = append(rollup.ReadbackRefs, path)
		if packet["status"] == "blocked" || packet["status"] == "denied" || packet["safe_to_execute"] == true || packet["executes_work"] == true || packet["approves_work"] == true || packet["mutation_authority"] == true {
			rollup.BlockedReadbacks++
			rollup.Status = "blocked"
		}
	}
	if rollup.ReadbackCount == 0 {
		return GatewayReadinessRollup{}, fmt.Errorf("gateway readiness rollup requires at least one readback")
	}
	return rollup, nil
}
