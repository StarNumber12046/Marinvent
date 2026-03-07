package api

import (
	"net/http"

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
// @Success 200 {object} ChartList
// @Router /api/v1/charts/{icao} [get]
func (h *Handler) GetCharts(c *gin.Context) {
	icao := c.Param("icao")
	typeQuery := c.Query("type")
	search := c.Query("search")

	results := h.catalog.Filter(icao, typeQuery, search)

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
		})
	}

	c.JSON(http.StatusOK, ChartList{
		ICAO:   icao,
		Total:  len(chartList),
		Charts: chartList,
	})
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

// GetChartPDF exports a chart to PDF
// @Summary Export chart to PDF
// @Description Exports a specific chart to PDF format. Returns the chart as a PDF file.
// @Tags charts
// @Accept json
// @Produce application/pdf
// @Param icao path string true "ICAO airport code"
// @Param filename path string true "Chart filename (e.g., KJFK225)"
// @Success 200 {file} pdf "PDF file containing the chart"
// @Router /api/v1/charts/{icao}/export/{filename} [get]
func (h *Handler) GetChartPDF(c *gin.Context) {
	icao := c.Param("icao")
	filename := c.Param("filename")

	chart := h.catalog.GetChart(filename)
	if chart == nil || chart.ICAO != icao {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	if chart.TCLPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "TCL file not found"})
		return
	}

	pdfBytes, err := h.catalog.ExportToPDF(chart.TCLPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename+".pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// GetHealth returns health check
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
		"version": "1.0.0",
	})
}
