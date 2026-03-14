package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"marinvent/internal/charts"

	"github.com/gin-gonic/gin"
)

// @title Marinvent Chart API
// @version 1.0
// @description API for accessing and exporting Jeppesen terminal charts
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/marinvent/marivent
// @contact.email support@marinvent.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// ChartInfo is the API response model
type ChartInfo struct {
	Filename  string `json:"filename"`
	ICAO      string `json:"icao"`
	ChartType string `json:"chart_type"`
	TypeName  string `json:"type_name"`
	Category  string `json:"category"`
	ProcID    string `json:"proc_id"`
	DateEff   string `json:"date_eff"`
	SheetID   string `json:"sheet_id"`
	HasTCL    bool   `json:"has_tcl"`
	IsVFR     bool   `json:"is_vfr"`
}

// ChartList is the API response for listing charts
// @description Response containing list of charts for an ICAO
type ChartList struct {
	ICAO   string      `json:"icao"`
	Total  int         `json:"total"`
	Charts []ChartInfo `json:"charts"`
}

// ChartTypesResponse is the API response for listing chart types
type ChartTypesResponse struct {
	Total int         `json:"total"`
	Types []ChartType `json:"types"`
}

// ChartType is a chart type entry
type ChartType struct {
	Code     string `json:"code"`
	Category string `json:"category"`
	Type     string `json:"type"`
}

// Airport is the API response model for airport info
type Airport struct {
	ICAO              string   `json:"icao"`
	IATA              string   `json:"iata,omitempty"`
	Name              string   `json:"name"`
	City              string   `json:"city,omitempty"`
	State             string   `json:"state,omitempty"`
	CountryCode       string   `json:"country_code,omitempty"`
	Latitude          float64  `json:"latitude"`
	Longitude         float64  `json:"longitude"`
	MagneticVariation float64  `json:"magnetic_variation"`
	LongestRunwayFt   int      `json:"longest_runway_ft"`
	Timezone          string   `json:"timezone,omitempty"`
	AirportUse        string   `json:"airport_use"`
	Customs           string   `json:"customs,omitempty"`
	Beacon            bool     `json:"beacon"`
	JetStartUnit      bool     `json:"jet_start_unit"`
	Oxygen            []string `json:"oxygen,omitempty"`
	RepairTypes       []string `json:"repair_types,omitempty"`
	FuelTypes         []string `json:"fuel_types,omitempty"`
}

// Config holds API configuration
type Config struct {
	ChartsDBFPath string
	TypesDBFPath  string
	TCLDir        string
}

// Handler holds the catalog and config
type Handler struct {
	catalog *charts.Catalog
	config  *Config
}

// NewHandler creates a new API handler
func NewHandler(catalog *charts.Catalog, config *Config) *Handler {
	return &Handler{
		catalog: catalog,
		config:  config,
	}
}

// GetCharts returns charts for an ICAO
// @Summary List charts for ICAO
// @Description Returns all charts for a given ICAO airport. Can be filtered by type (code or name) and search query.
// @Tags charts
// @Accept json
// @Produce json
// @Param icao path string true "ICAO airport code (e.g., KJFK, EGLL)"
// @Param type query string false "Chart type - can be code (1L, AP) or name (RNAV, ILS, AIRPORT). Looks up in ctypes.dbf"
// @Param search query string false "Search text to filter by PROC_ID (procedure name)"
// @Param types query string false "Chart types to include - can be 'vfr', 'ifr', or 'vfr,ifr' (default: both)"
// @Success 200 {object} ChartList
// @Router /api/v1/charts/{icao} [get]
func (h *Handler) GetCharts(c *gin.Context) {
	start := time.Now()

	icao := c.Param("icao")
	typeQuery := c.Query("type")
	search := c.Query("search")
	types := c.Query("types")

	t := time.Now()
	results := h.catalog.Filter(icao, typeQuery, search, types)
	filterTime := time.Since(t)

	chartList := make([]ChartInfo, 0, len(results))
	for _, r := range results {
		chartList = append(chartList, ChartInfo{
			Filename:  r.Filename,
			ICAO:      r.ICAO,
			ChartType: r.ChartType,
			TypeName:  r.TypeName,
			Category:  r.Category,
			ProcID:    r.ProcID,
			DateEff:   r.DateEff,
			SheetID:   r.SheetID,
			HasTCL:    r.TCLPath != "",
			IsVFR:     r.IsVFR,
		})
	}

	c.JSON(http.StatusOK, ChartList{
		ICAO:   icao,
		Total:  len(chartList),
		Charts: chartList,
	})

	log.Printf("[PERF] GetCharts(%s): filter=%.1fms total=%.1fms results=%d",
		icao, float64(filterTime.Microseconds())/1000, float64(time.Since(start).Microseconds())/1000, len(chartList))
}

// GetChartTypes returns all available chart types
// @Summary List all chart types
// @Description Returns all available chart types from ctypes.dbf with codes, categories, and descriptions
// @Tags charts
// @Accept json
// @Produce json
// @Success 200 {object} ChartTypesResponse
// @Router /api/v1/chart-types [get]
func (h *Handler) GetChartTypes(c *gin.Context) {
	db := h.catalog.GetDBF()
	types := db.GetAllChartTypes()

	typeList := make([]ChartType, 0, len(types))
	for _, t := range types {
		typeList = append(typeList, ChartType{
			Code:     t.Code,
			Category: t.Category,
			Type:     t.Type,
		})
	}

	c.JSON(http.StatusOK, ChartTypesResponse{
		Total: len(typeList),
		Types: typeList,
	})
}

// GetAirport gets an airport by ICAO code
// @Summary Get airport by ICAO
// @Description Returns airport details for a given ICAO code
// @Tags airports
// @Accept json
// @Produce json
// @Param icao path string true "ICAO airport code (e.g., KJFK, EGLL)"
// @Success 200 {object} Airport
// @Router /api/v1/airports/{icao} [get]
func (h *Handler) GetAirport(c *gin.Context) {
	icao := c.Param("icao")
	icao = strings.ToUpper(icao)

	airport := h.catalog.GetAirport(icao)
	if airport == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "airport not found"})
		return
	}

	c.JSON(http.StatusOK, Airport{
		ICAO:              airport.ICAO,
		IATA:              airport.IATA,
		Name:              airport.Name,
		City:              airport.City,
		State:             airport.State,
		CountryCode:       airport.CountryCode,
		Latitude:          airport.Latitude,
		Longitude:         airport.Longitude,
		MagneticVariation: airport.MagneticVariation,
		LongestRunwayFt:   airport.LongestRunwayFt,
		Timezone:          airport.Timezone,
		AirportUse:        airport.AirportUse,
		Customs:           airport.Customs,
		Beacon:            airport.Beacon,
		JetStartUnit:      airport.JetStartUnit,
		Oxygen:            airport.Oxygen,
		RepairTypes:       airport.RepairTypes,
		FuelTypes:         airport.FuelTypes,
	})
}

// GetChartPDF exports a chart to PDF
// @Summary Export chart to PDF
// @Description Exports a specific chart to PDF format. Returns the chart as a PDF file. Post-processing is enabled by default to remove waypoint overlays; use ?no_postprocess=1 to disable.
// @Tags charts
// @Accept json
// @Produce application/pdf
// @Param icao path string true "ICAO airport code"
// @Param filename path string true "Chart filename (e.g., KJFK225)"
// @Param no_postprocess query int false "Set to 1 to disable post-processing (default: 0)"
// @Success 200 {file} pdf "PDF file containing the chart"
// @Router /api/v1/charts/{icao}/export/{filename} [get]
func (h *Handler) GetChartPDF(c *gin.Context) {
	start := time.Now()
	timings := make(map[string]time.Duration)
	tick := func(name string, t time.Time) time.Duration {
		d := time.Since(t)
		timings[name] = d
		return d
	}

	t := time.Now()
	icao := c.Param("icao")
	filename := c.Param("filename")
	noPostProcess := c.Query("no_postprocess") == "1"
	tick("params", t)

	t = time.Now()
	chart := h.catalog.GetChart(filename)
	tick("GetChart", t)

	if chart == nil || chart.ICAO != icao {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	if chart.TCLPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "TCL file not found"})
		return
	}

	t = time.Now()
	pdfBytes, err := h.catalog.ExportToPDF(chart.TCLPath, !noPostProcess)
	tick("ExportToPDF", t)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	t = time.Now()
	c.Header("Content-Disposition", "attachment; filename="+filename+".pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
	tick("response", t)

	timings["total"] = time.Since(start)

	logTimings(timings, filename)
}

func logTimings(timings map[string]time.Duration, filename string) {
	total := timings["total"]
	delete(timings, "total")

	parts := make([]string, 0, len(timings)+1)
	for name, d := range timings {
		parts = append(parts, fmt.Sprintf("%s=%.1fms", name, float64(d.Microseconds())/1000))
	}
	parts = append(parts, fmt.Sprintf("total=%.1fms", float64(total.Microseconds())/1000))

	log.Printf("[PERF] %s: %s", filename, strings.Join(parts, " "))
}

// GetHealth returns health check
// ChartDataResponse is the API response for getting chart data
type ChartDataResponse struct {
	Filename string `json:"filename"`
	ICAO     string `json:"icao"`
	Width    int32  `json:"width"`
	Height   int32  `json:"height"`
	HasTCL   bool   `json:"has_tcl"`
}

// GetChartData returns data for a single chart
// @Summary Get chart data
// @Description Returns chart data including dimensions and georeferencing status for a specific chart
// @Tags charts
// @Accept json
// @Produce json
// @Param icao path string true "ICAO airport code"
// @Param filename path string true "Chart filename (e.g., KJFK225)"
// @Success 200 {object} ChartDataResponse
// @Router /api/v1/charts/{icao}/data/{filename} [get]
func (h *Handler) GetChartData(c *gin.Context) {
	icao := c.Param("icao")
	filename := c.Param("filename")

	chart := h.catalog.GetChart(filename)
	if chart == nil || chart.ICAO != icao {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	response := ChartDataResponse{
		Filename: chart.Filename,
		ICAO:     chart.ICAO,
		Width:    0,
		Height:   0,
		HasTCL:   chart.TCLPath != "",
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Health check
// @Description Returns API health status and version
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "1.3.0",
	})
}
