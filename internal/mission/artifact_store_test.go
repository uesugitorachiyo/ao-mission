package mission

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestRetainArtifactFirstCapture(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)
	body := []byte("first capture\x00with exact bytes\n")

	path, digest, err := store.retainArtifact(body)
	if err != nil {
		t.Fatal(err)
	}
	if digest != digestBytes(body) {
		t.Fatalf("digest=%q want %q", digest, digestBytes(body))
	}
	wantPath := filepath.Join(root, "artifacts", "sha256", strings.TrimPrefix(digest, "sha256:"))
	if path != wantPath {
		t.Fatalf("path=%q want %q", path, wantPath)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("retained object is not a regular non-symlink file: %s", info.Mode())
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("retained bytes=%q want %q", got, body)
	}
}

func TestRetainArtifactExactDeduplication(t *testing.T) {
	store := NewStore(t.TempDir())
	body := []byte("deduplicate these exact bytes")

	firstPath, firstDigest, err := store.retainArtifact(body)
	if err != nil {
		t.Fatal(err)
	}
	secondPath, secondDigest, err := store.retainArtifact(append([]byte(nil), body...))
	if err != nil {
		t.Fatal(err)
	}
	if secondPath != firstPath || secondDigest != firstDigest {
		t.Fatalf("deduplicated result=(%q, %q) first=(%q, %q)", secondPath, secondDigest, firstPath, firstDigest)
	}
	got, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("deduplicated bytes=%q want %q", got, body)
	}
}

func TestRetainArtifactRejectsMismatchedExistingObject(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)
	body := []byte("expected bytes")
	digest := digestBytes(body)
	path := filepath.Join(root, "artifacts", "sha256", strings.TrimPrefix(digest, "sha256:"))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	wantExisting := []byte("different bytes")
	if err := os.WriteFile(path, wantExisting, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := store.retainArtifact(body); err == nil {
		t.Fatal("mismatched existing object was accepted")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, wantExisting) {
		t.Fatalf("mismatched object changed to %q", got)
	}
}

func TestRetainArtifactRejectsSymlinkObject(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)
	body := []byte("symlink must not be followed")
	digest := digestBytes(body)
	path := filepath.Join(root, "artifacts", "sha256", strings.TrimPrefix(digest, "sha256:"))
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	wantTarget := []byte("target remains unchanged")
	if err := os.WriteFile(target, wantTarget, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if _, _, err := store.retainArtifact(body); err == nil {
		t.Fatal("symlink object was accepted")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, wantTarget) {
		t.Fatalf("symlink target changed to %q", got)
	}
}

func TestRetainArtifactConcurrentExactCapture(t *testing.T) {
	store := NewStore(t.TempDir())
	body := []byte("concurrent exact capture\x00")
	const workers = 32

	paths := make([]string, workers)
	digests := make([]string, workers)
	errs := make([]error, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			paths[i], digests[i], errs[i] = store.retainArtifact(body)
		}(i)
	}
	wg.Wait()

	wantDigest := digestBytes(body)
	wantPath := filepath.Join(store.Root, "artifacts", "sha256", strings.TrimPrefix(wantDigest, "sha256:"))
	for i := 0; i < workers; i++ {
		if errs[i] != nil {
			t.Fatalf("worker %d: %v", i, errs[i])
		}
		if paths[i] != wantPath || digests[i] != wantDigest {
			t.Fatalf("worker %d result=(%q, %q) want=(%q, %q)", i, paths[i], digests[i], wantPath, wantDigest)
		}
	}
	got, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("concurrent retained bytes=%q want %q", got, body)
	}
}
