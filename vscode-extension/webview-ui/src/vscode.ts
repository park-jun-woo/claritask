import type { MessageFromWebview } from './types';

declare function acquireVsCodeApi(): {
  postMessage(message: any): void;
  getState(): any;
  setState(state: any): void;
};

const vscodeApi = acquireVsCodeApi();

export const vscode = {
  postMessage: (message: MessageFromWebview) => vscodeApi.postMessage(message),
};

export function postMessage(message: MessageFromWebview): void {
  vscodeApi.postMessage(message);
}

export function getState<T>(): T | undefined {
  return vscodeApi.getState() as T | undefined;
}

export function setState<T>(state: T): void {
  vscodeApi.setState(state);
}

export function assignExpert(expertId: string): void {
  vscodeApi.postMessage({ type: 'assignExpert', expertId });
}

export function unassignExpert(expertId: string): void {
  vscodeApi.postMessage({ type: 'unassignExpert', expertId });
}

export function openExpertFile(expertId: string): void {
  vscodeApi.postMessage({ type: 'openExpertFile', expertId });
}

export function createExpert(expertId: string): void {
  vscodeApi.postMessage({ type: 'createExpert', expertId });
}
