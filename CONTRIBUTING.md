# 贡献指南

感谢您对 VideoDown-Go 项目的关注！我们欢迎所有形式的贡献，包括但不限于：

- 🐛 报告 Bug
- 💡 提出新功能建议
- 📝 改进文档
- 🔧 提交代码修复
- 🌟 添加新功能

## 📋 开始之前

在开始贡献之前，请：

1. 阅读我们的 [README.md](README.md)
2. 查看现有的 [Issues](https://github.com/oskey/VideoDown-Go/issues)
3. 确保您的贡献符合项目的目标和范围

## 🐛 报告 Bug

如果您发现了 Bug，请：

1. 检查是否已有相关的 Issue
2. 如果没有，请创建新的 Issue，包含：
   - 清晰的标题和描述
   - 重现步骤
   - 预期行为 vs 实际行为
   - 系统环境信息（操作系统、Go 版本等）
   - 相关的错误日志或截图

### Bug 报告模板

```markdown
**Bug 描述**
简洁清晰地描述这个 Bug。

**重现步骤**
1. 进入 '...'
2. 点击 '....'
3. 滚动到 '....'
4. 看到错误

**预期行为**
描述您期望发生的行为。

**实际行为**
描述实际发生的行为。

**环境信息**
- 操作系统: [例如 Windows 11]
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

再次感谢您的贡献！每一个贡献都让 VideoDown-Go 变得更好。🙏