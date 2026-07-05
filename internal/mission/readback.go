package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func BuildArtifactManifest(r Record) ArtifactManifest {
	return FinalizeArtifactManifest(ArtifactManifest{
		Schema:        "ao.mission.artifact-manifest.v0.1",
		MissionID:     r.MissionID,
		ArtifactRefs:  r.ArtifactRefs,
		SafeToExecute: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	})
}

func FinalizeArtifactManifest(manifest ArtifactManifest) ArtifactManifest {
	body, _ := json.Marshal(struct {
		MissionID    string        `json:"mission_id"`
		ArtifactRefs []ArtifactRef `json:"artifact_refs"`
	}{MissionID: manifest.MissionID, ArtifactRefs: manifest.ArtifactRefs})
	sum := sha256.Sum256(body)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	manifest.Schema = "ao.mission.artifact-manifest.v0.1"
	manifest.ManifestDigest = digest
	manifest.Signature = "ao-mission-local-digest:" + digest
	manifest.SafeToExecute = false
	manifest.ExecutesWork = false
	manifest.ApprovesWork = false
	manifest.GeneratedAtUTC = now(nil)
	return manifest
}

func ValidateArtifactManifestFile(path string) (ArtifactManifestValidation, error) {
	var manifest ArtifactManifest
	body, err := os.ReadFile(path)
	if err != nil {
		return ArtifactManifestValidation{Schema: "ao.mission.artifact-manifest-validation.v0.1", Status: "failed", GeneratedAtUTC: now(nil)}, err
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		return ArtifactManifestValidation{Schema: "ao.mission.artifact-manifest-validation.v0.1", Status: "failed", GeneratedAtUTC: now(nil)}, err
	}
	result := ArtifactManifestValidation{
		Schema:         "ao.mission.artifact-manifest-validation.v0.1",
		Status:         "passed",
		MissionID:      manifest.MissionID,
		ArtifactCount:  len(manifest.ArtifactRefs),
		ManifestDigest: manifest.ManifestDigest,
		ExecutesWork:   false,
		ApprovesWork:   false,
		GeneratedAtUTC: now(nil),
	}
	if manifest.Schema != "ao.mission.artifact-manifest.v0.1" {
		result.Status = "failed"
		return result, fmt.Errorf("artifact manifest schema must be ao.mission.artifact-manifest.v0.1")
	}
	if manifest.ExecutesWork || manifest.ApprovesWork || manifest.SafeToExecute {
		result.Status = "failed"
		return result, fmt.Errorf("artifact manifest must not claim execution or approval authority")
	}
	expected := FinalizeArtifactManifest(ArtifactManifest{MissionID: manifest.MissionID, ArtifactRefs: manifest.ArtifactRefs}).ManifestDigest
	if manifest.ManifestDigest != expected {
		result.Status = "failed"
		return result, fmt.Errorf("artifact manifest digest mismatch")
	}
	for _, ref := range manifest.ArtifactRefs {
		if strings.TrimSpace(ref.Ref) == "" || strings.TrimSpace(ref.Digest) == "" {
			result.Status = "failed"
			return result, fmt.Errorf("artifact manifest refs require ref and digest")
		}
		if !strings.HasPrefix(ref.Digest, "sha256:") {
			result.Status = "failed"
			return result, fmt.Errorf("artifact manifest ref %s digest must start with sha256:", ref.Ref)
		}
		actualPath := ref.Ref
		if !filepath.IsAbs(actualPath) {
			if _, err := os.Stat(actualPath); err != nil {
				actualPath = filepath.Join(filepath.Dir(path), actualPath)
			}
		}
		data, err := os.ReadFile(actualPath)
		if err != nil {
			result.Status = "failed"
			return result, err
		}
		sum := sha256.Sum256(normalizeTextArtifactDigestData(data))
		got := "sha256:" + hex.EncodeToString(sum[:])
		if got != ref.Digest {
			result.Status = "failed"
			return result, fmt.Errorf("artifact digest mismatch for %s", ref.Ref)
		}
	}
	return result, nil
}

func RepairArtifactManifestFile(path string) (ArtifactManifest, error) {
	var manifest ArtifactManifest
	body, err := os.ReadFile(path)
	if err != nil {
		return ArtifactManifest{}, err
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		return ArtifactManifest{}, err
	}
	for i, ref := range manifest.ArtifactRefs {
		actualPath := ref.Ref
		if !filepath.IsAbs(actualPath) {
			if _, err := os.Stat(actualPath); err != nil {
				actualPath = filepath.Join(filepath.Dir(path), actualPath)
			}
		}
		data, err := os.ReadFile(actualPath)
		if err != nil {
			return ArtifactManifest{}, err
		}
		sum := sha256.Sum256(normalizeTextArtifactDigestData(data))
		manifest.ArtifactRefs[i].Digest = "sha256:" + hex.EncodeToString(sum[:])
		if strings.TrimSpace(manifest.ArtifactRefs[i].Schema) == "" {
			manifest.ArtifactRefs[i].Schema = ArtifactRefSchema
		}
	}
	return FinalizeArtifactManifest(ArtifactManifest{
		MissionID:    manifest.MissionID,
		ArtifactRefs: manifest.ArtifactRefs,
	}), nil
}

func normalizeTextArtifactDigestData(data []byte) []byte {
	if !utf8.Valid(data) {
		return data
	}
	return []byte(strings.ReplaceAll(string(data), "\r\n", "\n"))
}

func BuildCommandStatus(r Record) CommandStatus {
	var atlasRecommendation *AtlasRecommendationReadbackCounts
	if r.Evidence.AtlasRecommendation != nil {
		copy := *r.Evidence.AtlasRecommendation
		atlasRecommendation = &copy
	}
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
		AtlasRecommendation: atlasRecommendation,
		Blockers:            r.Blockers,
		GeneratedAtUTC:      now(nil),
	}
}
