//go:build windows && 386

package mrvtcl

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	winspool = syscall.NewLazyDLL("winspool.drv")

	procLoadLibraryA     = kernel32.NewProc("LoadLibraryA")
	procFreeLibrary      = kernel32.NewProc("FreeLibrary")
	procGetProcAddress   = kernel32.NewProc("GetProcAddress")
	procGetFullPathNameA = kernel32.NewProc("GetFullPathNameA")
	procGetLastError     = kernel32.NewProc("GetLastError")

	procAddFontResourceExA    = gdi32.NewProc("AddFontResourceExA")
	procRemoveFontResourceExA = gdi32.NewProc("RemoveFontResourceExA")

	procCreateEnhMetaFileA = gdi32.NewProc("CreateEnhMetaFileA")
	procCloseEnhMetaFile   = gdi32.NewProc("CloseEnhMetaFile")
	procDeleteEnhMetaFile  = gdi32.NewProc("DeleteEnhMetaFile")
	procGetDeviceCaps      = gdi32.NewProc("GetDeviceCaps")
	procSetMapMode         = gdi32.NewProc("SetMapMode")
	procSetWindowExtEx     = gdi32.NewProc("SetWindowExtEx")
	procSetViewportExtEx   = gdi32.NewProc("SetViewportExtEx")
	procSetViewportOrgEx   = gdi32.NewProc("SetViewportOrgEx")
	procDeleteDC           = gdi32.NewProc("DeleteDC")

	procCreateDCA = gdi32.NewProc("CreateDCA")
	procStartDocA = gdi32.NewProc("StartDocA")
	procEndDoc    = gdi32.NewProc("EndDoc")
	procStartPage = gdi32.NewProc("StartPage")
	procEndPage   = gdi32.NewProc("EndPage")
	procAbortDoc  = gdi32.NewProc("AbortDoc")

	procEnumPrintersA = winspool.NewProc("EnumPrintersA")
	procOpenPrinterA  = winspool.NewProc("OpenPrinterA")
	procClosePrinter  = winspool.NewProc("ClosePrinter")

	procFindFirstFileA = kernel32.NewProc("FindFirstFileA")
	procFindNextFileA  = kernel32.NewProc("FindNextFileA")
	procFindClose      = kernel32.NewProc("FindClose")

	procSendMessageA = user32.NewProc("SendMessageA")
)

const (
	FR_PRIVATE                = 0x10
	WM_FONTCHANGE             = 0x001D
	HWND_BROADCAST            = 0xFFFF
	PRINTER_ENUM_LOCAL        = 0x00000002
	PRINTER_ENUM_CONNECTIONS  = 0x00000004
	ERROR_INSUFFICIENT_BUFFER = 122
	MM_ANISOTROPIC            = 8
	LOGPIXELSX                = 88
	HORZRES                   = 8
	VERTRES                   = 10
	MAX_PATH                  = 260
	FILE_ATTRIBUTE_DIRECTORY  = 0x00000010
	INVALID_HANDLE_VALUE      = ^uintptr(0)
)

type DOCINFOA struct {
	CbSize       int32
	LpszDocName  *byte
	LpszOutput   *byte
	LpszDatatype *byte
	FwType       uint32
}

type PRINTER_INFO_2A struct {
	PServerName         *byte
	PPrinterName        *byte
	PShareName          *byte
	PPortName           *byte
	PDriverName         *byte
	PComment            *byte
	PLocation           *byte
	PDevMode            uintptr
	PSepFile            *byte
	PPrintProcessor     *byte
	PDatatype           *byte
	PParameters         *byte
	PSecurityDescriptor uintptr
	Attributes          uint32
	Priority            uint32
	DefaultPriority     uint32
	StartTime           uint32
	UntilTime           uint32
	Status              uint32
	CJobs               uint32
	AveragePPM          uint32
}

type WIN32_FIND_DATAA struct {
	FileAttributes    uint32
	CreationTime      [8]byte
	LastAccessTime    [8]byte
	LastWriteTime     [8]byte
	FileSizeHigh      uint32
	FileSizeLow       uint32
	Reserved0         uint32
	Reserved1         uint32
	FileName          [MAX_PATH]byte
	AlternateFileName [14]byte
}

func LoadLibraryA(name *byte) uintptr {
	ret, _, _ := procLoadLibraryA.Call(uintptr(unsafe.Pointer(name)))
	return ret
}

func FreeLibrary(module uintptr) bool {
	ret, _, _ := procFreeLibrary.Call(module)
	return ret != 0
}

func GetProcAddress(module uintptr, name *byte) uintptr {
	ret, _, _ := procGetProcAddress.Call(module, uintptr(unsafe.Pointer(name)))
	return ret
}

func GetLastError() uint32 {
	ret, _, _ := procGetLastError.Call()
	return uint32(ret)
}

func GetFullPathNameA(fileName string, buffer []byte) (string, error) {
	var filePart *byte
	fn, _ := syscall.BytePtrFromString(fileName)
	ret, _, _ := procGetFullPathNameA.Call(
		uintptr(unsafe.Pointer(fn)),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&filePart)),
	)
	if ret == 0 {
		return "", errors.New("GetFullPathNameA failed")
	}
	n := ret
	if n > uintptr(len(buffer)) {
		n = uintptr(len(buffer))
	}
	return string(buffer[:n]), nil
}

func AddFontResourceExA(filename string, flags uint32, reserved uintptr) int {
	fn, _ := syscall.BytePtrFromString(filename)
	ret, _, _ := procAddFontResourceExA.Call(
		uintptr(unsafe.Pointer(fn)),
		uintptr(flags),
		reserved,
	)
	return int(ret)
}

func RemoveFontResourceExA(filename string, flags uint32, reserved uintptr) bool {
	fn, _ := syscall.BytePtrFromString(filename)
	ret, _, _ := procRemoveFontResourceExA.Call(
		uintptr(unsafe.Pointer(fn)),
		uintptr(flags),
		reserved,
	)
	return ret != 0
}

func BroadcastFontChange() {
	procSendMessageA.Call(HWND_BROADCAST, WM_FONTCHANGE, 0, 0)
}

func CreateEnhMetaFileA(hdcRef uintptr, filename string, bounds *RECT, description string) uintptr {
	fn, _ := syscall.BytePtrFromString(filename)
	desc, _ := syscall.BytePtrFromString(description)
	ret, _, _ := procCreateEnhMetaFileA.Call(
		hdcRef,
		uintptr(unsafe.Pointer(fn)),
		uintptr(unsafe.Pointer(bounds)),
		uintptr(unsafe.Pointer(desc)),
	)
	return ret
}

func CloseEnhMetaFile(hdc uintptr) uintptr {
	ret, _, _ := procCloseEnhMetaFile.Call(hdc)
	return ret
}

func DeleteEnhMetaFile(hemf uintptr) bool {
	ret, _, _ := procDeleteEnhMetaFile.Call(hemf)
	return ret != 0
}

func GetDeviceCaps(hdc uintptr, index int32) int32 {
	ret, _, _ := procGetDeviceCaps.Call(hdc, uintptr(index))
	return int32(ret)
}

func SetMapMode(hdc uintptr, mode int32) int32 {
	ret, _, _ := procSetMapMode.Call(hdc, uintptr(mode))
	return int32(ret)
}

func SetWindowExtEx(hdc uintptr, x, y int32) bool {
	ret, _, _ := procSetWindowExtEx.Call(hdc, uintptr(x), uintptr(y), 0)
	return ret != 0
}

func SetViewportExtEx(hdc uintptr, x, y int32) bool {
	ret, _, _ := procSetViewportExtEx.Call(hdc, uintptr(x), uintptr(y), 0)
	return ret != 0
}

func SetViewportOrgEx(hdc uintptr, x, y int32) bool {
	ret, _, _ := procSetViewportOrgEx.Call(hdc, uintptr(x), uintptr(y), 0)
	return ret != 0
}

func DeleteDC(hdc uintptr) bool {
	ret, _, _ := procDeleteDC.Call(hdc)
	return ret != 0
}

func CreateDCA(device, driver, output string, devMode uintptr) uintptr {
	var d, dr, o *byte
	if device != "" {
		d, _ = syscall.BytePtrFromString(device)
	}
	if driver != "" {
		dr, _ = syscall.BytePtrFromString(driver)
	}
	if output != "" {
		o, _ = syscall.BytePtrFromString(output)
	}
	ret, _, _ := procCreateDCA.Call(
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(dr)),
		uintptr(unsafe.Pointer(o)),
		devMode,
	)
	return ret
}

func CreateDCANull(printerName *byte) uintptr {
	ret, _, _ := procCreateDCA.Call(
		0,
		uintptr(unsafe.Pointer(printerName)),
		0,
		0,
	)
	return ret
}

func StartDocA(hdc uintptr, docInfo *DOCINFOA) int32 {
	ret, _, _ := procStartDocA.Call(hdc, uintptr(unsafe.Pointer(docInfo)))
	return int32(ret)
}

func EndDoc(hdc uintptr) int32 {
	ret, _, _ := procEndDoc.Call(hdc)
	return int32(ret)
}

func StartPage(hdc uintptr) int32 {
	ret, _, _ := procStartPage.Call(hdc)
	return int32(ret)
}

func EndPage(hdc uintptr) int32 {
	ret, _, _ := procEndPage.Call(hdc)
	return int32(ret)
}

func AbortDoc(hdc uintptr) int32 {
	ret, _, _ := procAbortDoc.Call(hdc)
	return int32(ret)
}

func EnumPrintersA(flags uint32, name *byte, level uint32, pPrinterEnum []byte, cbBuf uint32, pcbNeeded, pcReturned *uint32) bool {
	var pPrinterEnumPtr uintptr
	if len(pPrinterEnum) > 0 {
		pPrinterEnumPtr = uintptr(unsafe.Pointer(&pPrinterEnum[0]))
	}
	ret, _, _ := procEnumPrintersA.Call(
		uintptr(flags),
		uintptr(unsafe.Pointer(name)),
		uintptr(level),
		pPrinterEnumPtr,
		uintptr(cbBuf),
		uintptr(unsafe.Pointer(pcbNeeded)),
		uintptr(unsafe.Pointer(pcReturned)),
	)
	return ret != 0
}

func OpenPrinterA(printerName *byte, phPrinter *uintptr, pDefault uintptr) bool {
	ret, _, _ := procOpenPrinterA.Call(
		uintptr(unsafe.Pointer(printerName)),
		uintptr(unsafe.Pointer(phPrinter)),
		pDefault,
	)
	return ret != 0
}

func ClosePrinter(hPrinter uintptr) bool {
	ret, _, _ := procClosePrinter.Call(hPrinter)
	return ret != 0
}

func FindFirstFileA(pattern string, findData *WIN32_FIND_DATAA) uintptr {
	p, _ := syscall.BytePtrFromString(pattern)
	ret, _, _ := procFindFirstFileA.Call(
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(findData)),
	)
	return ret
}

func FindNextFileA(handle uintptr, findData *WIN32_FIND_DATAA) bool {
	ret, _, _ := procFindNextFileA.Call(handle, uintptr(unsafe.Pointer(findData)))
	return ret != 0
}

func FindClose(handle uintptr) bool {
	ret, _, _ := procFindClose.Call(handle)
	return ret != 0
}

func CString(s string) *byte {
	b, _ := syscall.BytePtrFromString(s)
	return b
}
