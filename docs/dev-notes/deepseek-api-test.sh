#!/usr/bin/env bash
# 用法: DEEPSEEK_API_KEY=sk-xxx ./docs/deepseek.sh

set -euo pipefail

: "${DEEPSEEK_API_KEY:?请设置环境变量 DEEPSEEK_API_KEY}"

curl https://api.deepseek.com/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${DEEPSEEK_API_KEY}" \
  -d '{
        "model": "deepseek-chat",
        "messages": [
          {"role": "system", "content": "You are a helpful assistant."},
          {"role": "user", "content": "Hello!"}
        ],
        "stream": false
      }'
