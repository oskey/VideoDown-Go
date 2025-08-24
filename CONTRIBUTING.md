# 贡献指南

感谢您对 X-KT 视频下载器 项目的关注！我们欢迎各种形式的贡献，包括但不限于代码贡献、文档改进、问题报告和功能建议。

## 🤝 贡献方式

### 1. 代码贡献
- 修复 Bug
- 添加新功能
- 性能优化
- 代码重构

### 2. 文档贡献
- 改进现有文档
- 添加使用示例
- 翻译文档
- 编写教程

### 3. 测试贡献
- 报告 Bug
- 测试新功能
- 编写测试用例
- 性能测试

### 4. 社区贡献
- 回答问题
- 分享使用经验
- 推广项目
- 组织活动

## 🚀 开始贡献

### 1. Fork 项目

1. 访问 [X-KT 视频下载器 GitHub 页面](https://github.com/oskey/VideoDown-Go)
2. 点击右上角的 "Fork" 按钮
3. 将项目 Fork 到您的 GitHub 账户

### 2. 克隆项目

```bash
# 克隆您 Fork 的项目
git clone https://github.com/YOUR_USERNAME/VideoDown-Go.git
cd VideoDown-Go

# 添加上游仓库
git remote add upstream https://github.com/oskey/VideoDown-Go.git
```

### 3. 设置开发环境

#### 安装依赖

```bash
# 安装 Go 依赖
go mod download

# 安装开发工具
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

#### 配置外部工具

```bash
# 创建 bin 目录
mkdir -p bin

# 下载 yt-dlp
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o bin/yt-dlp
chmod +x bin/yt-dlp

# 下载 FFmpeg (根据您的系统)
# 详见安装指南
```

### 4. 运行项目

```bash
# 运行开发服务器
go run main.go

# 或使用构建脚本
./build.sh  # Linux/macOS
build.bat   # Windows
```

## 🔧 开发流程

### 1. 创建功能分支

```bash
# 同步上游更改
git fetch upstream
git checkout main
git merge upstream/main

# 创建新分支
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/bug-description
```

### 2. 开发和测试

#### 代码规范

- 使用 `gofmt` 格式化代码
- 使用 `goimports` 管理导入
- 遵循 Go 语言惯例
- 添加适当的注释

```bash
# 格式化代码
go fmt ./...

# 整理导入
goimports -w .

# 运行 linter
golangci-lint run
```

#### 测试

```bash
# 运行测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 3. 提交更改

#### 提交信息规范

使用清晰、描述性的提交信息：

```bash
# 功能添加
git commit -m "feat: add batch download support"

# Bug 修复
git commit -m "fix: resolve memory leak in download manager"

# 文档更新
git commit -m "docs: update installation guide"

# 性能优化
git commit -m "perf: optimize video info extraction"

# 代码重构
git commit -m "refactor: simplify websocket handler"
```

#### 提交类型

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 代码重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

### 4. 推送和创建 PR

```bash
# 推送分支
git push origin feature/your-feature-name
```

然后在 GitHub 上创建 Pull Request。

## 📝 Pull Request 指南

### PR 标题

使用清晰、描述性的标题：
- `feat: Add support for TikTok video download`
- `fix: Fix memory leak in concurrent downloads`
- `docs: Update API documentation`

### PR 描述

请在 PR 描述中包含：

1. **更改摘要**：简要描述您的更改
2. **相关 Issue**：如果相关，请引用 Issue 编号
3. **测试说明**：描述如何测试您的更改
4. **截图**：如果有 UI 更改，请提供截图
5. **破坏性更改**：如果有，请明确说明

#### PR 模板示例

```markdown
## 更改摘要
添加了对 TikTok 视频下载的支持，包括无水印下载功能。

## 相关 Issue
Closes #123

## 更改类型
- [x] 新功能
- [ ] Bug 修复
- [ ] 文档更新
- [ ] 性能优化

## 测试
- [x] 单元测试通过
- [x] 手动测试 TikTok 视频下载
- [x] 测试无水印功能

## 截图
（如果适用）

## 检查清单
- [x] 代码遵循项目规范
- [x] 添加了适当的测试
- [x] 更新了相关文档
- [x] 所有测试通过
```

### 代码审查

我们会仔细审查每个 PR，可能会要求：
- 代码修改
- 添加测试
- 更新文档
- 性能优化

请耐心等待审查，并积极响应反馈。

## 🐛 报告 Bug

### 搜索现有 Issue

在创建新 Issue 之前，请搜索现有 Issue 确认问题未被报告。

### 创建 Bug 报告

请提供以下信息：

1. **Bug 描述**：清晰描述问题
2. **复现步骤**：详细的复现步骤
3. **预期行为**：您期望的正确行为
4. **实际行为**：实际发生的情况
5. **环境信息**：
   - 操作系统和版本
   - Go 版本
   - VideoDown-Go 版本
   - 浏览器版本（如果相关）
6. **错误日志**：相关的错误信息
7. **截图**：如果有助于理解问题

#### Bug 报告模板

```markdown
## Bug 描述
简要描述遇到的问题。

## 复现步骤
1. 打开应用程序
2. 输入视频链接 '...'
3. 点击下载按钮
4. 观察错误

## 预期行为
应该成功下载视频。

## 实际行为
显示错误信息并下载失败。

## 环境信息
- OS: Windows 10
- Go 版本: 1.21.0
- X-KT 视频下载器 版本: v1.0.0
- 浏览器: Firefox 118.0

## 错误日志
```
ERROR: Video unavailable
```

## 截图
（如果适用）
- Go 版本: [例如 1.21.0]
- 浏览器: [例如 Chrome 120]

**附加信息**
添加任何其他有助于解释问题的上下文信息。
```

## 💡 功能建议

我们欢迎新功能建议！请：

1. 检查是否已有相关的功能请求
2. 创建新的 Issue，标记为 "enhancement"
3. 详细描述您的想法和用例
4. 如果可能，提供设计草图或示例

## 🔧 代码贡献

### 开发环境设置

1. Fork 这个仓库
2. 克隆您的 Fork：
   ```bash
   git clone https://github.com/oskey/VideoDown-Go.git
   cd VideoDown-Go
   ```
3. 安装依赖：
   ```bash
   go mod download
   ```
4. 按照 [install.md](install.md) 设置外部工具

### 开发流程

1. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   # 或
   git checkout -b fix/your-bug-fix
   ```

2. **编写代码**
   - 遵循现有的代码风格
   - 添加必要的注释
   - 确保代码可读性

3. **测试**
   ```bash
   # 运行程序测试
   go run main.go
   
   # 测试各项功能
   # - 视频下载
   # - 缩略图生成
   # - 文件管理
   # - 响应式界面
   ```

4. **提交更改**
   ```bash
   git add .
   git commit -m "feat: 添加新功能描述"
   # 或
   git commit -m "fix: 修复某个问题"
   ```

5. **推送分支**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **创建 Pull Request**
   - 提供清晰的标题和描述
   - 说明更改的内容和原因
   - 如果修复了 Issue，请引用相关 Issue 编号

### 提交信息规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

- `feat:` 新功能
- `fix:` Bug 修复
- `docs:` 文档更新
- `style:` 代码格式化（不影响功能）
- `refactor:` 代码重构
- `test:` 添加或修改测试
- `chore:` 构建过程或辅助工具的变动

示例：
```
feat: 添加批量下载功能
fix: 修复缩略图生成失败的问题
docs: 更新安装指南
```

### 代码风格

- 使用 `go fmt` 格式化代码
- 遵循 Go 语言的命名约定
- 为公共函数和复杂逻辑添加注释
- 保持函数简洁，单一职责
- 使用有意义的变量和函数名

### Pull Request 检查清单

在提交 PR 之前，请确保：

- [ ] 代码已经过测试
- [ ] 遵循了项目的代码风格
- [ ] 添加了必要的注释
- [ ] 更新了相关文档（如果需要）
- [ ] 提交信息符合规范
- [ ] 没有引入新的警告或错误

## 📝 文档贡献

文档改进同样重要！您可以：

- 修正拼写错误或语法问题
- 改进现有文档的清晰度
- 添加缺失的文档
- 翻译文档到其他语言

## 🎨 UI/UX 改进

如果您有设计技能，欢迎贡献：

- 改进用户界面设计
- 优化用户体验
- 提供设计建议或原型
- 改进响应式布局

## 🌍 国际化

我们欢迎多语言支持的贡献：

- 翻译界面文本
- 添加新的语言支持
- 改进现有翻译

## 📞 联系我们

如果您有任何问题或需要帮助：

- 创建 [Issue](https://github.com/oskey/VideoDown-Go/issues)
- 参与 [Discussions](https://github.com/oskey/VideoDown-Go/discussions)

## 📜 许可证

通过贡献代码，您同意您的贡献将在 [MIT License](LICENSE) 下发布。

---

再次感谢您的贡献！每一个贡献都让 X-KT 视频下载器 变得更好。🙏