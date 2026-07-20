package mission

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseRehearsalWorkflowContract(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "release-rehearsal.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)

	for _, want := range []string{
		"workflow_dispatch:",
		"version:",
		"tag:",
		"source_sha:",
		"approved_manifest_digest:",
		"dry_run:",
		"default: true",
		"live_confirmation:",
		"ubuntu-latest",
		"macos-latest",
		"windows-latest",
		"linux-x86_64",
		"macos-aarch64",
		"windows-x86_64",
		"actions/upload-artifact",
		"candidate-summary.json",
		"SHA256SUMS",
		"immutable-promotion-plan.json",
		"immutable-promotion-plan.sha256",
		"dry-run-boundary.json",
		"tag_creation_attempted",
		"release_creation_attempted",
		"public_upload_attempted",
		"publication_performed",
		"if: ${{ inputs.dry_run == false",
		"inputs.live_confirmation == format(",
		"environment: ao-mission-release",
		"contents: write",
		"gh release create",
		"candidate archive checksum mismatch",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("release rehearsal workflow missing %q", want)
		}
	}

	if strings.Count(workflow, "contents: write") != 1 {
		t.Fatalf("release rehearsal workflow must grant contents: write only to the publisher job")
	}
	for _, forbidden := range []string{
		"pull_request:",
		"push:",
		"personal_access_token",
		"secrets.PAT",
	} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("release rehearsal workflow must not include %q", forbidden)
		}
	}
}
