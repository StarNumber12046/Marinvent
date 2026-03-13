#include <windows.h>
#include <stdio.h>
#include <math.h>

// DLL function pointers
typedef int (*TCL_Open_t)(const char*, unsigned int, const char*, void**);
typedef int (*TCL_ClosePict_t)(void*);
typedef int (*TCL_IsPictGeoRefd_t)(void*);
typedef int (*TCL_GetPictRect_t)(void*, void*);
typedef int (*TCL_GeoLatLon2XY_t)(void*, double, double, int*, int*);
typedef int (*TCL_GeoXY2LatLon_t)(void*, int, int, double*, double*);
typedef int (*TCL_GetNumPictsInFile_t)(const char*, unsigned int*);

// Global function pointers
static TCL_Open_t TCL_Open = NULL;
static TCL_ClosePict_t TCL_ClosePict = NULL;
static TCL_IsPictGeoRefd_t TCL_IsPictGeoRefd = NULL;
static TCL_GetPictRect_t TCL_GetPictRect = NULL;
static TCL_GeoLatLon2XY_t TCL_GeoLatLon2XY = NULL;
static TCL_GeoXY2LatLon_t TCL_GeoXY2LatLon = NULL;
static TCL_GetNumPictsInFile_t TCL_GetNumPictsInFile = NULL;

static HMODULE hTcl = NULL;
static HMODULE hDrv = NULL;

extern "C" {

// Initialize the DLL - must be called before any other functions
__declspec(dllexport) int GeoInit(const char* fontPath, const char* lineStylePath, const char* tclClassPath) {
    // Load mrvdrv.dll first
    hDrv = LoadLibraryA("mrvdrv.dll");
    if (!hDrv) {
        printf("Failed to load mrvdrv.dll\n");
        return 0;
    }

    // Load mrvtcl.dll
    hTcl = LoadLibraryA("mrvtcl.dll");
    if (!hTcl) {
        printf("Failed to load mrvtcl.dll\n");
        FreeLibrary(hDrv);
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

    if (!TCL_Open || !TCL_ClosePict || !TCL_IsPictGeoRefd || 
        !TCL_GetPictRect || !TCL_GeoLatLon2XY || !TCL_GeoXY2LatLon || !TCL_GetNumPictsInFile) {
        printf("Failed to get function addresses\n");
        FreeLibrary(hTcl);
        FreeLibrary(hDrv);
        return 0;
    }

    // Initialize MF_LibOpen from mrvdrv
    typedef void (*MF_LibOpen_t)();
    MF_LibOpen_t MF_LibOpen = (MF_LibOpen_t)GetProcAddress(hDrv, "MF_LibOpen");
    if (MF_LibOpen) {
        MF_LibOpen();
    }

    // Initialize TCL_LibInit
    typedef int (*TCL_LibInit_t)(const char*, const char*, const char*);
    TCL_LibInit_t TCL_LibInit = (TCL_LibInit_t)GetProcAddress(hTcl, "TCL_LibInit");
    if (TCL_LibInit) {
        TCL_LibInit(fontPath, lineStylePath, tclClassPath);
    }

    printf("GeoInit: DLLs loaded and initialized\n");
    return 1;
}

// Shutdown the DLL
__declspec(dllexport) void GeoShutdown() {
    if (hTcl) {
        typedef void (*TCL_LibClose_t)();
        TCL_LibClose_t TCL_LibClose = (TCL_LibClose_t)GetProcAddress(hTcl, "TCL_LibClose");
        if (TCL_LibClose) {
            TCL_LibClose();
        }
        FreeLibrary(hTcl);
        hTcl = NULL;
    }
    if (hDrv) {
        typedef void (*MF_LibClose_t)();
        MF_LibClose_t MF_LibClose = (MF_LibClose_t)GetProcAddress(hDrv, "MF_LibClose");
        if (MF_LibClose) {
            MF_LibClose();
        }
        FreeLibrary(hDrv);
        hDrv = NULL;
    }
    printf("GeoShutdown: DLLs unloaded\n");
}

// Open a chart file
// Returns: 1 on success, 0 on failure
// Output: pictHandle - pointer to store the picture handle
__declspec(dllexport) int GeoOpenChart(const char* tclPath, int pictIndex, void** pictHandle) {
    if (!TCL_Open || !pictHandle) {
        return 0;
    }
    
    unsigned int numPicts = 0;
    if (TCL_GetNumPictsInFile(tclPath, &numPicts) != 1 || numPicts == 0) {
        return 0;
    }
    
    if (pictIndex < 1 || pictIndex > (int)numPicts) {
        return 0;
    }
    
    int result = TCL_Open(tclPath, pictIndex, NULL, pictHandle);
    return (result == 1 && *pictHandle != NULL) ? 1 : 0;
}

// Close a chart
__declspec(dllexport) int GeoCloseChart(void* pictHandle) {
    if (!TCL_ClosePict || !pictHandle) {
        return 0;
    }
    TCL_ClosePict(pictHandle);
    return 1;
}

// Check if chart is georeferenced
// Returns: 1 if georeferenced, 0 if not, -1 on error
__declspec(dllexport) int GeoIsGeoreferenced(void* pictHandle) {
    if (!TCL_IsPictGeoRefd || !pictHandle) {
        return -1;
    }
    return TCL_IsPictGeoRefd(pictHandle);
}

// Convert lat/lon to pixel coordinates
// Returns: 1 on success, negative on error
// Error codes: -9 = invalid params, -21 = not georeferenced, -23 = out of bounds
__declspec(dllexport) int GeoCoordToPixel(void* pictHandle, double lat, double lon, int* x, int* y) {
    if (!TCL_GeoLatLon2XY || !pictHandle || !x || !y) {
        return -9;
    }
    return TCL_GeoLatLon2XY(pictHandle, lat, lon, x, y);
}

// Convert pixel coordinates to lat/lon
// Returns: 1 on success, negative on error
// Error codes: -9 = invalid params, -21 = not georeferenced, -23 = out of bounds
__declspec(dllexport) int GeoPixelToCoord(void* pictHandle, int x, int y, double* lat, double* lon) {
    if (!TCL_GeoXY2LatLon || !pictHandle || !lat || !lon) {
        return -9;
    }
    return TCL_GeoXY2LatLon(pictHandle, x, y, lat, lon);
}

} // extern "C"

BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved) {
    return TRUE;
}
