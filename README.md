# MarsX - 智能 Git 提交助手

**MarsX** 是一个基于 AI 的现代化命令行工具，旨在让 Git 提交变得极致简单、规范且优雅。它能自动分析您的暂存代码 (`git diff --staged`)，生成符合 [Conventional Commits](https://www.conventionalcommits.org/) 规范的提交信息。

![TUI Screenshot](https://via.placeholder.com/800x400.png?text=MarsX+Native+TUI)

> **核心理念**：`git add .` -> `marsx` -> 结束。

## ✨ 核心特性

*   🚀 **极速生成**：自动检测暂存区，一键生成 Commit Message。
*   🧠 **AI 驱动**：支持 OpenAI、DeepSeek 等兼容接口，理解您的代码逻辑。
*   📝 **规范强制**：默认生成 `feat`, `fix`, `docs` 等标准格式（支持自定义 Prompt）。
*   🎨 **原生 TUI**：基于原生终端实现，集成 [Glamour](https://github.com/charmbracelet/glamour) 渲染 Markdown，兼容性更强，响应更快。
*   💬 **对话模式**：不仅仅是提交工具，还是一个懂 Git 的 AI 聊天助手。
*   ⚡ **零依赖**：单文件二进制，跨平台支持 (Windows/Linux/macOS)。

## 📦 安装

### 预编译二进制
请前往 [Releases](https://github.com/ReturnMars/go-aic/releases) 页面下载对应平台的版本，解压后放入 PATH 路径即可。

### 源码编译
```bash
git clone https://github.com/ReturnMars/go-aic.git
cd marsx
go build -o marsx.exe main.go  # Windows
# 或
go build -o marsx main.go      # Linux/macOS
```

## ⚙️ 配置

首次运行前，请在用户主目录或当前目录下创建配置文件 `.marsx.yaml`：

```yaml
# 必填：您的 API Key (OpenAI / DeepSeek / Moonshot)
openai_api_key: "sk-xxxxxxxxxxxxxxxxxxxxxxxx"

# 选填：API 地址 (默认 https://api.openai.com/v1)
# DeepSeek 示例:
openai_api_base: "https://api.deepseek.com"

# 选填：模型名称 (默认 gpt-3.5-turbo)
# DeepSeek 示例:
openai_model: "deepseek-chat"
```

> 您也可以参考仓库中的 `.marsx.example.yaml`。

## 🎮 使用指南

### 1. 极速提交模式 (推荐)

这是 MarsX 最强大的工作流：

```bash
# 1. 暂存您的更改
git add .

# 2. 运行 MarsX
marsx
```

*   MarsX 会自动分析 Diff 并生成提交信息。
*   **交互操作**：
    *   按 `Enter`：直接提交并退出。
    *   按 `e`：进入编辑模式修改消息（`Ctrl+S` 保存）。
    *   按 `Esc` 或 `q`：取消。

#### 常用参数
*   `--quick` / `-q`：快速模式，跳过部分动画或确认步骤。

### 2. 智能辅助模式

如果您没有暂存更改，MarsX 会贴心地提示您：
> "No staged changes found. Press [Enter] to stage all changes (git add .), or type 'n' to skip: "

*   按 `Enter` 或 `y`：自动执行 `git add .` 并开始生成。
*   按 `n` 或其他键：进入**聊天模式**。

### 3. 聊天模式

您可以通过以下两种方式进入聊天模式：

1.  **命令行参数**：
    ```bash
    marsx --chat  # 或 marsx -c
    ```
    直接进入纯对话模式，不进行代码 Diff 分析。

2.  **交互式切换**：
    在生成 Commit Message 的界面，如果输入非空内容，会被视为与 AI 的对话。
    *   `> 怎么撤销上一次 commit？`
    *   `> 解释一下这段代码的逻辑`

### 4. 自定义 Prompt

想让 AI 用日文？或者改变 Emoji 风格？
修改项目根目录下的 `PROMPTS.md` 文件即可。MarsX 会优先读取该文件中的设定。

## 🛠️ 开发

本项目使用 Go 语言开发，核心库包括：
*   **Cobra**: CLI 骨架
*   **Viper**: 配置管理
*   **Glamour**: Markdown 渲染

### 构建发布

使用 [GoReleaser](https://goreleaser.com/) 进行构建：

```bash
# 测试构建 (生成到 dist/ 目录)
goreleaser release --snapshot --clean
```

## 📄 License

MIT
