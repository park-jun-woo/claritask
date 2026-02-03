import * as fs from 'fs';
import * as path from 'path';
import * as yaml from 'yaml';

export interface TTYConfig {
  max_parallel_sessions: number;
  terminal_close_delay: number;
}

export interface VSCodeConfig {
  sync_interval: number;
  watch_feature_files: boolean;
}

export interface ClaritaskConfig {
  tty: TTYConfig;
  vscode: VSCodeConfig;
}

export const DEFAULT_CONFIG: ClaritaskConfig = {
  tty: {
    max_parallel_sessions: 3,
    terminal_close_delay: 1,
  },
  vscode: {
    sync_interval: 1000,
    watch_feature_files: true,
  },
};

/**
 * Load configuration from .claritask/config.yaml
 */
export function loadConfig(workspacePath: string): ClaritaskConfig {
  const configPath = path.join(workspacePath, '.claritask', 'config.yaml');

  try {
    const content = fs.readFileSync(configPath, 'utf8');
    const loaded = yaml.parse(content) as Partial<ClaritaskConfig>;

    // Merge with defaults
    const config: ClaritaskConfig = {
      tty: { ...DEFAULT_CONFIG.tty, ...loaded?.tty },
      vscode: { ...DEFAULT_CONFIG.vscode, ...loaded?.vscode },
    };

    // Validate and clamp values
    if (config.tty.max_parallel_sessions < 1) {
      config.tty.max_parallel_sessions = 1;
    }
    if (config.tty.max_parallel_sessions > 10) {
      config.tty.max_parallel_sessions = 10;
    }
    if (config.vscode.sync_interval < 100) {
      config.vscode.sync_interval = 100;
    }

    return config;
  } catch {
    return DEFAULT_CONFIG;
  }
}

/**
 * Watch for config file changes
 */
export function watchConfig(
  workspacePath: string,
  callback: (config: ClaritaskConfig) => void
): fs.FSWatcher | null {
  const configPath = path.join(workspacePath, '.claritask', 'config.yaml');

  try {
    return fs.watch(configPath, (eventType) => {
      if (eventType === 'change') {
        const newConfig = loadConfig(workspacePath);
        callback(newConfig);
      }
    });
  } catch {
    return null;
  }
}
