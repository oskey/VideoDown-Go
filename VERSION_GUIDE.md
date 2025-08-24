# 版本管理指南

## 概述

X-KT 视频下载器现在支持版本管理功能，可以在编译时自动注入版本号和构建时间。

## 版本配置

### 版本文件
- **文件位置**: `version.txt`
- **格式**: 纯文本，一行版本号（如：V1.0.0）
- **用途**: 存储当前版本号，便于版本更新管理

### 页面标题格式
页面标题将自动显示为：`X-KT 视频下载器 V1.0.0 Build Time:20250124`

## 构建方式

### 1. 开发构建（推荐）
```bash
# Windows
.\build-dev.bat

# 或直接使用 go build
go build -ldflags "-X main.Version=V1.0.0 -X main.BuildTime=20250124" -o VideoDown-Go-dev.exe main.go
```

### 2. 生产构建
```bash
# Windows
.\build.bat
```

生产构建会自动：
- 从 `version.txt` 读取版本号
- 自动生成构建时间（YYYYMMDD格式）
- 构建多平台版本（Windows、Linux、macOS）

## 版本更新流程

1. **更新版本号**
   ```bash
   echo V1.0.1 > version.txt
   ```

2. **重新构建**
   ```bash
   .\build-dev.bat  # 开发测试
   # 或
   .\build.bat      # 生产发布
   ```

3. **验证版本**
   - 启动程序后，浏览器标题会显示新的版本信息
   - 可通过 `/api/app/info` API 获取版本信息

## API 接口

### 获取应用信息
- **URL**: `GET /api/app/info`
- **响应格式**:
  ```json
  {
    "name": "X-KT 视频下载器",
    "version": "V1.0.0",
    "buildTime": "20250124",
    "title": "X-KT 视频下载器 V1.0.0 Build Time:20250124"
  }
  ```

## 技术实现

### 编译时变量注入
使用 Go 的 `-ldflags` 参数在编译时注入版本信息：
```bash
-ldflags "-X main.Version=V1.0.0 -X main.BuildTime=20250124"
```

### 前端集成
- 页面加载时自动调用 `/api/app/info` 获取版本信息
- 动态更新浏览器标题
- 版本信息存储在全局变量中，便于其他功能使用

## 注意事项

1. **版本文件格式**: `version.txt` 必须是纯文本格式，只包含版本号
2. **构建时间格式**: 自动生成为 YYYYMMDD 格式（如：20250124）
3. **编码问题**: 如果批处理文件在 PowerShell 中出现编码问题，可直接使用 `go build` 命令
4. **版本一致性**: 确保 `version.txt` 中的版本号与实际发布版本保持一致

## 示例

假设当前版本为 V1.0.0，构建时间为 2025年1月24日：
- 浏览器标题：`X-KT 视频下载器 V1.0.0 Build Time:20250124`
- API 返回的版本信息包含完整的应用元数据
- 便于用户和开发者识别当前运行的版本