# Gap Analysis v2: specs vs 현재 구현

> **분석일**: 2026-02-03
> **현재 버전**: v0.0.4

---

## 요약

| 카테고리 | 이슈 수 | 심각도 |
|----------|---------|--------|
| Task 구조 | 4 | CRITICAL |
| DB 스키마 | 5 | HIGH |
| Expert 연동 | 3 | HIGH |
| Init/TTY | 4 | CRITICAL |
| FDL 구현 | 8 | MEDIUM |
| 기타 | 4 | LOW |
| **합계** | **28** | - |

---

## 완료된 Task (5개)

| Task ID | 내용 | 상태 |
|---------|------|------|
| TASK-DEV-074 | experts 테이블 스키마 업데이트 | ✓ 완료 |
| TASK-DEV-075 | _migrations 테이블 및 버전 관리 | ✓ 완료 |
| TASK-DEV-076 | 인덱스 추가 | ✓ 완료 |
| TASK-DEV-077 | clari db 명령어 구현 | ✓ 완료 |
| TASK-DEV-078 | Expert 동기화 로직 | ✓ 완료 |

---

## 신규 Task (18개)

### Critical Priority

| Task ID | 내용 | 심각도 |
|---------|------|--------|
| TASK-DEV-079 | Task 구조체 필드 추가 (level, skill, references, parent_id) | CRITICAL |
| TASK-DEV-083 | Init TTY Handover Phase 2-5 구현 | CRITICAL |

### High Priority

| Task ID | 내용 | 심각도 |
|---------|------|--------|
| TASK-DEV-080 | expert_assignments 테이블 추가 | HIGH |
| TASK-DEV-081 | experts 테이블 스키마 추가 수정 | HIGH |
| TASK-DEV-082 | Feature.version 구조체 필드 추가 | HIGH |
| TASK-DEV-084 | Expert manifest 연동 | HIGH |
| TASK-DEV-093 | task pop 의존성 기반 순서 | HIGH |

### Medium Priority

| Task ID | 내용 | 심각도 |
|---------|------|--------|
| TASK-DEV-087 | FDL Data Layer 구조화 | MEDIUM |
| TASK-DEV-088 | FDL Logic Layer 구조화 | MEDIUM |
| TASK-DEV-089 | FDL Interface Layer 구조화 | MEDIUM |
| TASK-DEV-090 | FDL Presentation Layer 구조화 | MEDIUM |
| TASK-DEV-094 | FDL 스켈레톤 생성 개선 | MEDIUM |
| TASK-DEV-096 | FDL verify/diff 기능 완성 | MEDIUM |

### Low Priority

| Task ID | 내용 | 심각도 |
|---------|------|--------|
| TASK-DEV-085 | Project 메시지 수정 ("phase" → "feature") | LOW |
| TASK-DEV-086 | Memo summary, tags 지원 | LOW |
| TASK-DEV-091 | feature tasks 서브커맨드 정리 | LOW |
| TASK-DEV-092 | expert add 에디터 자동열기 | LOW |
| TASK-DEV-095 | Required 필드 검증 확인 | LOW |

---

## 의존성 그래프

```
TASK-DEV-079 (Task 구조체) ←── TASK-DEV-093 (의존성 기반 pop)
                          ←── TASK-DEV-087~090 (FDL 구조화)

TASK-DEV-080 (expert_assignments) ←── TASK-DEV-084 (manifest 연동)

TASK-DEV-081 (experts 스키마) ←── TASK-DEV-084 (manifest 연동)

TASK-DEV-087~090 (FDL 구조화) ←── TASK-DEV-094 (스켈레톤 생성)
                              ←── TASK-DEV-096 (verify/diff)
```

---

## 상세 이슈 목록

### 1. Task 구조 (CRITICAL)

- **Task.ID 타입**: string → int64 변경 필요
- **누락 필드**: level, skill, references, parent_id
- **의존성 순서**: task pop 시 edge 미고려

### 2. DB 스키마 (HIGH)

- **experts 테이블**:
  - description 컬럼 누락
  - content_backup 컬럼 누락
  - status 컬럼 누락
  - name UNIQUE 제약 누락
- **expert_assignments 테이블**: 완전 누락
- **Feature.version**: Go 구조체에 누락

### 3. Expert 연동 (HIGH)

- task pop manifest에 expert 미포함
- feature-expert 연결 없음
- 프로젝트/feature 레벨 expert 통합 필요

### 4. Init/TTY (CRITICAL)

- Phase 2-5 미구현
- TTY Handover 메커니즘 없음
- LLM 연동 미구현
- 상태 관리 불완전

### 5. FDL 구현 (MEDIUM)

- **Data Layer**: 15% 구현
  - 필드 타입 검증 없음
  - 제약조건 파싱 없음
  - 인덱스, 관계 미지원
- **Logic Layer**: 10% 구현
  - input 검증 규칙 미지원
  - steps 타입 미지원
- **Interface Layer**: 10% 구현
  - request 분해 미지원
  - response 상태코드 미지원
- **Presentation Layer**: 10% 구현
  - computed, methods 미지원
  - styles 미지원

### 6. 기타 (LOW)

- Project plan 메시지 오류
- Memo summary/tags 미지원
- feature tasks 서브커맨드 문서 불일치
- expert add 에디터 자동열기 미구현

---

## 권장 실행 순서

1. **Phase 1**: Critical (2개)
   - TASK-DEV-079: Task 구조체
   - TASK-DEV-083: Init TTY Handover

2. **Phase 2**: High (5개)
   - TASK-DEV-080: expert_assignments
   - TASK-DEV-081: experts 스키마
   - TASK-DEV-082: Feature.version
   - TASK-DEV-084: Expert manifest
   - TASK-DEV-093: task pop 의존성

3. **Phase 3**: Medium (6개)
   - TASK-DEV-087~090: FDL 4개 레이어
   - TASK-DEV-094: 스켈레톤 생성
   - TASK-DEV-096: verify/diff

4. **Phase 4**: Low (5개)
   - TASK-DEV-085, 086, 091, 092, 095

---

*Gap Analysis Report v2 - 2026-02-03*
