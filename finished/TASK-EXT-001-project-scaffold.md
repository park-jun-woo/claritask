# TASK-EXT-001: VSCode Extension 프로젝트 Scaffolding

## 목표
VSCode Extension 프로젝트 기본 구조 생성.

## 디렉토리
`vscode-extension/`

## 파일 목록

### 1. package.json (Extension Manifest)

```json
{
  "name": "claritask",
  "displayName": "Claritask",
  "description": "Visual editor for Claritask projects",
  "version": "0.1.0",
  "publisher": "claritask",
  "engines": {
    "vscode": "^1.85.0"
  },
  "categories": ["Other"],
  "activationEvents": [],
  "main": "./out/extension.js",
  "contributes": {
    "customEditors": [
      {
        "viewType": "claritask.cltEditor",
        "displayName": "Claritask Editor",
        "selector": [
          {
            "filenamePattern": "*.clt"
          }
        ],
        "priority": "default"
      }
    ]
  },
  "scripts": {
    "vscode:prepublish": "npm run compile",
    "compile": "tsc -p ./",
    "watch": "tsc -watch -p ./",
    "build:webview": "cd webview-ui && npm run build",
    "package": "vsce package"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/vscode": "^1.85.0",
    "typescript": "^5.3.0"
  },
  "dependencies": {
    "better-sqlite3": "^9.4.0"
  }
}
```

### 2. tsconfig.json

```json
{
  "compilerOptions": {
    "module": "commonjs",
    "target": "ES2022",
    "outDir": "out",
    "lib": ["ES2022"],
    "sourceMap": true,
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "exclude": ["node_modules", "webview-ui"]
}
```

### 3. .vscodeignore

```
.vscode/**
node_modules/**
webview-ui/node_modules/**
webview-ui/src/**
src/**
!webview-ui/dist/**
*.map
.gitignore
tsconfig.json
```

### 4. .gitignore

```
node_modules/
out/
*.vsix
webview-ui/dist/
```

## 디렉토리 구조

```
vscode-extension/
├── package.json
├── tsconfig.json
├── .vscodeignore
├── .gitignore
├── src/
│   └── (다음 TASK에서 생성)
└── webview-ui/
    └── (다음 TASK에서 생성)
```

## 완료 조건
- [ ] package.json 생성
- [ ] tsconfig.json 생성
- [ ] .vscodeignore 생성
- [ ] .gitignore 생성
- [ ] 디렉토리 구조 생성
