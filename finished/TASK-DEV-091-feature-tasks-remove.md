# TASK-DEV-091: feature tasks 서브커맨드 정리

## 개요

`clari feature tasks` 서브커맨드가 specs에 없음. 제거하거나 문서화 필요

## 현재 상태

cmd/feature.go에 `feature tasks` 서브커맨드 존재:
```go
var featureTasksCmd = &cobra.Command{
    Use:   "tasks <id>",
    Short: "List or generate tasks for feature",
    ...
}
```

## 스펙 확인 (07-Feature.md)

정의된 서브커맨드:
- `clari feature list`
- `clari feature add`
- `clari feature get`
- `clari feature spec`
- `clari feature start`

**`feature tasks`는 없음**

## 관련 명령어

`clari fdl tasks <feature_id>` (09-FDL.md)가 동일 기능:
> FDL에서 Task 자동 생성

## 결정 필요

### 옵션 1: 제거
- feature tasks 삭제
- fdl tasks만 사용

### 옵션 2: 문서화
- specs/CLI/07-Feature.md에 추가
- feature tasks와 fdl tasks 구분

## 작업 내용 (옵션 1 선택 시)

### cmd/feature.go 수정

```go
// 제거할 코드
// var featureTasksCmd = ...
// featureCmd.AddCommand(featureTasksCmd)
// func runFeatureTasks(...)
```

## 완료 조건

- [ ] 옵션 결정 (제거 or 문서화)
- [ ] 선택한 옵션 구현
