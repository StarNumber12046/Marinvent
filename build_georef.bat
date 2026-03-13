@echo off
call "C:\Program Files\Microsoft Visual Studio\18\Community\VC\Auxiliary\Build\vcvars32.bat"
echo Compiling georef.cpp as 32-bit C++...
cl.exe /EHsc /MD georef.cpp /Fe:georef.exe gdi32.lib user32.lib winspool.lib
if %ERRORLEVEL% EQU 0 (
    echo Build successful!
    echo Output: georef.exe
) else (
    echo Build failed with error %ERRORLEVEL%
)
