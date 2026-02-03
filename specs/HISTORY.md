# Claritask Specs 변경이력

> 모든 specs 문서의 통합 변경이력

---

## v0.0.6 (2026-02-03)

| 문서 | 변경 내용 |
|------|----------|
| CLI/07-Feature.md | `clari feature create` 통합 생성 명령어 추가 (Feature+FDL+Task 한 번에) |
| VSCode/05-FeaturesTab.md | Feature 생성 다이얼로그 UI 추가, CLI 호출 연동 |
| VSCode/11-MessageProtocol.md | CLI 호출 메시지 (`createFeature`, `validateFDL`, `generateTasks` 등) 추가 |
| VSCode/14-CLICompatibility.md | CLI 호출 아키텍처, cliService 구현 스펙 추가 |

---

## v0.0.5 (2026-02-03)

| 문서 | 변경 내용 |
|------|----------|
| CLI/07-Feature.md | Feature md 파일 자동 생성 기능 추가 (`features/<name>.md`) |
| DB/02-A-Core.md | features 테이블에 `file_path`, `content`, `content_hash` 필드 추가 |
| DB/02-C-Content.md | 버전 업데이트 |
| VSCode/05-FeaturesTab.md | Feature md 파일 양방향 동기화, Markdown 렌더링 뷰 추가 |

---

## v0.0.4 (2026-02-03)

| 문서 | 변경 내용 |
|------|----------|
| Claritask.md | TTY Handover 아키텍처 반영, 구조 간소화 |
| CLI/* (14개) | 버전 v0.0.4 통일, 헤더/푸터 형식 통일, 상호 참조 추가 |
| FDL/* (8개) | 버전 v0.0.4 통일, 헤더/푸터 형식 통일, 파일명 변경 (02a→02-A) |
| TTY/* (7개) | 버전 v0.0.4 통일, 헤더/푸터 형식 통일, 상호 참조 추가 |
| VSCode/* (15개) | 버전 v0.0.4 통일, 헤더/푸터 형식 통일 |
| DB/* (5개) | DB 스키마 문서 신규 생성 |
| HISTORY.md | 변경이력 통합 문서 신규 생성 |

---

## v0.0.3 (2026-02-03)

| 문서 | 변경 내용 |
|------|----------|
| CLI/01-Overview.md | Expert DB 스키마 백업 필드 추가, 동기화 정책 |
| CLI/02-Init.md | ClariInit.md 통합 |
| CLI/03-Project.md | TTY Handover 연동 명세 추가 |
| VSCode/01-Overview.md | Execution Status 섹션 추가 |

---

## v0.0.2 (2026-02-03)

| 문서 | 변경 내용 |
|------|----------|
| CLI/01-Overview.md | Expert 명령어 추가 |
| VSCode/01-Overview.md | Project 탭 설계 추가 |

---

## v0.0.1 (2026-02-03)

| 문서 | 변경 내용 |
|------|----------|
| Claritask.md | 최초 작성 |
| CLI/* | 최초 작성 |
| FDL/* | 최초 작성 |
| TTY/* | 최초 작성 |
| VSCode/* | 최초 작성 |
| DB/* | 최초 작성 |

---

*Claritask Specs History v0.0.6 - 2026-02-03*
