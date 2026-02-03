# VSCode Extension Features 탭

> **버전**: v0.0.4

## Features 탭 레이아웃

```
┌─────────────────────────────────────────────────────────┐
│  [Project]  [Features]  [Tasks]                         │
├────────────┬────────────────────────────────────────────┤
│            │                                            │
│  Features  │      Feature Detail                        │
│  ──────    │      ──────────────                        │
│  ▸ user_auth │    Name: user_auth                       │
│  ▸ blog_post │    Status: active                        │
│  + Add...    │    Description: ...                      │
│              │                                          │
│              │    ┌─ Spec ───────────────────────────┐  │
│              │    │ # User Authentication            │  │
│              │    │ ...                              │  │
│              │    └──────────────────────────────────┘  │
│              │                                          │
│              │    ┌─ FDL ────────────────────────────┐  │
│              │    │ feature: user_auth               │  │
│              │    │ ...                              │  │
│              │    └──────────────────────────────────┘  │
├────────────┴────────────────────────────────────────────┤
│  Status: Connected │ Last sync: 2s ago │ WAL mode: ON   │
└─────────────────────────────────────────────────────────┘
```

## Feature 관리 기능

- Feature 목록 트리 뷰
- Feature 추가/삭제/편집
- Feature 스펙 (Markdown) 편집
- FDL 코드 편집 (코드 에디터 내장)

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
