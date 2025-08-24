#!/bin/bash

echo "========================================"
echo "VideoDown-Go 构建脚本"
echo "========================================"
echo

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 语言环境，请先安装 Go"
    echo "下载地址: https://golang.org/dl/"
    exit 1
fi

echo "检测到 Go 环境:"
go version
echo

# 清理之前的构建
echo "清理之前的构建文件..."
rm -rf build
mkdir -p build
echo

# 下载依赖
echo "下载 Go 依赖..."
go mod download
if [ $? -ne 0 ]; then
    echo "错误: 依赖下载失败"
    exit 1
fi
echo

# 构建 Windows 版本
echo "构建 Windows 64位 版本..."
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o "build/VideoDown-Go-windows-amd64.exe" main.go
if [ $? -ne 0 ]; then
    echo "错误: Windows 版本构建失败"
    exit 1
fi
echo "Windows 64位版本构建完成"

# 构建 Linux 版本
echo "构建 Linux 64位 版本..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "build/VideoDown-Go-linux-amd64" main.go
if [ $? -ne 0 ]; then
    echo "错误: Linux 版本构建失败"
    exit 1
fi
echo "Linux 64位版本构建完成"

# 构建 macOS 版本
echo "构建 macOS 64位 版本..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o "build/VideoDown-Go-darwin-amd64" main.go
if [ $? -ne 0 ]; then
    echo "错误: macOS 版本构建失败"
    exit 1
fi
echo "macOS 64位版本构建完成"

# 构建 macOS ARM64 版本 (Apple Silicon)
echo "构建 macOS ARM64 版本..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o "build/VideoDown-Go-darwin-arm64" main.go
if [ $? -ne 0 ]; then
    echo "错误: macOS ARM64 版本构建失败"
    exit 1
fi
echo "macOS ARM64版本构建完成"

# 复制必要文件到构建目录
echo
echo "复制必要文件..."
cp "README.md" "build/"
cp "LICENSE" "build/"
cp "install.md" "build/"
cp -r "templates" "build/"
if [ -d "static" ]; then
    cp -r "static" "build/"
fi

# 创建 bin 目录结构说明
echo "创建工具目录说明..."
mkdir -p "build/bin"
cat > "build/bin/README.txt" << EOF
请将以下工具放入此目录:

Windows:
- ffmpeg.exe
- ffplay.exe
- ffprobe.exe
- yt-dlp.exe

macOS/Linux:
- ffmpeg
- ffplay
- ffprobe
- yt-dlp

下载地址:
FFmpeg (optimized for video downloaders): https://github.com/yt-dlp/FFmpeg-Builds?tab=readme-ov-file#ffmpeg-static-auto-builds
yt-dlp: https://github.com/yt-dlp/yt-dlp/releases

对于 macOS 用户，推荐使用 Homebrew 安装:
brew install ffmpeg yt-dlp

对于 Linux 用户，可以使用包管理器安装:
# Ubuntu/Debian
sudo apt install ffmpeg
# 然后从 GitHub 下载 yt-dlp
EOF

# 设置可执行权限
chmod +x "build/VideoDown-Go-linux-amd64"
chmod +x "build/VideoDown-Go-darwin-amd64"
chmod +x "build/VideoDown-Go-darwin-arm64"

# 显示构建结果
echo
echo "========================================"
echo "构建完成！"
echo "========================================"
echo
echo "构建文件位置: build/"
ls -la "build/"
echo
echo "注意事项:"
echo "1. 用户需要自行下载 FFmpeg 和 yt-dlp 工具"
echo "2. 将工具放入 bin/ 目录中"
echo "3. 详细安装说明请查看 install.md"
echo
echo "构建完成！"