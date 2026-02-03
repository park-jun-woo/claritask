import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import * as crypto from 'crypto';
import { CltEditorProvider } from './CltEditorProvider';
import { Database, setExtensionPath } from './database';
import { TTYSessionManager } from './ttySessionManager';

let fdlWatcher: vscode.FileSystemWatcher | undefined;
let expertWatcher: vscode.FileSystemWatcher | undefined;
let database: Database | undefined;
let sessionStatusBar: vscode.StatusBarItem | undefined;
export let ttySessionManager: TTYSessionManager | undefined;

export function activate(context: vscode.ExtensionContext) {
  // Register the custom editor provider
  context.subscriptions.push(CltEditorProvider.register(context));

  // Set extension path for sql.js
  setExtensionPath(context.extensionPath);

  // Setup file watchers if workspace contains .claritask/db.clt
  setupFileWatchers(context);

  // Setup session status bar
  setupSessionStatusBar(context);

  // Setup TTY session manager
  setupTTYSessionManager(context);

  // Register session status command
  context.subscriptions.push(
    vscode.commands.registerCommand('claritask.showSessionStatus', () => {
      if (ttySessionManager) {
        const status = ttySessionManager.getStatus();
        const activeIds = ttySessionManager.getActiveSessionIds().join(', ') || '없음';
        const waitingIds = ttySessionManager.getWaitingSessionIds().join(', ') || '없음';
        vscode.window.showInformationMessage(
          `Claude Code 세션 상태\n` +
          `활성: ${status.active}/${status.max}\n` +
          `대기: ${status.waiting}\n\n` +
          `활성 세션: ${activeIds}\n` +
          `대기 세션: ${waitingIds}`
        );
      } else {
        vscode.window.showInformationMessage('TTY 세션 매니저가 초기화되지 않았습니다.');
      }
    })
  );
}

async function setupFileWatchers(context: vscode.ExtensionContext) {
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  if (!workspaceFolder) return;

  const dbPath = path.join(workspaceFolder.uri.fsPath, '.claritask', 'db.clt');
  if (!fs.existsSync(dbPath)) return;

  // Initialize database for file sync
  database = new Database(dbPath);
  try {
    await database.init();
  } catch (err) {
    console.error('Failed to initialize database for file watchers:', err);
    return;
  }

  // FDL file watcher (features/*.fdl.yaml)
  fdlWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(workspaceFolder, 'features/*.fdl.yaml')
  );

  fdlWatcher.onDidChange(uri => syncFDLFile(uri, database!));
  fdlWatcher.onDidCreate(uri => syncFDLFile(uri, database!));
  fdlWatcher.onDidDelete(uri => handleFDLFileDeleted(uri, database!));

  context.subscriptions.push(fdlWatcher);

  // Expert file watcher
  expertWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(workspaceFolder, '.claritask/experts/**/EXPERT.md')
  );

  expertWatcher.onDidChange(uri => syncExpertFile(uri, database!));
  expertWatcher.onDidCreate(uri => syncExpertFile(uri, database!));
  expertWatcher.onDidDelete(uri => restoreExpertFromDB(uri, database!));

  context.subscriptions.push(expertWatcher);
}

function setupSessionStatusBar(context: vscode.ExtensionContext) {
  sessionStatusBar = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100
  );
  sessionStatusBar.command = 'claritask.showSessionStatus';
  sessionStatusBar.tooltip = 'Claude Code 세션 상태 (클릭하여 상세 보기)';
  context.subscriptions.push(sessionStatusBar);

  // Initially hidden
  updateSessionStatusBar(0, 0, 3);
}

function updateSessionStatusBar(active: number, waiting: number, max: number) {
  if (!sessionStatusBar) return;

  if (active === 0 && waiting === 0) {
    sessionStatusBar.hide();
    return;
  }

  sessionStatusBar.text = `$(terminal) Claude: ${active}/${max}`;
  if (waiting > 0) {
    sessionStatusBar.text += ` (${waiting} 대기)`;
  }
  sessionStatusBar.tooltip = `활성 세션: ${active}\n대기 중: ${waiting}\n최대: ${max}\n\n클릭하여 상세 보기`;
  sessionStatusBar.show();
}

function setupTTYSessionManager(context: vscode.ExtensionContext) {
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  if (!workspaceFolder) return;

  ttySessionManager = new TTYSessionManager(workspaceFolder.uri.fsPath);

  // Subscribe to status changes
  ttySessionManager.onStatusChange((status) => {
    updateSessionStatusBar(status.active, status.waiting, status.max);
  });

  context.subscriptions.push({
    dispose: () => ttySessionManager?.dispose()
  });
}

function calculateHash(content: string): string {
  return crypto.createHash('sha256').update(content).digest('hex');
}

async function syncFDLFile(uri: vscode.Uri, db: Database) {
  try {
    const filePath = uri.fsPath;
    // Extract feature name from <name>.fdl.yaml
    const fileName = path.basename(filePath).replace('.fdl.yaml', '');
    const content = fs.readFileSync(filePath, 'utf-8');
    const fdlHash = calculateHash(content);

    // Find feature by name
    const feature = db.getFeatureByName(fileName);
    if (!feature) {
      console.log(`Feature not found for FDL file: ${fileName}`);
      return;
    }

    // Check if FDL content has changed
    if (feature.fdl_hash !== fdlHash) {
      db.updateFeatureFDL(feature.id, content, fdlHash);
      console.log(`Synced FDL file: ${fileName}`);
    }
  } catch (err) {
    console.error('Failed to sync FDL file:', err);
  }
}

function handleFDLFileDeleted(uri: vscode.Uri, db: Database) {
  try {
    const fileName = path.basename(uri.fsPath).replace('.fdl.yaml', '');
    const feature = db.getFeatureByName(fileName);
    if (feature) {
      // Clear FDL content in DB
      db.clearFeatureFDL(feature.id);
      console.log(`Cleared FDL for feature: ${fileName}`);
    }
  } catch (err) {
    console.error('Failed to handle FDL file deletion:', err);
  }
}

async function syncExpertFile(uri: vscode.Uri, db: Database) {
  try {
    const filePath = uri.fsPath;
    const expertId = path.basename(path.dirname(filePath));
    const content = fs.readFileSync(filePath, 'utf-8');
    const contentHash = calculateHash(content);

    // Check if content has changed
    const existingHash = db.getExpertContentHash(expertId);
    if (existingHash !== contentHash) {
      db.updateExpertContent(expertId, content, contentHash);
      console.log(`Synced expert file: ${expertId}`);
    }
  } catch (err) {
    console.error('Failed to sync expert file:', err);
  }
}

function restoreExpertFromDB(uri: vscode.Uri, db: Database) {
  try {
    const filePath = uri.fsPath;
    const expertId = path.basename(path.dirname(filePath));

    // Get content from DB
    const content = db.getExpertContent(expertId);
    if (content) {
      // Recreate the file
      const dir = path.dirname(filePath);
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }
      fs.writeFileSync(filePath, content);
      console.log(`Restored expert file from DB: ${expertId}`);
    }
  } catch (err) {
    console.error('Failed to restore expert file:', err);
  }
}

export function deactivate() {
  if (fdlWatcher) {
    fdlWatcher.dispose();
  }
  if (expertWatcher) {
    expertWatcher.dispose();
  }
  if (database) {
    database.close();
  }
  if (ttySessionManager) {
    ttySessionManager.dispose();
  }
  if (sessionStatusBar) {
    sessionStatusBar.dispose();
  }
}
