package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func BuildArtifactManifest(r Record) ArtifactManifest {
	body, _ := json.Marshal(struct {
		MissionID    string        `json:"mission_id"`
		ArtifactRefs []ArtifactRef `json:"artifact_refs"`
	}{MissionID: r.MissionID, ArtifactRefs: r.ArtifactRefs})
	sum := sha256.Sum256(body)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	return ArtifactManifest{
		Schema:         "ao.mission.artifact-manifest.v0.1",
		MissionID:      r.MissionID,
		ArtifactRefs:   r.ArtifactRefs,
		ManifestDigest: digest,
		Signature:      "ao-mission-local-digest:" + digest,
		SafeToExecute:  false,
		ExecutesWork:   false,
		ApprovesWork:   false,
		GeneratedAtUTC: now(nil),
	}
}

func BuildCommandStatus(r Record) CommandStatus {
	return CommandStatus{
		Schema:              "ao.command.mission-status.v0.1",
		MissionID:           r.MissionID,
		Status:              r.Status,
		CurrentRoute:        r.CurrentRoute,
		CurrentPhase:        r.CurrentPhase,
		ExactNextAction:     r.ExactNextAction,
		ReadOnly:            true,
		SafeToExecute:       false,
		ExecutesWork:        false,
		ApprovesWork:        false,
		MutatesRepositories: false,
		Blockers:            r.Blockers,
		GeneratedAtUTC:      now(nil),
	}
}
