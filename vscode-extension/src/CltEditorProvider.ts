import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import * as crypto from 'crypto';
import { Database, setExtensionPath } from './database';
import { SyncManager } from './sync';
import { createFeature as createFeatureCLI, validateFDL, generateTasks, generateSkeleton, CreateFeatureInput } from './cliService';
import { ttySessionManager } from './extension';

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
    const dbDir = path.dirname(dbPath);
    let database: Database | null = null;
    let sync: SyncManager | null = null;
    let expertWatcher: vscode.FileSystemWatcher | null = null;

    try {
      setExtensionPath(this.context.extensionUri.fsPath);
      database = new Database(dbPath);
      await database.init();
      sync = new SyncManager(database, webviewPanel.webview, dbPath);
      sync.start();

      // Setup Expert file watcher
      expertWatcher = this.setupExpertWatcher(dbDir, database, sync);

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
      expertWatcher?.dispose();
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

      case 'deleteFeature':
        this.handleDeleteFeature(message, database, webview, sync);
        break;

      case 'saveContext':
        this.handleSaveContext(message, database, webview, sync);
        break;

      case 'saveTech':
        this.handleSaveTech(message, database, webview, sync);
        break;

      case 'saveDesign':
        this.handleSaveDesign(message, database, webview, sync);
        break;

      case 'assignExpert':
        this.handleAssignExpert(message, database, webview, sync);
        break;

      case 'unassignExpert':
        this.handleUnassignExpert(message, database, webview, sync);
        break;

      case 'createExpert':
        this.handleCreateExpert(message, database, webview, sync);
        break;

      case 'openExpertFile':
        this.handleOpenExpertFile(message, database);
        break;

      // CLI integration for Feature creation
      case 'createFeatureWithFDL':
        this.handleCreateFeatureWithFDL(message, webview, sync);
        break;

      case 'validateFDL':
        this.handleValidateFDL(message, webview);
        break;

      case 'generateTasks':
        this.handleGenerateTasks(message, webview, sync);
        break;

      case 'generateSkeleton':
        this.handleGenerateSkeleton(message, webview, sync);
        break;

      case 'openFeatureFile':
        this.handleOpenFeatureFile(message, database);
        break;

      case 'sendMessage':
        this.handleSendMessage(message, database, webview, sync);
        break;

      case 'deleteMessage':
        this.handleDeleteMessage(message, database, webview, sync);
        break;

      case 'sendMessageCLI':
        this.handleSendMessageCLI(message, webview);
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

  private async handleCreateFeature(
    message: { name: string; description: string },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): Promise<void> {
    try {
      // Escape for bash single quotes
      const escapeName = message.name.replace(/'/g, "'\\''");
      const escapeDesc = message.description.replace(/'/g, "'\\''");
      const command = `~/bin/clari feature add --name '${escapeName}' --description '${escapeDesc}'`;

      // Build full command with cd for WSL
      const isWindows = process.platform === 'win32';
      const workspacePath = vscode.workspace.workspaceFolders?.[0].uri.fsPath;
      let fullCommand = command;

      if (isWindows && workspacePath) {
        const wslPath = windowsToWslPath(workspacePath);
        fullCommand = `cd '${wslPath}' && ${command}`;
      }

      // Use TTY Session Manager if available (respects max_parallel_sessions)
      if (ttySessionManager) {
        const sessionId = `feature-add-${escapeName}-${Date.now()}`;
        await ttySessionManager.startSession(sessionId, fullCommand);
      } else {
        // Fallback: Create terminal directly
        const terminal = vscode.window.createTerminal({
          name: 'Claritask - Create Feature',
          shellPath: isWindows ? 'wsl.exe' : undefined,
        });
        terminal.show();
        terminal.sendText(fullCommand);
      }

      // Send notification to webview
      webview.postMessage({
        type: 'cliStarted',
        command: 'feature.add',
        message: 'Claude Code will generate FDL for the feature...',
      });

      // Note: Actual feature creation and DB sync happens via CLI
      // The FDL file watcher will sync changes when FDL is generated
    } catch (err) {
      webview.postMessage({
        type: 'createResult',
        success: false,
        error: String(err),
      });
    }
  }

  private handleDeleteFeature(
    message: { featureId: number },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      database.deleteFeature(message.featureId);
      webview.postMessage({
        type: 'deleteResult',
        success: true,
        table: 'features',
        id: message.featureId,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'deleteResult',
        success: false,
        error: String(err),
      });
    }
  }

  private handleSaveContext(
    message: { data: Record<string, any> },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      database.saveContext(message.data);
      webview.postMessage({
        type: 'settingSaveResult',
        section: 'context',
        success: true,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'settingSaveResult',
        section: 'context',
        success: false,
        error: String(err),
      });
    }
  }

  private handleSaveTech(
    message: { data: Record<string, any> },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      database.saveTech(message.data);
      webview.postMessage({
        type: 'settingSaveResult',
        section: 'tech',
        success: true,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'settingSaveResult',
        section: 'tech',
        success: false,
        error: String(err),
      });
    }
  }

  private handleSaveDesign(
    message: { data: Record<string, any> },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      database.saveDesign(message.data);
      webview.postMessage({
        type: 'settingSaveResult',
        section: 'design',
        success: true,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'settingSaveResult',
        section: 'design',
        success: false,
        error: String(err),
      });
    }
  }

  private handleAssignExpert(
    message: { expertId: string },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      const project = database.getProject();
      if (!project) {
        webview.postMessage({
          type: 'expertResult',
          success: false,
          action: 'assign',
          error: 'No project found',
        });
        return;
      }

      database.assignExpert(project.id, message.expertId);
      webview.postMessage({
        type: 'expertResult',
        success: true,
        action: 'assign',
        expertId: message.expertId,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'expertResult',
        success: false,
        action: 'assign',
        error: String(err),
      });
    }
  }

  private handleUnassignExpert(
    message: { expertId: string },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      const project = database.getProject();
      if (!project) {
        webview.postMessage({
          type: 'expertResult',
          success: false,
          action: 'unassign',
          error: 'No project found',
        });
        return;
      }

      database.unassignExpert(project.id, message.expertId);
      webview.postMessage({
        type: 'expertResult',
        success: true,
        action: 'unassign',
        expertId: message.expertId,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'expertResult',
        success: false,
        action: 'unassign',
        error: String(err),
      });
    }
  }

  private async handleCreateExpert(
    message: { expertId: string },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): Promise<void> {
    try {
      const dbDir = path.dirname(database['dbPath']);
      const expertsDir = path.join(dbDir, 'experts', message.expertId);
      const expertFile = path.join(expertsDir, 'EXPERT.md');

      await fs.promises.mkdir(expertsDir, { recursive: true });

      const template = `# ${message.expertId}

## Role
TODO: Define the expert's role

## Tech Stack
- Language:
- Framework:

## Coding Rules
TODO: Define coding conventions

## Best Practices
TODO: Define best practices
`;

      await fs.promises.writeFile(expertFile, template);

      const contentHash = crypto.createHash('sha256').update(template).digest('hex');
      database.createExpert(message.expertId, message.expertId, expertFile, template, contentHash);

      webview.postMessage({
        type: 'expertResult',
        success: true,
        action: 'create',
        expertId: message.expertId,
      });
      sync.refresh();

      const uri = vscode.Uri.file(expertFile);
      await vscode.window.showTextDocument(uri);
    } catch (err) {
      webview.postMessage({
        type: 'expertResult',
        success: false,
        action: 'create',
        error: String(err),
      });
    }
  }

  private async handleOpenExpertFile(
    message: { expertId: string },
    database: Database
  ): Promise<void> {
    const dbDir = path.dirname(database['dbPath']);
    const expertFile = path.join(dbDir, 'experts', message.expertId, 'EXPERT.md');

    try {
      const uri = vscode.Uri.file(expertFile);
      await vscode.window.showTextDocument(uri);
    } catch (err) {
      vscode.window.showErrorMessage(`Failed to open expert file: ${err}`);
    }
  }

  private setupExpertWatcher(
    dbDir: string,
    database: Database,
    sync: SyncManager
  ): vscode.FileSystemWatcher {
    const expertsPattern = new vscode.RelativePattern(dbDir, 'experts/*/EXPERT.md');
    const watcher = vscode.workspace.createFileSystemWatcher(expertsPattern);

    watcher.onDidChange(async (uri) => {
      await this.syncExpertFromFile(uri, database, sync);
    });

    watcher.onDidCreate(async (uri) => {
      await this.syncExpertFromFile(uri, database, sync);
    });

    watcher.onDidDelete(async (uri) => {
      await this.handleExpertFileDeleted(uri, database);
    });

    return watcher;
  }

  private async syncExpertFromFile(
    uri: vscode.Uri,
    database: Database,
    sync: SyncManager
  ): Promise<void> {
    const filePath = uri.fsPath;
    const expertId = this.extractExpertId(filePath);

    if (!expertId) return;

    try {
      const content = await fs.promises.readFile(filePath, 'utf-8');
      const contentHash = crypto.createHash('sha256').update(content).digest('hex');

      const existingHash = database.getExpertContentHash(expertId);

      if (existingHash !== contentHash) {
        database.updateExpertContent(expertId, content, contentHash);

        const metadata = this.parseExpertMetadata(content);
        if (metadata) {
          database.updateExpertMetadata(
            expertId,
            metadata.name,
            metadata.domain,
            metadata.language,
            metadata.framework
          );
        }

        sync.refresh();
      }
    } catch (err) {
      console.error('Error syncing expert file:', err);
    }
  }

  private async handleExpertFileDeleted(uri: vscode.Uri, database: Database): Promise<void> {
    const filePath = uri.fsPath;
    const expertId = this.extractExpertId(filePath);

    if (!expertId) return;

    const content = database.getExpertContent(expertId);

    if (content) {
      const expertsDir = path.dirname(filePath);
      try {
        await fs.promises.mkdir(expertsDir, { recursive: true });
        await fs.promises.writeFile(filePath, content);
        vscode.window.showInformationMessage(`Expert file '${expertId}' was restored from backup.`);
      } catch (err) {
        console.error('Error restoring expert file:', err);
      }
    }
  }

  private extractExpertId(filePath: string): string {
    const parts = filePath.split(path.sep);
    const expertIndex = parts.indexOf('experts');
    if (expertIndex >= 0 && parts.length > expertIndex + 1) {
      return parts[expertIndex + 1];
    }
    return '';
  }

  private parseExpertMetadata(
    content: string
  ): { name: string; domain: string; language: string; framework: string } | null {
    const nameMatch = content.match(/^#\s+(.+)$/m);
    const name = nameMatch ? nameMatch[1].trim() : '';

    const roleMatch = content.match(/##\s+Role\s*\n([^\n#]+)/i);
    const domain = roleMatch ? roleMatch[1].trim() : '';

    const langMatch = content.match(/Language:\s*(.+)/i);
    const language = langMatch ? langMatch[1].trim() : '';

    const fwMatch = content.match(/Framework:\s*(.+)/i);
    const framework = fwMatch ? fwMatch[1].trim() : '';

    return { name, domain, language, framework };
  }

  // CLI integration handlers

  private async handleCreateFeatureWithFDL(
    message: { data: CreateFeatureInput },
    webview: vscode.Webview,
    sync: SyncManager
  ): Promise<void> {
    try {
      // Send progress
      webview.postMessage({
        type: 'cliProgress',
        command: 'feature.create',
        step: 'creating',
        message: 'Creating feature...',
      });

      const result = await createFeatureCLI(message.data);

      webview.postMessage({
        type: 'cliResult',
        command: 'feature.create',
        ...result,
      });

      if (result.success) {
        sync.refresh();
      }
    } catch (err) {
      webview.postMessage({
        type: 'cliResult',
        command: 'feature.create',
        success: false,
        error: String(err),
      });
    }
  }

  private async handleValidateFDL(
    message: { featureId: number },
    webview: vscode.Webview
  ): Promise<void> {
    try {
      const result = await validateFDL(message.featureId);

      webview.postMessage({
        type: 'cliResult',
        command: 'fdl.validate',
        ...result,
      });
    } catch (err) {
      webview.postMessage({
        type: 'cliResult',
        command: 'fdl.validate',
        success: false,
        error: String(err),
      });
    }
  }

  private async handleGenerateTasks(
    message: { featureId: number },
    webview: vscode.Webview,
    sync: SyncManager
  ): Promise<void> {
    try {
      webview.postMessage({
        type: 'cliProgress',
        command: 'fdl.tasks',
        step: 'generating',
        message: 'Generating tasks from FDL...',
      });

      const result = await generateTasks(message.featureId);

      webview.postMessage({
        type: 'cliResult',
        command: 'fdl.tasks',
        ...result,
      });

      if (result.success) {
        sync.refresh();
      }
    } catch (err) {
      webview.postMessage({
        type: 'cliResult',
        command: 'fdl.tasks',
        success: false,
        error: String(err),
      });
    }
  }

  private async handleGenerateSkeleton(
    message: { featureId: number; dryRun?: boolean },
    webview: vscode.Webview,
    sync: SyncManager
  ): Promise<void> {
    try {
      webview.postMessage({
        type: 'cliProgress',
        command: 'fdl.skeleton',
        step: 'generating',
        message: 'Generating skeleton files...',
      });

      const result = await generateSkeleton(message.featureId, message.dryRun);

      webview.postMessage({
        type: 'cliResult',
        command: 'fdl.skeleton',
        ...result,
      });

      if (result.success && !message.dryRun) {
        sync.refresh();
      }
    } catch (err) {
      webview.postMessage({
        type: 'cliResult',
        command: 'fdl.skeleton',
        success: false,
        error: String(err),
      });
    }
  }

  private async handleOpenFeatureFile(
    message: { featureId: number },
    database: Database
  ): Promise<void> {
    try {
      const features = database.getFeatures();
      const feature = features.find(f => f.id === message.featureId);

      if (!feature || !feature.file_path) {
        vscode.window.showWarningMessage('Feature file not found');
        return;
      }

      const uri = vscode.Uri.file(feature.file_path);
      await vscode.window.showTextDocument(uri);
    } catch (err) {
      vscode.window.showErrorMessage(`Failed to open feature file: ${err}`);
    }
  }

  // Message handlers
  private handleSendMessage(
    message: { content: string; featureId?: number },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      const project = database.getProject();
      if (!project) {
        webview.postMessage({
          type: 'messageResult',
          success: false,
          action: 'send',
          error: 'No project found',
        });
        return;
      }

      const messageId = database.createMessage(
        project.id,
        message.content,
        message.featureId
      );

      webview.postMessage({
        type: 'messageResult',
        success: true,
        action: 'send',
        messageId,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'messageResult',
        success: false,
        action: 'send',
        error: String(err),
      });
    }
  }

  private handleDeleteMessage(
    message: { messageId: number },
    database: Database,
    webview: vscode.Webview,
    sync: SyncManager
  ): void {
    try {
      database.deleteMessage(message.messageId);
      webview.postMessage({
        type: 'messageResult',
        success: true,
        action: 'delete',
        messageId: message.messageId,
      });
      sync.refresh();
    } catch (err) {
      webview.postMessage({
        type: 'messageResult',
        success: false,
        action: 'delete',
        error: String(err),
      });
    }
  }

  private async handleSendMessageCLI(
    message: { content: string; featureId?: number },
    webview: vscode.Webview
  ): Promise<void> {
    try {
      // Escape content for shell (single quotes)
      const escapedContent = message.content.replace(/'/g, "'\\''");

      // Build command
      let command = `~/bin/clari message send '${escapedContent}'`;
      if (message.featureId) {
        command += ` --feature ${message.featureId}`;
      }

      // Build full command with cd for WSL
      const isWindows = process.platform === 'win32';
      const workspacePath = vscode.workspace.workspaceFolders?.[0].uri.fsPath;
      let fullCommand = command;

      if (isWindows && workspacePath) {
        const wslPath = windowsToWslPath(workspacePath);
        fullCommand = `cd '${wslPath}' && ${command}`;
      }

      // Use TTY Session Manager if available
      if (ttySessionManager) {
        const sessionId = `message-send-${Date.now()}`;
        await ttySessionManager.startSession(sessionId, fullCommand);
      } else {
        // Fallback: Create terminal directly
        const terminal = vscode.window.createTerminal({
          name: 'Claritask - Send Message',
          shellPath: isWindows ? 'wsl.exe' : undefined,
        });
        terminal.show();
        terminal.sendText(fullCommand);
      }

      webview.postMessage({
        type: 'cliStarted',
        command: 'message.send',
        message: 'Claude Code will analyze your request...',
      });
    } catch (err) {
      webview.postMessage({
        type: 'messageResult',
        success: false,
        action: 'send',
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

/**
 * Convert Windows path to WSL path
 * C:\Users\mail\git\claritask → /mnt/c/Users/mail/git/claritask
 */
function windowsToWslPath(windowsPath: string): string {
  const match = windowsPath.match(/^([A-Za-z]):\\(.*)$/);
  if (match) {
    const drive = match[1].toLowerCase();
    const rest = match[2].replace(/\\/g, '/');
    return `/mnt/${drive}/${rest}`;
  }
  return windowsPath;
}
