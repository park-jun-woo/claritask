# TASK-DEV-083: Init TTY Handover 구현

## 개요

specs/CLI/02-Init.md의 Phase 2-5 구현

## 스펙 요구사항

### Phase 구조
1. **Phase 1**: DB 초기화 ✓ (구현됨)
2. **Phase 2**: TTY Handover → Claude Code
   - ScanProjectFiles 호출
   - LLM으로 tech/design 분석
3. **Phase 3**: 사용자 승인
4. **Phase 4**: Specs 문서 생성
5. **Phase 5**: 피드백 루프

## 현재 상태

- Phase 1만 동작
- TTY Handover 미구현
- LLM 연동 미구현

## 작업 내용

### 1. Phase 2 구현 (init_service.go)

```go
func (s *InitService) InitPhase2_Analysis(projectID string) (*AnalysisResult, error) {
    // 1. 프로젝트 파일 스캔
    files, err := s.scanner.ScanProjectFiles(".")
    if err != nil {
        return nil, err
    }

    // 2. 분석 프롬프트 생성
    prompt := s.prompt.BuildContextAnalysisPrompt(files)

    // 3. LLM 호출 또는 TTY Handover
    // TTY Handover: 프롬프트와 지시사항 반환
    return &AnalysisResult{
        Prompt: prompt,
        Instructions: "Use Claude Code to analyze the project...",
        Files: files,
    }, nil
}
```

### 2. Phase 3 구현

```go
func (s *InitService) InitPhase3_Approval(analysis *AnalysisResult) error {
    // 1. 사용자에게 분석 결과 표시
    // 2. 수정 가능하도록
    // 3. 승인 대기
    // JSON 응답으로 approval_pending 상태 반환
    return nil
}
```

### 3. Phase 4 구현

```go
func (s *InitService) InitPhase4_SpecsGeneration(approved *ApprovedAnalysis) (*SpecsResult, error) {
    // 1. specs 문서 생성 프롬프트
    prompt := s.prompt.BuildSpecsGenerationPrompt(approved)

    // 2. TTY Handover 또는 LLM 호출
    return &SpecsResult{
        Prompt: prompt,
        Instructions: "Generate specs document...",
    }, nil
}
```

### 4. Phase 5 구현

```go
func (s *InitService) InitPhase5_Feedback(specs *SpecsResult) error {
    // 피드백 루프
    // 사용자가 만족할 때까지 반복
    return nil
}
```

### 5. State 관리

```go
// 상태 저장 키
const (
    StateInitPhase = "init_phase"  // 현재 phase
    StateInitData  = "init_data"   // phase별 데이터
)
```

### 6. 명령어 수정 (cmd/init.go)

```go
// --resume 플래그로 중단된 곳에서 재개
if resumeFlag {
    return service.ResumeInit(database)
}
```

## 응답 스키마

### Phase 2 응답 (분석 완료)
```json
{
  "success": true,
  "phase": 2,
  "status": "analysis_complete",
  "analysis": {
    "files_scanned": 42,
    "suggested_tech": {...},
    "suggested_design": {...}
  },
  "next_action": "Run 'clari init --resume' to continue"
}
```

### Phase 3 응답 (승인 대기)
```json
{
  "success": true,
  "phase": 3,
  "status": "approval_pending",
  "context": {...},
  "tech": {...},
  "design": {...},
  "message": "Review and approve the analysis"
}
```

## 완료 조건

- [ ] Phase 2: 파일 스캔 및 분석 프롬프트 생성
- [ ] Phase 3: 승인 대기 상태 처리
- [ ] Phase 4: Specs 생성 프롬프트 생성
- [ ] Phase 5: 피드백 루프 구현
- [ ] --resume 플래그 동작
- [ ] 각 Phase의 상태 저장/복원
- [ ] 테스트 작성
