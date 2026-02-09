import {createMMKV} from 'react-native-mmkv';
import type {MMKV} from 'react-native-mmkv';

export const mmkv: MMKV = createMMKV({id: 'claribot-cache'});

const SERVER_URL_KEY = 'server_url';
const PROJECT_KEY = 'selected_project';

export function setCachedServerUrl(url: string) {
  mmkv.set(SERVER_URL_KEY, url);
}

export function getCachedServerUrl(): string | undefined {
  return mmkv.getString(SERVER_URL_KEY);
}

export function clearCachedServerUrl() {
  mmkv.remove(SERVER_URL_KEY);
}

export function getSelectedProject(): string | undefined {
  return mmkv.getString(PROJECT_KEY);
}

export function setSelectedProject(projectId: string) {
  mmkv.set(PROJECT_KEY, projectId);
}

export function clearSelectedProject() {
  mmkv.remove(PROJECT_KEY);
}
