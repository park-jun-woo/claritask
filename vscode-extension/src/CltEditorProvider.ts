import * as vscode from 'vscode';
import { Database } from './database';
import { SyncManager } from './sync';

export class CltEditorProvider implements vscode.CustomReadonlyEditorProvider {
  public static register(context: vscode.ExtensionContext): vscode.Disposable {
    const provider = new CltEditorProvider(context);
    return vscode.window.registerCustomEditorProvider(
      'claritask.cltEditor',
      provider,
      {
        webviewOptions: { retainContextWhenHidden: true },
        supportsMultipleEditorsPerDocument: false,
      }
    );
  }

  constructor(private readonly context: vscode.ExtensionContext) {}

  public async openCustomDocument(
    uri: vscode.Uri,
    _openContext: vscode.CustomDocumentOpenContext,
    _token: vscode.CancellationToken
  ): Promise<vscode.CustomDocument> {
    return { uri, dispose: () => {} };
  }

  public async resolveCustomEditor(
    document: vscode.CustomDocument,
    webviewPanel: vscode.WebviewPanel,
    _token: vscode.CancellationToken
  ): Promise<void> {
    webviewPanel.webview.options = {
      enableScripts: true,
      localResourceRoots: [
        vscode.Uri.joinPath(this.context.extensionUri, 'webview-ui', 'dist'),
      ],
    };

    webviewPanel.webview.html = this.getHtmlForWebview(webviewPanel.webview);

    const dbPath = document.uri.fsPath;
    let database: Database | null = null;
    let sync: SyncManager | null = null;

    try {
      database = new Database(dbPath);
      sync = new SyncManager(database, webviewPanel.webview, dbPath);
      sync.start();

      webviewPanel.webview.onDidReceiveMessage(
        (message) => this.handleMessage(message, database!, webviewPanel.webview, sync!),
        undefined,
        []
      );
    } catch (err) {
      webviewPanel.webview.postMessage({
        type: 'error',
        message: `Failed to open database: ${err}`,
      });
    }

    webviewPanel.onDidDispose(() => {
      sync?.stop();
      database?.close();
    });
  }

  private getHtmlForWebview(webview: vscode.Webview): string {
    const scriptUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this.context.extensionUri, 'webview-ui', 'dist', 'index.js')
    );
    const styleUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this.context.extensionUri, 'webview-ui', 'dist', 'index.css')
    );

    const nonce = getNonce();

    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource} 'unsafe-inline'; script-src 'nonce-${nonce}';">
  <link rel="stylesheet" href="${styleUri}">
  <title>Claritask</title>
</head>
<body>
  <div id="root"></div>
  <script nonce="${nonce}" src="${scriptUri}"></script>
</body>
</html>`;
  }

  private handleMessage(
    message: any,
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    switch (message.type) {
      case 'save':
        this.handleSave(message, database, webview, sync);
        break;

      case 'refresh':
        sync.refresh();
        break;

      case 'addEdge':
        this.handleAddEdge(message, database, webview, sync);
        break;

      case 'removeEdge':
        this.handleRemoveEdge(message, database, webview, sync);
        break;

      case 'createTask':
        this.handleCreateTask(message, database, webview, sync);
        break;

      case 'createFeature':
        this.handleCreateFeature(message, database, webview, sync);
        break;
    }
  }

  private handleSave(
    message: { table: string; id: number; data: any; version: number },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    let success = false;

    try {
      if (message.table === 'tasks') {
        success = database.updateTask(message.id, message.data, message.version);
      } else if (message.table === 'features') {
        success = database.updateFeature(message.id, message.data, message.version);
      }

      if (success) {
        webview.postMessage({
          type: 'saveResult',
          success: true,
          table: message.table,
          id: message.id,
        });
        sync.refresh();
      } else {
        webview.postMessage({
          type: 'conflict',
          table: message.table,
          id: message.id,
        });
      }
    } catch (err) {
      webview.postMessage({
        type: 'saveResult',
        success: false,
        error: String(err),
      });
    }
  }

  private handleAddEdge(
    message: { fromId: number; toId: number },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      database.addTaskEdge(message.fromId, message.toId);
      webview.postMessage({
        type: 'edgeResult',
        success: true,
        action: 'add',
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'edgeResult',
        success: false,
        error: String(err),
      });
    }
  }

  private handleRemoveEdge(
    message: { fromId: number; toId: number },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      database.removeTaskEdge(message.fromId, message.toId);
      webview.postMessage({
        type: 'edgeResult',
        success: true,
        action: 'remove',
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'edgeResult',
        success: false,
        error: String(err),
      });
    }
  }

  private handleCreateTask(
    message: { featureId: number; title: string; content: string },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      const id = database.createTask(message.featureId, message.title, message.content);
      webview.postMessage({
        type: 'createResult',
        success: true,
        table: 'tasks',
        id,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'createResult',
        success: false,
        error: String(err),
      });
    }
  }

  private handleCreateFeature(
    message: { name: string; description: string },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      const id = database.createFeature(message.name, message.description);
      webview.postMessage({
        type: 'createResult',
        success: true,
        table: 'features',
        id,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'createResult',
        success: false,
        error: String(err),
      });
    }
  }
}

function getNonce() {
  let text = '';
  const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  for (let i = 0; i < 32; i++) {
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  }
  return text;
}
