import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import * as crypto from 'crypto';
import { CltEditorProvider } from './CltEditorProvider';
import { Database, setExtensionPath } from './database';

let featureWatcher: vscode.FileSystemWatcher | undefined;
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

  // Feature file watcher
  featureWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(workspaceFolder, 'features/*.md')
  );

  featureWatcher.onDidChange(uri => syncFeatureFile(uri, database!));
  featureWatcher.onDidCreate(uri => syncFeatureFile(uri, database!));
  featureWatcher.onDidDelete(uri => clearFeatureFilePath(uri, database!));

  context.subscriptions.push(featureWatcher);

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

async function syncFeatureFile(uri: vscode.Uri, db: Database) {
  try {
    const filePath = uri.fsPath;
    const fileName = path.basename(filePath, '.md');
    const content = fs.readFileSync(filePath, 'utf-8');
    const contentHash = calculateHash(content);

    // Find feature by name
    const feature = db.getFeatureByName(fileName);
    if (!feature) {
      console.log(`Feature not found for file: ${fileName}`);
      return;
    }

    // Check if content has changed
    if (feature.content_hash !== contentHash) {
      db.updateFeatureContent(feature.id, content, contentHash, filePath);
      console.log(`Synced feature file: ${fileName}`);
    }
  } catch (err) {
    console.error('Failed to sync feature file:', err);
  }
}

function clearFeatureFilePath(uri: vscode.Uri, db: Database) {
  try {
    const fileName = path.basename(uri.fsPath, '.md');
    const feature = db.getFeatureByName(fileName);
    if (feature) {
      db.clearFeatureFilePath(feature.id);
      console.log(`Cleared file path for feature: ${fileName}`);
    }
  } catch (err) {
    console.error('Failed to clear feature file path:', err);
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
  if (featureWatcher) {
    featureWatcher.dispose();
  }
  if (expertWatcher) {
    expertWatcher.dispose();
  }
  if (database) {
    database.close();
  }
}
