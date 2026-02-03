import { useEffect, useState } from 'react';
import { useStore } from '../store';
import { postMessage } from '../vscode';
import type { MessageToWebview } from '../types';

interface SyncState {
  isConnected: boolean;
  lastSync: Date | null;
  error: string | null;
}

export function useSync(): SyncState {
  const [state, setState] = useState<SyncState>({
    isConnected: false,
    lastSync: null,
    error: null,
  });

  const { setData, addConflict, removePendingSave } = useStore();

  useEffect(() => {
    const handleMessage = (event: MessageEvent<MessageToWebview>) => {
      const message = event.data;

      switch (message.type) {
        case 'sync':
          setData(message.data);
          setState({
            isConnected: true,
            lastSync: new Date(message.timestamp),
            error: null,
          });
          break;

        case 'error':
          setState((prev) => ({
            ...prev,
            error: message.message,
          }));
          break;

        case 'saveResult':
          if (message.success && message.table && message.id !== undefined) {
            removePendingSave(`${message.table}:${message.id}`);
          }
          break;

        case 'conflict':
          addConflict(`${message.table}:${message.id}`);
          break;

        case 'edgeResult':
          if (!message.success && message.error) {
            setState((prev) => ({
              ...prev,
              error: message.error!,
            }));
          }
          break;

        case 'createResult':
          if (!message.success && message.error) {
            setState((prev) => ({
              ...prev,
              error: message.error!,
            }));
          }
          break;
      }
    };

    window.addEventListener('message', handleMessage);

    return () => {
      window.removeEventListener('message', handleMessage);
    };
  }, [setData, addConflict, removePendingSave]);

  return state;
}

export function refresh(): void {
  postMessage({ type: 'refresh' });
}

export function saveTask(id: number, data: Partial<any>, version: number): void {
  postMessage({
    type: 'save',
    table: 'tasks',
    id,
    data,
    version,
  });
}

export function saveFeature(id: number, data: Partial<any>, version: number): void {
  postMessage({
    type: 'save',
    table: 'features',
    id,
    data,
    version,
  });
}

export function addTaskEdge(fromId: number, toId: number): void {
  postMessage({
    type: 'addEdge',
    fromId,
    toId,
  });
}

export function removeTaskEdge(fromId: number, toId: number): void {
  postMessage({
    type: 'removeEdge',
    fromId,
    toId,
  });
}

export function createTask(featureId: number, title: string, content: string): void {
  postMessage({
    type: 'createTask',
    featureId,
    title,
    content,
  });
}

export function createFeature(name: string, description: string): void {
  postMessage({
    type: 'createFeature',
    name,
    description,
  });
}
