@echo off
chcp 65001 >nul
echo ========================================
echo VideoDown-Go Build Script
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

:: Read version from file
set /p version=<version.txt
echo Version: %version%
echo.

:: Build Windows version
echo Building Windows 64-bit version...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w -X main.Version=%version%" -o "build/VideoDown-Go-windows-amd64.exe" main.go
if %errorlevel% neq 0 (
    echo Error: Windows build failed
    pause
    exit /b 1
)
echo Windows 64-bit version build completed

:: Build Linux version
echo Building Linux 64-bit version...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w -X main.Version=%version%" -o "build/VideoDown-Go-linux-amd64" main.go
if %errorlevel% neq 0 (    echo Error: Linux build failed
    pause
    exit /b 1
)
echo Linux 64-bit version build completed

:: Build macOS version (skip if fails)
echo Building macOS 64-bit version...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w -X main.Version=%version%" -o "build/VideoDown-Go-darwin-amd64" main.go
if %errorlevel% neq 0 (
    echo Warning: macOS build failed, skipping...
) else (
    echo macOS 64-bit version build completed
)

:: Build macOS ARM64 version (skip if fails)
echo Building macOS ARM64 version...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-s -w" -o "build/VideoDown-Go-darwin-arm64" main.go
if %errorlevel% neq 0 (
    echo Warning: macOS ARM64 build failed, skipping...
) else (
    echo macOS ARM64 version build completed
)

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
echo macOS/Linux: >> "build/bin/README.txt"
echo - ffmpeg >> "build/bin/README.txt"
echo - ffplay >> "build/bin/README.txt"
echo - ffprobe >> "build/bin/README.txt"
echo - yt-dlp >> "build/bin/README.txt"
echo. >> "build/bin/README.txt"
echo Download links: >> "build/bin/README.txt"
echo FFmpeg (optimized for video downloaders): https://github.com/yt-dlp/FFmpeg-Builds?tab=readme-ov-file#ffmpeg-static-auto-builds >> "build/bin/README.txt"
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
echo 1. Users need to download FFmpeg and yt-dlp tools manually
echo 2. Place tools in the bin/ directory
echo 3. See install.md for detailed installation instructions
echo.
echo Press any key to exit...
pause >nul