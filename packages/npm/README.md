# @opentmd/cli

OpenTMD — 开源终端 AI 编码助手。用自然语言描述任务，自动阅读代码、编辑文件、执行命令、验证结果。

## 安装

```bash
npm install -g @opentmd/cli
```

安装完成后即可使用：

```bash
opentmd
```

> npm 会自动下载匹配当前平台的预编译二进制（darwin / linux / windows × arm64/x64）。

## 使用

```bash
# 交互模式（TUI）
opentmd

# 指定项目目录
opentmd -C /path/to/project

# 非交互模式（headless）
opentmd -p "解释这个仓库的架构"

# 继续上次对话
opentmd --continue
```

## 卸载

```bash
npm uninstall -g @opentmd/cli

# 或使用脚本卸载（可保留/删除配置）
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/uninstall.sh | bash
```

## curl 安装（无需 Node.js）

```bash
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/install.sh | bash
```

## 版本对应

npm 版本号与 GitHub Release 一致。详见 [Releases](https://github.com/opentmd/opentmd-cli/releases)。

## 链接

- [源码仓库](https://github.com/opentmd/opentmd-cli)
- [文档](https://github.com/opentmd/opentmd-cli#readme)
