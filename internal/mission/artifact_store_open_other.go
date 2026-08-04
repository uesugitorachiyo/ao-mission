//go:build !darwin && !linux && !windows

package mission

import "os"

func openRetainedArtifactFileNoFollow(path string) (*os.File, error) {
	return os.Open(path)
}

func validateRetainedArtifactDirectoryPlatform(string) error {
	return nil
}
