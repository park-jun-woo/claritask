# 2026-02-03 FDL 통합 설계 대화

## 주제
Claritask 프로젝트 평가 및 FDL(Feature Definition Language) 통합 설계

---

## 1. Claritask 현재 상태 평가

### 구현 완성도: 약 45%

**완료된 항목:**
- 인프라 (SQLite, Cobra CLI, JSON 입출력)
- 기본 CRUD (Project, Phase, Task, Memo, Context/Tech/Design)
- 테스트 7개 파일 통과

**미구현 핵심 기능:**
- 그래프 기반 의존성 (task_edges, feature_edges)
- Edge 관리 명령어
- 자동 실행 (project start/stop/status)
- Feature 관리 (스펙에서는 Feature, 현재는 Phase)

### 스펙대로 구현해도 쓸만할까?

**결론: 의문**

핵심 문제점:
```
Task 1 result: "createComment 함수 구현 완료"
Task 2: createComments로 오타 → 불일치 발생
```

- result 요약만으로 다음 Task에 충분한 컨텍스트 전달 어려움
- LLM 호출 비용 상당 (Feature 20개 기준 ~160회)
- 워크플로우 복잡

---

## 2. FDL 제안 검토

### FDL(ClariSpec)이란?
YAML 기반 DSL로 Feature를 정의 (모델 → 서비스 → API → UI)

```yaml
feature: comment_system
models:
  - name: Comment
    fields: [id, content, post_id, user_id]
service:
  - name: createComment
    input: { userId, postId, content }
api:
  - path: /posts/{postId}/comments
    use: service.createComment
```

### FDL의 효과
- 함수명, API 경로, 필드명이 명세에 확정
- Task 간 공유하면 불일치 상당히 줄어듦
- 하지만 LLM은 여전히 오타 가능

---

## 3. FDL + 스켈레톤 통합 설계

### 핵심 아이디어

```
FDL (YAML)  →  Python Parser  →  Skeleton Code  →  Task (TODO 채우기)
     ↓              ↓                  ↓                    ↓
  계약 정의      AST 변환         코드 틀 생성        LLM이 내용만 작성
```

**LLM 역할이 "코드 전체 작성" → "TODO 채우기"로 축소**

### 생성되는 스켈레톤 예시

```python
async def createComment(userId: UUID, postId: UUID, content: str) -> Comment:
    """
    Steps (from FDL):
    - validate: "content 길이 검증"
    - db: "INSERT INTO comments"
    """
    # TODO: 위 Steps를 구현하세요
    raise NotImplementedError("createComment not implemented")
```

### 워크플로우

1. `clari fdl create` - FDL 템플릿 생성
2. `clari fdl register` - FDL 등록, Feature 생성
3. `clari fdl skeleton` - Python이 스켈레톤 코드 생성 (확정적)
4. `clari fdl tasks` - TODO에서 Task 자동 추출 (확정적)
5. `clari project start` - LLM이 TODO만 채움
6. `clari fdl verify` - FDL과 구현 일치 검증

### LLM 호출 횟수 감소

| 단계 | 기존 | FDL |
|------|------|-----|
| Feature Spec/FDL | 20회 | 20회 |
| Task 생성 | 20회 | 0회 (Python) |
| Task Edge 추출 | 20회 | 0회 (자동) |
| **총** | **~60회** | **~22회** |

### 장점

| 기존 | FDL + 스켈레톤 |
|------|---------------|
| LLM이 함수명 결정 | FDL에서 확정 |
| Task 간 불일치 가능 | 스켈레톤이 Single Source |
| 전체 코드 작성 | TODO만 채우기 |
| 검증 불가 | FDL 기반 검증 가능 |

---

## 4. Claritask.md 업데이트 내용

### 추가된 섹션
- FDL 통합 섹션 (예시, 워크플로우, 비교표)
- FDL 명령어 (create, register, skeleton, tasks, verify)

### 스키마 변경
- `features` 테이블: `fdl`, `fdl_hash`, `skeleton_generated` 추가
- `skeletons` 테이블 신규
- `tasks` 테이블: `skeleton_id`, `target_file`, `target_line`, `target_function` 추가

### Manifest 응답 변경
- `fdl`: 현재 Task의 FDL 정의
- `skeleton`: 스켈레톤 현재 상태 (TODO 위치)

### 핵심 가치 업데이트
```
사람: FDL로 "무엇을" 정의
      ↓
Python: "어떤 구조로" 스켈레톤 생성
      ↓
LLM: "어떻게" TODO 채우기
```

---

## 결론

**"LLM의 창의성은 로직 구현에만, 구조는 확정적으로"**

FDL 통합으로 Claritask의 실용성이 크게 개선될 수 있음:
- 함수명/타입 불일치 원천 차단
- LLM 호출 횟수 약 60% 감소
- 검증 가능한 파이프라인
