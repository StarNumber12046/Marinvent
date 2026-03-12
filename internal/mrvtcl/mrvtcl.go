//go:build windows && 386

package mrvtcl

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	tclLibInit           uintptr
	tclLibClose          uintptr
	tclOpen              uintptr
	tclClosePict         uintptr
	tclGetNumPictsInFile uintptr
	tclGetPictRect       uintptr
	tclDisplay           uintptr
	tclIsPictGeoRefd     uintptr
	tclGeoLatLon2XY      uintptr
	tclGeoXY2LatLon      uintptr
)

func TCL_LibInit(fontPath, lineStylePath, tclClassPath uintptr) int32 {
	if tclLibInit == 0 {
		return -1
	}
	ret, _, _ := syscall.Syscall6(tclLibInit, 4, fontPath, lineStylePath, tclClassPath, 0, 0, 0)
	return int32(ret)
}

func TCL_LibClose() int32 {
	if tclLibClose == 0 {
		return -1
	}
	ret, _, _ := syscall.Syscall(tclLibClose, 0, 0, 0, 0)
	return int32(ret)
}

func TCL_Open(filePath uintptr, pictIndex uint32, pictName uintptr, pictHandle *uintptr) int32 {
	if tclOpen == 0 {
		return -1
	}
	ret, _, _ := syscall.Syscall6(tclOpen, 4, filePath, uintptr(pictIndex), pictName, uintptr(unsafe.Pointer(pictHandle)), 0, 0)
	return int32(ret)
}

func TCL_ClosePict(pictHandle uintptr) int32 {
	if tclClosePict == 0 {
		return -1
	}
	ret, _, _ := syscall.Syscall(tclClosePict, 1, pictHandle, 0, 0)
	return int32(ret)
}

func TCL_GetNumPictsInFile(filePath uintptr, numPicts *uint32) int32 {
	if tclGetNumPictsInFile == 0 {
		return -1
	}
	ret, _, _ := syscall.Syscall(tclGetNumPictsInFile, 2, filePath, uintptr(unsafe.Pointer(numPicts)), 0)
	return int32(ret)
}

func TCL_GetPictRect(pictHandle uintptr, rect *RECT) int32 {
	if tclGetPictRect == 0 {
		return -1
	}
	ret, _, _ := syscall.Syscall(tclGetPictRect, 2, pictHandle, uintptr(unsafe.Pointer(rect)), 0)
	return int32(ret)
}

func TCL_Display(pictHandle, hdc uintptr, scaleX, scaleY float32, srcRect, offset uintptr, flags uint16) int32 {
	if tclDisplay == 0 {
		return -1
	}

	ret, _, _ := syscall.Syscall9(tclDisplay, 7,
		pictHandle,
		hdc,
		uintptr(*(*uint32)(unsafe.Pointer(&scaleX))),
		uintptr(*(*uint32)(unsafe.Pointer(&scaleY))),
		srcRect,
		offset,
		uintptr(flags),
		0, 0,
	)
	return int32(ret)
}

func TCL_IsPictGeoRefd(pictHandle uintptr) int32 {
	if tclIsPictGeoRefd == 0 {
		return -1
	}
	ret, _, _ := syscall.Syscall(tclIsPictGeoRefd, 1, pictHandle, 0, 0)
	return int32(ret)
}

func TCL_GeoLatLon2XY(pictHandle uintptr, lat, lon float64, x, y *int32) int32 {
	if tclGeoLatLon2XY == 0 {
		return -1
	}
	ret, _, _ := syscall.Syscall6(tclGeoLatLon2XY, 5,
		pictHandle,
		uintptr(*(*uint64)(unsafe.Pointer(&lat))),
		uintptr(*(*uint64)(unsafe.Pointer(&lon))),
		uintptr(unsafe.Pointer(x)),
		uintptr(unsafe.Pointer(y)),
		0,
	)
	return int32(ret)
}

func TCL_GeoXY2LatLon(pictHandle uintptr, x, y int32, lat, lon *float64) int32 {
	if tclGeoXY2LatLon == 0 {
		return -1
	}
	ret, _, _ := syscall.Syscall6(tclGeoXY2LatLon, 5,
		pictHandle,
		uintptr(x),
		uintptr(y),
		uintptr(unsafe.Pointer(lat)),
		uintptr(unsafe.Pointer(lon)),
		0,
	)
	return int32(ret)
}

func loadMrvTclProc(module uintptr, name string) (uintptr, error) {
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

func loadMrvTcl(module uintptr) error {
	var err error

	tclLibInit, err = loadMrvTclProc(module, "TCL_LibInit")
	if err != nil {
		return err
	}

	tclLibClose, err = loadMrvTclProc(module, "TCL_LibClose")
	if err != nil {
		return err
	}

	tclOpen, err = loadMrvTclProc(module, "TCL_Open")
	if err != nil {
		return err
	}

	tclClosePict, err = loadMrvTclProc(module, "TCL_ClosePict")
	if err != nil {
		return err
	}

	tclGetNumPictsInFile, err = loadMrvTclProc(module, "TCL_GetNumPictsInFile")
	if err != nil {
		return err
	}

	tclGetPictRect, err = loadMrvTclProc(module, "TCL_GetPictRect")
	if err != nil {
		return err
	}

	tclDisplay, err = loadMrvTclProc(module, "TCL_Display")
	if err != nil {
		return err
	}

	tclIsPictGeoRefd, err = loadMrvTclProc(module, "TCL_IsPictGeoRefd")
	if err != nil {
		return err
	}

	tclGeoLatLon2XY, err = loadMrvTclProc(module, "TCL_GeoLatLon2XY")
	if err != nil {
		return err
	}

	tclGeoXY2LatLon, err = loadMrvTclProc(module, "TCL_GeoXY2LatLon")
	if err != nil {
		return err
	}

	return nil
}

func unloadMrvTcl() {
	tclLibInit = 0
	tclLibClose = 0
	tclOpen = 0
	tclClosePict = 0
	tclGetNumPictsInFile = 0
	tclGetPictRect = 0
	tclDisplay = 0
	tclIsPictGeoRefd = 0
	tclGeoLatLon2XY = 0
	tclGeoXY2LatLon = 0
}
