//go:build darwin || linux

package mission

import (
	"errors"
	"os"
	"syscall"
)

func openRetainedArtifactFileNoFollow(path string) (*os.File, error) {
	descriptor, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NONBLOCK|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(descriptor), path)
	if file == nil {
		_ = syscall.Close(descriptor)
		return nil, errors.New("open retained artifact")
	}
	return file, nil
}

func validateRetainedArtifactDirectoryPlatform(string) error {
	return nil
}
