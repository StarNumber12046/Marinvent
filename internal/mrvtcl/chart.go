//go:build windows && 386

package mrvtcl

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

var (
	chartMu      sync.Mutex
	activeCharts = make(map[uintptr]*Chart)
)

var (
	ErrInvalidPictureIndex = errors.New("invalid picture index")
	ErrOpenFailed          = errors.New("failed to open picture")
	ErrGetRectFailed       = errors.New("failed to get picture rect")
)

type Chart struct {
	handle    uintptr
	filePath  string
	pictIndex uint32
	bounds    ChartBounds
}

func OpenChart(tclPath string, pictIndex int) (*Chart, error) {
	chartMu.Lock()
	defer chartMu.Unlock()

	if !initialized {
		return nil, ErrNotInitialized
	}

	absPath, err := GetFullPathNameA(tclPath, make([]byte, MAX_PATH))
	if err != nil {
		absPath = tclPath
	}

	var numPicts uint32
	result := TCL_GetNumPictsInFile(uintptr(unsafe.Pointer(CString(absPath))), &numPicts)
	if result != 1 || numPicts == 0 {
		return nil, fmt.Errorf("failed to get picture count: result=%d", result)
	}

	if pictIndex < 1 || pictIndex > int(numPicts) {
		return nil, fmt.Errorf("%w: index %d not in range 1-%d", ErrInvalidPictureIndex, pictIndex, numPicts)
	}

	var pictHandle uintptr
	result = TCL_Open(uintptr(unsafe.Pointer(CString(absPath))), uint32(pictIndex), 0, &pictHandle)
	if result != 1 || pictHandle == 0 {
		return nil, fmt.Errorf("%w: result=%d", ErrOpenFailed, result)
	}

	var rect RECT
	result = TCL_GetPictRect(pictHandle, &rect)
	if result != 1 {
		TCL_ClosePict(pictHandle)
		return nil, fmt.Errorf("%w: result=%d", ErrGetRectFailed, result)
	}

	chart := &Chart{
		handle:    pictHandle,
		filePath:  absPath,
		pictIndex: uint32(pictIndex),
		bounds: ChartBounds{
			Left:   rect.Left,
			Top:    rect.Top,
			Right:  rect.Right,
			Bottom: rect.Bottom,
			Width:  rect.Right - rect.Left,
			Height: rect.Bottom - rect.Top,
		},
	}

	activeCharts[pictHandle] = chart
	return chart, nil
}

func (c *Chart) Close() error {
	chartMu.Lock()
	defer chartMu.Unlock()

	if c.handle != 0 {
		TCL_ClosePict(c.handle)
		delete(activeCharts, c.handle)
		c.handle = 0
	}
	return nil
}

func (c *Chart) Handle() uintptr {
	return c.handle
}

func (c *Chart) FilePath() string {
	return c.filePath
}

func (c *Chart) PictIndex() uint32 {
	return c.pictIndex
}

func (c *Chart) Bounds() ChartBounds {
	return c.bounds
}

func (c *Chart) Width() int32 {
	return c.bounds.Width
}

func (c *Chart) Height() int32 {
	return c.bounds.Height
}
