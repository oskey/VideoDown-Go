# 安装指南

本文档提供了 VideoDown-Go 的详细安装步骤。

## 📋 系统要求

- **操作系统**：Windows 10/11, macOS 10.15+, Linux (Ubuntu 18.04+)
- **Go 版本**：1.19 或更高版本
- **内存**：至少 512MB RAM
- **存储空间**：至少 1GB 可用空间

## 🚀 快速安装

### 方法一：从源码编译

1. **安装 Go 语言环境**
   - 访问 [Go 官网](https://golang.org/dl/) 下载并安装
   - 验证安装：`go version`

2. **克隆项目**
   ```bash
   git clone https://github.com/oskey/VideoDown-Go.git
   cd VideoDown-Go
   ```

3. **安装 Go 依赖**
   ```bash
   go mod download
   ```

4. **下载外部工具**
   
   **Windows 用户：**
   - 下载 [FFmpeg](https://ffmpeg.org/download.html#build-windows) 并解压
   - 将 `ffmpeg.exe`, `ffplay.exe`, `ffprobe.exe` 复制到项目的 `bin/` 目录
   - 下载 [yt-dlp](https://github.com/yt-dlp/yt-dlp/releases) 的 Windows 版本
   - 将 `yt-dlp.exe` 复制到项目的 `bin/` 目录
   
   **macOS 用户：**
   ```bash
   # 使用 Homebrew 安装
   brew install ffmpeg yt-dlp
   
   # 创建符号链接到 bin 目录
   mkdir -p bin
   ln -s $(which ffmpeg) bin/ffmpeg
   ln -s $(which ffplay) bin/ffplay
   ln -s $(which ffprobe) bin/ffprobe
   ln -s $(which yt-dlp) bin/yt-dlp
   ```
   
   **Linux 用户：**
   ```bash
   # Ubuntu/Debian
   sudo apt update
   sudo apt install ffmpeg
   
   # 安装 yt-dlp
   sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
   sudo chmod a+rx /usr/local/bin/yt-dlp
   
   # 创建符号链接
   mkdir -p bin
   ln -s $(which ffmpeg) bin/ffmpeg
   ln -s $(which ffplay) bin/ffplay
   ln -s $(which ffprobe) bin/ffprobe
   ln -s $(which yt-dlp) bin/yt-dlp
   ```

5. **运行程序**
   ```bash
   go run main.go
   ```

6. **访问界面**
   打开浏览器访问：http://127.0.0.1:8888

### 方法二：使用预编译版本（推荐）

1. 访问 [Releases 页面](https://github.com/oskey/VideoDown-Go/releases)
2. 下载适合你系统的预编译版本
3. 解压到任意目录
4. 按照上述步骤 4 安装外部工具
5. 运行可执行文件

## 🔧 配置说明

### 端口配置

默认端口为 8888，如需修改请编辑 `main.go` 文件：

```go
log.Println("服务器启动在 http://127.0.0.1:8888")
log.Fatal(http.ListenAndServe(":8888", nil))
```

将 `:8888` 改为你想要的端口，如 `:3000`。

### 存储路径配置

默认情况下，视频文件和缩略图存储在项目根目录。如需修改存储路径，请在 `main.go` 中查找相关配置。

## 🐛 常见问题

### Q: 提示找不到 ffmpeg 或 yt-dlp
**A:** 确保已正确安装并将可执行文件放在 `bin/` 目录中，或者确保它们在系统 PATH 中。

### Q: 下载视频失败
**A:** 
1. 检查网络连接
2. 确保 yt-dlp 是最新版本
3. 某些网站可能需要特殊配置或代理

### Q: 缩略图生成失败
**A:** 
1. 确保 FFmpeg 正确安装
2. 检查视频文件是否损坏
3. 确保有足够的磁盘空间

### Q: 端口被占用
**A:** 
1. 修改 `main.go` 中的端口号
2. 或者关闭占用 8888 端口的其他程序

### Q: 在 macOS 上提示安全警告
**A:** 
1. 系统偏好设置 → 安全性与隐私 → 通用
2. 点击「仍要打开」允许运行

## 🔄 更新程序

### 从源码更新
```bash
git pull origin main
go mod download
go run main.go
```

### 更新外部工具
定期更新 yt-dlp 以支持最新的网站：

```bash
# Windows: 重新下载最新版本
# macOS: brew upgrade yt-dlp
# Linux: 重新下载最新版本
```

## 📞 获取帮助

如果遇到问题，请：

1. 查看 [FAQ](https://github.com/oskey/VideoDown-Go/wiki/FAQ)
2. 搜索 [Issues](https://github.com/oskey/VideoDown-Go/issues)
3. 提交新的 [Issue](https://github.com/oskey/VideoDown-Go/issues/new)

---

安装完成后，你就可以开始使用 VideoDown-Go 下载和管理视频了！🎉