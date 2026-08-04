//go:build windows

package mission

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsRetainArtifactFailsClosedOnEveryDurabilityPath(t *testing.T) {
	body := []byte("Windows durability test")
	digest := digestBytes(body)
	objectName := filepath.Join(retainedArtifactDirectory, strings.TrimPrefix(digest, "sha256:"))

	tests := []struct {
		name  string
		setup func(t *testing.T, root string)
	}{
		{name: "first publication"},
		{
			name: "exact deduplication",
			setup: func(t *testing.T, root string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Join(root, retainedArtifactDirectory), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, objectName), body, 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "link collision exact reuse",
			setup: func(t *testing.T, _ string) {
				t.Helper()
				previous := beforeRetainedArtifactLink
				t.Cleanup(func() { beforeRetainedArtifactLink = previous })
				beforeRetainedArtifactLink = func(root retainedArtifactRoot, _, path string) error {
					return root.WriteFile(path, body, 0o644)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			if test.setup != nil {
				test.setup(t, root)
			}
			path, _, err := NewStore(root).retainArtifact(body)
			if path != "" {
				t.Fatalf("retention returned path despite missing durability: %q", path)
			}
			if !errors.Is(err, errRetainedArtifactWindowsDurabilityUnsupported) {
				t.Fatalf("retention error=%v, want %v", err, errRetainedArtifactWindowsDurabilityUnsupported)
			}
		})
	}
}
