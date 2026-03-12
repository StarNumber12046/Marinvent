package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"marinvent/internal/api"
	"marinvent/internal/mrvtcl"
)

func main() {
	port := flag.String("port", "8080", "API server port")
	host := flag.String("host", "0.0.0.0", "API server host")
	chartsDBF := flag.String("charts", "C:\\ProgramData\\Jeppesen\\Common\\TerminalCharts\\charts.dbf", "Path to charts.dbf")
	vfrChartsDBF := flag.String("vfrcharts", "C:\\ProgramData\\Jeppesen\\Common\\TerminalCharts\\vfrchrts.dbf", "Path to vfrchrts.dbf")
	typesDBF := flag.String("types", "C:\\ProgramData\\Jeppesen\\Common\\TerminalCharts\\ctypes.dbf", "Path to ctypes.dbf")
	airportsDBF := flag.String("airports", "C:\\ProgramData\\Jeppesen\\Common\\TerminalCharts\\Airports.dbf", "Path to Airports.dbf")
	tclDir := flag.String("tcls", "TCLs", "Directory containing TCL files")
	dllPath := flag.String("dll", "", "Path to DLL directory (default: auto-detect)")
	fontDir := flag.String("fonts", "C:\\ProgramData\\Jeppesen\\Common\\Fonts", "Path to Jeppesen fonts directory")
	skipDllInit := flag.Bool("skip-dll", false, "Skip DLL initialization (for non-Windows or 64-bit builds)")
	flag.Parse()

	if !*skipDllInit {
		log.Println("Loading Marinvent DLLs...")
		if err := mrvtcl.Load(*dllPath); err != nil {
			log.Printf("Warning: Failed to load DLLs: %v", err)
			log.Println("Chart export and georeferencing will not be available.")
			log.Println("To run with DLL support, build with: GOARCH=386 go build")
		} else {
			log.Println("DLLs loaded successfully.")

			log.Println("Initializing TCL library...")
			if err := mrvtcl.Init(*fontDir, *fontDir); err != nil {
				log.Printf("Warning: Failed to initialize TCL library: %v", err)
				log.Println("Chart export and georeferencing may not work correctly.")
			} else {
				log.Println("TCL library initialized successfully.")
			}

			defer mrvtcl.Close()
		}
	} else {
		log.Println("Skipping DLL initialization (--skip-dll flag)")
	}

	cfg := api.ServerConfig{
		Host:         *host,
		Port:         *port,
		ChartsDBF:    *chartsDBF,
		VFRChartsDBF: *vfrChartsDBF,
		TypesDBF:     *typesDBF,
		AirportsDBF:  *airportsDBF,
		TCLDir:       *tclDir,
		DLLPath:      *dllPath,
		FontDir:      *fontDir,
	}

	server, err := api.NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
