$msvcPath = "C:\Program Files\Microsoft Visual Studio\18\Community\VC\Tools\MSVC\14.50.35717\bin\Hostx86\x86"
$msvcInclude = "C:\Program Files\Microsoft Visual Studio\18\Community\VC\Tools\MSVC\14.50.35717\include"
$msvcLib = "C:\Program Files\Microsoft Visual Studio\18\Community\VC\Tools\MSVC\14.50.35717\lib\x86"
$sdkPath = "C:\Program Files (x86)\Windows Kits\10\Include\10.0.26100.0"
$sdkLib = "C:\Program Files (x86)\Windows Kits\10\Lib\10.0.26100.0"

$env:Path = "$msvcPath;$env:Path"
$env:INCLUDE = "$msvcInclude;$sdkPath\ucrt;$sdkPath\um;$sdkPath\shared"
$env:LIB = "$msvcLib;$sdkLib\ucrt\x86;$sdkLib\um\x86"

Write-Host "INCLUDE: $env:INCLUDE"
Write-Host "LIB: $env:LIB"

Set-Location "C:\Users\StarNumber\Documents\Marinvent"
& cl.exe /EHsc /MD tcl2emf.cpp /Fe:tcl2emf.exe gdi32.lib user32.lib winspool.lib 2>&1
