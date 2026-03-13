@echo off
cd /d C:\Users\StarNumber\Documents\Marinvent
echo Init:
georef.exe init
echo.
echo Open:
georef.exe open TCLs\lirz109.tcl 1
echo.
echo Coord2Pixel:
georef.exe coord2pixel 00000001 43.095464 12.502877
echo.
echo Pixel2Coord:
georef.exe pixel2coord 00000001 667 1192
echo.
echo Close:
georef.exe close 00000001
echo.
echo Shutdown:
georef.exe shutdown
