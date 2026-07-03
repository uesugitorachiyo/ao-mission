package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

type ImportReadback struct {
	Schema          string      `json:"schema"`
	MissionID       string      `json:"mission_id"`
	Kind            string      `json:"kind"`
	Status          string      `json:"status"`
	Artifact        ArtifactRef `json:"artifact"`
	ExactNextAction string      `json:"exact_next_action"`
	SafeToExecute   bool        `json:"safe_to_execute"`
	ExecutesWork    bool        `json:"executes_work"`
	ApprovesWork    bool        `json:"approves_work"`
	GeneratedAtUTC  string      `json:"generated_at_utc"`
}

func ImportArtifact(s Store, missionID, kind, path string) (ImportReadback, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return ImportReadback{}, err
	}
	if err := ValidatePublicSafeText(string(body)); err != nil {
		return ImportReadback{}, err
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return ImportReadback{}, err
	}
	ref := ArtifactRef{Schema: ArtifactRefSchema, Ref: path, Digest: digestBytes(body), Kind: kind}
	r, err := s.Update(missionID, func(rec *Record) error {
		rec.ArtifactRefs = append(rec.ArtifactRefs, ref)
		switch kind {
		case "blueprint-authorization":
			rec.CurrentRoute = "ao-atlas"
			rec.CurrentPhase = "blueprint_authorized"
			rec.ExactNextAction = "send authorized Blueprint pack to AO Atlas"
		case "atlas-workgraph":
			rec.CurrentRoute = "ao-foundry"
			rec.CurrentPhase = "atlas_workgraph_ready"
			rec.ExactNextAction = "send first safe Atlas node to AO Foundry"
		case "foundry-run-link":
			rec.CurrentPhase = "foundry_run_link_recorded"
			rec.ExactNextAction = "read next Atlas dependency-unblocked node or final rollup"
		default:
			return fmt.Errorf("unsupported import kind %q", kind)
		}
		return nil
	})
	if err != nil {
		return ImportReadback{}, err
	}
	return ImportReadback{
		Schema:          "ao.mission.import-readback.v0.1",
		MissionID:       r.MissionID,
		Kind:            kind,
		Status:          "recorded",
		Artifact:        ref,
		ExactNextAction: r.ExactNextAction,
		SafeToExecute:   false,
		ExecutesWork:    false,
		ApprovesWork:    false,
		GeneratedAtUTC:  now(nil),
	}, nil
}

func digestBytes(body []byte) string {
	sum := sha256.Sum256(body)
	return "sha256:" + hex.EncodeToString(sum[:])
}
