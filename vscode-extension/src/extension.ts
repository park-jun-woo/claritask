import * as vscode from 'vscode';
import { CltEditorProvider } from './CltEditorProvider';

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(CltEditorProvider.register(context));
}

export function deactivate() {}
