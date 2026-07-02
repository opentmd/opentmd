# 设置向导

## 概述

设置向导位于 `internal/setup/`，提供首次启动的交互式配置向导（3 步）和 Provider 登录功能。

## 首次启动向导

首次运行 `opentmd` 时自动触发（需终端交互）：

### Step 1: 选择 Provider

```
  ┌──────────────────────────────────────┐
  │       OpenTMD 首次启动向导 (1/3)      │
  └──────────────────────────────────────┘

  欢迎使用 OpenTMD — 终端 AI 编码助手
  请选择大模型 Provider：

    1) DeepSeek (deepseek)
    2) OpenAI (openai)
    3) Claude (via OpenRouter) (claude)
    4) GLM (智谱) (glm)
    5) Qwen (通义千问) (qwen)
    6) Ollama (本地) (ollama)
    0) 退出

  请输入序号 [1]:
```

### Step 2: 输入 API Key

```
  ┌──────────────────────────────────────┐
  │       OpenTMD 首次启动向导 (2/3)      │
  └──────────────────────────────────────┘

  Provider: DeepSeek
  API 地址: https://api.deepseek.com

  请输入 API Key（输入不可见）：
```

Ollama 跳过此步（无需 API Key）。

### Step 3: 确认模型

```
  ┌──────────────────────────────────────┐
  │       OpenTMD 首次启动向导 (3/3)      │
  └──────────────────────────────────────┘

  模型: deepseek-chat（直接回车确认）
  或输入自定义模型名:
```

## 登录命令

### `opentmd login`

交互式配置 Provider 和 API Key。

```bash
opentmd login                     # 交互式选择
opentmd login --provider deepseek # 直接指定 Provider
opentmd login --oauth             # OpenTMD OAuth（即将上线）
```

## OAuth

`RunOAuth()` 当前为占位实现，提示即将上线。

## 触发条件

`config.NeedsWizard()` 判断：

```go
func (c *Config) NeedsWizard() bool {
    if c.Setup.Completed && c.HasConfiguredProvider() {
        return false
    }
    return !c.HasConfiguredProvider()
}
```

首次运行或 Provider 未配置时触发。
