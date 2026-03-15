/*
 * TCL to PDF Converter
 * 
 * Exports Jeppesen TCL terminal chart files to PDF format.
 * Uses custom paper sizing to eliminate borders.
 * 
 * Build: build.bat (uses Visual Studio 32-bit compiler)
 * Usage: tcl2emf <input.tcl> <output.pdf> [picture_index]
 */

#include <windows.h>
#include <winspool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define JEPPVIEW_PATH "C:\\Program Files (x86)\\Jeppesen\\JeppView for Windows"
#define MAX_LOADED_FONTS 100

static HMODULE hMrvDrv = NULL;
static HMODULE hMrvTcl = NULL;

static FARPROC MF_LibOpen = NULL;
static FARPROC MF_LibClose = NULL;
static FARPROC MF_BeginPainting = NULL;
static FARPROC MF_EndPainting = NULL;
static FARPROC TCL_LibInit = NULL;
static FARPROC TCL_LibClose = NULL;
static FARPROC TCL_Open = NULL;
static FARPROC TCL_ClosePict = NULL;
static FARPROC TCL_GetNumPictsInFile = NULL;
static FARPROC TCL_GetPictRect = NULL;
static FARPROC TCL_Display = NULL;

static char gLoadedFonts[MAX_LOADED_FONTS][260];
static int gNumLoadedFonts = 0;

static HMODULE TryLoadLibrary(const char *name)
{
    HMODULE hMod = LoadLibraryA(name);
    if (hMod) return hMod;

    char path[MAX_PATH];
    snprintf(path, MAX_PATH, "%s\\%s", JEPPVIEW_PATH, name);
    hMod = LoadLibraryA(path);
    if (hMod) return hMod;

    fprintf(stderr, "  Tried: %s\n", name);
    fprintf(stderr, "  Tried: %s\n", path);
    return NULL;
}

static int LoadDLLs(void)
{
    hMrvDrv = TryLoadLibrary("mrvdrv.dll");
    if (!hMrvDrv) {
        fprintf(stderr, "Failed to load mrvdrv.dll\n");
        return 0;
    }

    hMrvTcl = TryLoadLibrary("mrvtcl.dll");
    if (!hMrvTcl) {
        fprintf(stderr, "Failed to load mrvtcl.dll\n");
        FreeLibrary(hMrvDrv);
        return 0;
    }

    MF_LibOpen = GetProcAddress(hMrvDrv, "MF_LibOpen");
    MF_LibClose = GetProcAddress(hMrvDrv, "MF_LibClose");
    MF_BeginPainting = GetProcAddress(hMrvDrv, "MF_BeginPainting");
    MF_EndPainting = GetProcAddress(hMrvDrv, "MF_EndPainting");

    TCL_LibInit = GetProcAddress(hMrvTcl, "TCL_LibInit");
    TCL_LibClose = GetProcAddress(hMrvTcl, "TCL_LibClose");
    TCL_Open = GetProcAddress(hMrvTcl, "TCL_Open");
    TCL_ClosePict = GetProcAddress(hMrvTcl, "TCL_ClosePict");
    TCL_GetNumPictsInFile = GetProcAddress(hMrvTcl, "TCL_GetNumPictsInFile");
    TCL_GetPictRect = GetProcAddress(hMrvTcl, "TCL_GetPictRect");
    TCL_Display = GetProcAddress(hMrvTcl, "TCL_Display");

    if (!MF_LibOpen || !TCL_LibInit || !TCL_Open || !TCL_Display) {
        fprintf(stderr, "Missing required DLL functions\n");
        return 0;
    }

    return 1;
}

static void UnloadDLLs(void)
{
    if (hMrvTcl) FreeLibrary(hMrvTcl);
    if (hMrvDrv) FreeLibrary(hMrvDrv);
    hMrvTcl = hMrvDrv = NULL;
}

static void LoadJeppesenFonts(const char *fontDir)
{
    char searchPath[MAX_PATH];
    WIN32_FIND_DATAA findData;
    HANDLE hFind;

    snprintf(searchPath, MAX_PATH, "%s\\*.jtf", fontDir);
    hFind = FindFirstFileA(searchPath, &findData);
    if (hFind != INVALID_HANDLE_VALUE) {
        do {
            char fontPath[MAX_PATH];
            snprintf(fontPath, MAX_PATH, "%s\\%s", fontDir, findData.cFileName);
            if (AddFontResourceExA(fontPath, FR_PRIVATE, 0) > 0) {
                if (gNumLoadedFonts < MAX_LOADED_FONTS) {
                    strncpy(gLoadedFonts[gNumLoadedFonts], fontPath, 259);
                    gNumLoadedFonts++;
                }
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
                if (gNumLoadedFonts < MAX_LOADED_FONTS) {
                    strncpy(gLoadedFonts[gNumLoadedFonts], fontPath, 259);
                    gNumLoadedFonts++;
                }
            }
        } while (FindNextFileA(hFind, &findData));
        FindClose(hFind);
    }

    printf("Loaded %d fonts from %s\n", gNumLoadedFonts, fontDir);
}

static void UnloadJeppesenFonts(void)
{
    for (int i = 0; i < gNumLoadedFonts; i++) {
        RemoveFontResourceExA(gLoadedFonts[i], FR_PRIVATE, 0);
    }
    gNumLoadedFonts = 0;
}

static const char* FindPDFPrinter(void)
{
    DWORD cbNeeded = 0;
    DWORD cReturned = 0;

    EnumPrintersA(PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS, NULL, 2, NULL, 0, &cbNeeded, &cReturned);
    if (cbNeeded == 0) return NULL;

    PRINTER_INFO_2A *pPrinters = (PRINTER_INFO_2A *)malloc(cbNeeded);
    if (!pPrinters) return NULL;

    if (!EnumPrintersA(PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS, NULL, 2, (LPBYTE)pPrinters, cbNeeded, &cbNeeded, &cReturned)) {
        free(pPrinters);
        return NULL;
    }

    static char printerName[256];

    for (DWORD i = 0; i < cReturned; i++) {
        if (pPrinters[i].pPrinterName) {
            if (strcmp(pPrinters[i].pPrinterName, "Microsoft Print to PDF") == 0) {
                strncpy(printerName, pPrinters[i].pPrinterName, 255);
                free(pPrinters);
                return printerName;
            }
        }
    }

    for (DWORD i = 0; i < cReturned; i++) {
        if (pPrinters[i].pPrinterName) {
            if (strstr(pPrinters[i].pPrinterName, "Print to PDF")) {
                strncpy(printerName, pPrinters[i].pPrinterName, 255);
                free(pPrinters);
                return printerName;
            }
        }
    }

    free(pPrinters);
    return NULL;
}

static HDC CreatePDFPrinterDC(int chartWidth, int chartHeight)
{
    const char *printerName = FindPDFPrinter();
    if (!printerName) {
        fprintf(stderr, "PDF printer not found\n");
        return NULL;
    }

    printf("Printer: %s\n", printerName);

    HANDLE hPrinter;
    if (!OpenPrinterA((LPSTR)printerName, &hPrinter, NULL)) {
        return NULL;
    }

    int devmodeSize = DocumentPropertiesA(NULL, hPrinter, (LPSTR)printerName, NULL, NULL, 0);
    if (devmodeSize < 1) {
        ClosePrinter(hPrinter);
        return NULL;
    }

    DEVMODEA *pDevmode = (DEVMODEA *)calloc(1, devmodeSize);
    if (!pDevmode) {
        ClosePrinter(hPrinter);
        return NULL;
    }

    if (DocumentPropertiesA(NULL, hPrinter, (LPSTR)printerName, pDevmode, NULL, DM_OUT_BUFFER) != IDOK) {
        free(pDevmode);
        ClosePrinter(hPrinter);
        return NULL;
    }

    int paperWidth = (int)((double)chartWidth * 2.646 + 0.5);
    int paperHeight = (int)((double)chartHeight * 2.646 + 0.5);

    if (0xc670 < paperWidth) paperWidth = 0xc670;
    if (0xc670 < paperHeight) paperHeight = 0xc670;

    pDevmode->dmFields |= DM_PAPERSIZE | DM_PAPERWIDTH | DM_PAPERLENGTH | DM_ORIENTATION;
    pDevmode->dmPaperSize = DMPAPER_USER;

    if (chartWidth < chartHeight) {
        pDevmode->dmPaperWidth = (short)paperWidth;
        pDevmode->dmPaperLength = (short)paperHeight;
        pDevmode->dmOrientation = DMORIENT_PORTRAIT;
    } else {
        pDevmode->dmPaperWidth = (short)paperHeight;
        pDevmode->dmPaperLength = (short)paperWidth;
        pDevmode->dmOrientation = DMORIENT_LANDSCAPE;
    }

    HDC hdc = CreateDCA(NULL, printerName, NULL, pDevmode);;
}

static int InitTCLLib(void)
{
    char tflPath[MAX_PATH];
    char tlsPath[MAX_PATH];
    char lssPath[MAX_PATH];
    char fontPath[MAX_PATH];

    FILE *f = fopen("jeppesen.tfl", "r");
    if (f) {
        fclose(f);
        GetFullPathNameA("jeppesen.tfl", MAX_PATH, tflPath, NULL);
        GetFullPathNameA("jeppesen.tls", MAX_PATH, tlsPath, NULL);
        GetFullPathNameA("lssdef.tcl", MAX_PATH, lssPath, NULL);
        strcpy(fontPath, ".");
    } else {
        strcpy(fontPath, "C:\\ProgramData\\Jeppesen\\Common\\Fonts");
        snprintf(tflPath, MAX_PATH, "%s\\jeppesen.tfl", fontPath);
        snprintf(tlsPath, MAX_PATH, "%s\\jeppesen.tls", fontPath);
        snprintf(lssPath, MAX_PATH, "%s\\lssdef.tcl", fontPath);
    }

    LoadJeppesenFonts(fontPath);
    MF_LibOpen();

    typedef int (__cdecl *TCL_LibInit_t)(const char*, const char*, const char*, void*);
    TCL_LibInit_t pTCL_LibInit = (TCL_LibInit_t)TCL_LibInit;
    return pTCL_LibInit(tflPath, tlsPath, lssPath, NULL) == 1;
}

static int ExportToPDF(const char *tclFile, const char *pdfFile, int pictIndex)
{
    void *pictHandle = NULL;
    unsigned int numPicts = 0;
    RECT pictRect;
    HDC hdcPrinter = NULL;
    int result;

    char absPath[MAX_PATH];
    GetFullPathNameA(tclFile, MAX_PATH, absPath, NULL);

    typedef unsigned int (__cdecl *TCL_GetNumPictsInFile_t)(const char*, unsigned int*);
    TCL_GetNumPictsInFile_t pTCL_GetNumPictsInFile = (TCL_GetNumPictsInFile_t)TCL_GetNumPictsInFile;

    result = pTCL_GetNumPictsInFile(absPath, &numPicts);
    if (result != 1 || numPicts == 0) {
        fprintf(stderr, "Failed to get picture count\n");
        return -1;
    }

    if (pictIndex < 1 || numPicts < (unsigned int)pictIndex) {
        fprintf(stderr, "Invalid picture index %d (valid: 1-%u)\n", pictIndex, numPicts);
        return -1;
    }

    typedef int (__cdecl *TCL_Open_t)(const char*, unsigned int, const char*, void**);
    typedef int (__cdecl *TCL_GetPictRect_t)(void*, RECT*);
    typedef void (__cdecl *TCL_ClosePict_t)(void*);

    TCL_Open_t pTCL_Open = (TCL_Open_t)TCL_Open;
    TCL_GetPictRect_t pTCL_GetPictRect = (TCL_GetPictRect_t)TCL_GetPictRect;
    TCL_ClosePict_t pTCL_ClosePict = (TCL_ClosePict_t)TCL_ClosePict;

    result = pTCL_Open(absPath, pictIndex, NULL, &pictHandle);
    if (result != 1 || !pictHandle) {
        fprintf(stderr, "TCL_Open failed\n");
        return -1;
    }

    result = pTCL_GetPictRect(pictHandle, &pictRect);
    if (result != 1) {
        fprintf(stderr, "TCL_GetPictRect failed\n");
        pTCL_ClosePict(pictHandle);
        return -1;
    }

    int width = pictRect.right - pictRect.left;
    int height = pictRect.bottom - pictRect.top;
    printf("Chart: %d x %d\n", width, height);

    hdcPrinter = CreatePDFPrinterDC(width, height);
    if (!hdcPrinter) {
        pTCL_ClosePict(pictHandle);
        return -1;
    }

    int pageWidth = GetDeviceCaps(hdcPrinter, HORZRES);
    int pageHeight = GetDeviceCaps(hdcPrinter, VERTRES);
    int dpi = GetDeviceCaps(hdcPrinter, LOGPIXELSX);

    printf("Page:  %d x %d @ %d dpi\n", pageWidth, pageHeight, dpi);

    DOCINFOA di = {0};
    di.cbSize = sizeof(DOCINFOA);
    di.lpszDocName = pdfFile;
    di.lpszOutput = pdfFile;

    if (StartDocA(hdcPrinter, &di) <= 0) {
        fprintf(stderr, "StartDoc/StartPage failed\n");
        DeleteDC(hdcPrinter);
        pTCL_ClosePict(pictHandle);
        return -1;
    }

    if (StartPage(hdcPrinter) <= 0) {
        fprintf(stderr, "StartDoc/StartPage failed\n");
        AbortDoc(hdcPrinter);
        DeleteDC(hdcPrinter);
        pTCL_ClosePict(pictHandle);
        return -1;
    }

    SetMapMode(hdcPrinter, MM_ANISOTROPIC);
    SetWindowOrgEx(hdcPrinter, pictRect.left, pictRect.top, NULL);
    SetWindowExtEx(hdcPrinter, width, height, NULL);
    SetViewportOrgEx(hdcPrinter, 0, 0, NULL);
    SetViewportExtEx(hdcPrinter, pageWidth, pageHeight, NULL);

    typedef int (__cdecl *MF_BeginPainting_t)(HDC);
    typedef int (__cdecl *TCL_Display_t)(void*, HDC, float, float, RECT*, POINT*, unsigned short);
    typedef int (__cdecl *MF_EndPainting_t)(HDC);

    MF_BeginPainting_t pMF_BeginPainting = (MF_BeginPainting_t)MF_BeginPainting;
    TCL_Display_t pTCL_Display = (TCL_Display_t)TCL_Display;
    MF_EndPainting_t pMF_EndPainting = (MF_EndPainting_t)MF_EndPainting;

    pMF_BeginPainting(hdcPrinter);
    result = pTCL_Display(pictHandle, hdcPrinter, 1.0f, 1.0f, NULL, NULL, 0xFFFF);
    pMF_EndPainting(hdcPrinter);

    EndPage(hdcPrinter);
    EndDoc(hdcPrinter);
    DeleteDC(hdcPrinter);

    pTCL_ClosePict(pictHandle);

    printf("PDF: %s\n", pdfFile);
    return (result == 1) ? 0 : -1;
}

int main(int argc, char *argv[])
{
    if (argc < 3) {
        printf("Usage: %s <input.tcl> <output.pdf> [picture_index]\n", argv[0]);
        return 1;
    }

    const char *tclFile = argv[1];
    const char *pdfFile = argv[2];
    int pictIndex = (argc > 3) ? atoi(argv[3]) : 1;

    printf("TCL->PDF  in=%s  out=%s  pict=%d\n", tclFile, pdfFile, pictIndex);

    if (!LoadDLLs()) {
        return 1;
    }

    if (!InitTCLLib()) {
        typedef void (__cdecl *TCL_LibClose_t)(void);
        TCL_LibClose_t pTCL_LibClose = (TCL_LibClose_t)TCL_LibClose;
        pTCL_LibClose();
        UnloadJeppesenFonts();
        UnloadDLLs();
        return 1;
    }

    int result = ExportToPDF(tclFile, pdfFile, pictIndex);

    typedef void (__cdecl *TCL_LibClose_t)(void);
    typedef void (__cdecl *MF_LibClose_t)(void);

    TCL_LibClose_t pTCL_LibClose = (TCL_LibClose_t)TCL_LibClose;
    MF_LibClose_t pMF_LibClose = (MF_LibClose_t)MF_LibClose;

    pTCL_LibClose();
    pMF_LibClose();
    UnloadJeppesenFonts();
    UnloadDLLs();

    return result;
}
