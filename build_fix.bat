@echo off
set "VSCOMNTOOLS=C:\Program Files\Microsoft Visual Studio\18\Community\Common7\Tools"
call "C:\Program Files\Microsoft Visual Studio\18\Community\VC\Auxiliary\Build\vcvars32.bat"
cd /d C:\Users\StarNumber\Documents\Marinvent
cl.exe /EHsc /MD tcl2emf.cpp /Fe:tcl2emf.exe gdi32.lib user32.lib winspool.lib
