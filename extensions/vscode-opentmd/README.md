# OpenTMD VS Code Extension

Minimal VS Code integration for OpenTMD daemon.

## Setup

1. Build and install CLI: `./scripts/build.sh && ./scripts/install.sh`
2. Compile extension: `cd extensions/vscode-opentmd && npm install && npm run compile`
3. Press F5 in VS Code to launch Extension Development Host

## Usage

- **OpenTMD: Start Daemon** — starts `opentmd daemon --port 13456`
- **OpenTMD: Open Chat** — sidebar chat connected to daemon SSE `/chat`
- **OpenTMD: Ask About Selection** — sends selected code to daemon
- **LSP sidebar** — tree view showing language server status (running / idle / missing)
- **MCP sidebar** — tree view showing MCP server connection status
- **Chat LSP events** — SSE `lsp_connect` events shown in chat panel during tool runs
- **OpenTMD: Reload Config** — `POST /config/reload` hot-reload from `~/.opentmd/config.toml`
- **OpenTMD: Reload LSP** — restart LSP clients via daemon

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `opentmd.daemonPort` | 13456 | Daemon HTTP port |
| `opentmd.binaryPath` | opentmd | Path to CLI binary |

## Daemon endpoints used

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Check daemon is running |
| `POST /chat` | SSE chat |
| `GET /lsp/status` | LSP tree view |
| `POST /lsp/reload` | Restart LSP |
| `GET /mcp/status` | MCP tree view |
| `POST /mcp/reload` | Reload MCP |
| `POST /config/reload` | Reload config.toml |
