package charts

import (
	"marinvent/internal/dbf"
	"marinvent/internal/export"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ChartInfo struct {
	Filename  string
	ICAO      string
	ChartType string
	TypeName  string
	Category  string
	ProcID    string
	DateEff   string
	SheetID   string
	TCLPath   string
}

type Catalog struct {
	db         *dbf.DBF
	tclDir     string
	chartCache map[string]*ChartInfo
	mu         sync.RWMutex
}

func NewCatalog(dbf *dbf.DBF, tclDir string) *Catalog {
	return &Catalog{
		db:         dbf,
		tclDir:     tclDir,
		chartCache: make(map[string]*ChartInfo),
	}
}

func (c *Catalog) buildCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.chartCache) > 0 {
		return
	}

	available := c.getAvailableTCLs()
	allCharts := c.db.GetAllCharts()
	for _, chart := range allCharts {
		tclPath := ""

		filename := strings.TrimSpace(chart.Filename)
		filenameLower := strings.ToLower(filename)

		if path, ok := available[filename]; ok {
			tclPath = path
		} else if path, ok := available[filenameLower]; ok {
			tclPath = path
		} else if path, ok := available[filenameLower+".tcl"]; ok {
			tclPath = path
		} else {
			for k, v := range available {
				if strings.TrimSpace(strings.ToLower(k)) == filenameLower ||
					strings.TrimSpace(strings.ToLower(strings.TrimSuffix(k, ".tcl"))) == filenameLower {
					tclPath = v
					break
				}
			}
		}

		chartType := c.db.GetChartType(chart.ChartType)
		info := &ChartInfo{
			Filename:  filename,
			ICAO:      strings.TrimSpace(chart.ICAO),
			ChartType: strings.TrimSpace(chart.ChartType),
			ProcID:    strings.TrimSpace(chart.ProcID),
			DateEff:   strings.TrimSpace(chart.DateEff),
			SheetID:   strings.TrimSpace(chart.SheetID),
			TCLPath:   tclPath,
		}
		if chartType != nil {
			info.TypeName = strings.TrimSpace(chartType.Type)
			info.Category = strings.TrimSpace(chartType.Category)
		}
		c.chartCache[info.Filename] = info
	}
}

func (c *Catalog) getAvailableTCLs() map[string]string {
	result := make(map[string]string)
	files, err := os.ReadDir(c.tclDir)
	if err != nil {
		return result
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := strings.ToUpper(f.Name())
		if strings.HasSuffix(name, ".TCL") {
			fullPath := filepath.Join(c.tclDir, f.Name())
			result[strings.TrimSuffix(f.Name(), ".tcl")+strings.Repeat(" ", 25-len(f.Name()))] = fullPath
			result[strings.ToLower(f.Name())] = fullPath
			result[f.Name()] = fullPath
		}
	}
	return result
}

func (c *Catalog) GetChart(filename string) *ChartInfo {
	c.mu.RLock()
	if len(c.chartCache) == 0 {
		c.mu.RUnlock()
		c.buildCache()
		c.mu.RLock()
	}
	defer c.mu.RUnlock()
	return c.chartCache[filename]
}

func (c *Catalog) GetAllCharts() []*ChartInfo {
	c.mu.RLock()
	if len(c.chartCache) == 0 {
		c.mu.RUnlock()
		c.buildCache()
		c.mu.RLock()
	}
	defer c.mu.RUnlock()
	charts := make([]*ChartInfo, 0, len(c.chartCache))
	for _, info := range c.chartCache {
		charts = append(charts, info)
	}
	return charts
}

func (c *Catalog) Search(query string) []*ChartInfo {
	c.mu.RLock()
	if len(c.chartCache) == 0 {
		c.mu.RUnlock()
		c.buildCache()
		c.mu.RLock()
	}
	defer c.mu.RUnlock()
	query = strings.ToUpper(query)
	var results []*ChartInfo
	for _, info := range c.chartCache {
		if strings.Contains(info.Filename, query) ||
			strings.Contains(info.ICAO, query) ||
			strings.Contains(info.ProcID, query) ||
			strings.Contains(info.TypeName, query) {
			results = append(results, info)
		}
	}
	return results
}

func (c *Catalog) FilterByICAO(icao string) []*ChartInfo {
	c.mu.RLock()
	if len(c.chartCache) == 0 {
		c.mu.RUnlock()
		c.buildCache()
		c.mu.RLock()
	}
	defer c.mu.RUnlock()
	icao = strings.ToUpper(icao)
	var results []*ChartInfo
	for _, info := range c.chartCache {
		if info.ICAO == icao {
			results = append(results, info)
		}
	}
	return results
}

func (c *Catalog) FilterByCategory(category string) []*ChartInfo {
	c.mu.RLock()
	if len(c.chartCache) == 0 {
		c.mu.RUnlock()
		c.buildCache()
		c.mu.RLock()
	}
	defer c.mu.RUnlock()
	category = strings.ToUpper(category)
	var results []*ChartInfo
	for _, info := range c.chartCache {
		if strings.Contains(strings.ToUpper(info.Category), category) {
			results = append(results, info)
		}
	}
	return results
}

func (c *Catalog) FilterByChartType(chartType string) []*ChartInfo {
	c.mu.RLock()
	if len(c.chartCache) == 0 {
		c.mu.RUnlock()
		c.buildCache()
		c.mu.RLock()
	}
	defer c.mu.RUnlock()
	var results []*ChartInfo
	for _, info := range c.chartCache {
		if info.ChartType == chartType {
			results = append(results, info)
		}
	}
	return results
}

func (c *Catalog) FilterByTypeName(typeName string) []*ChartInfo {
	c.mu.RLock()
	if len(c.chartCache) == 0 {
		c.mu.RUnlock()
		c.buildCache()
		c.mu.RLock()
	}
	defer c.mu.RUnlock()

	codes := c.db.ResolveChartTypes(typeName)
	if len(codes) == 0 {
		return nil
	}

	codeSet := make(map[string]bool)
	for _, code := range codes {
		codeSet[code] = true
	}

	var results []*ChartInfo
	for _, info := range c.chartCache {
		if codeSet[info.ChartType] {
			results = append(results, info)
		}
	}
	return results
}

func (c *Catalog) Filter(icao, typeName, search string) []*ChartInfo {
	c.mu.RLock()
	if len(c.chartCache) == 0 {
		c.mu.RUnlock()
		c.buildCache()
		c.mu.RLock()
	}
	defer c.mu.RUnlock()

	var candidates []*ChartInfo

	if icao != "" {
		icao = strings.ToUpper(icao)
		for _, info := range c.chartCache {
			if info.ICAO == icao {
				candidates = append(candidates, info)
			}
		}
	} else {
		for _, info := range c.chartCache {
			candidates = append(candidates, info)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	if typeName != "" {
		codes := c.db.ResolveChartTypes(typeName)
		if len(codes) > 0 {
			codeSet := make(map[string]bool)
			for _, code := range codes {
				codeSet[code] = true
			}
			var filtered []*ChartInfo
			for _, info := range candidates {
				if codeSet[info.ChartType] {
					filtered = append(filtered, info)
				}
			}
			candidates = filtered
		}
	}

	if search != "" {
		search = strings.ToUpper(search)
		var filtered []*ChartInfo
		for _, info := range candidates {
			if strings.Contains(strings.ToUpper(info.ProcID), search) ||
				strings.Contains(strings.ToUpper(info.Filename), search) ||
				strings.Contains(strings.ToUpper(info.TypeName), search) {
				filtered = append(filtered, info)
			}
		}
		candidates = filtered
	}

	return candidates
}

func (c *Catalog) NumCharts() int {
	c.mu.RLock()
	if len(c.chartCache) == 0 {
		c.mu.RUnlock()
		c.buildCache()
		c.mu.RLock()
	}
	defer c.mu.RUnlock()
	return len(c.chartCache)
}

func (c *Catalog) GetDBF() *dbf.DBF {
	return c.db
}

func (c *Catalog) ExportToPDF(tclPath string, postProcess bool) ([]byte, error) {
	exporter := export.NewExporter(c.tclDir)
	return exporter.ExportToPDFBytes(tclPath, postProcess)
}
