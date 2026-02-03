# Claritask CLI Overview

> **현재 버전**: v0.0.10 ([변경이력](../HISTORY.md))

---

## 개요

Claritask CLI의 모든 명령어 레퍼런스. 현재 구현 상태와 향후 계획을 구분하여 기술합니다.

**바이너리**: `clari`
**기술 스택**: Go + Cobra + SQLite (modernc.org/sqlite, Pure Go)

---

## 빌드 및 설치

### 요구사항
- Go 1.21+
- CGO 불필요 (Pure Go SQLite 사용)

### 빌드 방법

**Linux/Mac:**
```bash
cd cli
make build
make install  # /usr/local/bin에 설치
```

**Windows:**
```powershell
cd cli
.\build.ps1 all    # 또는: build.bat all
# 빌드 후 %USERPROFILE%\bin에 설치, PATH 자동 추가
```

### 빌드 스크립트
| 파일 | 플랫폼 | 명령어 |
|------|--------|--------|
| `Makefile` | Linux/Mac | `make build`, `make install`, `make clean` |
| `build.ps1` | Windows (PowerShell) | `.\build.ps1 [build\|install\|clean\|all]` |
| `build.bat` | Windows (CMD) | `build.bat [build\|install\|clean\|all]` |

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
│   ├── list / add / get / spec / start / delete / fdl
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
├── message                 # 메시지 관리
│   ├── send / list / get / delete
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
| Feature | 8 | 구현 완료 |
| Edge | 4 | 구현 완료 |
| FDL | 8 | 구현 완료 |
| Plan | 1 | 구현 완료 |
| Expert | 7 | 구현 완료 |
| Memo | 4 | 구현 완료 |
| Message | 4 | 미구현 |
| Context/Tech/Design | 6 | 구현 완료 |
| Required | 1 | 구현 완료 |
| **총계** | **58** | - |

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [TTY/01-Overview.md](../TTY/01-Overview.md) | TTY Handover 아키텍처 |
| [DB/01-Overview.md](../DB/01-Overview.md) | 데이터베이스 스키마 |
| [FDL/01-Overview.md](../FDL/01-Overview.md) | Feature Definition Language |
| [15-Message.md](15-Message.md) | Message 명령어 |
| [16-Config.md](16-Config.md) | Config 설정 파일 |

---

*Claritask Commands Reference v0.0.10*
