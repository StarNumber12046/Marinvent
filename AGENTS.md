# Marinvent TCL DLL Reverse Engineering

## Project Goal
Reverse engineer the Marinvent tcl .dll libraries and write a program that can load a .tcl file and export it to a printer job.

## Important DLLs
- **mrvtcl.dll** (port 8193): Core TCL parsing and rendering library
- **mrvdrv.dll** (port 8194): Low-level painting driver (GDI abstraction)
- **Terminal.dll** (port 8195): Application layer using mrvtcl.dll

## Ghidra Instances
- mrvtcl.dll: `localhost:8193`
- mrvdrv.dll: `localhost:8194`
- Terminal.dll: `localhost:8195`

## Ghidra HTTP API Usage
Use `ghydra` CLI or direct HTTP requests to interact with Ghidra.

### Common Commands
```bash
# List functions
ghydra --port 8193 functions list

# Decompile a function
ghydra --port 8193 functions decompile --name <function_name>

# Get function details
ghydra --port 8193 functions get --address 0x<address>

# List exports
ghydra --port 8193 functions list --name-contains "export"
```

### Direct HTTP API
```bash
# Get program info
curl http://localhost:8193/program

# List functions
curl http://localhost:8193/functions

# Get strings
curl http://localhost:8193/strings
```

## Findings

### mrvtcl.dll (Port 8193)
- **Status**: Analyzed
- **Purpose**: Terminal Chart Library - parses and renders TCL chart files
- **Key Exports** (50 functions):
  - Initialization: `TCL_LibInit`, `TCL_LibClose`
  - File Operations: `TCL_Open`, `TCL_ClosePict`, `TCL_CloseAllPicts`
  - Picture Info: `TCL_GetNumPictsInFile`, `TCL_GetPictName`, `TCL_GetPictRect`, `TCL_GetVisibleRect`
  - Display: `TCL_Display`, `TCL_DisplayEx`
  - Groups: `TCL_GetGroupList`, `TCL_ShowGroup`, `TCL_HighlightGroup`
  - Geographic: `TCL_GeoLatLon2XY`, `TCL_GeoXY2LatLon`, `TCL_IsPictGeoRefd`
  - Palette: `TCL_SetPalette`, `TCL_GetPaletteHandle`, `TCL_GetNumColors`
  - Rotation: `TCL_Rotate`, `TCL_RotateLeftRight`
- **Calling Convention**: __cdecl (32-bit)
- **Image Base**: 0x10000000
- **Detailed API**: See `doc/mrvtcl_api.md`

### mrvdrv.dll (Port 8194)
- **Status**: Analyzed
- **Purpose**: Marinvent Framework Driver - GDI abstraction layer
- **Key Exports** (53 functions):
  - Initialization: `MF_LibOpen`, `MF_LibClose`
  - Painting Context: `MF_BeginPainting`, `MF_EndPainting`
  - Object Creation: `MF_CreatePen`, `MF_CreateBrush`, `MF_CreateFont`, `MF_CreatePalette`
  - Drawing: `MF_MoveTo`, `MF_LineTo`, `MF_DrawArc`, `MF_DrawEllipse`, `MF_DrawText`, etc.
  - Raster: `MF_LoadRaster`, `MF_PaintRaster`, `MF_CreateCompatibleRaster`
  - File I/O: `MF_OpenFile`, `MF_Read`, `MF_CloseFile`
  - Compression: `MF_DecompressData`
- **Calling Convention**: __cdecl (32-bit)
- **Image Base**: 0x10000000
- **Detailed API**: See `doc/mrvdrv_api.md`

### Terminal.dll (Port 8195)
- **Status**: Analyzed
- **Purpose**: Application framework using mrvtcl.dll
- **Key Functions**:
  - `DoTripKitPrinting`: Main printing function
  - `IPrinting`: Printing interface
  - `CPrintableView`, `RichPrintView`: MFC printing views
- **Imports**: Uses MRVTCL.dll for TCL rendering

### TCL File Format
- **Magic Bytes**: "OX" (0x4f58) at offset 0
- **Version**: Byte 2 indicates version (must be > 3)
- **Flags**: Byte 3 has flags (bit 7 = encryption/compression)
- **Checksum**: 4-byte CRC at end of data
- **Structure**:
  - Header (6 bytes): Magic + Version + Flags + Reserved
  - Picture Directory: List of picture names and offsets
  - Picture Data: Compressed chart data
- **Pictures**: A TCL file can contain multiple "pictures" (charts)

## Printing Workflow

To print a TCL file to a printer or export as EMF/PDF:

1. Initialize libraries:
   ```c
   MF_LibOpen();
   TCL_LibInit(0, 0, 0, NULL);
   ```

2. Open TCL file:
   ```c
   uint fileHandle;
   int result = TCL_Open(&fileHandle, 0, NULL, NULL);  // Get count
   uint numPicts;
   TCL_GetNumPictsInFile(fileHandle, &numPicts);
   ```

3. Select picture:
   ```c
   void* pictHandle;
   TCL_Open(&fileHandle, 1, NULL, &pictHandle);  // Open first picture
   ```

4. Create printer/metafile DC:
   ```c
   HDC hdc = CreateEnhMetaFile(...);  // Or printer DC
   ```

5. Render:
   ```c
   MF_BeginPainting(hdc);
   TCL_Display(pictHandle, hdc, 1.0f, 1.0f, NULL, NULL, 0xFFFF);
   MF_EndPainting(hdc);
   ```

6. Cleanup:
   ```c
   TCL_ClosePict(pictHandle);
   TCL_LibClose();
   MF_LibClose();
   ```

## Task Progress
- [x] Analyze mrvtcl.dll exports and key functions
- [x] Analyze mrvdrv.dll exports and key functions
- [x] Analyze Terminal.dll for reference usage
- [x] Understand TCL file format (partial)
- [x] Create 32-bit test program (tcl2emf.exe)
- [ ] Debug TCL_Open error -28 (file parsing issue)
- [ ] Implement printer export functionality
- [ ] Test with sample TCL files

## Current Issue
TCL_GetNumPictsInFile works correctly and returns the picture count.
However, TCL_Open fails with error code -28 (0xFFFFFFE4) when trying to open a picture.

Possible causes:
- Missing font resource file
- Picture parsing error
- Checksum verification issue
- Exception in record type handler (FUN_100071a0)
- Need to investigate exception handling in FUN_100046b0/FUN_10006f70

Known issues:
- TCL_GetVersion crashes when called (skip this function)
- All tested TCL files produce error -28

Debug findings:
- TCL_LibInit with NULL params works (returns 1)
- Font files (jeppesen.tfl, jeppesen.tls) are found but not required for basic init
- Error -28 occurs during picture data parsing, not during file open
- The error might be from FUN_100046b0 (picture parsing) returning failure
- Copied all 113 DLLs from JeppView installation - still same error

Font files found:
- C:\ProgramData\Jeppesen\Common\TerminalCharts\jeppesen.tfl (font list, magic "YO")
- C:\ProgramData\Jeppesen\Common\TerminalCharts\jeppesen.tls (line styles, magic "ZO")

Next steps:
- Try running from JeppView installation directory
- Check if font definition files are required during picture parsing
- Analyze FUN_100046b0 exception handling
- Test with uncompressed TCL files (flags bit 7 = 0)
- Add more detailed error tracing
