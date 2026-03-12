//go:build windows && 386

package mrvtcl

import (
	"fmt"
	"syscall"
)

var (
	mfLibOpen       uintptr
	mfLibClose      uintptr
	mfBeginPainting uintptr
	mfEndPainting   uintptr
)

func MF_LibOpen() {
	if mfLibOpen != 0 {
		syscall.Syscall(mfLibOpen, 0, 0, 0, 0)
	}
}

func MF_LibClose() {
	if mfLibClose != 0 {
		syscall.Syscall(mfLibClose, 0, 0, 0, 0)
	}
}

func MF_BeginPainting(hdc uintptr) int32 {
	if mfBeginPainting == 0 {
		return 0
	}
	ret, _, _ := syscall.Syscall(mfBeginPainting, 1, hdc, 0, 0)
	return int32(ret)
}

func MF_EndPainting(hdc uintptr) int32 {
	if mfEndPainting == 0 {
		return 0
	}
	ret, _, _ := syscall.Syscall(mfEndPainting, 1, hdc, 0, 0)
	return int32(ret)
}

func loadMrvDrvProc(module uintptr, name string) (uintptr, error) {
	n, err := syscall.BytePtrFromString(name)
	if err != nil {
		return 0, err
	}
	proc := GetProcAddress(module, n)
	if proc == 0 {
		return 0, fmt.Errorf("GetProcAddress failed for %s: %d", name, GetLastError())
	}
	return proc, nil
}

func loadMrvDrv(module uintptr) error {
	var err error

	mfLibOpen, err = loadMrvDrvProc(module, "MF_LibOpen")
	if err != nil {
		return err
	}

	mfLibClose, err = loadMrvDrvProc(module, "MF_LibClose")
	if err != nil {
		return err
	}

	mfBeginPainting, err = loadMrvDrvProc(module, "MF_BeginPainting")
	if err != nil {
		return err
	}

	mfEndPainting, err = loadMrvDrvProc(module, "MF_EndPainting")
	if err != nil {
		return err
	}

	return nil
}

func unloadMrvDrv() {
	mfLibOpen = 0
	mfLibClose = 0
	mfBeginPainting = 0
	mfEndPainting = 0
}
