import * as http from 'http';

export interface ChatRequest {
  message: string;
  session_id?: string;
  work_dir?: string;
}

export interface LspServerEntry {
  ext: string;
  command: string;
  active: boolean;
  in_path: boolean;
}

export interface LspStatus {
  enabled: boolean;
  auto_detect: boolean;
  settle_delay_ms: number;
  warmup_file_limit: number;
  servers: LspServerEntry[];
  text: string;
}

export interface McpServerEntry {
  name: string;
  status: string;
  tool_count?: number;
  error?: string;
  command?: string;
}

export interface McpStatus {
  servers: McpServerEntry[];
  text: string;
}

export interface ConfigSummary {
  provider: string;
  model: string;
  lsp?: {
    enabled: boolean;
    auto_detect: boolean;
    settle_delay_ms: number;
  };
}

export type ChatEventHandler = (event: string, data: Record<string, string>) => void;

export class DaemonClient {
  constructor(private port: number) {}

  private request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const payload = body ? JSON.stringify(body) : undefined;
    return new Promise((resolve, reject) => {
      const req = http.request({
        hostname: '127.0.0.1',
        port: this.port,
        path,
        method,
        headers: payload
          ? {
              'Content-Type': 'application/json',
              'Content-Length': Buffer.byteLength(payload),
            }
          : {},
        timeout: 30000,
      }, (res) => {
        const chunks: Buffer[] = [];
        res.on('data', (c) => chunks.push(c));
        res.on('end', () => {
          const raw = Buffer.concat(chunks).toString('utf-8');
          if ((res.statusCode ?? 500) >= 400) {
            reject(new Error(`HTTP ${res.statusCode}: ${raw}`));
            return;
          }
          if (!raw) {
            resolve({} as T);
            return;
          }
          try {
            resolve(JSON.parse(raw) as T);
          } catch {
            reject(new Error(raw));
          }
        });
      });
      req.on('error', reject);
      req.on('timeout', () => { req.destroy(); reject(new Error('timeout')); });
      if (payload) req.write(payload);
      req.end();
    });
  }

  health(): Promise<boolean> {
    return new Promise((resolve) => {
      const req = http.get({ hostname: '127.0.0.1', port: this.port, path: '/health', timeout: 3000 }, (res) => {
        resolve((res.statusCode ?? 500) < 400);
      });
      req.on('error', () => resolve(false));
      req.on('timeout', () => { req.destroy(); resolve(false); });
    });
  }

  getConfig(): Promise<ConfigSummary> {
    return this.request<ConfigSummary>('GET', '/config');
  }

  reloadConfig(): Promise<{ status: string; provider: string; model: string }> {
    return this.request('POST', '/config/reload');
  }

  lspStatus(): Promise<LspStatus> {
    return this.request<LspStatus>('GET', '/lsp/status');
  }

  reloadLsp(): Promise<void> {
    return this.request('POST', '/lsp/reload').then(() => undefined);
  }

  mcpStatus(): Promise<McpStatus> {
    return this.request<McpStatus>('GET', '/mcp/status');
  }

  reloadMcp(): Promise<void> {
    return this.request('POST', '/mcp/reload').then(() => undefined);
  }

  newSession(): Promise<string> {
    return this.request<{ session_id: string }>('POST', '/sessions/new').then((r) => r.session_id);
  }

  respondPermission(requestId: string, decision: 'allow_once' | 'allow_session' | 'deny'): Promise<void> {
    return this.request('POST', '/permission/respond', {
      request_id: requestId,
      decision,
    }).then(() => undefined);
  }

  chat(
    message: string,
    onEvent: ChatEventHandler,
    opts?: { sessionId?: string; workDir?: string },
  ): Promise<{ sessionId?: string }> {
    const payload = JSON.stringify({
      message,
      session_id: opts?.sessionId,
      work_dir: opts?.workDir,
    } satisfies ChatRequest);

    let sessionId = opts?.sessionId;

    return new Promise((resolve, reject) => {
      const req = http.request({
        hostname: '127.0.0.1',
        port: this.port,
        path: '/chat',
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(payload),
        },
      }, (res) => {
        if ((res.statusCode ?? 500) >= 400) {
          reject(new Error(`HTTP ${res.statusCode}`));
          return;
        }
        let buffer = '';
        res.on('data', (chunk: Buffer) => {
          buffer += chunk.toString('utf-8');
          const parts = buffer.split('\n\n');
          buffer = parts.pop() ?? '';
          for (const block of parts) {
            const ev = parseSSE(block);
            if (!ev) continue;
            if (ev.event === 'session' && ev.data.session_id) {
              sessionId = ev.data.session_id;
            }
            if (ev.event === 'done' && ev.data.session_id) {
              sessionId = ev.data.session_id;
            }
            onEvent(ev.event, ev.data);
          }
        });
        res.on('end', () => {
          if (buffer.trim()) {
            const ev = parseSSE(buffer);
            if (ev) {
              if (ev.event === 'session' && ev.data.session_id) sessionId = ev.data.session_id;
              if (ev.event === 'done' && ev.data.session_id) sessionId = ev.data.session_id;
              onEvent(ev.event, ev.data);
            }
          }
          resolve({ sessionId });
        });
      });
      req.on('error', reject);
      req.write(payload);
      req.end();
    });
  }
}

function parseSSE(block: string): { event: string; data: Record<string, string> } | null {
  let event = 'message';
  let data = '';
  for (const line of block.split('\n')) {
    if (line.startsWith('event:')) event = line.slice(6).trim();
    if (line.startsWith('data:')) data = line.slice(5).trim();
  }
  if (!data) return null;
  try {
    return { event, data: JSON.parse(data) };
  } catch {
    return { event, data: { content: data } };
  }
}

export function spawnDaemon(binary: string, port: number, workDir?: string): Promise<boolean> {
  return new Promise((resolve) => {
    const { spawn } = require('child_process') as typeof import('child_process');
    const args = ['daemon', '--port', String(port)];
    spawn(binary, args, {
      detached: true,
      stdio: 'ignore',
      cwd: workDir,
    }).unref();
    setTimeout(() => {
      new DaemonClient(port).health().then(resolve);
    }, 800);
  });
}

export function daemonPort(): number {
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  const vscode = require('vscode') as typeof import('vscode');
  return vscode.workspace.getConfiguration('opentmd').get<number>('daemonPort', 13456);
}

export async function withDaemonClient<T>(fn: (client: DaemonClient) => Promise<T>): Promise<T> {
  const client = new DaemonClient(daemonPort());
  if (!(await client.health())) {
    throw new Error('Daemon 未运行，请执行 OpenTMD: Start Daemon');
  }
  return fn(client);
}
