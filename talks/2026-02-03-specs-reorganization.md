# 2026-02-03 Specs 문서 재구성 및 UseCase 정의

## 요약

specs 폴더의 문서들을 재구성하고, Claritask의 2-Phase 아키텍처를 명확히 정의함.

---

## 1. FDL 문서 분할

### 02-Schema.md를 레이어별로 분할

기존 02-Schema.md를 개요로 유지하고, 4개 레이어별 상세 문서 생성:

| 파일 | 내용 |
|------|------|
| `02-Schema.md` | DSL 전체 구조 개요, 계층간 연결(Wiring) 다이어그램 |
| `02a-DataLayer.md` | DATA LAYER - 필드 타입, 제약조건, 인덱스, 관계 정의, Soft Delete, Timestamps |
| `02b-LogicLayer.md` | LOGIC LAYER - 입출력 검증, 스텝 타입, 에러 처리, 트랜잭션, 권한 |
| `02c-InterfaceLayer.md` | INTERFACE LAYER - HTTP 메서드, 경로 설계, 요청/응답, 인증, Rate Limiting |
| `02d-PresentationLayer.md` | PRESENTATION LAYER - Atomic Design, Props/State/Computed, View 정의 |

---

## 2. Claritask.md 간소화

중복 내용 제거하고 개요만 유지:
- FDL 상세 → `specs/FDL/`
- 명령어 레퍼런스 → `specs/CLI/`
- DB 스키마 상세 제거
- 관련 문서 링크 테이블 추가

**파일 크기**: ~1200줄 → ~200줄 (83% 감소)

---

## 3. 폴더 이름 변경

```
specs/CLICommands/ → specs/CLI/
specs/vscodeGUI/   → specs/VSCode/
```

관련 참조 링크 모두 업데이트.

---

## 4. UseCase.md 생성 및 2-Phase 아키텍처 정의

### 핵심 논의: 대화형 모드의 역할

**기존 이해**: 대화형 모드 = 디버깅용
**수정된 이해**: 대화형 모드 = 요구사항 수립용 (Phase 1)

### 2-Phase 구조

```
Phase 1: 요구사항 수립 (대화형)
├─ clari init → TTY Handover → Claude Code
├─ 사용자와 대화하며 Features 확정
├─ clari feature add로 DB 저장
└─ "개발해" → clari project start

Phase 2: 자동 실행 (Claritask 드라이버)
├─ clari project start
├─ Plan → Task List → Edge Link
├─ Task별 TTY Handover → Claude Code
└─ 최종 보고
```

### 프로세스 아키텍처 논의

**질문**: `clari → Claude Code → clari → Claude Code` 구조가 괜찮은가?

**결론**: 괜찮음. 이유:
1. **순차 실행** - 중첩이 아니라 하나 끝나면 다음
2. **Stateless** - 각 clari 프로세스는 stateless, 상태는 DB에
3. **실패 복구 용이** - 어디서든 재개 가능
4. **Unix 철학** - 각 프로그램이 한 가지 일을 잘 함

### Phase 1 구현 방식 논의

**질문**: 대화형 모드를 직접 구현해야 하나?

**결론**: TTY Handover로 Claude Code에게 위임
- 대화 UI 구현 불필요
- Claude Code가 `clari feature add`로 상태 저장
- `clari project start` 실행하면 Phase 2로 전환

### clari project start 감지 문제

**질문**: Claude Code가 실행한 `clari project start`를 부모 clari가 감지할 수 있나?

**결론**: 감지 불필요
- `clari project start`가 새 프로세스로 실행되어 새 오케스트레이터가 됨
- 부모 clari는 그냥 종료
- 가장 단순한 설계

---

## 5. 최종 specs/ 폴더 구조

```
specs/
├── Claritask.md      # 전체 시스템 개요
├── UseCase.md        # 2-Phase 사용 흐름
├── TTY-Handover.md   # TTY 핸드오버 명세
├── CLI/              # 14개 파일 - 명령어 레퍼런스
├── FDL/              # 8개 파일 - Feature Definition Language
└── VSCode/           # 15개 파일 - VSCode 확장 명세
```

---

## 핵심 결정 사항

1. **Phase 1은 TTY Handover로 구현** - 직접 대화 UI 구현 X
2. **clari → Claude Code → clari → Claude Code 구조 채택** - 순차, stateless, 복구 용이
3. **Phase 전환은 새 프로세스로** - 감지 불필요, 단순함 유지
