//go:build windows && 386

package mrvtcl

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"unsafe"
)

var (
	mu          sync.Mutex
	mrvDrvH     uintptr
	mrvTclH     uintptr
	initialized bool
	loadedFonts []string
)

var (
	ErrNotInitialized = errors.New("mrvtcl library not initialized")
	ErrAlreadyInit    = errors.New("mrvtcl library already initialized")
	ErrLoadFailed     = errors.New("failed to load DLLs")
)

func Load(dllPath string) error {
	mu.Lock()
	defer mu.Unlock()

	if mrvDrvH != 0 || mrvTclH != 0 {
		return ErrAlreadyInit
	}

	paths := []string{
		".",
		dllPath,
		JEPPVIEW_PATH,
	}

	var mrvdrvPath, mrvtclPath string
	for _, p := range paths {
		candidate := fmt.Sprintf("%s\\mrvdrv.dll", p)
		if _, err := os.Stat(candidate); err == nil {
			mrvdrvPath = candidate
			break
		}
	}
	for _, p := range paths {
		candidate := fmt.Sprintf("%s\\mrvtcl.dll", p)
		if _, err := os.Stat(candidate); err == nil {
			mrvtclPath = candidate
			break
		}
	}

	if mrvdrvPath == "" {
		return fmt.Errorf("mrvdrv.dll not found in PATH or %s", JEPPVIEW_PATH)
	}
	if mrvtclPath == "" {
		return fmt.Errorf("mrvtcl.dll not found in PATH or %s", JEPPVIEW_PATH)
	}

	mrvDrvH = LoadLibraryA(CString(mrvdrvPath))
	if mrvDrvH == 0 {
		return fmt.Errorf("failed to load mrvdrv.dll: %d", GetLastError())
	}

	mrvTclH = LoadLibraryA(CString(mrvtclPath))
	if mrvTclH == 0 {
		err := GetLastError()
		FreeLibrary(mrvDrvH)
		mrvDrvH = 0
		return fmt.Errorf("failed to load mrvtcl.dll: %d", err)
	}

	if err := loadMrvDrv(mrvDrvH); err != nil {
		FreeLibrary(mrvTclH)
		FreeLibrary(mrvDrvH)
		mrvDrvH = 0
		mrvTclH = 0
		return fmt.Errorf("failed to resolve mrvdrv functions: %w", err)
	}

	if err := loadMrvTcl(mrvTclH); err != nil {
		FreeLibrary(mrvTclH)
		FreeLibrary(mrvDrvH)
		mrvDrvH = 0
		mrvTclH = 0
		return fmt.Errorf("failed to resolve mrvtcl functions: %w", err)
	}

	return nil
}

func Init(fontDir, configDir string) error {
	mu.Lock()
	defer mu.Unlock()

	if mrvDrvH == 0 || mrvTclH == 0 {
		return ErrNotInitialized
	}

	if initialized {
		return nil
	}

	fontPath, lineStylePath, tclClassPath := resolveConfigPaths(fontDir, configDir)

	n, err := LoadFonts(fontDir)
	if err != nil {
		return fmt.Errorf("failed to load fonts: %w", err)
	}

	MF_LibOpen()

	fontPathPtr := uintptr(0)
	lineStylePathPtr := uintptr(0)
	tclClassPathPtr := uintptr(0)

	if fontPath != "" {
		fontPathPtr = uintptr(unsafe.Pointer(CString(fontPath)))
	}
	if lineStylePath != "" {
		lineStylePathPtr = uintptr(unsafe.Pointer(CString(lineStylePath)))
	}
	if tclClassPath != "" {
		tclClassPathPtr = uintptr(unsafe.Pointer(CString(tclClassPath)))
	}

	result := TCL_LibInit(fontPathPtr, lineStylePathPtr, tclClassPathPtr)
	if result != 1 {
		UnloadFonts()
		return fmt.Errorf("TCL_LibInit failed: %d (loaded %d fonts)", result, n)
	}

	initialized = true
	return nil
}

func Close() {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		TCL_LibClose()
		initialized = false
	}

	UnloadFonts()

	MF_LibClose()

	unloadMrvTcl()
	unloadMrvDrv()

	if mrvTclH != 0 {
		FreeLibrary(mrvTclH)
		mrvTclH = 0
	}
	if mrvDrvH != 0 {
		FreeLibrary(mrvDrvH)
		mrvDrvH = 0
	}
}

func IsLoaded() bool {
	mu.Lock()
	defer mu.Unlock()
	return mrvDrvH != 0 && mrvTclH != 0
}

func IsInitialized() bool {
	mu.Lock()
	defer mu.Unlock()
	return initialized
}

func resolveConfigPaths(fontDir, configDir string) (fontPath, lineStylePath, tclClassPath string) {
	dirs := []string{".", fontDir, configDir, JEPPESEN_FONTS_DIR}

	for _, d := range dirs {
		p := fmt.Sprintf("%s\\jeppesen.tfl", d)
		if _, err := os.Stat(p); err == nil {
			fontPath = p
			break
		}
	}

	for _, d := range dirs {
		p := fmt.Sprintf("%s\\jeppesen.tls", d)
		if _, err := os.Stat(p); err == nil {
			lineStylePath = p
			break
		}
	}

	for _, d := range dirs {
		p := fmt.Sprintf("%s\\lssdef.tcl", d)
		if _, err := os.Stat(p); err == nil {
			tclClassPath = p
			break
		}
	}

	return
}
