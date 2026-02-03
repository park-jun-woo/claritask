import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import * as crypto from 'crypto';
import { CltEditorProvider } from './CltEditorProvider';
import { Database, setExtensionPath } from './database';

let fdlWatcher: vscode.FileSystemWatcher | undefined;
let expertWatcher: vscode.FileSystemWatcher | undefined;
let database: Database | undefined;

export function activate(context: vscode.ExtensionContext) {
  // Register the custom editor provider
  context.subscriptions.push(CltEditorProvider.register(context));

  // Set extension path for sql.js
  setExtensionPath(context.extensionPath);

  // Setup file watchers if workspace contains .claritask/db.clt
  setupFileWatchers(context);
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
}
