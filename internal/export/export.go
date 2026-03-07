package export

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Exporter struct {
	tclDir  string
	emfTool string
}

func NewExporter(tclDir string) *Exporter {
	return &Exporter{
		tclDir:  tclDir,
		emfTool: "tcl2emf.exe",
	}
}

func (e *Exporter) getToolPath() string {
	exePath := e.emfTool
	if !filepath.IsAbs(exePath) {
		absPath, err := filepath.Abs(exePath)
		if err != nil {
			return exePath
		}
		return absPath
	}
	return exePath
}

func (e *Exporter) ExportToEMF(tclPath, emfPath string) error {
	if _, err := os.Stat(tclPath); err != nil {
		return fmt.Errorf("TCL file not found: %s", tclPath)
	}

	cmd := exec.Command(e.getToolPath(), tclPath, emfPath)
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("export failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (e *Exporter) ExportToPDFBytes(tclPath string) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "marinvent-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	pdfPath := filepath.Join(tempDir, "output.pdf")

	toolPath := e.getToolPath()
	workDir := filepath.Dir(toolPath)
	if workDir == "." {
		absPath, err := filepath.Abs(".")
		if err == nil {
			workDir = absPath
		}
	}

	cmd := exec.Command(toolPath, tclPath, pdfPath)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("tcl2emf failed: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	if _, err := os.Stat(pdfPath); err != nil {
		return nil, fmt.Errorf("PDF file not created: %w", err)
	}

	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	return pdfData, nil
}

func (e *Exporter) ExportToPDF(tclPath, pdfPath string) error {
	if _, err := os.Stat(tclPath); err != nil {
		return fmt.Errorf("TCL file not found: %s", tclPath)
	}

	if filepath.Ext(pdfPath) != ".pdf" {
		pdfPath = pdfPath + ".pdf"
	}

	cmd := exec.Command(e.getToolPath(), tclPath, pdfPath)
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("export failed: %w\nOutput: %s", err, string(output))
	}

	return nil
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
