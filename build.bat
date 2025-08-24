@echo off
echo ========================================
echo VideoDown-Go 构建脚本
echo ========================================
echo.

:: 检查 Go 是否安装
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 Go 语言环境，请先安装 Go
    echo 下载地址: https://golang.org/dl/
    pause
    exit /b 1
)

echo 检测到 Go 环境:
go version
echo.

:: 清理之前的构建
echo 清理之前的构建文件...
if exist "build" rmdir /s /q "build"
mkdir "build"
echo.

:: 下载依赖
echo 下载 Go 依赖...
go mod download
if %errorlevel% neq 0 (
    echo 错误: 依赖下载失败
    pause
    exit /b 1
)
echo.

:: 构建 Windows 版本
echo 构建 Windows 64位 版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o "build/VideoDown-Go-windows-amd64.exe" main.go
if %errorlevel% neq 0 (
    echo 错误: Windows 版本构建失败
    pause
    exit /b 1
)
echo Windows 64位版本构建完成

:: 构建 Linux 版本
echo 构建 Linux 64位 版本...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o "build/VideoDown-Go-linux-amd64" main.go
if %errorlevel% neq 0 (
    echo 错误: Linux 版本构建失败
    pause
    exit /b 1
)
echo Linux 64位版本构建完成

:: 构建 macOS 版本
echo 构建 macOS 64位 版本...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w" -o "build/VideoDown-Go-darwin-amd64" main.go
if %errorlevel% neq 0 (
    echo 错误: macOS 版本构建失败
    pause
    exit /b 1
)
echo macOS 64位版本构建完成

:: 构建 macOS ARM64 版本 (Apple Silicon)
echo 构建 macOS ARM64 版本...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-s -w" -o "build/VideoDown-Go-darwin-arm64" main.go
if %errorlevel% neq 0 (
    echo 错误: macOS ARM64 版本构建失败
    pause
    exit /b 1
)
echo macOS ARM64版本构建完成

:: 复制必要文件到构建目录
echo.
echo 复制必要文件...
copy "README.md" "build/"
copy "LICENSE" "build/"
copy "install.md" "build/"
xcopy "templates" "build/templates/" /E /I /Q
if exist "static" xcopy "static" "build/static/" /E /I /Q

:: 创建 bin 目录结构说明
echo 创建工具目录说明...
mkdir "build/bin"
echo 请将以下工具放入此目录: > "build/bin/README.txt"
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
echo 下载地址: >> "build/bin/README.txt"
echo FFmpeg: https://ffmpeg.org/download.html >> "build/bin/README.txt"
echo yt-dlp: https://github.com/yt-dlp/yt-dlp/releases >> "build/bin/README.txt"

:: 显示构建结果
echo.
echo ========================================
echo 构建完成！
echo ========================================
echo.
echo 构建文件位置: build/
dir "build" /B
echo.
echo 注意事项:
echo 1. 用户需要自行下载 FFmpeg 和 yt-dlp 工具
echo 2. 将工具放入 bin/ 目录中
echo 3. 详细安装说明请查看 install.md
echo.
echo 按任意键退出...
pause >nul