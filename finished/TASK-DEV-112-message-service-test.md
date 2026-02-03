# TASK-DEV-112: Message Service 테스트

## 목표
`message_service_test.go` 파일 생성 - Message Service 단위 테스트

## 변경 파일
- `cli/test/message_service_test.go` (신규)

## 작업 내용

### 1. TestCreateMessage
```go
func TestCreateMessage(t *testing.T) {
    // 1. Setup test DB
    // 2. Create project
    // 3. Create message
    // 4. Verify message exists with status 'pending'
}
```

### 2. TestGetMessage
```go
func TestGetMessage(t *testing.T) {
    // 1. Setup test DB with message and tasks
    // 2. Get message detail
    // 3. Verify tasks are included
}
```

### 3. TestListMessages
```go
func TestListMessages(t *testing.T) {
    // 1. Setup test DB with multiple messages
    // 2. Test list without filters
    // 3. Test list with status filter
    // 4. Test list with feature filter
    // 5. Test list with limit
}
```

### 4. TestUpdateMessageStatus
```go
func TestUpdateMessageStatus(t *testing.T) {
    // 1. Setup test DB with pending message
    // 2. Update to processing
    // 3. Update to completed with response
    // 4. Verify completed_at is set
}
```

### 5. TestDeleteMessage
```go
func TestDeleteMessage(t *testing.T) {
    // 1. Setup test DB with message
    // 2. Delete message
    // 3. Verify message is deleted
}
```

### 6. TestLinkMessageTask
```go
func TestLinkMessageTask(t *testing.T) {
    // 1. Setup test DB with message and task
    // 2. Link them
    // 3. Verify link exists
}
```

## 테스트 실행
```bash
go test ./test/message_service_test.go -v
```
