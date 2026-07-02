# Task02 — 安装与首次配置

## 01 运行安装命令

脚本自动安装到 `/usr/local/bin` 或 `~/.local/bin`：

```bash
# 远程安装（需 GitHub Release）
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/install.sh | bash

# 本地源码安装
./scripts/install.sh
```

## 02 配置 Provider

支持 Claude / OpenAI / DeepSeek / GLM / Qwen / Ollama：

```bash
# 交互式选择 Provider 并输入 API Key
opentmd login

# 指定 Provider
opentmd login --provider deepseek

# OpenTMD OAuth（即将上线）
opentmd login --oauth
```

## 03 进入项目运行

```bash
cd your-project
opentmd
```

首次运行 `opentmd` 会进入 **3 步启动向导**：

1. 选择 Provider
2. 输入 API Key（Ollama 跳过）
3. 确认模型

配置保存在 `~/.opentmd/config.toml`。
