//go:build !windows || !386

package mrvtcl

import (
	"errors"
)

var (
	ErrNotSupported = errors.New("mrvtcl requires 32-bit Windows")
)

func Load(dllPath string) error {
	return ErrNotSupported
}

func Init(fontDir, configDir string) error {
	return ErrNotSupported
}

func Close() {}

func IsLoaded() bool {
	return false
}

func IsInitialized() bool {
	return false
}

type Chart struct{}

func OpenChart(tclPath string, pictIndex int) (*Chart, error) {
	return nil, ErrNotSupported
}

func (c *Chart) Close() error {
	return nil
}

func (c *Chart) Handle() uintptr {
	return 0
}

func (c *Chart) FilePath() string {
	return ""
}

func (c *Chart) PictIndex() uint32 {
	return 0
}

func (c *Chart) Bounds() ChartBounds {
	return ChartBounds{}
}

func (c *Chart) Width() int32 {
	return 0
}

func (c *Chart) Height() int32 {
	return 0
}

func (c *Chart) IsGeoreferenced() bool {
	return false
}

func (c *Chart) GetGeoRefStatus() (*GeoRefStatus, error) {
	return nil, ErrNotSupported
}

func (c *Chart) PixelToCoord(x, y int32) (lat, lon float64, err error) {
	return 0, 0, ErrNotSupported
}

func (c *Chart) CoordToPixel(lat, lon float64) (x, y int32, err error) {
	return 0, 0, ErrNotSupported
}

func (c *Chart) BatchPixelToCoord(points []PixelPoint) ([]GeoPoint, []error) {
	return nil, nil
}

func (c *Chart) BatchCoordToPixel(points []GeoPoint) ([]PixelPoint, []error) {
	return nil, nil
}

func (c *Chart) ExportToEMF(emfPath string) error {
	return ErrNotSupported
}

func (c *Chart) ExportToPDF(pdfPath string) error {
	return ErrNotSupported
}

func (c *Chart) ExportToPDFBytes() ([]byte, error) {
	return nil, ErrNotSupported
}
