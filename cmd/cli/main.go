package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"marinvent/internal/charts"
	"marinvent/internal/dbf"
	"marinvent/internal/export"
)

var (
	JeppDataPath = `C:\ProgramData\Jeppesen\Common\TerminalCharts`
)

func main() {
	chartsPath := flag.String("charts", filepath.Join(JeppDataPath, "charts.dbf"), "Path to charts.dbf")
	ctypesPath := flag.String("types", filepath.Join(JeppDataPath, "ctypes.dbf"), "Path to ctypes.dbf")
	tclDir := flag.String("tcls", "TCLs", "Directory containing TCL files")
	search := flag.String("search", "", "Search query (matches ICAO, filename, proc ID)")
	icao := flag.String("icao", "", "Filter by ICAO code")
	category := flag.String("category", "", "Filter by category (e.g., APPROACH, AIRPORT)")
	chartType := flag.String("type", "", "Filter by chart type code (e.g., AP, 1L)")
	listTypes := flag.Bool("list-types", false, "List all chart types")
	listICAOs := flag.Bool("list-icaos", false, "List all ICAO codes")
	exportPath := flag.String("export", "", "Export matching charts to EMF")
	quiet := flag.Bool("q", false, "Quiet mode (less verbose output)")
	flag.Parse()

	dbf, err := dbf.New(*chartsPath, *ctypesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading DBF files: %v\n", err)
		os.Exit(1)
	}

	cat := charts.NewCatalog(dbf, *tclDir)

	if *listTypes {
		listAllChartTypes(dbf)
		return
	}

	if *listICAOs {
		listAllICAOs(cat)
		return
	}

	var results []*charts.ChartInfo

	if *search != "" {
		results = cat.Search(*search)
	} else if *icao != "" {
		results = cat.FilterByICAO(*icao)
	} else if *category != "" {
		results = cat.FilterByCategory(*category)
	} else if *chartType != "" {
		results = cat.FilterByChartType(*chartType)
	} else {
		results = cat.GetAllCharts()
	}

	if len(results) == 0 {
		fmt.Println("No charts found matching criteria.")
		return
	}

	if !*quiet {
		fmt.Printf("Found %d charts:\n\n", len(results))
	}

	for _, c := range results {
		if *quiet {
			fmt.Printf("%s\t%s\t%s\t%s\n", c.Filename, c.ICAO, c.ChartType, c.TypeName)
		} else {
			fmt.Printf("  %s\n", c.Filename)
			fmt.Printf("    ICAO: %s  Type: %s (%s)\n", c.ICAO, c.ChartType, c.TypeName)
			fmt.Printf("    Category: %s\n", c.Category)
			fmt.Printf("    Effective: %s\n", c.DateEff)
			if *exportPath != "" {
				fmt.Printf("    Path: %s\n", c.TCLPath)
			}
			fmt.Println()
		}
	}

	if *exportPath != "" && len(results) > 0 {
		exp := export.NewExporter(*tclDir)

		os.MkdirAll(*exportPath, 0755)

		exported := 0
		failed := 0
		for _, c := range results {
			if c.TCLPath == "" {
				fmt.Printf("  Skipping %s (no TCL file found)\n", c.Filename)
				continue
			}

			ext := filepath.Ext(c.Filename)
			baseName := strings.TrimSuffix(c.Filename, ext)
			outPath := filepath.Join(*exportPath, baseName+".emf")

			fmt.Printf("Exporting %s -> %s\n", c.Filename, outPath)

			if err := exp.ExportToEMF(c.TCLPath, outPath); err != nil {
				fmt.Printf("  Error: %v\n", err)
				failed++
			} else {
				exported++
			}
		}

		fmt.Printf("\nExported %d charts, %d failed\n", exported, failed)
	}
}

func listAllChartTypes(d *dbf.DBF) {
	types := d.GetAllChartTypes()
	fmt.Printf("Chart Types (%d):\n\n", len(types))
	fmt.Printf("%-6s %-20s %s\n", "Code", "Category", "Type")
	fmt.Println(strings.Repeat("-", 80))
	for _, t := range types {
		fmt.Printf("%-6s %-20s %s\n", t.Code, t.Category, t.Type)
	}
}

func listAllICAOs(c *charts.Catalog) {
	seen := make(map[string]bool)
	icaos := []string{}

	for _, c := range c.GetAllCharts() {
		if !seen[c.ICAO] {
			seen[c.ICAO] = true
			icaos = append(icaos, c.ICAO)
		}
	}

	fmt.Printf("ICAO Codes (%d):\n\n", len(icaos))
	for i, icao := range icaos {
		if i > 0 && i%8 == 0 {
			fmt.Println()
		}
		fmt.Printf("%-6s ", icao)
	}
	fmt.Println()
}
