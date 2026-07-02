# DeepSeek API 连通性测试

## 说明

该脚本用于测试 DeepSeek Chat API 的连通性，通过 curl 发送一个简单的非流式对话请求。

## 用法

```bash
DEEPSEEK_API_KEY=sk-xxx ./docs/dev-notes/deepseek-api-test.sh
```

或通过环境变量文件：

```bash
export DEEPSEEK_API_KEY=sk-xxx
./docs/dev-notes/deepseek-api-test.sh
```

## API 参考

- **端点**: `POST https://api.deepseek.com/chat/completions`
- **模型**: `deepseek-chat`
- **流式**: 不支持（`stream: false`）

## 示例请求

```json
{
  "model": "deepseek-chat",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "stream": false
}
```

## 响应格式

返回标准的 OpenAI-compatible 非流式聊天完成响应。
