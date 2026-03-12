//go:build windows && 386

package mrvtcl

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type timer struct {
	last    time.Time
	timings []string
}

func newTimer() *timer {
	return &timer{last: time.Now()}
}

func (t *timer) tick(name string) time.Duration {
	now := time.Now()
	d := now.Sub(t.last)
	t.last = now
	t.timings = append(t.timings, fmt.Sprintf("%s=%.0fms", name, float64(d.Milliseconds())))
	return d
}

func (t *timer) log(prefix string) {
	log.Printf("[TIMING] %s: %s", prefix, strings.Join(t.timings, " "))
}

func (c *Chart) ExportToEMF(emfPath string) error {
	chartMu.Lock()
	defer chartMu.Unlock()

	if c.handle == 0 {
		return ErrNotInitialized
	}

	tm := newTimer()

	width := c.bounds.Width
	height := c.bounds.Height
	tm.tick("bounds")

	emfRect := RECT{
		Left:   0,
		Top:    0,
		Right:  width * 100,
		Bottom: height * 100,
	}

	hdcMeta := CreateEnhMetaFileA(0, emfPath, &emfRect, "Marinvent TCL Chart\\0Chart\\0")
	if hdcMeta == 0 {
		return fmt.Errorf("CreateEnhMetaFile failed: %d", GetLastError())
	}
	tm.tick("CreateEnhMetaFile")

	SetMapMode(hdcMeta, MM_ANISOTROPIC)
	SetWindowExtEx(hdcMeta, width, height)
	SetViewportExtEx(hdcMeta, width, height)
	tm.tick("SetupDC")

	MF_BeginPainting(hdcMeta)
	tm.tick("MF_BeginPainting")

	result := TCL_Display(c.handle, hdcMeta, 1.0, 1.0, 0, 0, 0xFFFF)
	tm.tick("TCL_Display")

	MF_EndPainting(hdcMeta)
	tm.tick("MF_EndPainting")

	hEmf := CloseEnhMetaFile(hdcMeta)
	if hEmf != 0 {
		DeleteEnhMetaFile(hEmf)
	}
	tm.tick("CloseEnhMetaFile")

	tm.log("ExportToEMF")

	if result != 1 {
		return fmt.Errorf("TCL_Display failed: %d", result)
	}

	return nil
}

func (c *Chart) ExportToPDF(pdfPath string) error {
	chartMu.Lock()
	defer chartMu.Unlock()

	if c.handle == 0 {
		return ErrNotInitialized
	}

	tm := newTimer()
	log.Printf("[EXPORT] ExportToPDF: %s (chart size: %dx%d)", pdfPath, c.bounds.Width, c.bounds.Height)

	hdcPrinter, err := createPDFPrinterDC(pdfPath)
	if err != nil {
		return err
	}
	defer DeleteDC(hdcPrinter)
	tm.tick("CreateDC")

	width := c.bounds.Width
	height := c.bounds.Height

	pageWidth := GetDeviceCaps(hdcPrinter, HORZRES)
	pageHeight := GetDeviceCaps(hdcPrinter, VERTRES)
	tm.tick("GetDeviceCaps")
	log.Printf("[EXPORT] Page size: %dx%d", pageWidth, pageHeight)

	scaleX := float64(pageWidth) / float64(width)
	scaleY := float64(pageHeight) / float64(height)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	scaledWidth := int32(float64(width) * scale)
	scaledHeight := int32(float64(height) * scale)
	offsetX := (pageWidth - scaledWidth) / 2
	offsetY := (pageHeight - scaledHeight) / 2

	docName, _ := syscall.BytePtrFromString(filepath.Base(pdfPath))
	outputPath, _ := syscall.BytePtrFromString(pdfPath)

	docInfo := DOCINFOA{
		CbSize:      int32(unsafe.Sizeof(DOCINFOA{})),
		LpszDocName: docName,
		LpszOutput:  outputPath,
	}
	tm.tick("DocInfo")

	if StartDocA(hdcPrinter, &docInfo) <= 0 {
		err := GetLastError()
		log.Printf("[EXPORT] StartDoc failed: %d", err)
		return fmt.Errorf("StartDoc failed: %d", err)
	}
	tm.tick("StartDoc")

	if StartPage(hdcPrinter) <= 0 {
		AbortDoc(hdcPrinter)
		log.Printf("[EXPORT] StartPage failed")
		return errors.New("StartPage failed")
	}
	tm.tick("StartPage")

	SetMapMode(hdcPrinter, MM_ANISOTROPIC)
	SetWindowExtEx(hdcPrinter, width, height)
	SetViewportExtEx(hdcPrinter, scaledWidth, scaledHeight)
	SetViewportOrgEx(hdcPrinter, offsetX, offsetY)
	tm.tick("SetupDC")

	MF_BeginPainting(hdcPrinter)
	tm.tick("MF_BeginPainting")

	result := TCL_Display(c.handle, hdcPrinter, 1.0, 1.0, 0, 0, 0xFFFF)
	tm.tick("TCL_Display")

	MF_EndPainting(hdcPrinter)
	tm.tick("MF_EndPainting")

	EndPage(hdcPrinter)
	tm.tick("EndPage")

	EndDoc(hdcPrinter)
	tm.tick("EndDoc")

	tm.log("ExportToPDF")

	if result != 1 {
		return fmt.Errorf("TCL_Display failed: %d", result)
	}

	log.Printf("[EXPORT] PDF should be at: %s", pdfPath)
	return nil
}

func (c *Chart) ExportToPDFBytes() ([]byte, error) {
	tm := newTimer()

	tempDir, err := os.MkdirTemp("", "marinvent-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)
	tm.tick("MkdirTemp")

	pdfPath := filepath.Join(tempDir, "output.pdf")

	if err := c.ExportToPDF(pdfPath); err != nil {
		return nil, err
	}
	tm.tick("ExportToPDF")

	var pdfData []byte
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	pollStart := time.Now()
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout: PDF at %s never finalized (Spooler hang)", pdfPath)
		case <-ticker.C:
			info, err := os.Stat(pdfPath)
			if err != nil || info.Size() < 100 {
				continue
			}

			f, err := os.OpenFile(pdfPath, os.O_RDWR, 0)
			if err != nil {
				continue
			}

			buf := make([]byte, 1024)
			fileSize := info.Size()
			offset := fileSize - 1024
			if offset < 0 {
				offset = 0
			}

			_, readErr := f.ReadAt(buf, offset)
			f.Close()

			if readErr == nil || readErr == io.EOF {
				content := string(buf)
				if strings.Contains(content, "%%EOF") {
					pdfData, err = os.ReadFile(pdfPath)
					if err == nil && len(pdfData) > 0 {
						tm.tick(fmt.Sprintf("Poll(%.0fms)", float64(time.Since(pollStart).Milliseconds())))
						tm.tick("ReadFile")
						tm.log("ExportToPDFBytes")
						return pdfData, nil
					}
				}
			}
		}
	}
}

func createPDFPrinterDC(pdfPath string) (uintptr, error) {
	_ = pdfPath
	tm := newTimer()
	var needed, returned uint32

	EnumPrintersA(PRINTER_ENUM_LOCAL|PRINTER_ENUM_CONNECTIONS, nil, 2, nil, 0, &needed, &returned)
	tm.tick("EnumPrinters1")

	if needed == 0 {
		return 0, errors.New("no printers found (EnumPrinters returned 0 bytes needed)")
	}

	buffer := make([]byte, needed)
	if !EnumPrintersA(PRINTER_ENUM_LOCAL|PRINTER_ENUM_CONNECTIONS, nil, 2, buffer, needed, &needed, &returned) {
		return 0, fmt.Errorf("EnumPrinters failed: %d", GetLastError())
	}
	tm.tick("EnumPrinters2")

	if returned == 0 {
		return 0, errors.New("no printers found (returned count is 0)")
	}

	var printerName *byte
	printerInfoSlice := unsafe.Slice((*PRINTER_INFO_2A)(unsafe.Pointer(&buffer[0])), returned)

	for i := uint32(0); i < returned; i++ {
		name := printerInfoSlice[i].PPrinterName
		if name != nil {
			nameStr := nullTerminatedString((*[MAX_PATH]byte)(unsafe.Pointer(name))[:])
			if nameStr == "Microsoft Print to PDF" {
				printerName = name
				break
			}
		}
	}

	if printerName == nil {
		for i := uint32(0); i < returned; i++ {
			name := printerInfoSlice[i].PPrinterName
			if name != nil {
				nameStr := nullTerminatedString((*[MAX_PATH]byte)(unsafe.Pointer(name))[:])
				if len(nameStr) > 0 && contains(nameStr, "Print to PDF") {
					printerName = name
					break
				}
			}
		}
	}

	if printerName == nil {
		return 0, errors.New("'Microsoft Print to PDF' printer not found. Enable it in Windows Settings > Devices > Printers")
	}

	printerNameStr := nullTerminatedString((*[MAX_PATH]byte)(unsafe.Pointer(printerName))[:])
	tm.tick("FindPrinter")
	log.Printf("[EXPORT] Using printer: %s", printerNameStr)

	var hPrinter uintptr
	if !OpenPrinterA(printerName, &hPrinter, 0) {
		return 0, fmt.Errorf("OpenPrinter failed: %d", GetLastError())
	}
	defer ClosePrinter(hPrinter)
	tm.tick("OpenPrinter")

	hdc := CreateDCANull(printerName)
	if hdc == 0 {
		return 0, fmt.Errorf("CreateDC failed: %d", GetLastError())
	}
	tm.tick("CreateDC")

	tm.log("createPDFPrinterDC")
	return hdc, nil
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
