//go:build windows && 386

package mrvtcl

import (
	"errors"
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
	if c.handle == 0 {
		return 0, 0, ErrNotInitialized
	}

	if !c.IsGeoreferenced() {
		return 0, 0, ErrNotGeoreferenced
	}

	geoPoints, errs := c.BatchPixelToCoord([]PixelPoint{{X: x, Y: y}})
	if errs[0] != nil {
		return 0, 0, errs[0]
	}

	return geoPoints[0].Latitude, geoPoints[0].Longitude, nil
}

func (c *Chart) CoordToPixel(lat, lon float64) (x, y int32, err error) {
	if c.handle == 0 {
		return 0, 0, ErrNotInitialized
	}

	if !c.IsGeoreferenced() {
		return 0, 0, ErrNotGeoreferenced
	}

	pixelPoints, errs := c.BatchCoordToPixel([]GeoPoint{{Latitude: lat, Longitude: lon}})
	if errs[0] != nil {
		return 0, 0, errs[0]
	}

	return pixelPoints[0].X, pixelPoints[0].Y, nil
}

func (c *Chart) BatchPixelToCoord(points []PixelPoint) ([]GeoPoint, []error) {
	results := make([]GeoPoint, len(points))
	errs := make([]error, len(points))

	if c.handle == 0 || c.filePath == "" {
		for i := range results {
			errs[i] = ErrNotInitialized
		}
		return results, errs
	}

	for i, p := range points {
		result, err := PixelToCoordCLI(c.filePath, int(c.pictIndex), p.X, p.Y)
		if err != nil {
			errs[i] = err
		} else {
			results[i] = GeoPoint{Latitude: result.Latitude, Longitude: result.Longitude}
		}
	}

	return results, errs
}

func (c *Chart) BatchCoordToPixel(points []GeoPoint) ([]PixelPoint, []error) {
	results := make([]PixelPoint, len(points))
	errs := make([]error, len(points))

	if c.handle == 0 || c.filePath == "" {
		for i := range results {
			errs[i] = ErrNotInitialized
		}
		return results, errs
	}

	for i, p := range points {
		x, y, err := CoordToPixelCLI(c.filePath, int(c.pictIndex), p.Latitude, p.Longitude)
		if err != nil {
			errs[i] = err
		} else {
			results[i] = PixelPoint{X: x, Y: y}
		}
	}

	return results, errs
}
