package mission

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (s Store) retainArtifact(body []byte) (string, string, error) {
	digest := digestBytes(body)
	objectPath := filepath.Join(s.Root, "artifacts", "sha256", strings.TrimPrefix(digest, "sha256:"))
	objectDir := filepath.Dir(objectPath)
	if err := os.MkdirAll(objectDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create artifact store: %w", err)
	}

	if _, err := os.Lstat(objectPath); err == nil {
		if err := verifyRetainedArtifact(objectPath, body); err != nil {
			return "", "", err
		}
		return objectPath, digest, nil
	} else if !os.IsNotExist(err) {
		return "", "", fmt.Errorf("inspect retained artifact: %w", err)
	}

	temporary, err := os.CreateTemp(objectDir, ".artifact-")
	if err != nil {
		return "", "", fmt.Errorf("create temporary retained artifact: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if _, err := temporary.Write(body); err != nil {
		_ = temporary.Close()
		return "", "", fmt.Errorf("write temporary retained artifact: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return "", "", fmt.Errorf("sync temporary retained artifact: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", "", fmt.Errorf("close temporary retained artifact: %w", err)
	}

	if err := os.Link(temporaryPath, objectPath); err != nil {
		if _, statErr := os.Lstat(objectPath); statErr == nil {
			if verifyErr := verifyRetainedArtifact(objectPath, body); verifyErr != nil {
				return "", "", verifyErr
			}
			return objectPath, digest, nil
		}
		return "", "", fmt.Errorf("publish retained artifact: %w", err)
	}
	if err := verifyRetainedArtifact(objectPath, body); err != nil {
		return "", "", err
	}
	return objectPath, digest, nil
}

func verifyRetainedArtifact(path string, expected []byte) error {
	pathInfo, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect retained artifact: %w", err)
	}
	if !pathInfo.Mode().IsRegular() || pathInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("retained artifact must be a regular non-symlink file")
	}
	if pathInfo.Size() != int64(len(expected)) {
		return fmt.Errorf("retained artifact size mismatch")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open retained artifact: %w", err)
	}
	openedInfo, statErr := file.Stat()
	if statErr != nil {
		_ = file.Close()
		return fmt.Errorf("stat retained artifact: %w", statErr)
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(pathInfo, openedInfo) {
		_ = file.Close()
		return fmt.Errorf("retained artifact changed while opening")
	}

	actual, readErr := io.ReadAll(io.LimitReader(file, int64(len(expected))))
	if readErr == nil {
		var extra [1]byte
		n, extraErr := file.Read(extra[:])
		if extraErr != nil && extraErr != io.EOF {
			readErr = extraErr
		}
		if n > 0 {
			actual = append(actual, extra[:n]...)
		}
	}
	closeErr := file.Close()
	if readErr != nil {
		return fmt.Errorf("read retained artifact: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close retained artifact: %w", closeErr)
	}

	afterInfo, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("reinspect retained artifact: %w", err)
	}
	if !afterInfo.Mode().IsRegular() || !os.SameFile(openedInfo, afterInfo) {
		return fmt.Errorf("retained artifact changed while reading")
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("retained artifact bytes mismatch")
	}
	return nil
}
