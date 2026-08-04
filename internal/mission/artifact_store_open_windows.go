//go:build windows

package mission

import (
	"errors"
	"os"
	"syscall"
)

func openRetainedArtifactFileNoFollow(path string) (*os.File, error) {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := syscall.CreateFile(
		pointer,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL|syscall.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return nil, err
	}
	var information syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(handle, &information); err != nil {
		_ = syscall.CloseHandle(handle)
		return nil, err
	}
	if information.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0 ||
		information.FileAttributes&syscall.FILE_ATTRIBUTE_DIRECTORY != 0 {
		_ = syscall.CloseHandle(handle)
		return nil, errors.New("retained artifact is a reparse point or directory")
	}
	if fileType, err := syscall.GetFileType(handle); err != nil || fileType != correlationWindowsFileTypeDisk {
		_ = syscall.CloseHandle(handle)
		if err != nil {
			return nil, err
		}
		return nil, errors.New("retained artifact is not a disk file")
	}
	file := os.NewFile(uintptr(handle), path)
	if file == nil {
		_ = syscall.CloseHandle(handle)
		return nil, errors.New("open retained artifact")
	}
	return file, nil
}

func validateRetainedArtifactDirectoryPlatform(path string) error {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return err
	}
	if attributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return errors.New("artifact store directory contains a reparse point")
	}
	return nil
}
