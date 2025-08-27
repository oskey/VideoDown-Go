# GitHub Actions 自动构建和发布指南

## 📋 概述

本项目已配置了完整的 GitHub Actions 工作流，支持：
- ✅ 自动多平台编译（Windows、Linux、macOS）
- ✅ 自动标签发布
- ✅ 自动生成 Release 页面
- ✅ 版本管理和更新日志

## 🚀 如何使用

### 1. 日常开发流程

当你推送代码到 `main` 或 `master` 分支时：
```bash
git add .
git commit -m "feat: 添加新功能"
git push origin main
```

**触发结果：**
- 自动运行测试
- 构建应用程序
- 生成多平台二进制文件（仅作为构建产物保存）

### 2. 发布新版本

#### 🚀 方法一：自动标签发布（推荐）

只需两步即可完成版本发布：

```bash
# 步骤 1：更新版本号
echo "V1.4.0" > version.txt

# 步骤 2：提交并推送
git add version.txt
git commit -m "chore: 发布版本 V1.4.0"
git push origin main
```

**✨ 自动执行：**
- 🔍 自动检测 `version.txt` 变化
- 📝 自动更新 README.md 版本号
- 💾 提交 README.md 更改
- 🏷️ 自动创建标签 `V1.4.0`
- 🚀 自动触发发布流程

#### 📋 方法二：手动标签发布（传统方式）

如果你喜欢手动控制：

```bash
# 步骤 1：更新版本号
echo "V1.4.0" > version.txt
git add version.txt
git commit -m "chore: 发布版本 V1.4.0"
git push origin main

# 步骤 2：手动创建标签
git tag V1.4.0
git push origin V1.4.0
```

**自动触发结果：**
- 🔄 自动构建 6 个平台的二进制文件
- 📦 自动创建 GitHub Release
- 📝 自动生成更新日志
- 🔗 自动添加下载链接

### 3. 支持的平台

| 平台 | 架构 | 文件名示例 |
|------|------|------------|
| Windows | x64 | `VideoDown-Go-V1.4.0-windows-amd64.exe` |
| Windows | ARM64 | `VideoDown-Go-V1.4.0-windows-arm64.exe` |
| Linux | x64 | `VideoDown-Go-V1.4.0-linux-amd64` |
| Linux | ARM64 | `VideoDown-Go-V1.4.0-linux-arm64` |
| macOS | x64 | `VideoDown-Go-V1.4.0-darwin-amd64` |
| macOS | ARM64 | `VideoDown-Go-V1.4.0-darwin-arm64` |

## ✨ 新功能：自动标签创建

### 🎯 核心优势

- **零手动操作**：只需修改 `version.txt` 即可完成发布
- **智能检测**：自动识别版本文件变化
- **防重复创建**：智能跳过已存在的标签
- **完整日志**：详细的执行日志便于调试

### 🔍 工作原理

1. **文件变化检测**
   ```bash
   git diff --name-only HEAD~1 HEAD | grep -q "version.txt"
   ```

2. **版本号提取**
   ```bash
   NEW_VERSION=$(cat version.txt)
   ```

3. **重复检查**
   ```bash
   git tag -l | grep -q "^$NEW_VERSION$"
   ```

4. **自动标签创建**
   ```bash
   git tag -a "$VERSION" -m "Auto-generated tag for version $VERSION"
   git push origin "$VERSION"
   ```

## 📝 README.md 自动版本更新

### 功能特性
- **智能匹配**：自动识别并更新 README.md 中的版本号
- **多位置更新**：同时更新标题和章节中的版本信息
- **格式保持**：保持原有的文档格式和样式
- **自动提交**：更新后自动提交并推送更改

### 更新范围
1. **主标题版本**：`# X-KT 视频下载器 🎥 v1.3.2` → `# X-KT 视频下载器 🎥 V1.4.0`
2. **BUG修复章节**：`## 🐛 v1.3.2 BUG修复` → `## 🐛 V1.4.0 BUG修复`
3. **新功能章节**：`## 🆕 v1.3.0 新功能` → `## 🆕 V1.4.0 新功能`
4. **更新内容章节**：`## 🔄 v1.2.0 更新内容` → `## 🔄 V1.4.0 更新内容`

### 版本号格式支持
- ✅ `V1.4.0` (推荐)
- ✅ `v1.4.0`
- ✅ `1.4.0`

## 🔧 工作流详解

### 触发条件

1. **构建测试**（每次推送到主分支）
   ```yaml
   on:
     push:
       branches: [ "main", "master" ]
     pull_request:
       branches: [ "main", "master" ]
   ```

2. **自动发布**（推送标签时）
   ```yaml
   on:
     push:
       tags:
         - 'v*.*.*'  # 如：v1.4.0
         - 'V*.*.*'  # 如：V1.4.0
   ```

### 工作流程

1. **build-and-test** 任务
   - 检出代码
   - 设置 Go 1.21 环境
   - 缓存依赖
   - 运行测试
   - 基础构建验证

2. **build-matrix** 任务
   - 多平台并行构建
   - 版本号注入
   - 上传构建产物

3. **release** 任务（仅标签触发）
   - 下载所有构建产物
   - 生成更新日志
   - 创建 GitHub Release

## 🛠️ 高级配置

### 自定义版本号

版本号获取优先级：
1. Git 标签（如果是标签推送）
2. `version.txt` 文件内容
3. `dev-{git-hash}` 格式

### 修改构建参数

在 `.github/workflows/go.yml` 中可以调整：

```yaml
# 修改 Go 版本
env:
  GO_VERSION: '1.21'
  APP_NAME: 'VideoDown-Go'

# 添加构建标志
run: |
  go build -ldflags "-s -w -X main.Version=${{ steps.version.outputs.version }}" -o "$BINARY_NAME" .
```

### 添加新平台

在 `build-matrix` 的 `strategy.matrix.include` 中添加：

```yaml
- goos: freebsd
  goarch: amd64
  suffix: ''
```

## 🐛 常见问题

### 1. Actions 失败原因

**权限问题：**
- 确保仓库设置中启用了 Actions
- 检查 `GITHUB_TOKEN` 权限

**构建失败：**
- 检查 Go 版本兼容性
- 确认依赖项完整
- 查看构建日志中的具体错误

**发布失败：**
- 确认标签格式正确（v1.0.0 或 V1.0.0）
- 检查是否有重复的标签

### 2. 手动触发构建

如果需要手动触发，可以在 GitHub 仓库页面：
1. 进入 "Actions" 标签页
2. 选择 "Build and Release" 工作流
3. 点击 "Run workflow"

### 3. 查看构建状态

- 📊 在仓库主页可以看到构建状态徽章
- 📝 在 Actions 页面查看详细日志
- 📦 在 Releases 页面查看发布版本

## 📚 相关链接

- [README.md 版本自动更新指南](README_VERSION_AUTO_UPDATE.md)
- [自动版本标签详细指南](AUTO_VERSION_GUIDE.md)
- [GitHub Actions 版本修复说明](ACTIONS_VERSION_FIX.md)
- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [Go 交叉编译指南](https://golang.org/doc/install/source#environment)
- [语义化版本规范](https://semver.org/lang/zh-CN/)

## 🎯 最佳实践

1. **版本号管理**
   - 使用语义化版本（如 V1.2.3）
   - 主版本号.次版本号.修订号

2. **提交信息**
   - 使用清晰的提交信息
   - 遵循约定式提交规范

3. **发布频率**
   - 功能完善后再发布
   - 重要修复及时发布

4. **测试覆盖**
   - 发布前确保测试通过
   - 添加必要的单元测试

---

**注意：** 首次使用时，请确保你的 GitHub 仓库已启用 Actions 功能，并且具有适当的权限设置。