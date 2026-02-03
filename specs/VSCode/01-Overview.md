# VSCode Extension Overview

> **현재 버전**: v0.0.4 ([변경이력](../HISTORY.md))

---

## 개요

`.clt` 확장자의 SQLite 데이터베이스 파일을 VSCode 내에서 시각적으로 편집하는 GUI 확장 프로그램.

**목표**: CLI와 GUI가 동일한 DB 파일을 실시간으로 공유하며 편집

---

## 아키텍처

### Phase 1: MVP (WAL + Polling)

```
┌─────────────────┐         ┌─────────────────┐
│  Claude Code    │         │  VSCode GUI     │
│  (clari cli)    │         │  (Webview)      │
└────────┬────────┘         └────────┬────────┘
         │                           │
         │ write                     │ read (1초 polling)
         ▼                           ▼
      ┌─────────────────────────────────┐
      │     db.clt (SQLite + WAL)       │
      └─────────────────────────────────┘
```

### Phase 2: File Watcher

```
db.clt 변경 감지 (fs.watch)
    ↓
즉시 SQLite 재읽기
    ↓
GUI 업데이트
```

### Phase 3: Daemon (향후)

```
┌──────────┐     ┌──────────┐
│  CLI     │     │   GUI    │
└────┬─────┘     └────┬─────┘
     │  WebSocket     │
     ▼                ▼
┌─────────────────────────┐
│   clari daemon          │
│   - SQLite 단독 접근    │
│   - 변경 시 broadcast   │
└─────────────────────────┘
```

---

## 파일 확장자

| 확장자 | 설명 |
|--------|------|
| `.clt` | Claritask SQLite 데이터베이스 |

**기존 경로 변경:**
```
Before: .claritask/db
After:  .claritask/db.clt
```

VSCode에서 `.clt` 파일 열면 Custom Editor가 활성화됨.

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [DB/01-Overview.md](../DB/01-Overview.md) | 데이터베이스 스키마 |
| [CLI/01-Overview.md](../CLI/01-Overview.md) | CLI 명령어 레퍼런스 |
| [DB/02-C-Content.md](../DB/02-C-Content.md) | Experts 테이블 스키마 |

---

*Claritask VSCode Extension Spec v0.0.4*
