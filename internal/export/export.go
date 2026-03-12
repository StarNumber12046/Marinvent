package export

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"marinvent/internal/mrvtcl"
)

type Exporter struct {
	tclDir      string
	postProcess string
	mu          sync.Mutex
}

func NewExporter(tclDir string) *Exporter {
	return &Exporter{
		tclDir:      tclDir,
		postProcess: "pdf_fixup_threshold.py",
	}
}

func (e *Exporter) ExportToEMF(tclPath, emfPath string) error {
	if _, err := os.Stat(tclPath); err != nil {
		return fmt.Errorf("TCL file not found: %s", tclPath)
	}

	if !mrvtcl.IsInitialized() {
		return fmt.Errorf("mrvtcl library not initialized")
	}

	chart, err := mrvtcl.OpenChart(tclPath, 1)
	if err != nil {
		return fmt.Errorf("failed to open chart: %w", err)
	}
	defer chart.Close()

	return chart.ExportToEMF(emfPath)
}

func (e *Exporter) ExportToPDFBytes(tclPath string, postProcess bool) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var timings []string
	tick := func(name string, t time.Time) time.Time {
		now := time.Now()
		d := now.Sub(t).Milliseconds()
		timings = append(timings, fmt.Sprintf("%s=%dms", name, d))
		return now
	}

	t := time.Now()

	if !mrvtcl.IsInitialized() {
		return nil, fmt.Errorf("mrvtcl library not initialized")
	}
	t = tick("check", t)

	chart, err := mrvtcl.OpenChart(tclPath, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to open chart: %w", err)
	}
	defer chart.Close()
	t = tick("open_chart", t)

	pdfData, err := chart.ExportToPDFBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to export PDF: %w", err)
	}
	t = tick("export_pdf", t)

	if postProcess && len(pdfData) > 0 {
		tempDir, err := os.MkdirTemp("", "marinvent-*")
		if err == nil {
			pdfPath := filepath.Join(tempDir, "output.pdf")
			if os.WriteFile(pdfPath, pdfData, 0644) == nil {
				if e.runPostProcess(pdfPath) == nil {
					pdfData, _ = os.ReadFile(pdfPath)
				}
			}
			os.RemoveAll(tempDir)
		}
	}
	t = tick("postprocess", t)

	log.Printf("[EXPORT] %s: %s", filepath.Base(tclPath), strings.Join(timings, " "))
	return pdfData, nil
}

func (e *Exporter) ExportToPDF(tclPath, pdfPath string) error {
	if _, err := os.Stat(tclPath); err != nil {
		return fmt.Errorf("TCL file not found: %s", tclPath)
	}

	if filepath.Ext(pdfPath) != ".pdf" {
		pdfPath = pdfPath + ".pdf"
	}

	if !mrvtcl.IsInitialized() {
		return fmt.Errorf("mrvtcl library not initialized")
	}

	chart, err := mrvtcl.OpenChart(tclPath, 1)
	if err != nil {
		return fmt.Errorf("failed to open chart: %w", err)
	}
	defer chart.Close()

	return chart.ExportToPDF(pdfPath)
}

func (e *Exporter) runPostProcess(pdfPath string) error {
	ppPath := e.postProcess
	if !filepath.IsAbs(ppPath) {
		absPath, err := filepath.Abs(ppPath)
		if err != nil {
			return err
		}
		ppPath = absPath
	}

	if _, err := os.Stat(ppPath); err != nil {
		return fmt.Errorf("post-process script not found: %s", ppPath)
	}

	cmd := exec.Command("python", ppPath, pdfPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (e *Exporter) ExportAll(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	files, err := os.ReadDir(e.tclDir)
	if err != nil {
		return fmt.Errorf("failed to read TCL directory: %w", err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if len(name) > 4 && name[len(name)-4:] == ".tcl" {
			tclPath := filepath.Join(e.tclDir, name)
			pdfPath := filepath.Join(outputDir, name[:len(name)-4]+".pdf")

			fmt.Printf("Exporting %s -> %s\n", name, pdfPath)
			if err := e.ExportToPDF(tclPath, pdfPath); err != nil {
				fmt.Printf("  Error: %v\n", err)
			}
		}
	}

	return nil
}
