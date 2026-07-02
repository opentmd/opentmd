import * as vscode from 'vscode';
import { McpStatus, withDaemonClient } from './daemonClient';

type McpTreeNode =
  | { kind: 'section'; label: string }
  | { kind: 'server'; name: string; status: string; toolCount?: number; error?: string; command?: string }
  | { kind: 'message'; label: string };

export class McpTreeProvider implements vscode.TreeDataProvider<McpTreeNode> {
  private readonly onChangeEmitter = new vscode.EventEmitter<void>();
  readonly onDidChangeTreeData = this.onChangeEmitter.event;

  private status?: McpStatus;
  private error?: string;

  refresh(): void {
    this.onChangeEmitter.fire();
  }

  async load(): Promise<void> {
    this.error = undefined;
    try {
      this.status = await withDaemonClient((client) => client.mcpStatus());
    } catch (err) {
      this.status = undefined;
      this.error = err instanceof Error ? err.message : String(err);
    }
    this.refresh();
  }

  getTreeItem(element: McpTreeNode): vscode.TreeItem {
    switch (element.kind) {
      case 'section':
        return {
          label: element.label,
          collapsibleState: vscode.TreeItemCollapsibleState.Expanded,
        };
      case 'server': {
        const icon = element.status === 'connected'
          ? new vscode.ThemeIcon('plug', new vscode.ThemeColor('testing.iconPassed'))
          : element.status === 'error'
            ? new vscode.ThemeIcon('error', new vscode.ThemeColor('problemsErrorIcon.foreground'))
            : new vscode.ThemeIcon('circle-outline');
        const desc = element.status === 'connected'
          ? `${element.toolCount ?? 0} tools`
          : element.error ?? element.status;
        return {
          label: element.name,
          description: desc,
          tooltip: element.command,
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

  getChildren(element?: McpTreeNode): McpTreeNode[] {
    if (element) {
      return [];
    }
    if (this.error) {
      return [{ kind: 'message', label: this.error }];
    }
    if (!this.status) {
      return [{ kind: 'message', label: '加载中…' }];
    }

    const nodes: McpTreeNode[] = [{ kind: 'section', label: 'Servers' }];
    if (this.status.servers.length === 0) {
      nodes.push({ kind: 'message', label: '无 MCP 服务器配置' });
      return nodes;
    }
    for (const s of this.status.servers) {
      nodes.push({
        kind: 'server',
        name: s.name,
        status: s.status,
        toolCount: s.tool_count,
        error: s.error,
        command: s.command,
      });
    }
    return nodes;
  }
}

export function registerMcpView(context: vscode.ExtensionContext): McpTreeProvider {
  const provider = new McpTreeProvider();
  context.subscriptions.push(
    vscode.window.registerTreeDataProvider('opentmd.mcpView', provider),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.refreshMcp', async () => {
      await provider.load();
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.reloadMcp', async () => {
      try {
        await withDaemonClient((client) => client.reloadMcp());
        vscode.window.showInformationMessage('OpenTMD MCP 已重载');
        await provider.load();
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        vscode.window.showErrorMessage(message);
      }
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('opentmd.showMcpStatus', async () => {
      try {
        const st = await withDaemonClient((client) => client.mcpStatus());
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
