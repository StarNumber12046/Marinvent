# mrvdrv.dll API Documentation

## Overview
mrvdrv.dll (Marinvent Framework Driver) provides low-level painting primitives for rendering graphics. It abstracts the Windows GDI to provide a consistent interface for rendering to different output devices (screen, printer, metafile).

## Architecture
- **Language**: 32-bit x86 Windows DLL
- **Image Base**: 0x10000000
- **Compiler**: Visual C++ (cdecl calling convention)
- **Purpose**: Graphics driver abstraction layer

## Initialization Functions

### MF_LibOpen
```c
void MF_LibOpen(void);
```
- **Address**: 0x100011d0
- **Purpose**: Initialize the driver library
- **Note**: Checks Windows version for platform-specific behavior

### MF_LibClose
```c
void MF_LibClose(void);
```
- **Address**: 0x10001210
- **Purpose**: Close the driver library

## Painting Context Functions

### MF_BeginPainting
```c
int MF_BeginPainting(HDC hdc);
```
- **Address**: 0x10001220
- **Purpose**: Initialize painting context for a device context
- **Parameters**:
  - `hdc`: Windows device context handle
- **Returns**: 1 on success
- **Side Effects**:
  - Saves current GDI state (clip region, pen, brush, font, palette)
  - Sets default drawing mode (R2_COPYPEN)
  - Sets default fill mode (WINDING)
  - Sets default background mode (TRANSPARENT)
  - Sets default text alignment

### MF_EndPainting
```c
int MF_EndPainting(HDC hdc);
```
- **Address**: 0x100012f0
- **Purpose**: Finalize painting context and restore GDI state
- **Parameters**:
  - `hdc`: Windows device context handle
- **Returns**: 1 on success
- **Side Effects**:
  - Restores original GDI objects (pen, brush, font, palette)
  - Restores original GDI settings (ROP2, fill mode, background mode, text align, text color)
  - Restores clip region

## Object Creation Functions

### MF_CreatePen
```c
HPEN MF_CreatePen(int style, int width, COLORREF color);
```
- **Address**: 0x100013f0
- **Purpose**: Create a pen object

### MF_CreateBrush
```c
HBRUSH MF_CreateBrush(int style, COLORREF color);
```
- **Address**: 0x100014a0
- **Purpose**: Create a brush object

### MF_CreateFont
```c
HFONT MF_CreateFont(const char* faceName, int height, int style);
```
- **Address**: 0x10001540
- **Purpose**: Create a font object

### MF_CreatePalette
```c
HPALETTE MF_CreatePalette(PALETTEENTRY* entries, int count);
```
- **Address**: 0x10001680
- **Purpose**: Create a palette object

### MF_DeleteHandle
```c
void MF_DeleteHandle(HGDIOBJ handle);
```
- **Address**: 0x10001900
- **Purpose**: Delete a GDI object

## Object Selection Functions

### MF_SetPen
```c
void MF_SetPen(HDC hdc, HPEN pen);
```
- **Address**: 0x10001930

### MF_SetBrush
```c
void MF_SetBrush(HDC hdc, HBRUSH brush);
```
- **Address**: 0x10001990

### MF_SetFont
```c
void MF_SetFont(HDC hdc, HFONT font);
```
- **Address**: 0x100019f0

### MF_SetPalette
```c
void MF_SetPalette(HDC hdc, HPALETTE palette);
```
- **Address**: 0x10001a50

## Clipping Functions

### MF_SetClipRgn
```c
void MF_SetClipRgn(HDC hdc, HRGN region);
```
- **Address**: 0x10001ac0

### MF_CreateRectRgn
```c
HRGN MF_CreateRectRgn(int left, int top, int right, int bottom);
```
- **Address**: 0x10001740

### MF_CreatePolygonRgn
```c
HRGN MF_CreatePolygonRgn(POINT* points, int count, int mode);
```
- **Address**: 0x100017a0

### MF_CreateEllipticRgn
```c
HRGN MF_CreateEllipticRgn(int left, int top, int right, int bottom);
```
- **Address**: 0x10001870

### MF_CombineRgn
```c
int MF_CombineRgn(HRGN dest, HRGN src1, HRGN src2, int mode);
```
- **Address**: 0x10001ba0

### MF_PaintRgn
```c
void MF_PaintRgn(HDC hdc, HRGN region);
```
- **Address**: 0x10001ce0

## Drawing Mode Functions

### MF_SetDrawMode
```c
void MF_SetDrawMode(HDC hdc, int mode);
```
- **Address**: 0x10001d30

### MF_SetPolyFillMode
```c
void MF_SetPolyFillMode(HDC hdc, int mode);
```
- **Address**: 0x10001d80

### MF_SetTextBkMode
```c
void MF_SetTextBkMode(HDC hdc, int mode);
```
- **Address**: 0x10001dd0

### MF_SetTextAlignment
```c
void MF_SetTextAlignment(HDC hdc, int align);
```
- **Address**: 0x10001e20

### MF_SetTextColor
```c
void MF_SetTextColor(HDC hdc, COLORREF color);
```
- **Address**: 0x10001e70

## Drawing Primitives

### MF_MoveTo
```c
void MF_MoveTo(HDC hdc, int x, int y);
```
- **Address**: 0x10001fe0

### MF_LineTo
```c
void MF_LineTo(HDC hdc, int x, int y);
```
- **Address**: 0x10002020

### MF_DrawArc
```c
void MF_DrawArc(HDC hdc, int left, int top, int right, int bottom, 
                int startAngle, int endAngle);
```
- **Address**: 0x10002060

### MF_DrawEllipse
```c
void MF_DrawEllipse(HDC hdc, int left, int top, int right, int bottom);
```
- **Address**: 0x100023a0

### MF_DrawChord
```c
void MF_DrawChord(HDC hdc, int left, int top, int right, int bottom,
                  int startAngle, int endAngle);
```
- **Address**: 0x10002440

### MF_DrawPolygon
```c
void MF_DrawPolygon(HDC hdc, POINT* points, int count);
```
- **Address**: 0x100024a0

### MF_DrawPie
```c
void MF_DrawPie(HDC hdc, int left, int top, int right, int bottom,
                int startAngle, int endAngle);
```
- **Address**: 0x10002520

## Text Functions

### MF_DrawText
```c
void MF_DrawText(HDC hdc, const char* text, int x, int y);
```
- **Address**: 0x10001fa0

### MF_BeginTextExtentsCalc
```c
void MF_BeginTextExtentsCalc(HDC hdc);
```
- **Address**: 0x10001ec0

### MF_EndTextExtentsCalc
```c
void MF_EndTextExtentsCalc(HDC hdc);
```
- **Address**: 0x10001ef0

### MF_TextExtentsCalc
```c
void MF_TextExtentsCalc(HDC hdc, const char* text, int* width, int* height);
```
- **Address**: 0x10004380

### MF_GetTextExtents
```c
void MF_GetTextExtents(HDC hdc, const char* text, SIZE* size);
```
- **Address**: 0x10004690

## Raster/Image Functions

### MF_LoadRaster
```c
int MF_LoadRaster(const char* filename, void** raster);
```
- **Address**: 0x10002df0
- **Purpose**: Load a raster image

### MF_DeleteRaster
```c
void MF_DeleteRaster(void* raster);
```
- **Address**: 0x10002640

### MF_BeginPaintingRaster
```c
void MF_BeginPaintingRaster(HDC hdc, void* raster);
```
- **Address**: 0x10002690

### MF_PaintRaster
```c
void MF_PaintRaster(HDC hdc, void* raster, int x, int y, int width, int height);
```
- **Address**: 0x10002700

### MF_EndPaintingRaster
```c
void MF_EndPaintingRaster(HDC hdc, void* raster);
```
- **Address**: 0x10002a00

### MF_CreateCompatibleRaster
```c
void* MF_CreateCompatibleRaster(HDC hdc, int width, int height);
```
- **Address**: 0x100032f0

## File I/O Functions

### MF_OpenFile
```c
int MF_OpenFile(const char* filename, int* handle);
```
- **Address**: 0x10002a80

### MF_CloseFile
```c
void MF_CloseFile(int handle);
```
- **Address**: 0x10002b90

### MF_Read
```c
int MF_Read(int handle, void* buffer, int size);
```
- **Address**: 0x10002b40

### MF_GetFileSize
```c
int MF_GetFileSize(int handle);
```
- **Address**: 0x10002ad0

### MF_SetFilePosition
```c
void MF_SetFilePosition(int handle, int position);
```
- **Address**: 0x10002b10

## Utility Functions

### MF_DecompressData
```c
int MF_DecompressData(void* src, void* dst, int srcSize, int* dstSize);
```
- **Address**: 0x10002c90
- **Purpose**: Decompress data (likely used for compressed chart data)

### MF_GetFilePathLength
```c
int MF_GetFilePathLength(int handle);
```
- **Address**: 0x10002be0

### MF_GetNearestPaletteIndex
```c
int MF_GetNearestPaletteIndex(HPALETTE palette, COLORREF color);
```
- **Address**: 0x10001f40

## Usage Pattern

Typical usage for rendering to a printer:

```c
// Initialize
MF_LibOpen();
MF_BeginPainting(printerDC);

// Create objects
HPEN pen = MF_CreatePen(PS_SOLID, 1, RGB(0,0,0));
HBRUSH brush = MF_CreateBrush(BS_SOLID, RGB(255,255,255));
HFONT font = MF_CreateFont("Arial", 12, 0);

// Set objects
MF_SetPen(printerDC, pen);
MF_SetBrush(printerDC, brush);
MF_SetFont(printerDC, font);

// Drawing operations
MF_MoveTo(printerDC, 100, 100);
MF_LineTo(printerDC, 200, 200);
MF_DrawText(printerDC, "Hello", 100, 100);

// Cleanup
MF_DeleteHandle(pen);
MF_DeleteHandle(brush);
MF_DeleteHandle(font);
MF_EndPainting(printerDC);
MF_LibClose();
```
