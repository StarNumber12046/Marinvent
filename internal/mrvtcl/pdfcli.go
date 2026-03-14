//go:build windows && 386

package mrvtcl

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ExportToPDFCLI(tclPath string, pdfPath string, pictIndex int) error {
	log.Printf("[PDF CLI] ExportToPDFCLI: tclPath=%s, pdfPath=%s, pictIndex=%d", tclPath, pdfPath, pictIndex)

	// Check if TCL file exists
	if _, err := os.Stat(tclPath); err != nil {
		return fmt.Errorf("TCL file not found: %s (error: %v)", tclPath, err)
	}

	// Get absolute paths to ensure they work from any directory
	absTclPath, err := filepath.Abs(tclPath)
	if err != nil {
		absTclPath = tclPath
	}
	absPdfPath, err := filepath.Abs(pdfPath)
	if err != nil {
		absPdfPath = pdfPath
	}

	// Use absolute path to tcl2emf.exe in the app directory
	exePath := "C:\\Users\\StarNumber\\Documents\\Marinvent\\tcl2emf.exe"
	cmd := exec.Command(exePath, absTclPath, absPdfPath, strconv.Itoa(pictIndex))

	// Set working directory to where tcl2emf.exe and fonts are located
	cmd.Dir = "C:\\Users\\StarNumber\\Documents\\Marinvent"

	log.Printf("[PDF CLI] Running: %s %s %s %d", exePath, absTclPath, absPdfPath, pictIndex)

	// Set a reasonable timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		log.Printf("[PDF CLI] Command finished, err=%v", err)
		if err != nil {
			return fmt.Errorf("tcl2emf.exe failed: %v", err)
		}
	case <-time.After(60 * time.Second):
		cmd.Process.Kill()
		return fmt.Errorf("tcl2emf.exe timed out after 60 seconds")
	}

	// Poll for the PDF file to be created and finalized
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	pollStart := time.Now()
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout: PDF at %s never finalized (Spooler hang)", absPdfPath)
		case <-ticker.C:
			info, err := os.Stat(absPdfPath)
			if err != nil || info.Size() < 100 {
				continue
			}

			// Try to open the file to check if it's ready
			f, err := os.OpenFile(absPdfPath, os.O_RDWR, 0)
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

			if readErr == nil {
				content := string(buf)
				if strings.Contains(content, "%%EOF") {
					log.Printf("[PDF CLI] PDF finalized after %v", time.Since(pollStart))

					// We don't have easy access to the chart bounds here, so we do it in ExportToPDFBytes instead!
					return nil
				}
			}
		}
	}
}

func cropPDFFileMediaBox(pdfPath string, width, height float64) error {
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return err
	}

	hPt := 792.0
	wPt := hPt * (width / height)
	if wPt > 612.0 {
		wPt = 612.0
		hPt = wPt * (height / width)
	}

	re := regexp.MustCompile(`/MediaBox\s*\[\s*[\d.]+\s+[\d.]+\s+[\d.]+\s+[\d.]+\s*\]`)
	replacement := fmt.Sprintf("/MediaBox [0 0 %.2f %.2f]", wPt, hPt)
	result := re.ReplaceAllString(string(data), replacement)

	re2 := regexp.MustCompile(`/CropBox\s+\[\s*[\d.]+\s+[\d.]+\s+[\d.]+\s+[\d.]+\s*\]`)
	if re2.MatchString(result) {
		replacement2 := fmt.Sprintf("/CropBox [0 0 %.2f %.2f]", wPt, hPt)
		result = re2.ReplaceAllString(result, replacement2)
	}

	log.Printf("[PDF CLI] Fixed PDF bounds to %.2fx%.2f", wPt, hPt)
	return os.WriteFile(pdfPath, []byte(result), 0644)
}

func ExportToEMFCLI(tclPath string, emfPath string, pictIndex int) error {
	log.Printf("[PDF CLI] ExportToEMFCLI: tclPath=%s, emfPath=%s, pictIndex=%d", tclPath, emfPath, pictIndex)

	// Check if TCL file exists
	if _, err := os.Stat(tclPath); err != nil {
		return fmt.Errorf("TCL file not found: %s (error: %v)", tclPath, err)
	}

	// Get absolute paths
	absTclPath, err := filepath.Abs(tclPath)
	if err != nil {
		absTclPath = tclPath
	}
	absEmfPath, err := filepath.Abs(emfPath)
	if err != nil {
		absEmfPath = emfPath
	}

	// Use absolute path to tcl2emf.exe
	exePath := "C:\\Users\\StarNumber\\Documents\\Marinvent\\tcl2emf.exe"
	cmd := exec.Command(exePath, absTclPath, absEmfPath, strconv.Itoa(pictIndex))
	cmd.Dir = "C:\\Users\\StarNumber\\Documents\\Marinvent"

	log.Printf("[PDF CLI] Running: %s %s %s %d", exePath, absTclPath, absEmfPath, pictIndex)

	output, err := cmd.CombinedOutput()
	result := string(output)
	log.Printf("[PDF CLI] Output: %s", result)

	if err != nil {
		// Check for error patterns in output
		if strings.Contains(result, "Error") || strings.Contains(result, "Failed") {
			return fmt.Errorf("tcl2emf.exe failed: %s", result)
		}
		return fmt.Errorf("tcl2emf.exe failed: %v", err)
	}

	// Check if the EMF was created
	if _, err := os.Stat(absEmfPath); err != nil {
		return fmt.Errorf("EMF file not created: %v", err)
	}

	return nil
}
