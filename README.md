# VideoDown-Go 🎥

![VideoDown-Go Screenshot](https://private-user-images.githubusercontent.com/13282035/481362403-6948cc61-83d6-4b4f-90cc-b3d00eb67f4a.png?jwt=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NTYwMjc5NDQsIm5iZiI6MTc1NjAyNzY0NCwicGF0aCI6Ii8xMzI4MjAzNS80ODEzNjI0MDMtNjk0OGNjNjEtODNkNi00YjRmLTkwY2MtYjNkMDBlYjY3ZjRhLnBuZz9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUFLSUFWQ09EWUxTQTUzUFFLNFpBJTJGMjAyNTA4MjQlMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjUwODI0VDA5MjcyNFomWC1BbXotRXhwaXJlcz0zMDAmWC1BbXotU2lnbmF0dXJlPTE0MWI2MjZlNTZlYjViYTY3YzkxZTg5NTVjYjEwY2UyYzFlODJmZjlmY2FlMzczZDM5NTU1YmU0YmJiZDJjZDcmWC1BbXotU2lnbmVkSGVhZGVycz1ob3N0In0.z5JrxvUW8w30iqnk3HPrBljzNOT54oMN2lC7Zqd8XRs)

一个基于 Go 语言开发的现代化视频下载和管理工具，支持多平台视频下载、智能缩略图生成和响应式 Web 界面。

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
1. 勾选需要操作的视频
2. 点击「删除选中项」进行批量删除
3. 支持全选/取消全选操作

## 🏗️ 项目结构

```
VideoDown-Go/
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

## 🔧 配置说明

### 服务器配置
- **端口**：默认 8888，可在 `main.go` 中修改
- **存储路径**：视频和缩略图默认存储在项目根目录

### 缩略图配置
程序会根据视频宽高比自动选择最佳的缩略图显示方式：
- 竖屏视频（宽高比 < 0.8）：使用竖向缩略图
- 方形视频（宽高比 0.8-1.3）：使用方形缩略图  
- 横屏视频（宽高比 > 1.3）：使用横向缩略图

## 🌟 技术亮点

### 智能缩略图系统
- **FFmpeg 优化**：使用 `scale='min(320,iw)':-1` 保持原始宽高比
- **CSS 自适应**：结合 `object-fit: contain` 确保完整显示
- **JavaScript 检测**：动态检测图片宽高比并应用相应样式

### 现代化前端
- **响应式设计**：适配各种屏幕尺寸
- **实时更新**：WebSocket 实现实时状态同步
- **用户体验**：流畅的动画效果和交互反馈

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
- [FFmpeg](https://ffmpeg.org/) - 优秀的多媒体处理框架
- [Go](https://golang.org/) - 高效的编程语言

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 [Issue](https://github.com/oskey/VideoDown-Go/issues)
- 访问 [Releases 页面](https://github.com/oskey/VideoDown-Go/releases) 获取最新版本

---

⭐ 如果这个项目对你有帮助，请给它一个 Star！