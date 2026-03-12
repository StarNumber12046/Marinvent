package mrvtcl

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type POINT struct {
	X int32
	Y int32
}

type SIZE struct {
	CX int32
	CY int32
}

type GeoPoint struct {
	Latitude  float64
	Longitude float64
}

type PixelPoint struct {
	X int32
	Y int32
}

type ChartBounds struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
	Width  int32
	Height int32
}

type GeoRefStatus struct {
	Georeferenced bool        `json:"georeferenced"`
	Bounds        ChartBounds `json:"bounds"`
}

type CoordToPixelRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type PixelToCoordRequest struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

type CoordToPixelResponse struct {
	X     int32  `json:"x"`
	Y     int32  `json:"y"`
	Error string `json:"error,omitempty"`
}

type PixelToCoordResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Error     string  `json:"error,omitempty"`
}

type BatchCoordToPixelRequest struct {
	Points []CoordToPixelRequest `json:"points"`
}

type BatchPixelToCoordRequest struct {
	Points []PixelToCoordRequest `json:"points"`
}

type BatchCoordToPixelResponse struct {
	Points []CoordToPixelResponse `json:"points"`
}

type BatchPixelToCoordResponse struct {
	Points []PixelToCoordResponse `json:"points"`
}

const (
	JEPPVIEW_PATH      = "C:\\Program Files (x86)\\Jeppesen\\JeppView for Windows"
	JEPPESEN_FONTS_DIR = "C:\\ProgramData\\Jeppesen\\Common\\Fonts"
)

type Config struct {
	DLLPath   string
	FontDir   string
	ConfigDir string
}

var DefaultConfig = Config{
	DLLPath:   JEPPVIEW_PATH,
	FontDir:   JEPPESEN_FONTS_DIR,
	ConfigDir: JEPPESEN_FONTS_DIR,
}
