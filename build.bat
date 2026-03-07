@echo off
REM Build script for TCL to EMF converter
REM Compiles as 32-bit Windows application

setlocal enabledelayedexpansion

REM Set up Visual Studio 32-bit environment
call "C:\Program Files\Microsoft Visual Studio\18\Community\VC\Auxiliary\Build\vcvars32.bat"

REM Compile the program as C++
echo Compiling tcl2emf.cpp as 32-bit C++...
cl.exe /EHsc /MD tcl2emf.cpp /Fe:tcl2emf.exe gdi32.lib user32.lib winspool.lib

if %ERRORLEVEL% EQU 0 (
    echo Build successful!
    echo Output: tcl2emf.exe
) else (
    echo Build failed with error %ERRORLEVEL%
)

endlocal
