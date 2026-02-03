# TASK-DEV-109: Message 모델 정의

## 목표
Message 관련 모델 구조체 정의

## 변경 파일
- `cli/internal/model/models.go`

## 작업 내용

### 1. Message 구조체 추가
```go
// Message represents a user modification request
type Message struct {
    ID          int64   `json:"id"`
    ProjectID   string  `json:"project_id"`
    FeatureID   *int64  `json:"feature_id,omitempty"`
    Content     string  `json:"content"`
    Response    string  `json:"response,omitempty"`
    Status      string  `json:"status"`
    Error       string  `json:"error,omitempty"`
    CreatedAt   string  `json:"created_at"`
    CompletedAt *string `json:"completed_at,omitempty"`
}
```

### 2. MessageTask 구조체 추가
```go
// MessageTask represents the relationship between a message and its created tasks
type MessageTask struct {
    MessageID int64  `json:"message_id"`
    TaskID    int64  `json:"task_id"`
    CreatedAt string `json:"created_at"`
}
```

### 3. MessageListItem 구조체 추가
```go
// MessageListItem represents a message in list view
type MessageListItem struct {
    ID         int64   `json:"id"`
    Content    string  `json:"content"`
    Status     string  `json:"status"`
    FeatureID  *int64  `json:"feature_id,omitempty"`
    TasksCount int     `json:"tasks_count"`
    CreatedAt  string  `json:"created_at"`
}
```

### 4. MessageDetail 구조체 추가
```go
// MessageDetail represents a message with associated tasks
type MessageDetail struct {
    Message
    Tasks []TaskListItem `json:"tasks,omitempty"`
}
```

## 테스트
- `go test ./test/models_test.go -v` 실행
