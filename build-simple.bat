@echo off
chcp 65001 >nul
echo ========================================
echo VideoDown-Go Simple Build Script
echo ========================================
echo.

:: Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go not found, please install Go first
    echo Download: https://golang.org/dl/
    pause
    exit /b 1
)

echo Go environment detected:
go version
echo.

:: Clean previous builds
echo Cleaning previous build files...
if exist "build" rmdir /s /q "build"
mkdir "build"
echo.

:: Download dependencies
echo Downloading Go dependencies...
go mod download
if %errorlevel% neq 0 (
    echo Error: Failed to download dependencies
    pause
    exit /b 1
)
echo.

:: Build Windows version only
echo Building Windows 64-bit version...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o "build/VideoDown-Go-windows-amd64.exe" main.go
if %errorlevel% neq 0 (
    echo Error: Windows build failed
    pause
    exit /b 1
)
echo Windows 64-bit version build completed

:: Copy necessary files
echo Copying project files...
copy "README.md" "build/"
copy "LICENSE" "build/"
copy "install.md" "build/"
xcopy "templates" "build/templates/" /E /I /Q
if exist "static" xcopy "static" "build/static/" /E /I /Q

:: Create bin directory structure description
echo Creating tool directory description...
mkdir "build/bin"
echo Please place the following tools in this directory: > "build/bin/README.txt"
echo. >> "build/bin/README.txt"
echo Windows: >> "build/bin/README.txt"
echo - ffmpeg.exe >> "build/bin/README.txt"
echo - ffplay.exe >> "build/bin/README.txt"
echo - ffprobe.exe >> "build/bin/README.txt"
echo - yt-dlp.exe >> "build/bin/README.txt"
echo. >> "build/bin/README.txt"
echo Download links: >> "build/bin/README.txt"
echo FFmpeg: https://ffmpeg.org/download.html >> "build/bin/README.txt"
echo yt-dlp: https://github.com/yt-dlp/yt-dlp/releases >> "build/bin/README.txt"

:: Display build results
echo.
echo ========================================
echo Build completed!
echo ========================================
echo.
echo Build files location: build/
dir "build" /B
echo.
echo Notes:
echo 1. This is a Windows-only build
echo 2. Users need to download FFmpeg and yt-dlp tools manually
echo 3. Place tools in the bin/ directory
echo 4. See install.md for detailed installation instructions
echo.
echo Press any key to exit...
pause >nul