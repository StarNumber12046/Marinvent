//go:build windows && 386

package mrvtcl

import (
	"errors"
	"fmt"
)

var (
	ErrNotGeoreferenced = errors.New("chart is not georeferenced")
	ErrOutOfBounds      = errors.New("coordinates out of bounds")
	ErrInvalidParam     = errors.New("invalid parameters")
)

func (c *Chart) IsGeoreferenced() bool {
	if c.handle == 0 {
		return false
	}
	result := TCL_IsPictGeoRefd(c.handle)
	return result == 1
}

func (c *Chart) GetGeoRefStatus() (*GeoRefStatus, error) {
	if c.handle == 0 {
		return nil, ErrNotInitialized
	}

	isGeo := c.IsGeoreferenced()

	return &GeoRefStatus{
		Georeferenced: isGeo,
		Bounds:        c.bounds,
	}, nil
}

func (c *Chart) PixelToCoord(x, y int32) (lat, lon float64, err error) {
	chartMu.Lock()
	defer chartMu.Unlock()

	if c.handle == 0 {
		return 0, 0, ErrNotInitialized
	}

	if !c.IsGeoreferenced() {
		return 0, 0, ErrNotGeoreferenced
	}

	result := TCL_GeoXY2LatLon(c.handle, x, y, &lat, &lon)

	switch result {
	case 1:
		return lat, lon, nil
	case -9:
		return 0, 0, ErrInvalidParam
	case -21:
		return 0, 0, ErrNotGeoreferenced
	case -23:
		return 0, 0, ErrOutOfBounds
	default:
		return 0, 0, fmt.Errorf("TCL_GeoXY2LatLon failed: %d", result)
	}
}

func (c *Chart) CoordToPixel(lat, lon float64) (x, y int32, err error) {
	chartMu.Lock()
	defer chartMu.Unlock()

	if c.handle == 0 {
		return 0, 0, ErrNotInitialized
	}

	if !c.IsGeoreferenced() {
		return 0, 0, ErrNotGeoreferenced
	}

	result := TCL_GeoLatLon2XY(c.handle, lat, lon, &x, &y)

	switch result {
	case 1:
		return x, y, nil
	case -9:
		return 0, 0, ErrInvalidParam
	case -21:
		return 0, 0, ErrNotGeoreferenced
	case -23:
		return 0, 0, ErrOutOfBounds
	default:
		return 0, 0, fmt.Errorf("TCL_GeoLatLon2XY failed: %d", result)
	}
}

func (c *Chart) BatchPixelToCoord(points []PixelPoint) ([]GeoPoint, []error) {
	results := make([]GeoPoint, len(points))
	errs := make([]error, len(points))

	for i, p := range points {
		lat, lon, err := c.PixelToCoord(p.X, p.Y)
		results[i] = GeoPoint{Latitude: lat, Longitude: lon}
		errs[i] = err
	}

	return results, errs
}

func (c *Chart) BatchCoordToPixel(points []GeoPoint) ([]PixelPoint, []error) {
	results := make([]PixelPoint, len(points))
	errs := make([]error, len(points))

	for i, p := range points {
		x, y, err := c.CoordToPixel(p.Latitude, p.Longitude)
		results[i] = PixelPoint{X: x, Y: y}
		errs[i] = err
	}

	return results, errs
}
