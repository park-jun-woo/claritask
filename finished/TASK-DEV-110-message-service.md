# TASK-DEV-110: Message Service 구현

## 목표
`message_service.go` 파일 생성 - Message CRUD 및 TTY 연동

## 변경 파일
- `cli/internal/service/message_service.go` (신규)

## 작업 내용

### 1. CreateMessage 함수
```go
func CreateMessage(database *db.DB, projectID string, featureID *int64, content string) (*model.Message, error) {
    // 1. INSERT into messages (status: 'pending')
    // 2. Return created message
}
```

### 2. GetMessage 함수
```go
func GetMessage(database *db.DB, messageID int64) (*model.MessageDetail, error) {
    // 1. SELECT message by ID
    // 2. SELECT associated tasks via message_tasks
    // 3. Return MessageDetail with tasks
}
```

### 3. ListMessages 함수
```go
func ListMessages(database *db.DB, projectID string, status string, featureID *int64, limit int) ([]model.MessageListItem, error) {
    // 1. Build query with optional filters
    // 2. Count tasks for each message
    // 3. Return list
}
```

### 4. UpdateMessageStatus 함수
```go
func UpdateMessageStatus(database *db.DB, messageID int64, status, response, errMsg string) error {
    // 1. UPDATE status, response, error
    // 2. Set completed_at if status is 'completed' or 'failed'
}
```

### 5. DeleteMessage 함수
```go
func DeleteMessage(database *db.DB, messageID int64) error {
    // 1. DELETE from messages (message_tasks cascade)
}
```

### 6. LinkMessageTask 함수
```go
func LinkMessageTask(database *db.DB, messageID, taskID int64) error {
    // 1. INSERT into message_tasks
}
```

### 7. MessageAnalysisSystemPrompt 함수
```go
func MessageAnalysisSystemPrompt() string {
    // Return system prompt for Claude Code message analysis mode
}
```

### 8. RunMessageAnalysisWithTTY 함수
```go
func RunMessageAnalysisWithTTY(database *db.DB, message *model.Message) error {
    // 1. Update status to 'processing'
    // 2. Prepare system prompt and initial prompt
    // 3. Call RunWithTTYHandoverEx with complete file
    // 4. Read complete file content for response
    // 5. Update status to 'completed' with response
    // 6. Save report to reports/ folder
}
```

## 테스트
- `go test ./test/message_service_test.go -v` 실행
