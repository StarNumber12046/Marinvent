package api

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

import (
	"fmt"
	"log"
	"os"

	"marinvent/internal/charts"
	"marinvent/internal/dbf"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "marinvent/docs"
)

type Server struct {
	addr    string
	port    string
	router  *gin.Engine
	catalog *charts.Catalog
	config  ServerConfig
}

type ServerConfig struct {
	Host      string
	Port      string
	ChartsDBF string
	TypesDBF  string
	TCLDir    string
}

func NewServer(cfg ServerConfig) (*Server, error) {
	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	dbf, err := dbf.New(cfg.ChartsDBF, cfg.TypesDBF)
	if err != nil {
		return nil, fmt.Errorf("failed to load DBF files: %w", err)
	}

	catalog := charts.NewCatalog(dbf, cfg.TCLDir)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	handler := NewHandler(catalog, &Config{
		ChartsDBFPath: cfg.ChartsDBF,
		TypesDBFPath:  cfg.TypesDBF,
		TCLDir:        cfg.TCLDir,
	})

	registerRoutes(router, handler)

	return &Server{
		addr:    cfg.Host,
		port:    cfg.Port,
		router:  router,
		catalog: catalog,
		config:  cfg,
	}, nil
}

func registerRoutes(r *gin.Engine, h *Handler) {
	r.GET("/health", h.GetHealth)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		api.GET("/charts/:icao", h.GetCharts)
		api.GET("/charts/:icao/export/:filename", h.GetChartPDF)
		api.GET("/chart-types", h.GetChartTypes)
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.addr, s.port)
	log.Printf("Starting Marinvent API server on %s", addr)
	log.Printf("Charts DBF: %s", s.config.ChartsDBF)
	log.Printf("Types DBF: %s", s.config.TypesDBF)
	log.Printf("TCL Directory: %s", s.config.TCLDir)
	log.Printf("Swagger UI: http://%s/swagger/index.html", addr)
	return s.router.Run(addr)
}

func (s *Server) GetConfig() ServerConfig {
	return s.config
}

func LoadConfigFromEnv() ServerConfig {
	return ServerConfig{
		Host:      getEnv("HOST", "0.0.0.0"),
		Port:      getEnv("PORT", "8080"),
		ChartsDBF: getEnv("CHARTS_DBF", "C:\\ProgramData\\Jeppesen\\Common\\TerminalCharts\\charts.dbf"),
		TypesDBF:  getEnv("TYPES_DBF", "C:\\ProgramData\\Jeppesen\\Common\\TerminalCharts\\ctypes.dbf"),
		TCLDir:    getEnv("TCL_DIR", "TCLs"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
