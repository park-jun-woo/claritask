# TASK-EXT-002: Extension 진입점 및 Custom Editor Provider

## 목표
VSCode Extension 활성화 및 Custom Editor Provider 구현.

## 파일

### 1. src/extension.ts

Extension 진입점. `.clt` 파일 열 때 Custom Editor 활성화.

```typescript
import * as vscode from 'vscode';
import { CltEditorProvider } from './CltEditorProvider';

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    CltEditorProvider.register(context)
  );
}

export function deactivate() {}
```

### 2. src/CltEditorProvider.ts

Custom Editor Provider 구현.

```typescript
import * as vscode from 'vscode';
import { Database } from './database';
import { SyncManager } from './sync';

export class CltEditorProvider implements vscode.CustomTextEditorProvider {
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

  public async resolveCustomTextEditor(
    document: vscode.TextDocument,
    webviewPanel: vscode.WebviewPanel,
    _token: vscode.CancellationToken
  ): Promise<void> {
    // Webview 설정
    webviewPanel.webview.options = {
      enableScripts: true,
      localResourceRoots: [
        vscode.Uri.joinPath(this.context.extensionUri, 'webview-ui', 'dist')
      ]
    };

    // HTML 로드
    webviewPanel.webview.html = this.getHtmlForWebview(webviewPanel.webview);

    // Database 연결
    const dbPath = document.uri.fsPath;
    const database = new Database(dbPath);

    // Sync 시작
    const sync = new SyncManager(database, webviewPanel.webview);
    sync.start();

    // 메시지 핸들러
    webviewPanel.webview.onDidReceiveMessage(
      (message) => this.handleMessage(message, database, webviewPanel.webview),
      undefined,
      []
    );

    // 정리
    webviewPanel.onDidDispose(() => {
      sync.stop();
      database.close();
    });
  }

  private getHtmlForWebview(webview: vscode.Webview): string {
    const scriptUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this.context.extensionUri, 'webview-ui', 'dist', 'index.js')
    );
    const styleUri = webview.asWebviewUri(
      vscode.Uri.joinPath(this.context.extensionUri, 'webview-ui', 'dist', 'index.css')
    );

    return `<!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <link rel="stylesheet" href="${styleUri}">
      <title>Claritask</title>
    </head>
    <body>
      <div id="root"></div>
      <script src="${scriptUri}"></script>
    </body>
    </html>`;
  }

  private handleMessage(
    message: any,
    database: Database,
    webview: vscode.Webview
  ): void {
    switch (message.type) {
      case 'save':
        // 저장 처리
        break;
      case 'refresh':
        // 새로고침
        break;
      case 'addEdge':
        // Edge 추가
        break;
    }
  }
}
```

## 완료 조건
- [ ] src/extension.ts 생성
- [ ] src/CltEditorProvider.ts 생성
- [ ] Custom Editor 등록 확인
