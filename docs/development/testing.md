# 测试指南

## 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/config/...
go test ./internal/deepseek/...
go test ./internal/session/...

# 详细输出
go test -v ./...

# 查看覆盖率
go test -cover ./...
```

## 测试文件

| 文件 | 测试内容 |
|------|----------|
| `internal/config/config_test.go` | 配置读写、Provider 预设 |
| `internal/deepseek/deepseek_test.go` | SSE 流式响应解析 |
| `internal/session/session_test.go` | 会话创建、消息管理 |

### DeepSeek 流式解析测试

`internal/deepseek/deepseek_test.go` 测试 `ParseSSEData()` 函数：

```go
func TestParseSSEData(t *testing.T) {
    // 测试内容 chunk
    content, done, err := ParseSSEData(`{"choices":[{"delta":{"content":"Hello"}}]}`)
    assert content == "Hello"
    assert done == false

    // 测试结束标记
    _, done, err = ParseSSEData("[DONE]")
    assert done == true

    // 测试 finish_reason
    content, done, err = ParseSSEData(`{"choices":[{"delta":{},"finish_reason":"stop"}]}`)
    assert done == true
}
```

## 测试注意事项

1. API 调用需要有效的 API Key，单元测试不发起真实网络请求
2. Session 测试使用临时目录，避免影响真实会话数据
3. 所有测试使用 Go 标准 `testing` 包
