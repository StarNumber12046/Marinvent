# TCL Chart Georeferencing

## Overview

TCL chart files can be georeferenced, allowing bidirectional conversion between:

- **Chart pixels (X, Y)** - Integer coordinates within the chart image
- **Geographic coordinates (Lat, Lon)** - WGS84 decimal degrees

## API Functions

### TCL_IsPictGeoRefd

```c
int TCL_IsPictGeoRefd(void* pictHandle);
```

Returns:

- `1` - Chart is georeferenced
- `-21` - Chart is NOT georeferenced

### TCL_GeoLatLon2XY

```c
int TCL_GeoLatLon2XY(void* pictHandle, double lat, double lon, int* x, int* y);
```

Converts geographic coordinates to chart pixels.

Returns:

- `1` - Success
- `-9` - Invalid pointers
- `-21` - Not georeferenced
- `-23` - Coordinates out of chart bounds

### TCL_GeoXY2LatLon

```c
int TCL_GeoXY2LatLon(void* pictHandle, int x, int y, double* lat, double* lon);
```

Converts chart pixels to geographic coordinates.

Returns:

- `1` - Success
- `-9` - Invalid pointers
- `-21` - Not georeferenced
- `-23` - Pixel coordinates out of chart bounds

## Usage Example

```c
#include <windows.h>
#include <stdio.h>

typedef int (__cdecl *TCL_Open_t)(const char*, int, const char*, void**);
typedef int (__cdecl *TCL_ClosePict_t)(void*);
typedef int (__cdecl *TCL_IsPictGeoRefd_t)(void*);
typedef int (__cdecl *TCL_GeoLatLon2XY_t)(void*, double, double, int*, int*);
typedef int (__cdecl *TCL_GeoXY2LatLon_t)(void*, int, int, double*, double*);

int main() {
    HMODULE hTcl = LoadLibraryA("mrvtcl.dll");
    TCL_Open_t TCL_Open = (TCL_Open_t)GetProcAddress(hTcl, "TCL_Open");
    TCL_ClosePict_t TCL_ClosePict = (TCL_ClosePict_t)GetProcAddress(hTcl, "TCL_ClosePict");
    TCL_IsPictGeoRefd_t TCL_IsPictGeoRefd = (TCL_IsPictGeoRefd_t)GetProcAddress(hTcl, "TCL_IsPictGeoRefd");
    TCL_GeoLatLon2XY_t TCL_GeoLatLon2XY = (TCL_GeoLatLon2XY_t)GetProcAddress(hTcl, "TCL_GeoLatLon2XY");
    TCL_GeoXY2LatLon_t TCL_GeoXY2LatLon = (TCL_GeoXY2LatLon_t)GetProcAddress(hTcl, "TCL_GeoXY2LatLon");

    // Open chart
    void* pictHandle = NULL;
    TCL_Open("chart.tcl", 1, NULL, &pictHandle);

    // Check if georeferenced
    if (TCL_IsPictGeoRefd(pictHandle) == 1) {
        // Convert Lat/Lon to chart pixels
        int x, y;
        double lat = 43.095464;
        double lon = 12.502877;

        if (TCL_GeoLatLon2XY(pictHandle, lat, lon, &x, &y) == 1) {
            printf("Location (%.6f, %.6f) is at pixel (%d, %d)\n", lat, lon, x, y);
        }

        // Convert chart pixels to Lat/Lon
        double lat2, lon2;
        if (TCL_GeoXY2LatLon(pictHandle, 1000, 1500, &lat2, &lon2) == 1) {
            printf("Pixel (1000, 1500) is at location (%.6f, %.6f)\n", lat2, lon2);
        }
    }

    TCL_ClosePict(pictHandle);
    FreeLibrary(hTcl);
    return 0;
}
```

## Accuracy

Round-trip conversion accuracy is typically:

- **Latitude error**: < 1 meter
- **Longitude error**: 1-10 meters

Test results for LIRZ111 chart:

```
Input:  Lat=43.095464, Lon=12.502877
Output: X=1179, Y=822
Round-trip: Lat=43.095470, Lon=12.502955
Error: 0.7m lat, 6.3m lon
```

## Chart Types

Not all TCL charts are georeferenced:

| Chart Type              | Georeferenced | Example            |
| ----------------------- | ------------- | ------------------ |
| Area charts (10x)       | ✓ Yes         | lirz102, lirz109   |
| STAR/SID (11x, 12x)     | ✓ Yes         | lirz111, lirz121   |
| Approach plates (x9a/b) | ✗ No          | lirz109a, lirz109b |

**Note**: Approach plates are typically schematic diagrams and may not have proper geographic referencing.

## Dependencies

- `mrvtcl.dll` - Core TCL library
- `mrvdrv.dll` - Driver library (required for TCL_LibInit)
- Jeppesen fonts (for proper rendering)

## Error Codes

| Code | Meaning                   |
| ---- | ------------------------- |
| 1    | Success                   |
| -9   | Invalid pointer parameter |
| -15  | Library not initialized   |
| -21  | Chart not georeferenced   |
| -23  | Coordinates out of bounds |
