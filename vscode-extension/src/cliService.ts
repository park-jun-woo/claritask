import * as vscode from 'vscode';
import { spawn } from 'child_process';

export interface CLIResult {
  success: boolean;
  data?: any;
  error?: string;
  [key: string]: any;
}

/**
 * Execute a clari CLI command
 */
export async function executeCLI(
  command: string,
  subcommand: string,
  jsonArg?: object
): Promise<CLIResult> {
  return new Promise((resolve) => {
    const args = [command, subcommand];
    if (jsonArg) {
      args.push(JSON.stringify(jsonArg));
    }

    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      resolve({ success: false, error: 'No workspace folder' });
      return;
    }

    const proc = spawn('clari', args, {
      cwd: workspaceFolder.uri.fsPath,
      shell: true
    });

    let stdout = '';
    let stderr = '';

    proc.stdout.on('data', (data) => {
      stdout += data;
    });

    proc.stderr.on('data', (data) => {
      stderr += data;
    });

    proc.on('close', (code) => {
      try {
        const result = JSON.parse(stdout);
        resolve(result);
      } catch {
        resolve({
          success: false,
          error: stderr || stdout || `CLI execution failed with code ${code}`
        });
      }
    });

    proc.on('error', (err) => {
      resolve({ success: false, error: err.message });
    });
  });
}

/**
 * Create a feature with optional FDL and task generation
 */
export interface CreateFeatureInput {
  name: string;
  description: string;
  fdl?: string;
  generate_tasks?: boolean;
  generate_skeleton?: boolean;
}

export const createFeature = (data: CreateFeatureInput): Promise<CLIResult> =>
  executeCLI('feature', 'create', data);

/**
 * Validate FDL for a feature
 */
export const validateFDL = (featureId: number): Promise<CLIResult> =>
  executeCLI('fdl', 'validate', { id: featureId });

/**
 * Generate tasks from FDL
 */
export const generateTasks = (featureId: number): Promise<CLIResult> =>
  executeCLI('fdl', 'tasks', { id: featureId });

/**
 * Generate skeleton files from FDL
 */
export const generateSkeleton = (featureId: number, dryRun = false): Promise<CLIResult> => {
  const args = ['fdl', 'skeleton', String(featureId)];
  if (dryRun) {
    args.push('--dry-run');
  }
  return new Promise((resolve) => {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      resolve({ success: false, error: 'No workspace folder' });
      return;
    }

    const proc = spawn('clari', args, {
      cwd: workspaceFolder.uri.fsPath,
      shell: true
    });

    let stdout = '';
    let stderr = '';

    proc.stdout.on('data', (data) => {
      stdout += data;
    });

    proc.stderr.on('data', (data) => {
      stderr += data;
    });

    proc.on('close', () => {
      try {
        resolve(JSON.parse(stdout));
      } catch {
        resolve({ success: false, error: stderr || stdout || 'CLI execution failed' });
      }
    });

    proc.on('error', (err) => {
      resolve({ success: false, error: err.message });
    });
  });
};

/**
 * Add a feature (simple version)
 */
export const addFeature = (name: string, description: string): Promise<CLIResult> =>
  executeCLI('feature', 'add', { name, description });
