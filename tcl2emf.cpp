/*
 * TCL to EMF/PDF Converter
 * 
 * Exports Jeppesen TCL terminal chart files to EMF or PDF format.
 * 
 * Build: build.bat (uses Visual Studio 32-bit compiler)
 * Usage: tcl2emf <input.tcl> [output.emf|output.pdf] [picture_index]
 */

#include <windows.h>
#include <commdlg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>

static LARGE_INTEGER gFreq;
static LARGE_INTEGER gLastTime;

static void TimerInit(void) {
    QueryPerformanceFrequency(&gFreq);
    QueryPerformanceCounter(&gLastTime);
}

static double TimerTick(const char* label) {
    LARGE_INTEGER now;
    QueryPerformanceCounter(&now);
    double ms = (double)(now.QuadPart - gLastTime.QuadPart) * 1000.0 / gFreq.QuadPart;
    gLastTime = now;
    if (label) printf("[%8.2f ms] %s\n", ms, label);
    return ms;
}

#define JEPPVIEW_PATH "C:\\Program Files (x86)\\Jeppesen\\JeppView for Windows"
#define JEPPESEN_FONTS_PATH "C:\\ProgramData\\Jeppesen\\Common\\Fonts"

typedef int (__cdecl *TCL_LibInit_t)(int, int, int, void*);
typedef int (__cdecl *TCL_LibClose_t)(void);
typedef int (__cdecl *TCL_Open_t)(const char*, unsigned int, const char*, void**);
typedef int (__cdecl *TCL_ClosePict_t)(void*);
typedef unsigned int (__cdecl *TCL_GetNumPictsInFile_t)(const char*, unsigned int*);
typedef int (__cdecl *TCL_GetPictRect_t)(void*, RECT*);
typedef int (__cdecl *TCL_Display_t)(void*, HDC, float, float, RECT*, POINT*, unsigned short);
typedef int (__cdecl *TCL_IsPictGeoRefd_t)(void*);
typedef int (__cdecl *TCL_GeoLatLon2XY_t)(void*, double, double, int*, int*);
typedef int (__cdecl *TCL_GeoXY2LatLon_t)(void*, int, int, double*, double*);

typedef void (__cdecl *MF_LibOpen_t)(void);
typedef void (__cdecl *MF_LibClose_t)(void);
typedef int (__cdecl *MF_BeginPainting_t)(HDC);
typedef int (__cdecl *MF_EndPainting_t)(HDC);

static HMODULE hMrvDrv = NULL;
static HMODULE hMrvTcl = NULL;

static TCL_LibInit_t TCL_LibInit = NULL;
static TCL_LibClose_t TCL_LibClose = NULL;
static TCL_Open_t TCL_Open = NULL;
static TCL_ClosePict_t TCL_ClosePict = NULL;
static TCL_GetNumPictsInFile_t TCL_GetNumPictsInFile = NULL;
static TCL_GetPictRect_t TCL_GetPictRect = NULL;
static TCL_Display_t TCL_Display = NULL;
static TCL_IsPictGeoRefd_t TCL_IsPictGeoRefd = NULL;
static TCL_GeoLatLon2XY_t TCL_GeoLatLon2XY = NULL;
static TCL_GeoXY2LatLon_t TCL_GeoXY2LatLon = NULL;

static MF_LibOpen_t MF_LibOpen = NULL;
static MF_LibClose_t MF_LibClose = NULL;
static MF_BeginPainting_t MF_BeginPainting = NULL;
static MF_EndPainting_t MF_EndPainting = NULL;

static HMODULE TryLoadLibrary(const char* name) {
    HMODULE hMod = LoadLibraryA(name);
    if (hMod) return hMod;
    
    char path[MAX_PATH];
    snprintf(path, MAX_PATH, "%s\\%s", JEPPVIEW_PATH, name);
    hMod = LoadLibraryA(path);
    if (hMod) return hMod;
    
    DWORD err = GetLastError();
    fprintf(stderr, "  Tried: %s (error %lu)\n", name, err);
    fprintf(stderr, "  Tried: %s (error %lu)\n", path, err);
    return NULL;
}

static int LoadDLLs(void) {
    hMrvDrv = TryLoadLibrary("mrvdrv.dll");
    if (!hMrvDrv) {
        fprintf(stderr, "Failed to load mrvdrv.dll (not found locally or in " JEPPVIEW_PATH ")\n");
        fprintf(stderr, "Make sure DLLs and their dependencies are in PATH or the executable directory\n");
        return 0;
    }
    
    hMrvTcl = TryLoadLibrary("mrvtcl.dll");
    if (!hMrvTcl) {
        fprintf(stderr, "Failed to load mrvtcl.dll (not found locally or in " JEPPVIEW_PATH ")\n");
        FreeLibrary(hMrvDrv);
        return 0;
    }
    
    MF_LibOpen = (MF_LibOpen_t)GetProcAddress(hMrvDrv, "MF_LibOpen");
    MF_LibClose = (MF_LibClose_t)GetProcAddress(hMrvDrv, "MF_LibClose");
    MF_BeginPainting = (MF_BeginPainting_t)GetProcAddress(hMrvDrv, "MF_BeginPainting");
    MF_EndPainting = (MF_EndPainting_t)GetProcAddress(hMrvDrv, "MF_EndPainting");
    
    TCL_LibInit = (TCL_LibInit_t)GetProcAddress(hMrvTcl, "TCL_LibInit");
    TCL_LibClose = (TCL_LibClose_t)GetProcAddress(hMrvTcl, "TCL_LibClose");
    TCL_Open = (TCL_Open_t)GetProcAddress(hMrvTcl, "TCL_Open");
    TCL_ClosePict = (TCL_ClosePict_t)GetProcAddress(hMrvTcl, "TCL_ClosePict");
    TCL_GetNumPictsInFile = (TCL_GetNumPictsInFile_t)GetProcAddress(hMrvTcl, "TCL_GetNumPictsInFile");
    TCL_GetPictRect = (TCL_GetPictRect_t)GetProcAddress(hMrvTcl, "TCL_GetPictRect");
    TCL_Display = (TCL_Display_t)GetProcAddress(hMrvTcl, "TCL_Display");
    TCL_IsPictGeoRefd = (TCL_IsPictGeoRefd_t)GetProcAddress(hMrvTcl, "TCL_IsPictGeoRefd");
    TCL_GeoLatLon2XY = (TCL_GeoLatLon2XY_t)GetProcAddress(hMrvTcl, "TCL_GeoLatLon2XY");
    TCL_GeoXY2LatLon = (TCL_GeoXY2LatLon_t)GetProcAddress(hMrvTcl, "TCL_GeoXY2LatLon");
    
    if (!MF_LibOpen || !TCL_LibInit || !TCL_Open || !TCL_Display) {
        fprintf(stderr, "Missing required functions\n");
        return 0;
    }
    
    return 1;
}

static void UnloadDLLs(void) {
    if (hMrvTcl) FreeLibrary(hMrvTcl);
    if (hMrvDrv) FreeLibrary(hMrvDrv);
    hMrvTcl = hMrvDrv = NULL;
}

static char gLoadedFonts[100][MAX_PATH];
static int gNumLoadedFonts = 0;

static int LoadJeppesenFonts(const char* fontDir) {
    char searchPath[MAX_PATH];
    WIN32_FIND_DATAA findData;
    HANDLE hFind;
    int loaded = 0;
    
    snprintf(searchPath, MAX_PATH, "%s\\*.jtf", fontDir);
    hFind = FindFirstFileA(searchPath, &findData);
    if (hFind != INVALID_HANDLE_VALUE) {
        do {
            char fontPath[MAX_PATH];
            snprintf(fontPath, MAX_PATH, "%s\\%s", fontDir, findData.cFileName);
            if (AddFontResourceExA(fontPath, FR_PRIVATE, 0) > 0) {
                if (gNumLoadedFonts < 100) {
                    strncpy(gLoadedFonts[gNumLoadedFonts], fontPath, MAX_PATH - 1);
                    gNumLoadedFonts++;
                }
                loaded++;
            }
        } while (FindNextFileA(hFind, &findData));
        FindClose(hFind);
    }
    
    snprintf(searchPath, MAX_PATH, "%s\\*.ttf", fontDir);
    hFind = FindFirstFileA(searchPath, &findData);
    if (hFind != INVALID_HANDLE_VALUE) {
        do {
            char fontPath[MAX_PATH];
            snprintf(fontPath, MAX_PATH, "%s\\%s", fontDir, findData.cFileName);
            if (AddFontResourceExA(fontPath, FR_PRIVATE, 0) > 0) {
                if (gNumLoadedFonts < 100) {
                    strncpy(gLoadedFonts[gNumLoadedFonts], fontPath, MAX_PATH - 1);
                    gNumLoadedFonts++;
                }
                loaded++;
            }
        } while (FindNextFileA(hFind, &findData));
        FindClose(hFind);
    }
    
    return loaded;
}

static void UnloadJeppesenFonts(void) {
    for (int i = 0; i < gNumLoadedFonts; i++) {
        RemoveFontResourceExA(gLoadedFonts[i], FR_PRIVATE, 0);
    }
    gNumLoadedFonts = 0;
}

static int InitTCLLib(void) {
    char fontPath[MAX_PATH] = {0};
    char lineStylePath[MAX_PATH] = {0};
    char tclClassPath[MAX_PATH] = {0};
    char fontDir[MAX_PATH] = {0};
    
    FILE* f = fopen("jeppesen.tfl", "r");
    if (f) {
        fclose(f);
        GetFullPathNameA("jeppesen.tfl", MAX_PATH, fontPath, NULL);
        GetFullPathNameA("jeppesen.tls", MAX_PATH, lineStylePath, NULL);
        GetFullPathNameA("lssdef.tcl", MAX_PATH, tclClassPath, NULL);
        strcpy(fontDir, ".");
    } else {
        strcpy(fontDir, JEPPESEN_FONTS_PATH);
        snprintf(fontPath, MAX_PATH, "%s\\jeppesen.tfl", JEPPESEN_FONTS_PATH);
        snprintf(lineStylePath, MAX_PATH, "%s\\jeppesen.tls", JEPPESEN_FONTS_PATH);
        snprintf(tclClassPath, MAX_PATH, "%s\\lssdef.tcl", JEPPESEN_FONTS_PATH);
    }
    
    int fontsLoaded = LoadJeppesenFonts(fontDir);
    char label[64];
    snprintf(label, sizeof(label), "LoadJeppesenFonts (%d fonts)", fontsLoaded);
    TimerTick(label);
    
    MF_LibOpen();
    TimerTick("MF_LibOpen");
    
    int result = TCL_LibInit((int)fontPath, (int)lineStylePath, (int)tclClassPath, NULL) == 1;
    TimerTick("TCL_LibInit");
    
    return result;
}

static int IsPDFFile(const char* filename) {
    size_t len = strlen(filename);
    if (len < 4) return 0;
    return _stricmp(filename + len - 4, ".pdf") == 0;
}

static HDC CreatePDFPrinterDC(const char* pdfPath) {
    DWORD needed, returned;
    PRINTER_INFO_2A* pinfo = NULL;
    HANDLE hPrinter = NULL;
    HDC hdc = NULL;
    
    if (!EnumPrintersA(PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS, NULL, 2, NULL, 0, &needed, &returned)) {
        if (GetLastError() != ERROR_INSUFFICIENT_BUFFER) {
            fprintf(stderr, "EnumPrinters failed: %lu\n", GetLastError());
            return NULL;
        }
    }
    
    pinfo = (PRINTER_INFO_2A*)malloc(needed);
    if (!pinfo) return NULL;
    
    if (!EnumPrintersA(PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS, NULL, 2, (LPBYTE)pinfo, needed, &needed, &returned)) {
        free(pinfo);
        return NULL;
    }
    
    const char* printerName = NULL;
    
    // Try Microsoft Print to PDF first (most reliable)
    for (DWORD i = 0; i < returned; i++) {
        if (pinfo[i].pPrinterName && 
            strcmp(pinfo[i].pPrinterName, "Microsoft Print to PDF") == 0) {
            printerName = pinfo[i].pPrinterName;
            break;
        }
    }
    
    // Fall back to any PDF printer
    if (!printerName) {
        for (DWORD i = 0; i < returned; i++) {
            if (pinfo[i].pPrinterName && 
                strstr(pinfo[i].pPrinterName, "Print to PDF")) {
                printerName = pinfo[i].pPrinterName;
                break;
            }
        }
    }
    
    // Try Jeppesen Format as last resort
    if (!printerName) {
        for (DWORD i = 0; i < returned; i++) {
            if (pinfo[i].pPrinterName && 
                strcmp(pinfo[i].pPrinterName, "Jeppesen Format") == 0) {
                printerName = pinfo[i].pPrinterName;
                break;
            }
        }
    }
    
    if (!printerName) {
        fprintf(stderr, "'Microsoft Print to PDF' printer not found.\n");
        fprintf(stderr, "Enable it in: Settings > Devices > Printers > Add printer\n");
        free(pinfo);
        return NULL;
    }
    
    printf("Using printer: %s\n", printerName);
    
    if (!OpenPrinterA((LPSTR)printerName, &hPrinter, NULL)) {
        fprintf(stderr, "OpenPrinter failed: %lu\n", GetLastError());
        free(pinfo);
        return NULL;
    }
    
    hdc = CreateDCA(NULL, printerName, NULL, NULL);
    
    free(pinfo);
    ClosePrinter(hPrinter);
    
    return hdc;
}

static int ExportToPDF(const char* tclFile, const char* pdfFile, int pictIndex) {
    void* pictHandle = NULL;
    unsigned int numPicts = 0;
    RECT pictRect;
    HDC hdcPrinter = NULL;
    int result;
    
    char absPath[MAX_PATH];
    GetFullPathNameA(tclFile, MAX_PATH, absPath, NULL);
    
    if (TCL_GetNumPictsInFile(absPath, &numPicts) != 1 || numPicts == 0) {
        fprintf(stderr, "Failed to get picture count\n");
        return -1;
    }
    TimerTick("TCL_GetNumPictsInFile");
    
    if (pictIndex < 1 || pictIndex > (int)numPicts) {
        fprintf(stderr, "Invalid picture index: %d (valid: 1-%u)\n", pictIndex, numPicts);
        return -1;
    }
    
    __try {
        result = TCL_Open(absPath, pictIndex, NULL, &pictHandle);
    } __except(1) {
        fprintf(stderr, "Exception in TCL_Open\n");
        return -1;
    }
    
    if (result != 1 || !pictHandle) {
        fprintf(stderr, "TCL_Open failed: %d\n", result);
        return -1;
    }
    TimerTick("TCL_Open");
    
    __try {
        result = TCL_GetPictRect(pictHandle, &pictRect);
    } __except(1) {
        fprintf(stderr, "Exception in TCL_GetPictRect\n");
        TCL_ClosePict(pictHandle);
        return -1;
    }
    
    if (result != 1) {
        fprintf(stderr, "TCL_GetPictRect failed\n");
        TCL_ClosePict(pictHandle);
        return -1;
    }
    TimerTick("TCL_GetPictRect");
    
    int width = pictRect.right - pictRect.left;
    int height = pictRect.bottom - pictRect.top;
    printf("Picture size: %d x %d pixels\n", width, height);
    
    hdcPrinter = CreatePDFPrinterDC(pdfFile);
    if (!hdcPrinter) {
        TCL_ClosePict(pictHandle);
        return -1;
    }
    TimerTick("CreatePDFPrinterDC");
    
    int dpi = GetDeviceCaps(hdcPrinter, LOGPIXELSX);
    int pageWidth = GetDeviceCaps(hdcPrinter, HORZRES);
    int pageHeight = GetDeviceCaps(hdcPrinter, VERTRES);
    
    double scaleX = (double)pageWidth / width;
    double scaleY = (double)pageHeight / height;
    double scale = (scaleX < scaleY) ? scaleX : scaleY;
    
    int scaledWidth = (int)(width * scale);
    int scaledHeight = (int)(height * scale);
    int offsetX = (pageWidth - scaledWidth) / 2;
    int offsetY = (pageHeight - scaledHeight) / 2;
    
    printf("Page size: %d x %d at %d DPI\n", pageWidth, pageHeight, dpi);
    printf("Scaled to: %d x %d (offset: %d, %d)\n", scaledWidth, scaledHeight, offsetX, offsetY);
    
    DOCINFOA di = {0};
    di.cbSize = sizeof(DOCINFOA);
    di.lpszDocName = pdfFile;
    di.lpszOutput = pdfFile;
    
    if (StartDocA(hdcPrinter, &di) <= 0) {
        fprintf(stderr, "StartDoc failed: %lu\n", GetLastError());
        DeleteDC(hdcPrinter);
        TCL_ClosePict(pictHandle);
        return -1;
    }
    TimerTick("StartDoc");
    
    if (StartPage(hdcPrinter) <= 0) {
        fprintf(stderr, "StartPage failed\n");
        AbortDoc(hdcPrinter);
        DeleteDC(hdcPrinter);
        TCL_ClosePict(pictHandle);
        return -1;
    }
    TimerTick("StartPage");
    
    SetMapMode(hdcPrinter, MM_ANISOTROPIC);
    SetWindowExtEx(hdcPrinter, width, height, NULL);
    SetViewportExtEx(hdcPrinter, scaledWidth, scaledHeight, NULL);
    SetViewportOrgEx(hdcPrinter, offsetX, offsetY, NULL);
    
    MF_BeginPainting(hdcPrinter);
    TimerTick("MF_BeginPainting");
    
    result = TCL_Display(pictHandle, hdcPrinter, 1.0f, 1.0f, NULL, NULL, 0xFFFF);
    TimerTick("TCL_Display");
    
    MF_EndPainting(hdcPrinter);
    TimerTick("MF_EndPainting");
    
    EndPage(hdcPrinter);
    TimerTick("EndPage");
    
    EndDoc(hdcPrinter);
    TimerTick("EndDoc");
    
    DeleteDC(hdcPrinter);
    TimerTick("DeleteDC");
    
    TCL_ClosePict(pictHandle);
    TimerTick("TCL_ClosePict");
    
    printf("PDF created: %s\n", pdfFile);
    return (result == 1) ? 0 : -1;
}

static int ExportToEMF(const char* tclFile, const char* emfFile, int pictIndex) {
    void* pictHandle = NULL;
    unsigned int numPicts = 0;
    RECT pictRect;
    HDC hdcMeta = NULL;
    HENHMETAFILE hEmf = NULL;
    int result;
    
    char absPath[MAX_PATH];
    GetFullPathNameA(tclFile, MAX_PATH, absPath, NULL);
    
    if (TCL_GetNumPictsInFile(absPath, &numPicts) != 1 || numPicts == 0) {
        fprintf(stderr, "Failed to get picture count\n");
        return -1;
    }
    TimerTick("TCL_GetNumPictsInFile");
    
    if (pictIndex < 1 || pictIndex > (int)numPicts) {
        fprintf(stderr, "Invalid picture index: %d (valid: 1-%u)\n", pictIndex, numPicts);
        return -1;
    }
    
    __try {
        result = TCL_Open(absPath, pictIndex, NULL, &pictHandle);
    } __except(1) {
        fprintf(stderr, "Exception in TCL_Open\n");
        return -1;
    }
    
    if (result != 1 || !pictHandle) {
        fprintf(stderr, "TCL_Open failed: %d\n", result);
        return -1;
    }
    TimerTick("TCL_Open");
    
    __try {
        result = TCL_GetPictRect(pictHandle, &pictRect);
    } __except(1) {
        fprintf(stderr, "Exception in TCL_GetPictRect\n");
        TCL_ClosePict(pictHandle);
        return -1;
    }
    
    if (result != 1) {
        fprintf(stderr, "TCL_GetPictRect failed\n");
        TCL_ClosePict(pictHandle);
        return -1;
    }
    TimerTick("TCL_GetPictRect");
    
    int width = pictRect.right - pictRect.left;
    int height = pictRect.bottom - pictRect.top;
    printf("Picture size: %d x %d pixels\n", width, height);
    
    RECT emfRect = {0, 0, width * 100, height * 100};
    hdcMeta = CreateEnhMetaFileA(NULL, emfFile, &emfRect, "Marinvent TCL Chart\0Chart\0");
    if (!hdcMeta) {
        fprintf(stderr, "CreateEnhMetaFile failed: %lu\n", GetLastError());
        TCL_ClosePict(pictHandle);
        return -1;
    }
    TimerTick("CreateEnhMetaFile");
    
    SetMapMode(hdcMeta, MM_ANISOTROPIC);
    SetWindowExtEx(hdcMeta, width, height, NULL);
    SetViewportExtEx(hdcMeta, width, height, NULL);
    
    MF_BeginPainting(hdcMeta);
    TimerTick("MF_BeginPainting");
    
    result = TCL_Display(pictHandle, hdcMeta, 1.0f, 1.0f, NULL, NULL, 0xFFFF);
    TimerTick("TCL_Display");
    
    MF_EndPainting(hdcMeta);
    TimerTick("MF_EndPainting");
    
    hEmf = CloseEnhMetaFile(hdcMeta);
    TimerTick("CloseEnhMetaFile");
    
    if (hEmf) {
        DeleteEnhMetaFile(hEmf);
        TimerTick("DeleteEnhMetaFile");
        printf("EMF created: %s\n", emfFile);
    } else {
        fprintf(stderr, "CloseEnhMetaFile failed\n");
    }
    
    TCL_ClosePict(pictHandle);
    TimerTick("TCL_ClosePict");
    
    return (result == 1) ? 0 : -1;
}

static void PrintUsage(const char* progName) {
    printf("TCL to EMF/PDF Converter\n\n");
    printf("Usage:\n");
    printf("  %s <input.tcl> [output] [picture_index]   - Export to EMF/PDF\n", progName);
    printf("  %s -geo <input.tcl> [picture_index]      - Test georeferencing\n\n", progName);
    printf("Arguments:\n");
    printf("  input.tcl      Input TCL file\n");
    printf("  output         Output file (.emf or .pdf, default: output.emf)\n");
    printf("  picture_index  Picture index (default: 1)\n\n");
    printf("Examples:\n");
    printf("  %s chart.tcl chart.pdf\n", progName);
    printf("  %s chart.tcl chart.emf 1\n", progName);
    printf("  %s -geo chart.tcl 1\n", progName);
}

static int TestGeoref(const char* tclFile, int pictIndex) {
    void* pictHandle = NULL;
    unsigned int numPicts = 0;
    
    if (!TCL_Open(tclFile, 0, NULL, NULL)) {
        fprintf(stderr, "TCL_Open failed\n");
        return -1;
    }
    
    TCL_GetNumPictsInFile(tclFile, &numPicts);
    printf("Number of pictures in file: %u\n", numPicts);
    printf("Opening picture %d...\n", pictIndex);
    fflush(stdout);
    
    if (!TCL_Open(tclFile, pictIndex, NULL, &pictHandle)) {
        fprintf(stderr, "TCL_Open for picture %d failed\n", pictIndex);
        return -1;
    }
    
    printf("Picture %d opened (handle: %p)\n", pictIndex, pictHandle);
    printf("Checking georef status...\n");
    fflush(stdout);
    
    if (!TCL_IsPictGeoRefd) {
        fprintf(stderr, "TCL_IsPictGeoRefd not available\n");
        TCL_ClosePict(pictHandle);
        return -1;
    }
    
    int isGeoRefd = TCL_IsPictGeoRefd(pictHandle);
    printf("TCL_IsPictGeoRefd returned: %d\n", isGeoRefd);
    fflush(stdout);
    printf("Georeferenced: %s\n\n", isGeoRefd ? "YES" : "NO");
    fflush(stdout);
    
    if (!isGeoRefd) {
        printf("Chart is not georeferenced\n");
        TCL_ClosePict(pictHandle);
        return 0;
    }
    
    printf("Getting picture rect...\n");
    fflush(stdout);
    
    RECT pictRect;
    TCL_GetPictRect(pictHandle, & pictRect);
    printf("Picture rect done\n");
    fflush(stdout);
    printf("Chart pixel bounds: left=%d, top=%d, right=%d, bottom=%d\n",
           pictRect.left, pictRect.top, pictRect.right, pictRect.bottom);
    fflush(stdout);
    
    if (!TCL_GeoLatLon2XY || !TCL_GeoXY2LatLon) {
        printf("Geocoordinate conversion functions not available\n");
        TCL_ClosePict(pictHandle);
        return 0;
    }
    
    printf("\nGeocoordinate conversion functions: AVAILABLE\n");
    fflush(stdout);
    
    double testLat = 43.095464;
    double testLon = 12.502877;
    int x = 0, y = 0;
    
    printf("\n=== Test 1: LatLon -> XY ===\n");
    printf("Input: Lat=%.6f, Lon=%.6f\n", testLat, testLon);
    fflush(stdout);
    int result = TCL_GeoLatLon2XY(pictHandle, testLat, testLon, &x, &y);
    printf("TCL_GeoLatLon2XY result: %d, X=%d, Y=%d\n", result, x, y);
    fflush(stdout);
    
    printf("\n=== Test 2: XY -> LatLon (center) ===\n");
    int cx = (pictRect.right - pictRect.left) / 2;
    int cy = (pictRect.bottom - pictRect.top) / 2;
    double lat = 0, lon = 0;
    printf("Input: X=%d, Y=%d\n", cx, cy);
    fflush(stdout);
    int result2 = TCL_GeoXY2LatLon(pictHandle, cx, cy, &lat, &lon);
    printf("TCL_GeoXY2LatLon result: %d, Lat=%.6f, Lon=%.6f\n", result2, lat, lon);
    fflush(stdout);
    
    printf("\n=== Test 3: XY -> LatLon (corner) ===\n");
    int result3 = TCL_GeoXY2LatLon(pictHandle, 100, 100, &lat, &lon);
    printf("TCL_GeoXY2LatLon(100,100) result: %d, Lat=%.6f, Lon=%.6f\n", result3, lat, lon);
    fflush(stdout);
    
    if (result == 1 && x != 0) {
        printf("\n=== Round-trip test ===\n");
        double lat2 = 0, lon2 = 0;
        int rtResult = TCL_GeoXY2LatLon(pictHandle, x, y, &lat2, &lon2);
        printf("XY(%d,%d) -> Lat=%.6f, Lon=%.6f (result=%d)\n", x, y, lat2, lon2, rtResult);
        double latErr = fabs(testLat - lat2) * 111000.0;  // approx meters
        double lonErr = fabs(testLon - lon2) * 111000.0 * cos(testLat * 3.14159 / 180.0);
        printf("Error: %.1fm lat, %.1fm lon\n", latErr, lonErr);
    }
    
    printf("\nChart info: bounds=(%d,%d)-(%d,%d)\n",
           pictRect.left, pictRect.top, pictRect.right, pictRect.bottom);
    printf("Error codes: 1=success, -9=invalid, -21=not georef, -23=out of bounds\n");
    
    TCL_ClosePict(pictHandle);
    return 0;
}

int main(int argc, char* argv[]) {
    printf("TCL to EMF/PDF Converter\n");
    printf("========================\n\n");
    
    if (argc < 2) {
        PrintUsage(argv[0]);
        return 1;
    }
    
    if (strcmp(argv[1], "-geo") == 0) {
        if (argc < 3) {
            printf("Usage: %s -geo <input.tcl> [picture_index]\n", argv[0]);
            return 1;
        }
        
        const char* tclFile = argv[2];
        int pictIndex = (argc > 3) ? atoi(argv[3]) : 1;
        
        printf("Testing georeferencing for: %s (picture %d)\n\n", tclFile, pictIndex);
        fflush(stdout);
        
        FILE* f = fopen(tclFile, "rb");
        if (!f) {
            fprintf(stderr, "File not found: %s\n", tclFile);
            return 1;
        }
        fclose(f);
        
        printf("File found, loading DLLs...\n");
        fflush(stdout);
        
        TimerInit();
        printf("Calling LoadDLLs...\n");
        fflush(stdout);
        
        if (!LoadDLLs()) {
            fprintf(stderr, "Failed to load DLLs\n");
            return 1;
        }
        printf("DLLs loaded successfully\n");
        fflush(stdout);
        TimerTick("LoadDLLs");
        
    if (!InitTCLLib()) {
        fprintf(stderr, "TCL_LibInit failed\n");
        UnloadJeppesenFonts();
        UnloadDLLs();
        return 1;
    }
    printf("InitTCLLib succeeded\n");
    fflush(stdout);
    TimerTick("InitTCLLib");
        
        int result = TestGeoref(tclFile, pictIndex);
        
        TCL_LibClose();
        UnloadJeppesenFonts();
        UnloadDLLs();
        
        return result;
    }
    
    const char* tclFile = argv[1];
    const char* outFile = (argc > 2) ? argv[2] : "output.emf";
    int pictIndex = (argc > 3) ? atoi(argv[3]) : 1;
    
    printf("Input: %s\n", tclFile);
    printf("Output: %s\n", outFile);
    printf("Picture: %d\n\n", pictIndex);
    
    FILE* f = fopen(tclFile, "rb");
    if (!f) {
        fprintf(stderr, "File not found: %s\n", tclFile);
        return 1;
    }
    fclose(f);
    
    printf("=== Performance Timing ===\n");
    TimerInit();
    
    if (!LoadDLLs()) {
        fprintf(stderr, "Failed to load DLLs\n");
        return 1;
    }
    TimerTick("LoadDLLs");
    
    if (!InitTCLLib()) {
        fprintf(stderr, "TCL_LibInit failed\n");
        UnloadJeppesenFonts();
        UnloadDLLs();
        return 1;
    }
    TimerTick("InitTCLLib (fonts + MF_LibOpen + TCL_LibInit)");
    
    int result;
    if (IsPDFFile(outFile)) {
        result = ExportToPDF(tclFile, outFile, pictIndex);
    } else {
        result = ExportToEMF(tclFile, outFile, pictIndex);
    }
    
    TCL_LibClose();
    TimerTick("TCL_LibClose");
    
    UnloadJeppesenFonts();
    TimerTick("UnloadJeppesenFonts");
    
    UnloadDLLs();
    TimerTick("UnloadDLLs");
    
    printf("\n=== Done ===\n");
    return result;
}
