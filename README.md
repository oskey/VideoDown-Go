# X-KT 视频下载器 🎥 v1.3.2

![X-KT 视频下载器 Screenshot](https://private-user-images.githubusercontent.com/13282035/481367744-307e8eaf-b517-4a1a-8b3c-e66a056d1662.png?jwt=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NTYwMzI4NjYsIm5iZiI6MTc1NjAzMjU2NiwicGF0aCI6Ii8xMzI4MjAzNS80ODEzNjc3NDQtMzA3ZThlYWYtYjUxNy00YTFhLThiM2MtZTY2YTA1NmQxNjYyLnBuZz9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUFLSUFWQ09EWUxTQTUzUFFLNFpBJTJGMjAyNTA4MjQlMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjUwODI0VDEwNDkyNlomWC1BbXotRXhwaXJlcz0zMDAmWC1BbXotU2lnbmF0dXJlPTE5ZDU0ODBkNTBiNGE2MzM5ZmI4OTdkYTY4MGUwZmMzZjUzMTQ1ODQwM2M5M2E2MWZhMmRjNWIxYzg4ZDM0MDImWC1BbXotU2lnbmVkSGVhZGVycz1ob3N0In0.Gl1uEV0-g3FePtr2rjk0qnyYFCMCVVICAQj4lETvAro)

一个基于 Go 语言开发的现代化视频下载和管理工具，支持多平台视频下载、智能缩略图生成和响应式 Web 界面。

## 🐛 v1.3.2 BUG修复

### 🔧 临时文件清理优化
- **智能文件检测**：修复了任务停止时临时文件清理的BUG
- **动态格式识别**：现在根据实际下载文件的扩展名（如 `.mp4`）来检测临时文件
- **精准清理**：不再依赖用户选择的视频格式，而是基于yt-dlp实际下载的文件格式
- **完整清理**：能正确清理如 `视频名.mp4.part`、`视频名.mp4.ytdl` 等所有相关临时文件
- **日志优化**：调试信息现在显示实际检测的文件扩展名，便于问题排查

## 🆕 v1.3.0 新功能

### 📋 批量操作增强
- **视频格式选择**：新增视频播放在线解码及保存格式选择，无论下载的视频格式是什么，通过FFmpeg进行转码和编码
- **全选功能**：新增全选复选框，支持一键选择/取消选择所有视频
- **智能状态同步**：全选复选框会根据当前选中状态智能显示
  - 全部选中：显示为选中状态
  - 部分选中：显示为半选状态（indeterminate）
  - 全部未选：显示为未选状态
- **用户体验优化**：批量操作更加便捷，提升视频管理效率

## 🔄 v1.2.0 更新内容

### 🔧 FFmpeg 集成优化
- **自动版本检测**：新增 FFmpeg 版本显示功能，实时显示已安装的 FFmpeg 版本信息
- **跨平台支持**：完善 Linux 和 Windows 下的 FFmpeg 检测逻辑
- **智能下载**：根据操作系统自动选择合适的 FFmpeg 下载包格式
  - Windows：自动下载 `.zip` 格式
  - Linux：自动下载 `.tar.xz` 格式
- **自动解压**：支持多种压缩格式的自动解压和安装

### 🌐 网络连接改进
- **多地址支持**：服务器现在支持多种访问方式
  - `http://127.0.0.1:8888`
  - `http://localhost:8888`
  - `http://[本机IP]:8888`（局域网访问）
- **连接稳定性**：优化服务器监听逻辑，提升网络连接稳定性

### 🛠️ 跨平台兼容性
- **Linux 支持**：完善 Linux 系统下的 FFmpeg 管理功能
- **路径处理**：统一 Windows 和 Linux 下的可执行文件路径处理
- **文件格式**：智能识别不同操作系统所需的文件格式

## ✨ 功能特性

### 🚀 核心功能
- **多平台视频下载**：支持 YouTube、TikTok、Bilibili 等主流视频平台
- **智能缩略图生成**：自动生成视频缩略图，支持宽高比自适应显示
- **实时下载进度**：WebSocket 实时显示下载进度和状态
- **批量操作**：支持批量选择和删除视频文件
- **视频预览**：内置视频播放器，支持在线预览

### 🎨 界面特性
- **现代化 UI**：简洁美观的响应式界面设计
- **智能缩略图显示**：根据视频宽高比自动调整缩略图尺寸
  - 竖屏视频：60×107px（完整显示竖屏内容）
  - 方形视频：90×90px（正方形显示）
  - 横屏视频：160×90px（16:9 比例显示）
- **图片预览**：点击缩略图可弹窗查看大图
- **拖拽排序**：支持视频列表拖拽排序
- **移动端适配**：完美支持手机和平板设备

### 🛠️ 技术特性
- **高性能**：Go 语言开发，内存占用低，运行速度快
- **跨平台**：支持 Windows、macOS、Linux
- **零依赖部署**：单文件部署，无需额外安装依赖
- **实时通信**：WebSocket 实现实时状态更新

## 📦 安装说明

> ⚠️ **重要提示**：如果遇到软件无法正常使用的情况，请先尝试更新 yt-dlp.exe 到最新版本。下载地址：[yt-dlp Releases](https://github.com/yt-dlp/yt-dlp/releases)

### 环境要求
- Go 1.19 或更高版本
- FFmpeg（用于视频处理和缩略图生成）
- yt-dlp（用于视频下载）

### 快速开始

1. **克隆项目**
```bash
git clone https://github.com/oskey/VideoDown-Go.git
cd VideoDown-Go
```

2. **安装依赖**
```bash
go mod download
```

3. **准备工具**
   - 将 `ffmpeg.exe`、`ffplay.exe`、`ffprobe.exe` 放入 `bin/` 目录
   - 将 `yt-dlp.exe` 放入 `bin/` 目录

4. **运行程序**
```bash
go run main.go
```

5. **访问界面**
   打开浏览器访问：http://127.0.0.1:8888

## 🎯 使用方法

### ⚠️ 重要注意事项

本项目基于 **yt-dlp** 和 **FFmpeg** 构建，使用前请确保：

1. **工具配置**：需要自行下载 `yt-dlp.exe` 并放入项目根目录的 `bin/` 文件夹中
2. **浏览器配置**：推荐使用 **Firefox** 浏览器登录相关视频网站
   - yt-dlp 会自动通过 Firefox 获取 Cookie 信息
   - 这样可以确保下载到最高品质的视频内容
3. **浏览器选择**：不推荐使用 Chrome，因为 Cookie 获取配置相对复杂

> 💡 **提示**：如果下载失败，请检查是否已在 Firefox 中登录对应的视频网站

### 下载视频
1. 在输入框中粘贴视频链接
2. 点击「下载视频」按钮
3. 实时查看下载进度
4. 下载完成后自动刷新视频列表

### 管理视频
- **播放视频**：点击播放图标在线预览
- **查看大图**：点击缩略图弹窗查看
- **重命名**：右键菜单选择重命名
- **删除视频**：支持单个删除或批量删除
- **排序**：支持按名称、大小、时间排序

### 批量操作
1. **全选功能**：使用全选复选框一键选择/取消选择所有视频
2. **单项选择**：勾选需要操作的视频
3. **批量删除**：点击「删除选中项」进行批量删除
4. **智能状态**：全选复选框会根据当前选中情况智能显示状态
   - 全部选中时显示为选中状态
   - 部分选中时显示为半选状态
   - 全部未选时显示为未选状态

## 🌐 支持的下载平台

### 📺 视频平台
**综合视频**：YouTube（包括频道、播放列表、直播回放）、Vimeo、Dailymotion、Meta（Facebook、Instagram 的视频内容）、TikTok（支持视频和直播下载）、Twitter（X.com）、Reddit（内嵌视频）、Pinterest（视频内容）。

**影视/流媒体**：Netflix（需特定配置）、Amazon Prime Video（部分内容）、Hulu、Disney+、HBO Max、Paramount+、BBC iPlayer、ITV Hub、Discovery+。

**专业内容**：TED（演讲视频）、Coursera、edX（课程视频）、可汗学院（Khan Academy）。

### 🎵 音乐/音频平台
**音乐视频**：YouTube Music、Vevo、MTV、VEVO。

**音频平台**：SoundCloud、Bandcamp、Spotify（需特定配置，主要下载公开分享内容）、Apple Music（部分可下载内容）。

### 🎮 直播/游戏平台
**直播平台**：Twitch（直播和回放）、YouTube Live、Facebook Live、Periscope（已关闭，但历史内容可能支持）。

**游戏相关**：Steam 社区视频、Epic Games 相关视频、Mixer（已关闭，历史内容）。

### 📰 新闻/资讯平台
CNN、Fox News、NBC News、ABC News、BBC News、Al Jazeera（半岛电视台）、《纽约时报》官网视频、《华盛顿邮报》视频内容。

### 🔗 其他知名站点
**社交/UGC**：Snapchat（公开视频）、LinkedIn（视频内容）、Imgur（视频）。

> 📋 **完整支持列表**：更详细的支持列表请查询 [yt-dlp 官方支持站点列表](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md) <mcreference link="https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md" index="0">0</mcreference>

## 🏗️ 项目结构

```
X-KT 视频下载器/
├── main.go              # 主程序文件
├── go.mod              # Go 模块文件
├── go.sum              # 依赖校验文件
├── README.md           # 项目说明文档
├── bin/                # 外部工具目录
│   ├── ffmpeg.exe      # 视频处理工具
│   ├── ffplay.exe      # 视频播放工具
│   ├── ffprobe.exe     # 视频信息工具
│   └── yt-dlp.exe      # 视频下载工具
├── templates/          # 模板文件目录
│   └── index.html      # 主页面模板
├── static/             # 静态资源目录
├── thumbnails/         # 缩略图存储目录
└── *.mp4              # 下载的视频文件
```

## 📋 版本管理

### 版本信息显示
- **动态版本显示**：Web界面标题会自动显示当前版本号
- **API接口**：通过 `/api/app/info` 接口获取应用信息
- **版本文件**：版本号存储在 `version.txt` 文件中，便于独立管理

### 构建脚本

#### 生产构建
```bash
# Windows
build.bat

# 生成多平台版本：
# - build/X-KT 视频下载器-windows-amd64.exe
# - build/X-KT 视频下载器-linux-amd64
# - build/X-KT 视频下载器-darwin-amd64
```

### 版本更新流程
1. 修改 `version.txt` 文件中的版本号
2. 运行构建脚本重新编译
3. Web界面会自动显示新的版本信息

### 技术实现
- 使用 Go 的 `-ldflags` 参数在编译时注入版本号
- 前端通过 JavaScript 调用 API 动态更新页面标题
- 支持浏览器标题栏和页面内容的同步更新

## 🔧 配置说明

### 服务器配置
- **端口**：默认 8888，可在 `main.go` 中修改
- **存储路径**：视频和缩略图默认存储在项目根目录

### 缩略图配置
程序会根据视频宽高比自动选择最佳的缩略图显示方式：
- 竖屏视频（宽高比 < 0.8）：使用竖向缩略图
- 方形视频（宽高比 0.8-1.3）：使用方形缩略图  
- 横屏视频（宽高比 > 1.3）：使用横向缩略图

## 🚀 高级选项

### 下载设置
- **下载类型**：支持视频+音频、仅视频、仅音频三种模式
- **分离下载**：可选择分别下载视频和音频文件
- **视频分辨率**：支持最佳、720p、1080p、4K等多种分辨率选择
- **音频格式**：支持mp3、m4a、wav等多种音频格式

### 字幕设置
- **字幕下载**：可选择下载视频字幕
- **字幕语言**：支持多语言字幕选择
- **字幕格式**：支持srt、vtt等字幕格式

### 播放列表设置
- **播放列表下载**：支持批量下载播放列表中的视频
- **索引范围**：可指定下载播放列表中特定范围的视频

### 性能优化
- **线程下载**：可配置并发下载线程数（1-8个线程）
- **限速设置**：支持下载速度限制，避免占用过多带宽
- **错误处理**：可配置下载失败时的重试次数和处理方式

### 网络设置
- **Referer设置**：可启用Referer头，解决某些网站的访问限制
- **代理支持**：支持HTTP/HTTPS代理设置

## 🔄 版本自动更新

### yt-dlp版本管理
- **自动检查**：程序启动后自动检查yt-dlp最新版本
- **版本对比**：显示当前版本与最新版本的对比信息
- **一键更新**：支持一键下载并更新yt-dlp到最新版本
- **实时进度**：更新过程中显示实时下载进度
- **安全备份**：更新前自动备份当前版本，失败时可自动恢复

### 更新功能特性
- **智能检测**：自动识别是否已安装yt-dlp
- **断点续传**：支持更新过程中的网络中断恢复
- **取消更新**：更新过程中可随时取消操作
- **版本验证**：更新完成后自动验证新版本是否正常工作

## 🌟 技术亮点

### 智能缩略图系统
- **FFmpeg 优化**：使用 `scale='min(320,iw)':-1` 保持原始宽高比
- **CSS 自适应**：结合 `object-fit: contain` 确保完整显示
- **JavaScript 检测**：动态检测图片宽高比并应用相应样式

### 现代化前端
- **响应式设计**：适配各种屏幕尺寸
- **实时更新**：WebSocket 实现实时状态同步
- **用户体验**：流畅的动画效果和交互反馈
- **版本管理**：内置版本管理和自动更新功能
- **高级配置**：丰富的下载选项和性能优化设置

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - 强大的视频下载工具
- [FFmpeg](https://github.com/yt-dlp/FFmpeg-Builds?tab=readme-ov-file#ffmpeg-static-auto-builds) - 专为视频下载器优化的多媒体处理框架
- [Go](https://golang.org/) - 高效的编程语言

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 [Issue](https://github.com/oskey/VideoDown-Go/issues)
- 访问 [Releases 页面](https://github.com/oskey/VideoDown-Go/releases) 获取最新版本

---

⭐ 如果这个项目对你有帮助，请给它一个 Star！