# TASK-EXT-005: Webview React 기본 구조

## 목표
Webview용 React 앱 기본 구조 설정.

## 디렉토리
`webview-ui/`

## 파일 목록

### 1. webview-ui/package.json

```json
{
  "name": "claritask-webview",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "@vscode/webview-ui-toolkit": "^1.4.0",
    "zustand": "^4.5.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@types/vscode-webview": "^1.57.0",
    "@vitejs/plugin-react": "^4.2.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0",
    "tailwindcss": "^3.4.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

### 2. webview-ui/vite.config.ts

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    rollupOptions: {
      output: {
        entryFileNames: 'index.js',
        assetFileNames: 'index.css',
      },
    },
  },
});
```

### 3. webview-ui/tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"]
}
```

### 4. webview-ui/tailwind.config.js

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {},
  },
  plugins: [],
};
```

### 5. webview-ui/postcss.config.js

```javascript
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
};
```

### 6. webview-ui/index.html

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Claritask</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

### 7. webview-ui/src/main.tsx

```typescript
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './index.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

### 8. webview-ui/src/index.css

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
  --vscode-font-family: var(--vscode-editor-font-family, monospace);
}

body {
  margin: 0;
  padding: 0;
  font-family: var(--vscode-font-family);
  background-color: var(--vscode-editor-background);
  color: var(--vscode-editor-foreground);
}
```

### 9. webview-ui/src/App.tsx

```typescript
import React from 'react';
import { FeatureTree } from './components/FeatureTree';
import { StatusBar } from './components/StatusBar';
import { useProjectStore } from './stores/projectStore';
import { useSync } from './hooks/useSync';

function App() {
  useSync();
  const { project, features, tasks } = useProjectStore();

  return (
    <div className="h-screen flex flex-col">
      {/* Header */}
      <header className="h-10 px-4 flex items-center border-b border-gray-700">
        <h1 className="text-sm font-semibold">
          Claritask: {project?.name ?? 'Loading...'}
        </h1>
        <button
          className="ml-auto text-xs px-2 py-1 bg-blue-600 rounded hover:bg-blue-700"
          onClick={() => window.postMessage({ type: 'refresh' }, '*')}
        >
          ⟳ Refresh
        </button>
      </header>

      {/* Main Content */}
      <main className="flex-1 flex overflow-hidden">
        {/* Left: Feature Tree */}
        <aside className="w-64 border-r border-gray-700 overflow-auto">
          <FeatureTree features={features} tasks={tasks} />
        </aside>

        {/* Center: Canvas (Phase 2) */}
        <section className="flex-1 flex items-center justify-center text-gray-500">
          Canvas (Coming in Phase 2)
        </section>

        {/* Right: Inspector (Phase 3) */}
        <aside className="w-72 border-l border-gray-700 p-4 text-gray-500">
          Inspector (Coming in Phase 3)
        </aside>
      </main>

      {/* Status Bar */}
      <StatusBar />
    </div>
  );
}

export default App;
```

## 완료 조건
- [ ] package.json 생성
- [ ] vite.config.ts 생성
- [ ] tsconfig.json 생성
- [ ] Tailwind 설정
- [ ] index.html 생성
- [ ] main.tsx, App.tsx 생성
- [ ] npm install 실행
- [ ] npm run build 성공
