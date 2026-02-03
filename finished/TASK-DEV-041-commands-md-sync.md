# TASK-DEV-041: Commands.md 문서 업데이트

## 개요
- **파일**: `specs/Commands.md`
- **유형**: 문서 수정
- **스펙 참조**: 현재 코드베이스

## 배경
Commands.md 문서가 현재 구현 상태와 불일치:
- "미구현 명령어" 섹션에 이미 구현된 명령어들이 포함됨
- Feature, Edge, FDL 명령어들이 구현되었으나 문서에는 미구현으로 표시

## 수정 내용

### 1. 구현 완료로 이동해야 할 명령어
```
### Feature 관리 (구현 완료)
clari feature list
clari feature add '<json>'
clari feature get <id>
clari feature spec <id> '<spec>'
clari feature start <id>

### Edge 관리 (구현 완료)
clari edge add --from <id> --to <id>
clari edge add --feature --from <id> --to <id>
clari edge list
clari edge list --feature
clari edge list --task
clari edge list --phase <id>
clari edge remove --from <id> --to <id>

### FDL 관리 (구현 완료)
clari fdl create <name>
clari fdl register <file>
clari fdl validate <feature_id>
clari fdl show <feature_id>
clari fdl skeleton <feature_id>
clari fdl skeleton <feature_id> --dry-run
clari fdl skeleton <feature_id> --force
clari fdl tasks <feature_id>

### Project 실행 (구현 완료)
clari project start
clari project start --feature <id>
clari project start --dry-run
clari project start --fallback-interactive
clari project stop
clari project status
```

### 2. 미구현으로 남아있어야 할 명령어
```
### 미구현 (향후 계획)
clari fdl verify <feature_id>       # FDL-코드 일치 검증
clari fdl diff <feature_id>         # FDL-코드 차이점
clari edge infer --feature <id>     # Task Edge LLM 추론
clari edge infer --project          # Feature Edge LLM 추론
clari plan features                 # Feature 목록 LLM 산출
```

### 3. 명령어 구조 섹션 업데이트
```
clari
├── init
├── project
│   ├── set / get / plan / start / stop / status
├── phase
│   ├── create / list / plan / start
├── task
│   ├── push / pop / start / complete / fail
│   ├── status / get / list
├── feature               ← 추가
│   ├── list / add / get / spec / start
├── edge                  ← 추가
│   ├── add / list / remove
├── fdl                   ← 추가
│   ├── create / register / validate / show
│   ├── skeleton / tasks
├── memo
│   ├── set / get / list / del
├── context / tech / design
│   ├── set / get
└── required
```

## 완료 기준
- [ ] Feature 관리 섹션 추가
- [ ] Edge 관리 섹션 추가
- [ ] FDL 관리 섹션 추가
- [ ] 미구현 섹션 정리
- [ ] 명령어 구조 업데이트
