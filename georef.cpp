#include <windows.h>
#include <stdio.h>
#include <string.h>

// DLL function pointers
typedef int (*TCL_Open_t)(const char*, unsigned int, const char*, void**);
typedef int (*TCL_ClosePict_t)(void*);
typedef int (*TCL_IsPictGeoRefd_t)(void*);
typedef int (*TCL_GetPictRect_t)(void*, void*);
typedef int (*TCL_GeoLatLon2XY_t)(void*, double, double, int*, int*);
typedef int (*TCL_GeoXY2LatLon_t)(void*, int, int, double*, double*);
typedef int (*TCL_GetNumPictsInFile_t)(const char*, unsigned int*);
typedef void (*TCL_LibInit_t)(const char*, const char*, const char*);
typedef void (*TCL_LibClose_t)();
typedef void (*MF_LibOpen_t)();
typedef void (*MF_LibClose_t)();

// Global function pointers
static TCL_Open_t TCL_Open = NULL;
static TCL_ClosePict_t TCL_ClosePict = NULL;
static TCL_IsPictGeoRefd_t TCL_IsPictGeoRefd = NULL;
static TCL_GetPictRect_t TCL_GetPictRect = NULL;
static TCL_GeoLatLon2XY_t TCL_GeoLatLon2XY = NULL;
static TCL_GeoXY2LatLon_t TCL_GeoXY2LatLon = NULL;
static TCL_GetNumPictsInFile_t TCL_GetNumPictsInFile = NULL;
static TCL_LibInit_t TCL_LibInit = NULL;
static TCL_LibClose_t TCL_LibClose = NULL;
static MF_LibOpen_t MF_LibOpen = NULL;
static MF_LibClose_t MF_LibClose = NULL;

static HMODULE hTcl = NULL;
static HMODULE hDrv = NULL;

#define JEPPVIEW_PATH "C:\\Program Files (x86)\\Jeppesen\\JeppView for Windows"
#define FONTS_PATH "C:\\ProgramData\\Jeppesen\\Common\\Fonts"

static HMODULE TryLoadLibrary(const char* name) {
    // Try current directory first
    HMODULE h = LoadLibraryA(name);
    if (h) return h;
    
    // Try JeppView directory
    char path[512];
    snprintf(path, sizeof(path), "%s\\%s", JEPPVIEW_PATH, name);
    h = LoadLibraryA(path);
    if (h) return h;
    
    // Try system PATH
    h = LoadLibraryA(name);
    if (h) return h;
    
    return NULL;
}

static int LoadDLLs() {
    // Load mrvdrv.dll first
    hDrv = TryLoadLibrary("mrvdrv.dll");
    if (!hDrv) {
        DWORD err = GetLastError();
        fprintf(stderr, "Failed to load mrvdrv.dll, error: %lu\n", err);
        return 0;
    }

    // Load mrvdrv.dll first
    hTcl = TryLoadLibrary("mrvtcl.dll");
    if (!hTcl) {
        DWORD err = GetLastError();
        fprintf(stderr, "Failed to load mrvtcl.dll, error: %lu\n", err);
        FreeLibrary(hDrv);
        hDrv = NULL;
        return 0;
    }

    // Get function addresses
    TCL_Open = (TCL_Open_t)GetProcAddress(hTcl, "TCL_Open");
    TCL_ClosePict = (TCL_ClosePict_t)GetProcAddress(hTcl, "TCL_ClosePict");
    TCL_IsPictGeoRefd = (TCL_IsPictGeoRefd_t)GetProcAddress(hTcl, "TCL_IsPictGeoRefd");
    TCL_GetPictRect = (TCL_GetPictRect_t)GetProcAddress(hTcl, "TCL_GetPictRect");
    TCL_GeoLatLon2XY = (TCL_GeoLatLon2XY_t)GetProcAddress(hTcl, "TCL_GeoLatLon2XY");
    TCL_GeoXY2LatLon = (TCL_GeoXY2LatLon_t)GetProcAddress(hTcl, "TCL_GeoXY2LatLon");
    TCL_GetNumPictsInFile = (TCL_GetNumPictsInFile_t)GetProcAddress(hTcl, "TCL_GetNumPictsInFile");
    TCL_LibInit = (TCL_LibInit_t)GetProcAddress(hTcl, "TCL_LibInit");
    TCL_LibClose = (TCL_LibClose_t)GetProcAddress(hTcl, "TCL_LibClose");
    MF_LibOpen = (MF_LibOpen_t)GetProcAddress(hDrv, "MF_LibOpen");
    MF_LibClose = (MF_LibClose_t)GetProcAddress(hDrv, "MF_LibClose");

    if (!TCL_Open || !TCL_ClosePict || !TCL_IsPictGeoRefd || 
        !TCL_GetPictRect || !TCL_GeoLatLon2XY || !TCL_GeoXY2LatLon || 
        !TCL_GetNumPictsInFile || !TCL_LibInit || !MF_LibOpen) {
        fprintf(stderr, "Failed to get function addresses\n");
        return 0;
    }

    return 1;
}

static void UnloadDLLs() {
    if (hTcl) {
        FreeLibrary(hTcl);
        hTcl = NULL;
    }
    if (hDrv) {
        FreeLibrary(hDrv);
        hDrv = NULL;
    }
}

static int Init(const char* fontPath, const char* lineStylePath, const char* tclClassPath) {
    if (!LoadDLLs()) {
        return 0;
    }
    
    if (MF_LibOpen) MF_LibOpen();
    if (TCL_LibInit) TCL_LibInit(fontPath, lineStylePath, tclClassPath);
    
    return 1;
}

static void Shutdown() {
    if (TCL_LibClose) TCL_LibClose();
    if (MF_LibClose) MF_LibClose();
    UnloadDLLs();
}

static void* OpenChart(const char* tclPath, int pictIndex) {
    fprintf(stderr, "OpenChart: path=%s index=%d\n", tclPath, pictIndex);
    fprintf(stderr, "TCL_GetNumPictsInFile=%p\n", TCL_GetNumPictsInFile);
    unsigned int numPicts = 0;
    if (!TCL_GetNumPictsInFile) {
        fprintf(stderr, "TCL_GetNumPictsInFile is NULL!\n");
        return NULL;
    }
    int countResult = TCL_GetNumPictsInFile(tclPath, &numPicts);
    fprintf(stderr, "GetNumPictsInFile result=%d count=%u\n", countResult, numPicts);
    if (countResult != 1 || numPicts == 0) {
        fprintf(stderr, "Failed to get picture count\n");
        return NULL;
    }
    
    if (pictIndex < 1 || pictIndex > (int)numPicts) {
        fprintf(stderr, "Invalid picture index\n");
        return NULL;
    }
    
    void* handle = NULL;
    int result = TCL_Open(tclPath, pictIndex, NULL, &handle);
    if (result != 1 || handle == NULL) {
        fprintf(stderr, "Failed to open picture\n");
        return NULL;
    }
    
    return handle;
}

static void CloseChart(void* handle) {
    if (handle && TCL_ClosePict) {
        TCL_ClosePict(handle);
    }
}

static int IsGeoreferenced(void* handle) {
    if (!handle || !TCL_IsPictGeoRefd) return 0;
    return TCL_IsPictGeoRefd(handle);
}

static int CoordToPixel(void* handle, double lat, double lon, int* x, int* y) {
    if (!handle || !TCL_GeoLatLon2XY || !x || !y) return -9;
    return TCL_GeoLatLon2XY(handle, lat, lon, x, y);
}

static int PixelToCoord(void* handle, int x, int y, double* lat, double* lon) {
    if (!handle || !TCL_GeoXY2LatLon || !lat || !lon) return -9;
    return TCL_GeoXY2LatLon(handle, x, y, lat, lon);
}

static void PrintUsage() {
    printf("Usage: georef.exe <command> [args]\n");
    printf("\nCommands:\n");
    printf("  init <fontPath> <lineStylePath> <tclClassPath>\n");
    printf("    Initialize the georeferencing library\n");
    printf("  shutdown\n");
    printf("    Shutdown the library\n");
    printf("  open <tclPath> <pictIndex>\n");
    printf("    Open a chart and get a handle\n");
    printf("  close <handle>\n");
    printf("    Close a chart handle\n");
    printf("  isgeoref <handle>\n");
    printf("    Check if chart is georeferenced\n");
    printf("  coord2pixel <handle> <lat> <lon>\n");
    printf("    Convert lat/lon to pixel coordinates\n");
    printf("  pixel2coord <handle> <x> <y>\n");
    printf("    Convert pixel coordinates to lat/lon\n");
    printf("  convert <tclPath> <pictIndex> <lat> <lon>\n");
    printf("    Convert lat/lon to pixel (one-shot, init+open+convert+close+shutdown)\n");
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        PrintUsage();
        return 1;
    }

    if (strcmp(argv[1], "init") == 0) {
        const char* fontPath = (argc > 2) ? argv[2] : FONTS_PATH "\\jeppesen.tfl";
        const char* lineStylePath = (argc > 3) ? argv[3] : FONTS_PATH "\\jeppesen.tls";
        const char* tclClassPath = (argc > 4) ? argv[4] : FONTS_PATH "\\lssdef.tcl";
        
        if (Init(fontPath, lineStylePath, tclClassPath)) {
            printf("OK\n");
            return 0;
        } else {
            printf("FAILED\n");
            return 1;
        }
    }
    else if (strcmp(argv[1], "shutdown") == 0) {
        Shutdown();
        printf("OK\n");
        return 0;
    }
    else if (strcmp(argv[1], "open") == 0) {
        if (argc < 4) {
            fprintf(stderr, "Usage: open <tclPath> <pictIndex>\n");
            return 1;
        }
        fprintf(stderr, "Opening: %s index %d\n", argv[2], atoi(argv[3]));
        void* handle = OpenChart(argv[2], atoi(argv[3]));
        fprintf(stderr, "Handle result: %p\n", handle);
        if (handle) {
            printf("%p\n", handle);
            return 0;
        } else {
            printf("FAILED\n");
            return 1;
        }
    }
    else if (strcmp(argv[1], "close") == 0) {
        if (argc < 3) {
            fprintf(stderr, "Usage: close <handle>\n");
            return 1;
        }
        void* handle = (void*)strtoull(argv[2], NULL, 16);
        CloseChart(handle);
        printf("OK\n");
        return 0;
    }
    else if (strcmp(argv[1], "isgeoref") == 0) {
        if (argc < 3) {
            fprintf(stderr, "Usage: isgeoref <handle>\n");
            return 1;
        }
        void* handle = (void*)strtoull(argv[2], NULL, 16);
        int result = IsGeoreferenced(handle);
        printf("%d\n", result);
        return 0;
    }
    else if (strcmp(argv[1], "coord2pixel") == 0) {
        if (argc < 5) {
            fprintf(stderr, "Usage: coord2pixel <handle> <lat> <lon>\n");
            return 1;
        }
        void* handle = (void*)strtoull(argv[2], NULL, 16);
        double lat = atof(argv[3]);
        double lon = atof(argv[4]);
        int x = 0, y = 0;
        int result = CoordToPixel(handle, lat, lon, &x, &y);
        if (result == 1) {
            printf("%d,%d\n", x, y);
        } else {
            printf("ERROR:%d\n", result);
        }
        return 0;
    }
    else if (strcmp(argv[1], "pixel2coord") == 0) {
        if (argc < 5) {
            fprintf(stderr, "Usage: pixel2coord <handle> <x> <y>\n");
            return 1;
        }
        void* handle = (void*)strtoull(argv[2], NULL, 16);
        int x = atoi(argv[3]);
        int y = atoi(argv[4]);
        double lat = 0, lon = 0;
        int result = PixelToCoord(handle, x, y, &lat, &lon);
        if (result == 1) {
            printf("%.6f,%.6f\n", lat, lon);
        } else {
            printf("ERROR:%d\n", result);
        }
        return 0;
    }
    else if (strcmp(argv[1], "convert") == 0) {
        // One-shot: init + open + coord2pixel + close + shutdown
        if (argc < 6) {
            fprintf(stderr, "Usage: convert <tclPath> <pictIndex> <lat> <lon>\n");
            return 1;
        }
        
        const char* tclPath = argv[2];
        int pictIndex = atoi(argv[3]);
        double lat = atof(argv[4]);
        double lon = atof(argv[5]);
        
        // Initialize
        if (!Init(FONTS_PATH "\\jeppesen.tfl", FONTS_PATH "\\jeppesen.tls", FONTS_PATH "\\lssdef.tcl")) {
            printf("ERROR:init_failed\n");
            return 1;
        }
        
        // Open chart
        void* handle = OpenChart(tclPath, pictIndex);
        if (!handle) {
            printf("ERROR:open_failed\n");
            Shutdown();
            return 1;
        }
        
        // Convert
        int x = 0, y = 0;
        int result = CoordToPixel(handle, lat, lon, &x, &y);
        
        // Close and shutdown
        CloseChart(handle);
        Shutdown();
        
        if (result == 1) {
            printf("%d,%d\n", x, y);
            return 0;
        } else {
            printf("ERROR:%d\n", result);
            return 1;
        }
    }
    else if (strcmp(argv[1], "pixel2coord") == 0) {
        // One-shot: init + open + pixel2coord + close + shutdown
        if (argc < 6) {
            fprintf(stderr, "Usage: pixel2coord <tclPath> <pictIndex> <x> <y>\n");
            return 1;
        }
        
        const char* tclPath = argv[2];
        int pictIndex = atoi(argv[3]);
        int x = atoi(argv[4]);
        int y = atoi(argv[5]);
        
        // Initialize
        if (!Init(FONTS_PATH "\\jeppesen.tfl", FONTS_PATH "\\jeppesen.tls", FONTS_PATH "\\lssdef.tcl")) {
            printf("ERROR:init_failed\n");
            return 1;
        }
        
        // Open chart
        void* handle = OpenChart(tclPath, pictIndex);
        if (!handle) {
            printf("ERROR:open_failed\n");
            Shutdown();
            return 1;
        }
        
        // Check if georeferenced
        if (!IsGeoreferenced(handle)) {
            printf("ERROR:-21\n");
            CloseChart(handle);
            Shutdown();
            return 1;
        }
        
        // Convert
        double lat = 0, lon = 0;
        int result = PixelToCoord(handle, x, y, &lat, &lon);
        
        // Close and shutdown
        CloseChart(handle);
        Shutdown();
        
        if (result == 1) {
            printf("%.6f,%.6f\n", lat, lon);
            return 0;
        } else {
            printf("ERROR:%d\n", result);
            return 1;
        }
    }
    else {
        fprintf(stderr, "Unknown command: %s\n", argv[1]);
        PrintUsage();
        return 1;
    }
}
