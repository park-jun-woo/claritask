# 2026-02-03 Specs 문서 정리 완료

## 세션 개요

이전 세션에서 시작한 specs 문서 재구성 작업을 마무리한 세션.

---

## 이전 세션에서 완료된 작업

1. **FDL/02-Schema.md 분할**: 레이어별 4개 파일로 분할 (02a~02d)
2. **Claritask.md 간소화**: 개요만 남기고 중복 제거
3. **폴더 리네임**: CLICommands → CLI, vscodeGUI → VSCode
4. **TTY 폴더 생성**: TTY-Handover.md와 UseCase.md를 7개 파일로 분할
5. **DB 폴더 생성**: 스키마 문서 5개 파일 생성
6. **HISTORY.md 생성**: 변경이력 통합 문서

---

## 이번 세션에서 완료된 작업

### 1. CLI 문서 TTY Handover 통합

**CLI/02-Init.md**:
- `claude --print` 방식에서 TTY Handover 방식으로 업데이트
- Phase 구조 간소화 (5단계 → 2단계)
- TTY Handover 흐름도 추가

**CLI/03-Project.md**:
- `project start`에 TTY Handover 동작 흐름 추가
- Task별 루프 구조 명시

### 2. 버전 히스토리 통합

다음 파일들의 개별 변경이력 제거 → `HISTORY.md` 참조로 교체:
- CLI/01-Overview.md
- CLI/02-Init.md
- CLI/03-Project.md
- FDL/01-Overview.md
- TTY/01-Overview.md
- VSCode/01-Overview.md

버전 헤더 형식 통일:
```markdown
> **현재 버전**: v0.0.4 ([변경이력](../HISTORY.md))
```

### 3. 문서 간 상호 참조 추가

| 문서 | 추가된 참조 |
|------|------------|
| CLI/01-Overview.md | TTY, DB, FDL |
| CLI/02-Init.md | TTY/03-Phase1, TTY/05-Implementation, TTY/06-ClaudeCLI |
| CLI/03-Project.md | TTY/04-Phase2, TTY/05-Implementation |
| FDL/01-Overview.md | CLI/09-FDL, TTY/04-Phase2 |
| TTY/01-Overview.md | CLI/02-Init, CLI/03-Project, DB/01-Overview |
| VSCode/01-Overview.md | DB/01-Overview, CLI/01-Overview, DB/02-C-Content |

### 4. HISTORY.md v0.0.4 업데이트

```markdown
## v0.0.4 (2026-02-03)

| 문서 | 변경 내용 |
|------|----------|
| Claritask.md | TTY Handover 아키텍처 반영, 구조 간소화 |
| CLI/01-Overview.md | 버전 통합, 상호 참조 추가 |
| CLI/02-Init.md | TTY Handover 연동 명세 추가 |
| CLI/03-Project.md | TTY Handover 기반 자동 실행 명세 추가 |
| FDL/01-Overview.md | 버전 통합, 상호 참조 추가 |
| TTY/01-Overview.md | 버전 통합, 상호 참조 추가 |
| VSCode/01-Overview.md | Experts 탭 추가, FileSystemWatcher 동기화, 버전 통합 |
| DB/* | DB 스키마 문서 신규 생성 |
| HISTORY.md | 변경이력 통합 문서 신규 생성 |
```

---

## 최종 specs/ 폴더 구조

```
specs/
├── Claritask.md          # 메인 개요
├── HISTORY.md            # 통합 변경이력
├── CLI/                  # CLI 명령어 (14개 파일)
│   ├── 01-Overview.md
│   ├── 02-Init.md        # TTY Handover 연동
│   ├── 03-Project.md     # TTY Handover 연동
│   └── ...
├── DB/                   # 데이터베이스 스키마 (5개 파일)
│   ├── 01-Overview.md
│   ├── 02-A-Core.md
│   ├── 02-B-Settings.md
│   ├── 02-C-Content.md
│   └── 03-Migration.md
├── FDL/                  # Feature Definition Language (8개 파일)
│   ├── 01-Overview.md
│   ├── 02-Schema.md
│   ├── 02a~02d-*Layer.md
│   └── ...
├── TTY/                  # TTY Handover (7개 파일)
│   ├── 01-Overview.md
│   └── ...
└── VSCode/               # VSCode Extension (15개 파일)
    ├── 01-Overview.md
    └── ...
```

---

## 보류된 작업

- **FDL 스켈레톤 생성기 명세**: 별도 작업 예정

---

*2026-02-03 대화 저장*
