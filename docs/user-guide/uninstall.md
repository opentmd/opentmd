# 卸载指南

## curl / 脚本安装的用户

```bash
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/uninstall.sh | bash
```

卸载脚本按三组资源分类处理：

| 分组 | 内容 | 默认行为 |
|------|------|----------|
| **G1 — 二进制** | `opentmd` 可执行文件、shell rc PATH 行、遗留 `tmd` 软链 | 删除 |
| **G2 — credentials** | `config.toml`、`mcp.json`、`hooks.json` | 保留（删除前自动备份） |
| **G3 — 状态** | `sessions/`、`skills/`、`memory.md` 等 | 删除 |

### 非交互标志

| 标志 | 说明 |
|------|------|
| `--yes` | 跳过交互（G1=yes, G2=no, G3=yes） |
| `--purge` | 删除全部数据（含 credentials） |
| `--keep-data` | 仅删除二进制与 PATH 配置 |
| `--dry-run` | 仅显示计划，不执行 |

示例：

```bash
# 仅删除二进制，保留所有配置和数据
./scripts/uninstall.sh --keep-data --yes

# 完全清除
./scripts/uninstall.sh --purge --yes

# 预览将删除的内容
./scripts/uninstall.sh --dry-run
```

### shell rc 回滚

若安装时写入了 PATH，卸载脚本会自动删除 `# Added by opentmd-cli installer` 标记行，并备份原文件为 `~/.bashrc.opentmd-uninstall.bak`。

## npm 安装的用户

```bash
npm uninstall -g @opentmd/cli
```

npm 卸载仅移除 CLI 二进制与 Node shim，**不会**删除 `~/.opentmd/` 配置目录。如需清除配置，请额外运行卸载脚本：

```bash
curl -fsSL .../uninstall.sh | bash -s -- --purge --yes
```

## 手动卸载

```bash
rm -f ~/.local/bin/opentmd   # 或 /usr/local/bin/opentmd
rm -f ~/.local/bin/tmd       # 遗留短名软链（如有）
rm -rf ~/.opentmd            # 可选：删除全部配置（谨慎）
```
