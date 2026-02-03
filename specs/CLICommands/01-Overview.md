# Claritask CLI Overview

> **버전**: v0.0.3

## 개요

Claritask CLI의 모든 명령어 레퍼런스. 현재 구현 상태와 향후 계획을 구분하여 기술합니다.

**바이너리**: `clari`
**기술 스택**: Go + Cobra + SQLite

---

## 명령어 구조

```
clari
├── init                    # 프로젝트 초기화
├── project                 # 프로젝트 관리
│   ├── set / get / plan / start / stop / status
├── task                    # 작업 관리
│   ├── push / pop / start / complete / fail
│   ├── status / get / list
├── feature                 # Feature 관리
│   ├── list / add / get / spec / start
├── edge                    # Edge (의존성) 관리
│   ├── add / list / remove / infer
├── fdl                     # FDL 관리
│   ├── create / register / validate / show
│   ├── skeleton / tasks / verify / diff
├── plan                    # Planning 명령어
│   └── features
├── expert                  # Expert 관리
│   ├── add / list / get / edit / remove
│   ├── assign / unassign
├── memo                    # 메모 관리
│   ├── set / get / list / del
├── context                 # 컨텍스트 관리
│   ├── set / get
├── tech                    # 기술 스택 관리
│   ├── set / get
├── design                  # 설계 결정 관리
│   ├── set / get
└── required                # 필수 설정 확인
```

---

## 구현 상태

| 카테고리 | 명령어 수 | 상태 |
|----------|----------|------|
| 초기화 | 1 | 구현 완료 |
| Project | 6 | 구현 완료 |
| Task | 8 | 구현 완료 |
| Feature | 5 | 구현 완료 |
| Edge | 4 | 구현 완료 |
| FDL | 8 | 구현 완료 |
| Plan | 1 | 구현 완료 |
| Expert | 7 | 미구현 |
| Memo | 4 | 구현 완료 |
| Context/Tech/Design | 6 | 구현 완료 |
| Required | 1 | 구현 완료 |
| **총계** | **51** | - |

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
