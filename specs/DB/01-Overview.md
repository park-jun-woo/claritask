# Database Overview

> **현재 버전**: v0.0.4 ([변경이력](../HISTORY.md))

---

## 개요

Claritask는 SQLite 단일 파일로 모든 상태를 관리합니다.

---

## 파일 위치

```
<project>/
└── .claritask/
    └── db.clt          # SQLite 데이터베이스
```

**확장자**: `.clt` (Claritask)
- VSCode에서 Custom Editor로 열림
- CLI와 GUI가 동일 파일 공유

---

## 동시성

**WAL 모드 (Write-Ahead Logging)**:
- CLI 쓰기 중에도 GUI 읽기 가능
- 다중 읽기, 단일 쓰기
- 충돌 시 낙관적 잠금 (version 컬럼)

```sql
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;
```

---

## 테이블 구조

| 카테고리 | 테이블 | 설명 |
|----------|--------|------|
| Core | projects | 프로젝트 |
| Core | features | Feature (기능 단위) |
| Core | tasks | Task (실행 단위) |
| Core | task_edges | Task 간 의존성 |
| Core | feature_edges | Feature 간 의존성 |
| Settings | context | 프로젝트 컨텍스트 (싱글톤) |
| Settings | tech | 기술 스택 (싱글톤) |
| Settings | design | 설계 결정 (싱글톤) |
| Settings | state | 현재 상태 (key-value) |
| Content | memos | 메모 (scope 기반) |
| Content | skeletons | 생성된 스켈레톤 파일 |
| Content | experts | Expert 문서 |

---

## 스키마 문서

| 문서 | 내용 |
|------|------|
| [02-A-Core.md](02-A-Core.md) | 핵심 테이블 (projects, features, tasks, edges) |
| [02-B-Settings.md](02-B-Settings.md) | 설정 테이블 (context, tech, design, state) |
| [02-C-Content.md](02-C-Content.md) | 콘텐츠 테이블 (memos, skeletons, experts) |
| [03-Migration.md](03-Migration.md) | 마이그레이션 전략 |

---

*Database Specification v0.0.4*
