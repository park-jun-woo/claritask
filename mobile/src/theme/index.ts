import {MD3LightTheme, MD3DarkTheme} from 'react-native-paper';
import type {MD3Theme} from 'react-native-paper';

// Colors matching Web UI shadcn/ui HSL palette
// Light: primary=hsl(222.2, 47.4%, 11.2%) → #1a1f2e
// Dark:  primary=hsl(210, 40%, 98%) → #f5f7fa

const lightColors = {
  primary: '#1a1f2e',
  onPrimary: '#f5f7fa',
  primaryContainer: '#dde2ee',
  onPrimaryContainer: '#1a1f2e',
  secondary: '#eef1f6',
  onSecondary: '#1a1f2e',
  secondaryContainer: '#eef1f6',
  onSecondaryContainer: '#1a1f2e',
  tertiary: '#6b7280',
  onTertiary: '#ffffff',
  background: '#ffffff',
  onBackground: '#0f172a',
  surface: '#ffffff',
  onSurface: '#0f172a',
  surfaceVariant: '#f1f5f9',
  onSurfaceVariant: '#6b7280',
  error: '#dc2626',
  onError: '#ffffff',
  outline: '#e2e8f0',
  elevation: {
    level0: 'transparent',
    level1: '#f8fafc',
    level2: '#f1f5f9',
    level3: '#e2e8f0',
    level4: '#cbd5e1',
    level5: '#94a3b8',
  },
};

const darkColors = {
  primary: '#f5f7fa',
  onPrimary: '#1a1f2e',
  primaryContainer: '#2a3040',
  onPrimaryContainer: '#f5f7fa',
  secondary: '#232a38',
  onSecondary: '#f5f7fa',
  secondaryContainer: '#232a38',
  onSecondaryContainer: '#f5f7fa',
  tertiary: '#94a3b8',
  onTertiary: '#0f172a',
  background: '#0f172a',
  onBackground: '#f5f7fa',
  surface: '#0f172a',
  onSurface: '#f5f7fa',
  surfaceVariant: '#1e293b',
  onSurfaceVariant: '#94a3b8',
  error: '#7f1d1d',
  onError: '#f5f7fa',
  outline: '#232a38',
  elevation: {
    level0: 'transparent',
    level1: '#1e293b',
    level2: '#232a38',
    level3: '#2a3040',
    level4: '#334155',
    level5: '#475569',
  },
};

export const lightTheme: MD3Theme = {
  ...MD3LightTheme,
  colors: {
    ...MD3LightTheme.colors,
    ...lightColors,
  },
};

export const darkTheme: MD3Theme = {
  ...MD3DarkTheme,
  colors: {
    ...MD3DarkTheme.colors,
    ...darkColors,
  },
};

// Task status colors
export const statusColors = {
  todo: '#9ca3af',
  planned: '#3b82f6',
  split: '#a855f7',
  done: '#22c55e',
  failed: '#ef4444',
};
