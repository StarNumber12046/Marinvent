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

typedef int (__cdecl *TCL_LibInit_t)(int, int, int, void*);
typedef int (__cdecl *TCL_LibClose_t)(void);
typedef int (__cdecl *TCL_Open_t)(const char*, unsigned int, const char*, void**);
typedef int (__cdecl *TCL_ClosePict_t)(void*);
typedef unsigned int (__cdecl *TCL_GetNumPictsInFile_t)(const char*, unsigned int*);
typedef int (__cdecl *TCL_GetPictRect_t)(void*, RECT*);
typedef int (__cdecl *TCL_Display_t)(void*, HDC, float, float, RECT*, POINT*, unsigned short);

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

static MF_LibOpen_t MF_LibOpen = NULL;
static MF_LibClose_t MF_LibClose = NULL;
static MF_BeginPainting_t MF_BeginPainting = NULL;
static MF_EndPainting_t MF_EndPainting = NULL;

static int LoadDLLs(void) {
    hMrvDrv = LoadLibraryA("mrvdrv.dll");
    if (!hMrvDrv) {
        fprintf(stderr, "Failed to load mrvdrv.dll\n");
        return 0;
    }
    
    hMrvTcl = LoadLibraryA("mrvtcl.dll");
    if (!hMrvTcl) {
        fprintf(stderr, "Failed to load mrvtcl.dll\n");
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

static int InitTCLLib(void) {
    char fontPath[MAX_PATH] = {0};
    char lineStylePath[MAX_PATH] = {0};
    char tclClassPath[MAX_PATH] = {0};
    
    FILE* f = fopen("jeppesen.tfl", "r");
    if (f) {
        fclose(f);
        GetFullPathNameA("jeppesen.tfl", MAX_PATH, fontPath, NULL);
        GetFullPathNameA("jeppesen.tls", MAX_PATH, lineStylePath, NULL);
        GetFullPathNameA("lssdef.tcl", MAX_PATH, tclClassPath, NULL);
    } else {
        strcpy(fontPath, "C:\\ProgramData\\Jeppesen\\Common\\TerminalCharts\\jeppesen.tfl");
        strcpy(lineStylePath, "C:\\ProgramData\\Jeppesen\\Common\\TerminalCharts\\jeppesen.tls");
        strcpy(tclClassPath, "C:\\ProgramData\\Jeppesen\\Common\\TerminalCharts\\lssdef.tcl");
    }
    
    MF_LibOpen();
    return TCL_LibInit((int)fontPath, (int)lineStylePath, (int)tclClassPath, NULL) == 1;
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
    
    int width = pictRect.right - pictRect.left;
    int height = pictRect.bottom - pictRect.top;
    printf("Picture size: %d x %d pixels\n", width, height);
    
    hdcPrinter = CreatePDFPrinterDC(pdfFile);
    if (!hdcPrinter) {
        TCL_ClosePict(pictHandle);
        return -1;
    }
    
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
    
    if (StartPage(hdcPrinter) <= 0) {
        fprintf(stderr, "StartPage failed\n");
        AbortDoc(hdcPrinter);
        DeleteDC(hdcPrinter);
        TCL_ClosePict(pictHandle);
        return -1;
    }
    
    SetMapMode(hdcPrinter, MM_ANISOTROPIC);
    SetWindowExtEx(hdcPrinter, width, height, NULL);
    SetViewportExtEx(hdcPrinter, scaledWidth, scaledHeight, NULL);
    SetViewportOrgEx(hdcPrinter, offsetX, offsetY, NULL);
    
    MF_BeginPainting(hdcPrinter);
    result = TCL_Display(pictHandle, hdcPrinter, 1.0f, 1.0f, NULL, NULL, 0xFFFF);
    MF_EndPainting(hdcPrinter);
    
    EndPage(hdcPrinter);
    EndDoc(hdcPrinter);
    
    DeleteDC(hdcPrinter);
    TCL_ClosePict(pictHandle);
    
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
    
    SetMapMode(hdcMeta, MM_ANISOTROPIC);
    SetWindowExtEx(hdcMeta, width, height, NULL);
    SetViewportExtEx(hdcMeta, width, height, NULL);
    
    MF_BeginPainting(hdcMeta);
    result = TCL_Display(pictHandle, hdcMeta, 1.0f, 1.0f, NULL, NULL, 0xFFFF);
    MF_EndPainting(hdcMeta);
    
    hEmf = CloseEnhMetaFile(hdcMeta);
    if (hEmf) {
        DeleteEnhMetaFile(hEmf);
        printf("EMF created: %s\n", emfFile);
    } else {
        fprintf(stderr, "CloseEnhMetaFile failed\n");
    }
    
    TCL_ClosePict(pictHandle);
    return (result == 1) ? 0 : -1;
}

static void PrintUsage(const char* progName) {
    printf("TCL to EMF/PDF Converter\n\n");
    printf("Usage: %s <input.tcl> [output.emf|output.pdf] [picture_index]\n\n", progName);
    printf("Arguments:\n");
    printf("  input.tcl      Input TCL file\n");
    printf("  output         Output file (.emf or .pdf, default: output.emf)\n");
    printf("  picture_index  Picture index (default: 1)\n\n");
    printf("Examples:\n");
    printf("  %s chart.tcl chart.pdf\n", progName);
    printf("  %s chart.tcl chart.emf 1\n", progName);
}

int main(int argc, char* argv[]) {
    printf("TCL to EMF/PDF Converter\n");
    printf("========================\n\n");
    
    if (argc < 2) {
        PrintUsage(argv[0]);
        return 1;
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
    
    if (!LoadDLLs()) {
        fprintf(stderr, "Failed to load DLLs\n");
        return 1;
    }
    
    if (!InitTCLLib()) {
        fprintf(stderr, "TCL_LibInit failed\n");
        UnloadDLLs();
        return 1;
    }
    
    int result;
    if (IsPDFFile(outFile)) {
        result = ExportToPDF(tclFile, outFile, pictIndex);
    } else {
        result = ExportToEMF(tclFile, outFile, pictIndex);
    }
    
    TCL_LibClose();
    UnloadDLLs();
    
    return result;
}
