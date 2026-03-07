# Marinvent Chart API

A Go-based API for accessing and exporting Jeppesen terminal charts from TCL files.

## Requirements

- **Windows** (required for TCL rendering via mrvtcl.dll/mrvdrv.dll)
- **Go 1.25+**
- **Jeppesen data files**:
  - `charts.dbf` - Chart database
  - `ctypes.dbf` - Chart type definitions
  - TCL files (extracted charts)

## Quick Start

### 1. Clone and Build

```bash
# Build CLI and API server
go build -o marinvent.exe ./cmd/cli/
go build -o marinvent-api.exe ./cmd/server/
```

### 2. Directory Structure

```
Marinvent/
├── marinvent.exe          # CLI binary
├── marinvent-api.exe      # API server binary
├── mrvtcl.dll            # TCL rendering library (required)
├── mrvdrv.dll            # GDI driver library (required)
├── zlib.dll              # Compression library (required)
├── tcl2emf.exe           # TCL to EMF converter (required)
├── TCLs/                 # Extracted TCL chart files
├── cmd/
│   ├── cli/              # CLI source
│   └── server/           # API server source
└── internal/
    ├── api/              # HTTP handlers
    ├── charts/           # Chart catalog
    ├── dbf/              # DBF parsing
    └── export/            # Export functionality
```

### 3. Run the API Server

```bash
# Default configuration (uses paths below)
./marinvent-api.exe

# Custom configuration
./marinvent-api.exe -port 9000 -host 127.0.0.1 \
  -charts "C:\ProgramData\Jeppesen\Common\TerminalCharts\charts.dbf" \
  -types "C:\ProgramData\Jeppesen\Common\TerminalCharts\ctypes.dbf" \
  -tcls "C:\path\to\TCLs"
```

### 4. Use the API

```bash
# Health check
curl http://localhost:8080/health

# Get OpenAPI spec
curl http://localhost:8080/openapi.json

# List all KJFK charts
curl http://localhost:8080/api/v1/charts/KJFK

# Filter KJFK by type name (RNAV, ILS, etc.)
curl "http://localhost:8080/api/v1/charts/KJFK?type=RNAV"

# Filter KJFK by type code
curl "http://localhost:8080/api/v1/charts/KJFK?type=1L"

# Filter KJFK by procedure name
curl "http://localhost:8080/api/v1/charts/KJFK?search=RWY+30"

# Combine type and search
curl "http://localhost:8080/api/v1/charts/KJFK?type=RNAV&search=RWY+30"

# List chart types
curl http://localhost:8080/api/v1/chart-types

# Export chart to PDF
curl -o chart.pdf "http://localhost:8080/api/v1/charts/KJFK/export/KJFK225"
```

## Environment Variables

The API server respects these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `HOST` | `0.0.0.0` | HTTP server host |
| `CHARTS_DBF` | `C:\ProgramData\Jeppesen\Common\TerminalCharts\charts.dbf` | Charts DBF path |
| `TYPES_DBF` | `C:\ProgramData\Jeppesen\Common\TerminalCharts\ctypes.dbf` | Chart types DBF path |
| `TCL_DIR` | `TCLs` | Directory containing TCL files |

## CLI Usage

```bash
# List all charts for an ICAO
./marivent.exe -icao KJFK

# Search charts
./marivent.exe -search RNAV

# Filter by type name
./marivent.exe -type ILS

# List all chart types
./marivent.exe -list-types

# Export charts
./marivent.exe -icao KJFK -export output/
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/swagger/index.html` | Swagger UI (interactive API docs) |
| `GET` | `/swagger.json` | OpenAPI 3.0 specification |
| `GET` | `/api/v1/charts/{icao}` | List charts for ICAO |
| `GET` | `/api/v1/charts/{icao}/export/{filename}` | Export chart to PDF |
| `GET` | `/api/v1/chart-types` | List all chart types |

### Query Parameters for `/api/v1/charts/{icao}`

| Parameter | Type | Description |
|-----------|------|-------------|
| `type` | string | Filter by chart type (code like `1L` or name like `RNAV`, `ILS`) |
| `search` | string | Search in procedure name (PROC_ID) |

## Type Lookup

The `type` parameter supports both:
- **Raw codes**: `1L`, `AP`, `01`, etc.
- **Human-readable names**: `RNAV`, `ILS`, `VOR`, `AIRPORT`, etc.

When using a name, the API searches both the `TYPE` and `CATEGORY` fields in `ctypes.dbf` and returns all matching chart types.

Examples:
- `?type=RNAV` → matches codes `1L`, `1C`, etc. (all RNAV types)
- `?type=ILS` → matches codes `01`, `1K`, `2A`, etc. (all ILS types)
- `?type=AP` → matches code `AP` (Airport)

## Data Sources

This tool works with Jeppesen terminal chart data. The DBF files are typically located at:
- `C:\ProgramData\Jeppesen\Common\TerminalCharts\charts.dbf`
- `C:\ProgramData\Jeppesen\Common\TerminalCharts\ctypes.dbf`

TCL files need to be extracted from Jeppesen's `.tcz` archive files.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        API Server                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Gin     │  │ Charts   │  │   DBF    │  │ Export   │  │
│  │  Router  │──│ Service  │──│ Parser   │──│ Service  │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
└─────────────────────────────────────────────────────────────┘
         │                    │                   │
         ▼                    ▼                   ▼
    HTTP Client         DBF Files          tcl2emf.exe
                                          (Windows only)
```

## License

This is reverse-engineered software for educational purposes. Use in accordance with Jeppesen's terms of service.
