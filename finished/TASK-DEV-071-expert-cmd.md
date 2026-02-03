# TASK-DEV-071: Expert 명령어 구현

## 목표
`internal/cmd/expert.go` 신규 파일 생성

## 작업 내용

### 1. 기본 구조
```go
var expertCmd = &cobra.Command{
    Use:   "expert",
    Short: "Expert management commands",
}

func init() {
    expertCmd.AddCommand(expertAddCmd)
    expertCmd.AddCommand(expertListCmd)
    expertCmd.AddCommand(expertGetCmd)
    expertCmd.AddCommand(expertEditCmd)
    expertCmd.AddCommand(expertRemoveCmd)
    expertCmd.AddCommand(expertAssignCmd)
    expertCmd.AddCommand(expertUnassignCmd)
}
```

### 2. clari expert add
```go
var expertAddCmd = &cobra.Command{
    Use:   "add <expert-id>",
    Short: "Create a new expert",
    Args:  cobra.ExactArgs(1),
    RunE:  runExpertAdd,
}
```
- service.AddExpert 호출
- JSON 응답 출력

### 3. clari expert list
```go
var expertListCmd = &cobra.Command{
    Use:   "list",
    Short: "List experts",
    RunE:  runExpertList,
}
```
- `--assigned` 플래그: 할당된 것만
- `--available` 플래그: 미할당만
- service.ListExperts 호출

### 4. clari expert get
```go
var expertGetCmd = &cobra.Command{
    Use:   "get <expert-id>",
    Short: "Get expert details",
    Args:  cobra.ExactArgs(1),
    RunE:  runExpertGet,
}
```
- service.GetExpert 호출

### 5. clari expert edit
```go
var expertEditCmd = &cobra.Command{
    Use:   "edit <expert-id>",
    Short: "Edit expert file",
    Args:  cobra.ExactArgs(1),
    RunE:  runExpertEdit,
}
```
- $EDITOR 환경변수로 에디터 실행
- 없으면 vi (linux) 또는 notepad (windows)

### 6. clari expert remove
```go
var expertRemoveCmd = &cobra.Command{
    Use:   "remove <expert-id>",
    Short: "Remove an expert",
    Args:  cobra.ExactArgs(1),
    RunE:  runExpertRemove,
}
```
- `--force` 플래그
- service.RemoveExpert 호출

### 7. clari expert assign
```go
var expertAssignCmd = &cobra.Command{
    Use:   "assign <expert-id>",
    Short: "Assign expert to project",
    Args:  cobra.ExactArgs(1),
    RunE:  runExpertAssign,
}
```
- `--project` 플래그 (기본값: 현재 프로젝트)
- service.AssignExpert 호출

### 8. clari expert unassign
```go
var expertUnassignCmd = &cobra.Command{
    Use:   "unassign <expert-id>",
    Short: "Unassign expert from project",
    Args:  cobra.ExactArgs(1),
    RunE:  runExpertUnassign,
}
```
- `--project` 플래그
- service.UnassignExpert 호출

### 9. root.go 수정
- `rootCmd.AddCommand(expertCmd)` 추가

## 완료 조건
- [ ] expertCmd 정의
- [ ] expertAddCmd 구현
- [ ] expertListCmd 구현
- [ ] expertGetCmd 구현
- [ ] expertEditCmd 구현
- [ ] expertRemoveCmd 구현
- [ ] expertAssignCmd 구현
- [ ] expertUnassignCmd 구현
- [ ] root.go에 등록
- [ ] 컴파일 성공
