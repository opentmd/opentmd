import * as vscode from 'vscode';
import { LspStatus, withDaemonClient } from './daemonClient';

type LspTreeNode =
  | { kind: 'section'; label: string }
  | { kind: 'setting'; label: string; value: string }
  | { kind: 'server'; ext: string; command: string; active: boolean; inPath: boolean }
  | { kind: 'message'; label: string };

export class LspTreeProvider implements vscode.TreeDataProvider<LspTreeNode> {
  private readonly onChangeEmitter = new vscode.EventEmitter<void>();
  readonly onDidChangeTreeData = this.onChangeEmitter.event;

  private status?: LspStatus;
  private error?: string;

  refresh(): void {
    this.onChangeEmitter.fire();
  }

  async load(): Promise<void> {
    this.error = undefined;
    try {
      this.status = await withDaemonClient((client) => client.lspStatus());
    } catch (err) {
      this.status = undefined;
      this.error = err instanceof Error ? err.message : String(err);
    }
    this.refresh();
  }

  getTreeItem(element: LspTreeNode): vscode.TreeItem {
    switch (element.kind) {
      case 'section':
        return {
          label: element.label,
          collapsibleState: vscode.TreeItemCollapsibleState.Expanded,
        };
      case 'setting':
        return {
          label: element.label,
          description: element.value,
          collapsibleState: vscode.TreeItemCollapsibleState.None,
        };
      case 'server': {
        const icon = element.active
          ? new vscode.ThemeIcon('debug-start', new vscode.ThemeColor('testing.iconPassed'))
          : element.inPath
            ? new vscode.ThemeIcon('circle-outline')
            : new vscode.ThemeIcon('warning', new vscode.ThemeColor('problemsWarningIcon.foreground'));
        const state = element.active ? 'running' : element.inPath ? 'idle' : 'missing';
        return {
          label: `${element.command} (.${element.ext})`,
          description: state,
          iconPath: icon,
          collapsibleState: vscode.TreeItemCollapsibleState.None,
        };
      }
      default:
        return {
          label: element.label,
          collapsibleState: vscode.TreeItemCollapsibleState.None,
        };
    }
  }

  getChildren(element?: LspTreeNode): LspTreeNode[] {
    if (element) {
      return [];
    }
    if (this.error) {
      return [{ kind: 'message', label: this.error }];
    }
    if (!this.status) {
      return [{ kind: 'message', label: '加载中…' }];
    }

    const st = this.status;
    const nodes: LspTreeNode[] = [
      { kind: 'section', label: 'Settings' },
      { kind: 'setting', label: 'enabled', value: String(st.enabled) },
      { kind: 'setting', label: 'auto_detect', value: String(st.auto_detect) },
      { kind: 'setting', label: 'settle_ms', value: String(st.settle_delay_ms) },
      { kind: 'setting', label: 'warmup', value: String(st.warmup_file_limit) },
    ];

    nodes.push({ kind: 'section', label: 'Servers' });
    if (st.servers.length === 0) {
      nodes.push({ kind: 'message', label: st.enabled ? '无已配置服务器' : 'LSP 已禁用' });
    } else {
      for (const s of st.servers) {
        nodes.push({
          kind: 'server',
          ext: s.ext,
          command: s.command,
          active: s.active,
          inPath: s.in_path,
        });
      }
    }
    return nodes;
  }
}

export function registerLspView(context: vscode.ExtensionContext): LspTreeProvider {
  const provider = new LspTreeProvider();
  context.subscriptions.push(
    vscode.window.registerTreeDataProvider('opentmd.lspView', provider),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.refreshLsp', async () => {
      await provider.load();
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.reloadLsp', async () => {
      try {
        await withDaemonClient((client) => client.reloadLsp());
        vscode.window.showInformationMessage('OpenTMD LSP 已重启');
        await provider.load();
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        vscode.window.showErrorMessage(message);
      }
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.reloadConfig', async () => {
      try {
        const result = await withDaemonClient((client) => client.reloadConfig());
        vscode.window.showInformationMessage(
          `OpenTMD 配置已重载: ${result.provider} / ${result.model}`,
        );
        await provider.load();
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        vscode.window.showErrorMessage(message);
      }
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.showLspStatus', async () => {
      try {
        const st = await withDaemonClient((client) => client.lspStatus());
        const doc = await vscode.workspace.openTextDocument({
          content: st.text,
          language: 'plaintext',
        });
        await vscode.window.showTextDocument(doc, { preview: true });
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        vscode.window.showErrorMessage(message);
      }
    }),
  );

  void provider.load();
  return provider;
}
