//go:build windows

package mission

import (
	"syscall"
	"unsafe"
)

const (
	missionMoveFileReplaceExisting = 0x00000001
	missionMoveFileWriteThrough    = 0x00000008
)

var moveFileExProc = syscall.NewLazyDLL("kernel32.dll").NewProc("MoveFileExW")

func replaceAtomicFile(source, destination string) error {
	sourcePointer, err := syscall.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	destinationPointer, err := syscall.UTF16PtrFromString(destination)
	if err != nil {
		return err
	}
	result, _, callErr := moveFileExProc.Call(
		uintptr(unsafe.Pointer(sourcePointer)),
		uintptr(unsafe.Pointer(destinationPointer)),
		missionMoveFileReplaceExisting|missionMoveFileWriteThrough,
	)
	if result == 0 {
		return callErr
	}
	return nil
}

func syncDirectory(_ string) error {
	return nil
}
