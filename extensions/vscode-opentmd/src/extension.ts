import * as vscode from 'vscode';
import { DaemonClient, spawnDaemon } from './daemonClient';
import { registerLspView } from './lspView';
import { registerMcpView } from './mcpView';

class ChatViewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'opentmd.chatView';
  private sessionId?: string;
  private view?: vscode.WebviewView;

  constructor(private readonly ext: vscode.ExtensionContext) {
    this.sessionId = ext.globalState.get<string>('sessionId');
  }

  resolveWebviewView(webviewView: vscode.WebviewView): void {
    this.view = webviewView;
    webviewView.webview.options = { enableScripts: true };
    webviewView.webview.html = this.html(this.sessionId);

    webviewView.webview.onDidReceiveMessage(async (msg) => {
      if (msg.type === 'new_session') {
        await this.newSession(webviewView);
        return;
      }
      if (msg.type !== 'chat') return;

      const port = vscode.workspace.getConfiguration('opentmd').get<number>('daemonPort', 13456);
      const workDir = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      const client = new DaemonClient(port);

      if (!(await client.health())) {
        webviewView.webview.postMessage({ type: 'error', text: 'Daemon 未运行，请执行 OpenTMD: Start Daemon' });
        return;
      }

      try {
        const result = await client.chat(
          msg.text,
          async (event, data) => {
            if (event === 'text' && data.content) {
              webviewView.webview.postMessage({ type: 'delta', text: data.content });
            }
            if (event === 'tool_start' && data.status) {
              webviewView.webview.postMessage({ type: 'status', text: data.status });
            }
            if (event === 'lsp_connect') {
              const line = data.status === 'started'
                ? `✓ LSP ${data.command} started (.${data.ext})`
                : `× LSP ${data.command} failed (.${data.ext}): ${data.error ?? ''}`;
              webviewView.webview.postMessage({ type: 'lsp', text: line });
            }
            if (event === 'permission_request') {
              const decision = await this.askPermission(data);
              await client.respondPermission(data.request_id, decision);
              webviewView.webview.postMessage({
                type: 'status',
                text: decision === 'deny' ? '已拒绝权限' : '已批准: ' + data.tool,
              });
            }
            if (event === 'error' && data.message) {
              webviewView.webview.postMessage({ type: 'error', text: data.message });
            }
            if (event === 'done') {
              webviewView.webview.postMessage({ type: 'done' });
            }
          },
          { sessionId: this.sessionId, workDir },
        );
        if (result.sessionId) {
          this.sessionId = result.sessionId;
          await this.ext.globalState.update('sessionId', result.sessionId);
          webviewView.webview.postMessage({ type: 'session', id: result.sessionId.slice(0, 8) });
        }
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        webviewView.webview.postMessage({ type: 'error', text: message });
      }
    });
  }

  private async newSession(webviewView: vscode.WebviewView): Promise<void> {
    const port = vscode.workspace.getConfiguration('opentmd').get<number>('daemonPort', 13456);
    const client = new DaemonClient(port);
    if (!(await client.health())) {
      webviewView.webview.postMessage({ type: 'error', text: 'Daemon 未运行' });
      return;
    }
    this.sessionId = await client.newSession();
    await this.ext.globalState.update('sessionId', this.sessionId);
    webviewView.webview.postMessage({ type: 'clear' });
    webviewView.webview.postMessage({ type: 'session', id: this.sessionId.slice(0, 8) });
  }

  private async askPermission(data: Record<string, string>): Promise<'allow_once' | 'allow_session' | 'deny'> {
    const tool = data.tool ?? 'tool';
    const reason = data.reason ? `\n${data.reason}` : '';
    const args = data.args ? `\n${data.args.slice(0, 200)}` : '';
    const choice = await vscode.window.showWarningMessage(
      `OpenTMD 请求权限: ${tool}${reason}${args}`,
      { modal: true },
      '允许一次',
      '本会话允许',
      '拒绝',
    );
    switch (choice) {
      case '允许一次': return 'allow_once';
      case '本会话允许': return 'allow_session';
      default: return 'deny';
    }
  }

  prefill(text: string): void {
    this.view?.webview.postMessage({ type: 'prefill', text });
  }

  private html(sessionShort?: string): string {
    const sid = sessionShort ? sessionShort.slice(0, 8) : '—';
    return `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>
body{font-family:var(--vscode-font-family);color:var(--vscode-foreground);background:var(--vscode-editor-background);margin:0;padding:8px;display:flex;flex-direction:column;height:100vh;box-sizing:border-box}
.toolbar{display:flex;gap:8px;align-items:center;margin-bottom:6px;font-size:12px;color:var(--vscode-descriptionForeground)}
.toolbar button{font-size:11px;padding:2px 8px}
#log{flex:1;overflow:auto;white-space:pre-wrap;font-size:13px;line-height:1.5;margin-bottom:8px}
#input{display:flex;gap:6px}
textarea{flex:1;resize:none;height:64px;background:var(--vscode-input-background);color:var(--vscode-input-foreground);border:1px solid var(--vscode-input-border);padding:6px}
button{padding:6px 12px}
.status{color:var(--vscode-descriptionForeground);font-size:12px;margin:4px 0}
.lsp{color:var(--vscode-charts-green);font-size:12px;margin:2px 0}
.lsp.err{color:var(--vscode-errorForeground)}
.user{color:var(--vscode-textLink-foreground)}
</style></head><body>
<div class="toolbar">session: <span id="sid">${sid}</span> <button id="new">新会话</button></div>
<div id="log"></div>
<div id="lspLog"></div>
<div class="status" id="status"></div>
<div id="input"><textarea id="msg" placeholder="Ask OpenTMD…"></textarea><button id="send">Send</button></div>
<script>
const vscode = acquireVsCodeApi();
const log = document.getElementById('log');
const lspLog = document.getElementById('lspLog');
const status = document.getElementById('status');
const msg = document.getElementById('msg');
const sid = document.getElementById('sid');
function sendChat() {
  const text = msg.value.trim();
  if (!text) return;
  log.innerHTML += '<div class="user">\\n> ' + text + '</div>\\n';
  status.textContent = 'Thinking…';
  vscode.postMessage({ type: 'chat', text });
  msg.value = '';
}
document.getElementById('send').onclick = sendChat;
document.getElementById('new').onclick = () => vscode.postMessage({ type: 'new_session' });
msg.addEventListener('keydown', e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendChat(); }});
window.addEventListener('message', e => {
  const m = e.data;
  if (m.type === 'delta') log.textContent += m.text;
  if (m.type === 'status') status.textContent = m.text;
  if (m.type === 'lsp') {
    const div = document.createElement('div');
    div.className = 'lsp' + (String(m.text).startsWith('×') ? ' err' : '');
    div.textContent = m.text;
    lspLog.appendChild(div);
  }
  if (m.type === 'error') { status.textContent = m.text; log.textContent += '\\n[error] ' + m.text; }
  if (m.type === 'done') status.textContent = '';
  if (m.type === 'session') sid.textContent = m.id;
  if (m.type === 'clear') { log.textContent = ''; lspLog.textContent = ''; }
  if (m.type === 'prefill') { msg.value = m.text || ''; msg.focus(); }
});
</script></body></html>`;
  }
}

let chatProvider: ChatViewProvider | undefined;

export function activate(context: vscode.ExtensionContext): void {
  chatProvider = new ChatViewProvider(context);
  registerLspView(context);
  registerMcpView(context);
  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider(ChatViewProvider.viewType, chatProvider),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.openChat', () => {
      vscode.commands.executeCommand('opentmd.chatView.focus');
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.startDaemon', async () => {
      const cfg = vscode.workspace.getConfiguration('opentmd');
      const port = cfg.get<number>('daemonPort', 13456);
      const binary = cfg.get<string>('binaryPath', 'opentmd');
      const workDir = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      const client = new DaemonClient(port);
      if (await client.health()) {
        vscode.window.showInformationMessage(`OpenTMD daemon 已在端口 ${port} 运行`);
        return;
      }
      const ok = await spawnDaemon(binary, port, workDir);
      if (ok) {
        vscode.window.showInformationMessage(`OpenTMD daemon 已启动 (port ${port})`);
      } else {
        vscode.window.showErrorMessage('启动 daemon 失败，请确认 opentmd 在 PATH 中');
      }
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.newConversation', async () => {
      const port = vscode.workspace.getConfiguration('opentmd').get<number>('daemonPort', 13456);
      const client = new DaemonClient(port);
      if (!(await client.health())) {
        vscode.window.showErrorMessage('Daemon 未运行');
        return;
      }
      await client.newSession();
      vscode.window.showInformationMessage('已创建新会话');
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.sendSelection', async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor || editor.selection.isEmpty) {
        vscode.window.showWarningMessage('请先选中代码');
        return;
      }
      const text = editor.document.getText(editor.selection);
      const lang = editor.document.languageId;
      await vscode.commands.executeCommand('opentmd.chatView.focus');
      // Prefill via command that posts to webview - use executeCommand workaround
      vscode.commands.executeCommand('opentmd.openChat');
      const prompt = `Explain this ${lang} code:\n\`\`\`${lang}\n${text}\n\`\`\``;
      chatProvider?.prefill(prompt);
    }),
  );
}

export function deactivate(): void {}
