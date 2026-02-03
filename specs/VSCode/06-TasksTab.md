# VSCode Extension Tasks 탭

> **버전**: v0.0.4

## Tasks 탭 레이아웃

```
┌─────────────────────────────────────────────────────────┐
│  [Project]  [Features]  [Tasks]                         │
├────────────┬────────────────────────┬───────────────────┤
│            │                        │                   │
│  Features  │      Canvas            │    Inspector      │
│  ──────    │      (드래그앤드롭)     │    ──────────     │
│  □ user_auth │                      │    Feature:       │
│    ├ task1   │   ┌─────┐  ┌─────┐   │    - name         │
│    ├ task2   │   │ T1  │→→│ T2  │   │    - spec         │
│    └ task3   │   └─────┘  └─────┘   │    - status       │
│  □ blog_post │        ↓             │                   │
│              │   ┌─────┐            │    Task:          │
│              │   │ T3  │            │    - title        │
│              │   └─────┘            │    - content      │
│              │                      │    - skill        │
│              │                      │    - status       │
├────────────┴────────────────────────┴───────────────────┤
│  Status: Connected │ Last sync: 2s ago │ WAL mode: ON   │
└─────────────────────────────────────────────────────────┘
```

## 패널 구성

| 패널 | 기능 |
|------|------|
| **Left: Tree View** | Feature/Task 계층 구조 |
| **Center: Canvas** | Task 노드 시각화, 드래그앤드롭으로 Edge 연결 |
| **Right: Inspector** | 선택된 항목 속성 편집 |
| **Bottom: Status Bar** | 동기화 상태, WAL 모드, 마지막 업데이트 |

---

## Task 관리 기능

- Task 노드 시각화
- 드래그앤드롭으로 Task 생성
- Task 상태 표시 (색상 구분)
  - `pending`: 회색
  - `doing`: 파란색
  - `done`: 녹색
  - `failed`: 빨간색

## Edge (의존성) 관리

- 노드 간 드래그로 Edge 생성
- Edge 클릭으로 삭제
- 순환 의존성 감지 및 경고

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
