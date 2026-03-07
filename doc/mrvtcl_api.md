# mrvtcl.dll API Documentation

## Overview
mrvtcl.dll is the core library for parsing and rendering Marinvent TCL (Terminal Chart Library) files. It provides functions to open, manipulate, and display terminal charts.

## Architecture
- **Language**: 32-bit x86 Windows DLL
- **Image Base**: 0x10000000
- **Compiler**: Visual C++ (cdecl calling convention)
- **Dependencies**: mrvdrv.dll (painting driver)

## Initialization Functions

### TCL_LibInit
```c
int TCL_LibInit(int param1, int param2, int param3, void* param4);
```
- **Address**: 0x10033e40
- **Purpose**: Initialize the TCL library
- **Parameters**:
  - `param1`: Font path (optional, can be 0)
  - `param2`: Palette configuration (optional, can be 0)
  - `param3`: Additional config (optional)
  - `param4`: Callback/error handler
- **Returns**: 1 on success, negative error code on failure

### TCL_LibClose
```c
int TCL_LibClose(void);
```
- **Address**: 0x10034150
- **Purpose**: Close the TCL library and free resources
- **Returns**: Success code

## File Operations

### TCL_Open
```c
int TCL_Open(uint* fileHandle, uint pictIndex, char* pictName, undefined4** pictHandle);
```
- **Address**: 0x10034270
- **Purpose**: Open a TCL file and optionally select a picture by index or name
- **Parameters**:
  - `fileHandle`: Pointer to receive the file handle
  - `pictIndex`: Picture index (0 to get number of pictures, 1+ to select)
  - `pictName`: Picture name (can be NULL if using index)
  - `pictHandle`: Pointer to receive picture handle
- **Returns**: 1 on success, negative error code on failure
- **Error Codes**:
  - -15: Library not initialized
  - -9: Invalid parameter
  - -11: Picture not found

### TCL_ClosePict
```c
int TCL_ClosePict(undefined4 pictHandle);
```
- **Address**: 0x100344f0
- **Purpose**: Close a picture handle

### TCL_CloseAllPicts
```c
int TCL_CloseAllPicts(void);
```
- **Address**: 0x100345b0
- **Purpose**: Close all open pictures

## Picture Information

### TCL_GetNumPictsInFile
```c
uint TCL_GetNumPictsInFile(int fileHandle, uint* numPicts);
```
- **Address**: 0x100343f0
- **Purpose**: Get the number of pictures in a TCL file
- **Parameters**:
  - `fileHandle`: File handle from TCL_Open
  - `numPicts`: Pointer to receive the count
- **Returns**: 1 on success, negative error code on failure

### TCL_GetPictName
```c
int TCL_GetPictName(int pictHandle, char* buffer, int bufferSize);
```
- **Address**: 0x10034c50
- **Purpose**: Get the name of a picture

### TCL_GetPictNamesInFile
```c
int TCL_GetPictNamesInFile(int fileHandle, char** names, int* count);
```
- **Address**: 0x10034450
- **Purpose**: Get all picture names in a file

### TCL_GetPictRect
```c
int TCL_GetPictRect(int pictHandle, RECT* rect);
```
- **Address**: 0x10034b90
- **Purpose**: Get the bounding rectangle of a picture

### TCL_GetVisibleRect
```c
int TCL_GetVisibleRect(int pictHandle, RECT* rect);
```
- **Address**: 0x10032da0
- **Purpose**: Get the visible rectangle of a picture

## Display Functions

### TCL_Display
```c
int TCL_Display(undefined4 pictHandle, HDC hdc, float scaleX, float scaleY, 
                RECT* srcRect, POINT* offset, ushort flags);
```
- **Address**: 0x100345e0
- **Purpose**: Render a picture to a device context
- **Parameters**:
  - `pictHandle`: Picture handle from TCL_Open
  - `hdc`: Windows device context (can be screen, printer, or metafile)
  - `scaleX`, `scaleY`: Scale factors
  - `srcRect`: Source rectangle (can be NULL for entire picture)
  - `offset`: Output offset (can be NULL)
  - `flags`: Display flags (bitfield)
    - Bit 0: Show highlights
    - Bit 1: Show text
    - Bit 2: Show groups
    - Bit 3: Additional rendering
- **Returns**: 1 on success, negative error code on failure

### TCL_DisplayEx
```c
int TCL_DisplayEx(undefined4 pictHandle, HDC hdc, float scale, 
                  RECT* srcRect, POINT* offset, ushort flags);
```
- **Address**: 0x10034870
- **Purpose**: Extended display function with unified scaling
- **Note**: Similar to TCL_Display but uses same scale for X and Y

## Group Operations

### TCL_GetGroupList
```c
int TCL_GetGroupList(int pictHandle, int* groupList, int* count);
```
- **Address**: 0x10032a10
- **Purpose**: Get list of groups in a picture

### TCL_ShowGroup
```c
int TCL_ShowGroup(int pictHandle, int groupId, int show);
```
- **Address**: 0x10031f10
- **Purpose**: Show or hide a group

### TCL_HighlightGroup
```c
int TCL_HighlightGroup(int pictHandle, int groupId, int highlight);
```
- **Address**: 0x10032220
- **Purpose**: Highlight a group

### TCL_GetGroupInfo
```c
int TCL_GetGroupInfo(int pictHandle, int groupId, void* info);
```
- **Address**: 0x10033a60
- **Purpose**: Get information about a group

## Geographic Functions

### TCL_IsPictGeoRefd
```c
int TCL_IsPictGeoRefd(int pictHandle);
```
- **Address**: 0x1002ef20
- **Purpose**: Check if picture has geographic reference

### TCL_GeoLatLon2XY
```c
int TCL_GeoLatLon2XY(int pictHandle, double lat, double lon, double* x, double* y);
```
- **Address**: 0x1002fbe0
- **Purpose**: Convert latitude/longitude to picture coordinates

### TCL_GeoXY2LatLon
```c
int TCL_GeoXY2LatLon(int pictHandle, double x, double y, double* lat, double* lon);
```
- **Address**: 0x10030330
- **Purpose**: Convert picture coordinates to latitude/longitude

## Palette Functions

### TCL_SetPalette
```c
int TCL_SetPalette(int pictHandle, HPALETTE palette);
```
- **Address**: 0x10035410
- **Purpose**: Set the color palette

### TCL_GetPaletteHandle
```c
HPALETTE TCL_GetPaletteHandle(int pictHandle);
```
- **Address**: 0x100352b0
- **Purpose**: Get the current palette handle

### TCL_GetNumColors
```c
int TCL_GetNumColors(int pictHandle);
```
- **Address**: 0x10035250
- **Purpose**: Get the number of colors

## Rotation Functions

### TCL_Rotate
```c
int TCL_Rotate(int pictHandle, float angle);
```
- **Address**: 0x10034f10
- **Purpose**: Rotate the picture

### TCL_RotateLeftRight
```c
int TCL_RotateLeftRight(int pictHandle, int mirror);
```
- **Address**: 0x10034d50
- **Purpose**: Mirror the picture left-right

## Utility Functions

### TCL_GetVersion
```c
int TCL_GetVersion(void);
```
- **Address**: 0x10034230
- **Purpose**: Get library version

### TCL_MovePict
```c
int TCL_MovePict(int pictHandle, int dx, int dy);
```
- **Address**: 0x10035150
- **Purpose**: Move picture origin

## File Format Notes

Based on reverse engineering, TCL files have the following structure:
1. **Header**: Magic bytes "OX" followed by version and flags
2. **Checksum**: CRC or similar checksum for data integrity
3. **Picture Directory**: List of picture names and offsets
4. **Picture Data**: Compressed or uncompressed chart data

The file format appears to be designed for aviation terminal charts.
