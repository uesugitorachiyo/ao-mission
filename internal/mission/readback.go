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
	refs := append([]ArtifactRef(nil), r.ArtifactRefs...)
	for i := range refs {
		refs[i].ContentRef = ""
	}
	return FinalizeArtifactManifest(ArtifactManifest{
		Schema:        "ao.mission.artifact-manifest.v0.1",
		MissionID:     r.MissionID,
		ArtifactRefs:  refs,
		SafeToExecute: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	})
}

func FinalizeArtifactManifest(manifest ArtifactManifest) ArtifactManifest {
	if manifest.Schema == "" {
		manifest.Schema = "ao.mission.artifact-manifest.v0.1"
	}
	manifest.ManifestDigest = artifactManifestDigest(manifest)
	manifest.Signature = "ao-mission-local-digest:" + manifest.ManifestDigest
	manifest.SafeToExecute = false
	manifest.ExecutesWork = false
	manifest.ApprovesWork = false
	manifest.GeneratedAtUTC = now(nil)
	return manifest
}

func artifactManifestDigest(manifest ArtifactManifest) string {
	var body []byte
	if manifest.Schema == "ao.mission.artifact-manifest.v0.2" {
		body, _ = json.Marshal(struct {
			Schema       string        `json:"schema"`
			MissionID    string        `json:"mission_id"`
			ArtifactRefs []ArtifactRef `json:"artifact_refs"`
		}{Schema: manifest.Schema, MissionID: manifest.MissionID, ArtifactRefs: manifest.ArtifactRefs})
	} else {
		// Preserve the v0.1 digest representation for historical manifests.
		body, _ = json.Marshal(struct {
			MissionID    string        `json:"mission_id"`
			ArtifactRefs []ArtifactRef `json:"artifact_refs"`
		}{MissionID: manifest.MissionID, ArtifactRefs: manifest.ArtifactRefs})
	}
	sum := sha256.Sum256(body)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func MaterializeArtifactManifest(r Record, outPath string) (ArtifactManifest, error) {
	refs := append([]ArtifactRef(nil), r.ArtifactRefs...)
	for i := range refs {
		source := refs[i].ContentRef
		if source == "" {
			source = refs[i].Ref
		}
		data, err := readArtifactFile(source)
		if err != nil {
			return ArtifactManifest{}, fmt.Errorf("read retained artifact %s: %w", refs[i].Ref, err)
		}
		if digestBytes(data) != refs[i].Digest {
			return ArtifactManifest{}, fmt.Errorf("artifact digest mismatch for %s", refs[i].Ref)
		}
		refs[i].ContentRef = artifactManifestContentRef(refs[i].Digest)
		if err := writeArtifactManifestContent(outPath, refs[i].ContentRef, data); err != nil {
			return ArtifactManifest{}, err
		}
	}
	return FinalizeArtifactManifest(ArtifactManifest{
		Schema:       "ao.mission.artifact-manifest.v0.2",
		MissionID:    r.MissionID,
		ArtifactRefs: refs,
	}), nil
}

func ValidateArtifactManifestFile(path string) (ArtifactManifestValidation, error) {
	var manifest ArtifactManifest
	body, err := os.ReadFile(path)
	if err != nil {
		return ArtifactManifestValidation{Schema: "ao.mission.artifact-manifest-validation.v0.1", Status: "failed", GeneratedAtUTC: now(nil)}, err
	}
	if err := validateNoDuplicateJSONKeys(body); err != nil {
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
	if manifest.Schema != "ao.mission.artifact-manifest.v0.1" && manifest.Schema != "ao.mission.artifact-manifest.v0.2" {
		result.Status = "failed"
		return result, fmt.Errorf("artifact manifest schema must be ao.mission.artifact-manifest.v0.1 or ao.mission.artifact-manifest.v0.2")
	}
	if manifest.ExecutesWork || manifest.ApprovesWork || manifest.SafeToExecute {
		result.Status = "failed"
		return result, fmt.Errorf("artifact manifest must not claim execution or approval authority")
	}
	expected := artifactManifestDigest(manifest)
	if manifest.ManifestDigest != expected {
		result.Status = "failed"
		return result, fmt.Errorf("artifact manifest digest mismatch")
	}
	for _, ref := range manifest.ArtifactRefs {
		if err := validateArtifactManifestRef(ref, manifest.Schema); err != nil {
			result.Status = "failed"
			return result, err
		}
		actualPath, err := artifactManifestArtifactPath(path, ref, manifest.Schema)
		if err != nil {
			result.Status = "failed"
			return result, err
		}
		data, err := readArtifactFile(actualPath)
		if err != nil {
			result.Status = "failed"
			return result, err
		}
		got := digestBytes(data)
		if manifest.Schema == "ao.mission.artifact-manifest.v0.1" {
			sum := sha256.Sum256(normalizeTextArtifactDigestData(data))
			got = "sha256:" + hex.EncodeToString(sum[:])
		}
		if got != ref.Digest {
			result.Status = "failed"
			return result, fmt.Errorf("artifact digest mismatch for %s", ref.Ref)
		}
	}
	return result, nil
}

func RepairArtifactManifestFile(path string) (ArtifactManifest, error) {
	return repairArtifactManifestFile(path, path)
}

func repairArtifactManifestFile(path, outPath string) (ArtifactManifest, error) {
	var manifest ArtifactManifest
	body, err := os.ReadFile(path)
	if err != nil {
		return ArtifactManifest{}, err
	}
	if err := validateNoDuplicateJSONKeys(body); err != nil {
		return ArtifactManifest{}, err
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		return ArtifactManifest{}, err
	}
	if manifest.Schema != "ao.mission.artifact-manifest.v0.1" && manifest.Schema != "ao.mission.artifact-manifest.v0.2" {
		return ArtifactManifest{}, fmt.Errorf("artifact manifest schema must be ao.mission.artifact-manifest.v0.1 or ao.mission.artifact-manifest.v0.2")
	}
	if manifest.SafeToExecute || manifest.ExecutesWork || manifest.ApprovesWork {
		return ArtifactManifest{}, fmt.Errorf("artifact manifest must not claim execution or approval authority")
	}
	if manifest.ManifestDigest != artifactManifestDigest(manifest) {
		return ArtifactManifest{}, fmt.Errorf("artifact manifest digest mismatch")
	}
	refs := append([]ArtifactRef(nil), manifest.ArtifactRefs...)
	for i, ref := range refs {
		if err := validateArtifactManifestRef(ref, manifest.Schema); err != nil {
			return ArtifactManifest{}, err
		}
		actualPath, err := artifactManifestArtifactPath(path, ref, manifest.Schema)
		if err != nil {
			return ArtifactManifest{}, err
		}
		data, err := readArtifactFile(actualPath)
		if err != nil {
			return ArtifactManifest{}, err
		}
		if digestBytes(data) != ref.Digest {
			return ArtifactManifest{}, fmt.Errorf("artifact digest mismatch for %s", ref.Ref)
		}
		refs[i].ContentRef = artifactManifestContentRef(ref.Digest)
		if err := writeArtifactManifestContent(outPath, refs[i].ContentRef, data); err != nil {
			return ArtifactManifest{}, err
		}
	}
	return FinalizeArtifactManifest(ArtifactManifest{
		Schema:       "ao.mission.artifact-manifest.v0.2",
		MissionID:    manifest.MissionID,
		ArtifactRefs: refs,
	}), nil
}

func validateArtifactManifestRef(ref ArtifactRef, schema string) error {
	if strings.TrimSpace(ref.Ref) == "" || strings.TrimSpace(ref.Digest) == "" {
		return fmt.Errorf("artifact manifest refs require ref and digest")
	}
	if !strings.HasPrefix(ref.Digest, "sha256:") {
		return fmt.Errorf("artifact manifest ref %s digest must start with sha256:", ref.Ref)
	}
	if schema == "ao.mission.artifact-manifest.v0.2" {
		if ref.Schema != ArtifactRefSchema {
			return fmt.Errorf("artifact manifest ref %s artifact ref schema must be %s", ref.Ref, ArtifactRefSchema)
		}
		if ref.ContentRef != artifactManifestContentRef(ref.Digest) {
			return fmt.Errorf("artifact manifest ref %s content_ref must be contained and digest-addressed", ref.Ref)
		}
	}
	return nil
}

func artifactManifestArtifactPath(manifestPath string, ref ArtifactRef, schema string) (string, error) {
	if schema == "ao.mission.artifact-manifest.v0.2" {
		return artifactManifestContentPath(manifestPath, ref.ContentRef)
	}
	actualPath := ref.Ref
	if !filepath.IsAbs(actualPath) {
		if _, err := os.Stat(actualPath); err != nil {
			actualPath = filepath.Join(filepath.Dir(manifestPath), actualPath)
		}
	}
	return actualPath, nil
}

func artifactManifestContentRef(digest string) string {
	return filepath.ToSlash(filepath.Join(retainedArtifactDirectory, strings.TrimPrefix(digest, "sha256:")))
}

func artifactManifestContentPath(manifestPath, contentRef string) (string, error) {
	if contentRef != artifactManifestContentRef("sha256:"+filepath.Base(filepath.FromSlash(contentRef))) {
		return "", fmt.Errorf("artifact manifest content_ref must be contained and digest-addressed")
	}
	return filepath.Join(filepath.Dir(manifestPath), filepath.FromSlash(contentRef)), nil
}

func writeArtifactManifestContent(manifestPath, contentRef string, data []byte) error {
	contentPath, err := artifactManifestContentPath(manifestPath, contentRef)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(contentPath), 0o755); err != nil {
		return fmt.Errorf("create artifact manifest content directory: %w", err)
	}
	if existing, err := readArtifactFile(contentPath); err == nil {
		if string(existing) != string(data) {
			return fmt.Errorf("artifact manifest content collision for %s", contentRef)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.WriteFile(contentPath, data, 0o644); err != nil {
		return fmt.Errorf("write artifact manifest content: %w", err)
	}
	return nil
}

func readArtifactFile(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("artifact path is required")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("artifact must be a regular non-symlink file")
	}
	return os.ReadFile(path)
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
	var goalLease *GoalLease
	if r.GoalLease != nil {
		copy := *r.GoalLease
		goalLease = &copy
	}
	gate := r.ReturnGate
	if gate == nil {
		evaluated := EvaluateReturnGate(r)
		gate = &evaluated
	}
	checkpointFreshness := "missing"
	if len(r.Checkpoints) > 0 {
		checkpointFreshness = "fresh"
	} else if gate != nil && !gate.FinalResponseAllowed {
		checkpointFreshness = "stale_or_missing"
	} else if gate != nil && gate.FinalResponseAllowed {
		checkpointFreshness = "not_required"
	}
	returnGateStatus := ""
	if gate != nil {
		returnGateStatus = gate.Status
	}
	return CommandStatus{
		Schema:                    "ao.command.mission-status.v0.1",
		MissionID:                 r.MissionID,
		CorrelationID:             r.CorrelationID,
		Status:                    r.Status,
		CurrentRoute:              r.CurrentRoute,
		CurrentPhase:              r.CurrentPhase,
		ExactNextAction:           r.ExactNextAction,
		GoalLease:                 goalLease,
		CheckpointCount:           len(r.Checkpoints),
		CheckpointFreshnessStatus: checkpointFreshness,
		ReturnGateStatus:          returnGateStatus,
		ReadOnly:                  true,
		SafeToExecute:             false,
		ExecutesWork:              false,
		ApprovesWork:              false,
		MutatesRepositories:       false,
		AtlasRecommendation:       atlasRecommendation,
		Blockers:                  r.Blockers,
		GeneratedAtUTC:            now(nil),
	}
}
