# VSCode Extension 로드맵

> **버전**: v0.0.4

## Phase 1: MVP
- [x] Custom Editor Provider 구현
- [x] SQLite 읽기 (sql.js)
- [x] React Webview 기본 구조
- [x] Feature/Task 트리 뷰
- [x] 1초 polling 동기화
- [x] WAL 모드 활성화

## Phase 1.5: Project 탭
- [ ] Project 탭 UI 구현
- [ ] ProjectPanel 컴포넌트
- [ ] ProjectInfo (읽기 전용) 컴포넌트
- [ ] ContextSection 편집 컴포넌트
- [ ] TechSection 편집 컴포넌트
- [ ] DesignSection 편집 컴포넌트
- [ ] ExecutionStatus 컴포넌트
  - [ ] ProgressBar (진행률 시각화)
  - [ ] RecentTaskList (최근 Task 로그)
  - [ ] 상태 표시 (Running/Idle/Completed/Has Failures)
- [ ] Context/Tech/Design 저장 메시지 핸들러
- [ ] 필수 필드 검증 (required indicator)
- [ ] 사용자 정의 필드 추가/삭제

## Phase 2: Canvas
- [ ] React Flow 통합
- [ ] Task 노드 시각화
- [ ] 드래그앤드롭 Edge 생성
- [ ] 상태별 색상 표시

## Phase 3: 편집 기능
- [ ] Inspector 패널
- [ ] Task 속성 편집
- [ ] Feature 스펙 편집
- [ ] FDL 코드 편집기

## Phase 4: Experts 탭
- [ ] ExpertsPanel 컴포넌트
- [ ] ExpertCard 컴포넌트 (마크다운 렌더링)
- [ ] react-markdown 통합
- [ ] Expert 파일 열기 메시지 핸들러
- [ ] Assign/Unassign 기능
- [ ] Create New Expert 다이얼로그
- [ ] FileSystemWatcher 구현
  - [ ] Expert 파일 변경 감지 → DB 백업
  - [ ] Expert 파일 삭제 감지 → 자동 복구
- [ ] activationEvents 설정

## Phase 5: 동기화 강화
- [ ] db.clt File watcher 추가
- [ ] 낙관적 잠금 구현
- [ ] 충돌 해결 UI

## Phase 6: Daemon (선택)
- [ ] clari daemon 명령어
- [ ] WebSocket 서버
- [ ] 실시간 push 동기화

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
