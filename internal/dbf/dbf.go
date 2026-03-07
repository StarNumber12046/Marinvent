package dbf

import (
	"github.com/Bowbaq/dbf"
)

type Chart struct {
	ICAO      string
	Filename  string
	ChartType string
	IndexNo   string
	ProcID    string
	Action    string
	DateRev   string
	DateEff   string
	TrimSize  string
	GeoRef    string
	SheetID   string
	FtBk      string
}

type ChartType struct {
	Code      string
	Category  string
	Type      string
	Precision string
}

type DBF struct {
	charts   *dbf.DbfTable
	ctypes   *dbf.DbfTable
	chartMap map[string]*Chart
	typeMap  map[string]*ChartType
}

func New(chartsPath, ctypesPath string) (*DBF, error) {
	charts, err := dbf.LoadFile(chartsPath)
	if err != nil {
		return nil, err
	}

	ctypes, err := dbf.LoadFile(ctypesPath)
	if err != nil {
		return nil, err
	}

	d := &DBF{
		charts:   charts,
		ctypes:   ctypes,
		chartMap: make(map[string]*Chart),
		typeMap:  make(map[string]*ChartType),
	}

	d.buildMaps()
	return d, nil
}

func (d *DBF) buildMaps() {
	iter := d.charts.NewIterator()
	for iter.Next() {
		row := iter.Row()
		if len(row) >= 12 {
			chart := &Chart{
				ICAO:      trim(row[0]),
				Filename:  trim(row[1]),
				ChartType: trim(row[2]),
				IndexNo:   trim(row[3]),
				ProcID:    trim(row[4]),
				Action:    trim(row[5]),
				DateRev:   trim(row[6]),
				DateEff:   trim(row[7]),
				TrimSize:  trim(row[8]),
				GeoRef:    trim(row[9]),
				SheetID:   trim(row[10]),
				FtBk:      trim(row[11]),
			}
			d.chartMap[chart.Filename] = chart
		}
	}

	iter = d.ctypes.NewIterator()
	for iter.Next() {
		row := iter.Row()
		if len(row) >= 4 {
			ct := &ChartType{
				Code:      trim(row[0]),
				Category:  trim(row[1]),
				Type:      trim(row[2]),
				Precision: trim(row[3]),
			}
			d.typeMap[ct.Code] = ct
		}
	}
}

func (d *DBF) GetChart(filename string) *Chart {
	return d.chartMap[filename]
}

func (d *DBF) GetChartType(code string) *ChartType {
	return d.typeMap[code]
}

func (d *DBF) GetAllCharts() []*Chart {
	charts := make([]*Chart, 0, len(d.chartMap))
	for _, c := range d.chartMap {
		charts = append(charts, c)
	}
	return charts
}

func (d *DBF) GetAllChartTypes() []*ChartType {
	types := make([]*ChartType, 0, len(d.typeMap))
	for _, t := range d.typeMap {
		types = append(types, t)
	}
	return types
}

func (d *DBF) SearchCharts(query string) []*Chart {
	query = toUpper(query)
	var results []*Chart
	for _, c := range d.chartMap {
		if contains(c.ICAO, query) || contains(c.Filename, query) || contains(c.ProcID, query) {
			results = append(results, c)
		}
	}
	return results
}

func (d *DBF) FilterByType(chartType string) []*Chart {
	var results []*Chart
	for _, c := range d.chartMap {
		if c.ChartType == chartType {
			results = append(results, c)
		}
	}
	return results
}

func (d *DBF) FilterByICAO(icao string) []*Chart {
	icao = toUpper(icao)
	var results []*Chart
	for _, c := range d.chartMap {
		if c.ICAO == icao {
			results = append(results, c)
		}
	}
	return results
}

func (d *DBF) NumCharts() int {
	return len(d.chartMap)
}

func (d *DBF) NumChartTypes() int {
	return len(d.typeMap)
}

func (d *DBF) ResolveChartTypes(query string) []string {
	if query == "" {
		return nil
	}

	queryUpper := toUpper(query)

	if ct, ok := d.typeMap[query]; ok {
		return []string{ct.Code}
	}

	if ct, ok := d.typeMap[queryUpper]; ok {
		return []string{ct.Code}
	}

	var codes []string
	for _, ct := range d.typeMap {
		if contains(toUpper(ct.Type), queryUpper) || contains(toUpper(ct.Category), queryUpper) {
			codes = append(codes, ct.Code)
		}
	}
	return codes
}

func trim(s string) string {
	for len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	return s
}

func toUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
