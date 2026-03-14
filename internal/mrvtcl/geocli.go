//go:build windows && 386

package mrvtcl

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func CoordToPixelCLI(tclPath string, pictIndex int, lat, lon float64) (x, y int32, err error) {
	// Use absolute path to georef.exe in the app directory
	cmd := exec.Command("C:\\Users\\StarNumber\\Documents\\Marinvent\\georef.exe", "convert", tclPath, strconv.Itoa(pictIndex),
		strconv.FormatFloat(lat, 'f', 6, 64), strconv.FormatFloat(lon, 'f', 6, 64))
	// Only get stdout, ignore stderr (debug output)
	out, err := cmd.Output()
	result := strings.TrimSpace(string(out))

	if err != nil {
		return 0, 0, fmt.Errorf("georef convert failed: %s", result)
	}

	if strings.HasPrefix(result, "ERROR:") {
		code := strings.TrimPrefix(result, "ERROR:")
		code = strings.TrimSpace(code)
		switch code {
		case "-9":
			return 0, 0, ErrInvalidParam
		case "-21":
			return 0, 0, ErrNotGeoreferenced
		case "-23":
			return 0, 0, ErrOutOfBounds
		default:
			return 0, 0, fmt.Errorf("georef error: %s", code)
		}
	}

	parts := strings.Split(result, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid result: %s", result)
	}

	xVal, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("parse x failed: %v", err)
	}
	yVal, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("parse y failed: %v", err)
	}

	return int32(xVal), int32(yVal), nil
}

type GeoCoordResult struct {
	Latitude  float64
	Longitude float64
}

func PixelToCoordCLI(tclPath string, pictIndex int, x, y int32) (GeoCoordResult, error) {
	// Use georef.exe with pixel2coord mode
	cmd := exec.Command("C:\\Users\\StarNumber\\Documents\\Marinvent\\georef.exe", "pixel2coord", tclPath, strconv.Itoa(pictIndex),
		strconv.Itoa(int(x)), strconv.Itoa(int(y)))
	out, err := cmd.Output()
	result := strings.TrimSpace(string(out))

	if err != nil {
		return GeoCoordResult{}, fmt.Errorf("georef pixel2coord failed: %s", result)
	}

	if strings.HasPrefix(result, "ERROR:") {
		code := strings.TrimPrefix(result, "ERROR:")
		code = strings.TrimSpace(code)
		switch code {
		case "-9":
			return GeoCoordResult{}, ErrInvalidParam
		case "-21":
			return GeoCoordResult{}, ErrNotGeoreferenced
		case "-23":
			return GeoCoordResult{}, ErrOutOfBounds
		default:
			return GeoCoordResult{}, fmt.Errorf("georef error: %s", code)
		}
	}

	parts := strings.Split(result, ",")
	if len(parts) != 2 {
		return GeoCoordResult{}, fmt.Errorf("invalid result: %s", result)
	}

	latVal, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return GeoCoordResult{}, fmt.Errorf("parse lat failed: %v", err)
	}
	lonVal, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return GeoCoordResult{}, fmt.Errorf("parse lon failed: %v", err)
	}

	return GeoCoordResult{Latitude: latVal, Longitude: lonVal}, nil
}
