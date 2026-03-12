@echo off
echo Building Marinvent Server (32-bit for DLL support)
echo ==================================================
echo.

set GOARCH=386
set CGO_ENABLED=0

echo Building server...
go build -o server.exe ./cmd/server

if %ERRORLEVEL% neq 0 (
    echo.
    echo Build FAILED!
    exit /b 1
)

echo.
echo Build successful!
echo Output: server.exe
echo.
echo To run: server.exe
echo For help: server.exe -h
