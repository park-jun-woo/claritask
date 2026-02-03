# VSCode Extension 프로젝트 구조

> **버전**: v0.0.4

## 폴더 구조

```
claritask-vscode/
├── package.json              # Extension manifest
├── src/
│   ├── extension.ts          # Extension 진입점
│   ├── CltEditorProvider.ts  # Custom Editor Provider
│   ├── database.ts           # SQLite 읽기/쓰기
│   ├── sync.ts               # Polling 로직
│   └── expertWatcher.ts      # Expert 파일 감시 및 동기화
├── webview-ui/
│   ├── package.json
│   ├── src/
│   │   ├── App.tsx
│   │   ├── components/
│   │   │   ├── FeatureList.tsx      # Feature 목록
│   │   │   ├── TaskPanel.tsx        # Task 관리 패널
│   │   │   ├── StatusBar.tsx        # 하단 상태 바
│   │   │   ├── ProjectPanel.tsx     # Project 탭 메인
│   │   │   ├── ProjectInfo.tsx      # 프로젝트 기본 정보 (읽기 전용)
│   │   │   ├── ContextSection.tsx   # Context 편집 섹션
│   │   │   ├── TechSection.tsx      # Tech Stack 편집 섹션
│   │   │   ├── DesignSection.tsx    # Design Decisions 편집 섹션
│   │   │   ├── ExecutionStatus.tsx  # 실행 상태 표시
│   │   │   ├── ProgressBar.tsx      # 진행률 바
│   │   │   ├── RecentTaskList.tsx   # 최근 Task 로그
│   │   │   ├── EditableField.tsx    # Key-Value 편집 컴포넌트
│   │   │   ├── SectionCard.tsx      # 섹션 카드 래퍼
│   │   │   ├── ExpertsPanel.tsx     # Experts 탭 메인
│   │   │   ├── ExpertCard.tsx       # Expert 카드 (렌더링 + 버튼)
│   │   │   └── ExpertContent.tsx    # 마크다운 렌더링 컴포넌트
│   │   ├── hooks/
│   │   │   └── useSync.ts
│   │   └── stores/
│   │       └── store.ts
│   └── vite.config.ts
└── README.md
```

---

## package.json (Extension Manifest)

```json
{
  "name": "claritask",
  "displayName": "Claritask",
  "description": "Visual editor for Claritask projects",
  "version": "0.1.0",
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
  }
}
```

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
