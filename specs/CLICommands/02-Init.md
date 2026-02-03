# clari init - 프로젝트 초기화

> **버전**: v0.0.3

## 개요

프로젝트 초기화. LLM과 협업하여 프로젝트 설정 완성.

```bash
clari init <project-id> [options]
```

---

## 인자

- `project-id` (필수): 프로젝트 ID
  - 규칙: 영문 소문자, 숫자, 하이픈(`-`), 언더스코어(`_`)만 허용

---

## 옵션

| 옵션 | 단축 | 설명 |
|------|------|------|
| --name | -n | 프로젝트 이름 (기본값: project-id) |
| --description | -d | 프로젝트 설명 |
| --skip-analysis | | 컨텍스트 분석 건너뛰기 |
| --skip-specs | | Specs 생성 건너뛰기 |
| --non-interactive | | 비대화형 모드 (자동 승인) |
| --force | | 기존 DB 덮어쓰기 |
| --resume | | 중단된 초기화 재개 |

---

## 프로세스

1. **Phase 1**: DB 초기화 (.claritask/db.clt 생성)
2. **Phase 2**: 프로젝트 파일 분석 (claude --print)
3. **Phase 3**: tech/design 승인 (대화형)
4. **Phase 4**: Specs 초안 생성 (claude --print)
5. **Phase 5**: 피드백 루프 (승인까지 반복)

---

## 생성 구조

```
./
├── .claritask/
│   └── db
└── specs/
    └── <project-id>.md
```

---

## 응답

**성공:**
```json
{
  "success": true,
  "project_id": "my-api",
  "db_path": ".claritask/db.clt",
  "specs_path": "specs/my-api.md"
}
```

**에러:**
```json
{
  "success": false,
  "error": "database already exists at .claritask/db.clt (use --force to overwrite)"
}
```

---

## 예시

```bash
# 기본 사용 (전체 프로세스)
clari init my-api

# 옵션 지정
clari init my-api --name "My REST API" --description "사용자 관리 API"

# 빠른 초기화 (LLM 호출 없이)
clari init my-api --skip-analysis --skip-specs

# 기존 프로젝트 재초기화
clari init my-api --force

# 중단된 초기화 재개
clari init --resume

# 비대화형 모드 (CI/CD용)
clari init my-api --non-interactive
```

---

## Phase 상세

| Phase | 설명 | 건너뛰기 |
|-------|------|----------|
| 1 | .claritask/db.clt 생성, 프로젝트 레코드 | 불가 |
| 2 | 파일 스캔, LLM으로 tech/design 분석 | --skip-analysis |
| 3 | 분석 결과 사용자 승인 | --non-interactive (자동 승인) |
| 4 | LLM으로 specs 문서 생성 | --skip-specs |
| 5 | 피드백 반영, 최종 승인 | --non-interactive (자동 승인) |

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
