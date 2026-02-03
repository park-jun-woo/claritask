# TASK-DEV-111: Message 명령어 구현

## 목표
`clari message send/list/get/delete` 명령어 구현

## 변경 파일
- `cli/internal/cmd/message.go` (신규)
- `cli/internal/cmd/root.go` (import 추가)

## 작업 내용

### 1. messageCmd 부모 명령어
```go
var messageCmd = &cobra.Command{
    Use:   "message",
    Short: "Manage messages",
}

func init() {
    rootCmd.AddCommand(messageCmd)
    messageCmd.AddCommand(messageSendCmd)
    messageCmd.AddCommand(messageListCmd)
    messageCmd.AddCommand(messageGetCmd)
    messageCmd.AddCommand(messageDeleteCmd)
}
```

### 2. message send 명령어
```go
var messageSendCmd = &cobra.Command{
    Use:   "send <content>",
    Short: "Send a modification request and convert to tasks",
    Args:  cobra.ExactArgs(1),
    RunE:  runMessageSend,
}

func init() {
    messageSendCmd.Flags().Int64P("feature", "f", 0, "Related feature ID")
}

func runMessageSend(cmd *cobra.Command, args []string) error {
    // 1. Get current project
    // 2. Create message
    // 3. Run TTY handover for Claude analysis
    // 4. Output result JSON
}
```

### 3. message list 명령어
```go
var messageListCmd = &cobra.Command{
    Use:   "list",
    Short: "List messages",
    RunE:  runMessageList,
}

func init() {
    messageListCmd.Flags().StringP("status", "s", "", "Filter by status")
    messageListCmd.Flags().Int64P("feature", "f", 0, "Filter by feature ID")
    messageListCmd.Flags().IntP("limit", "l", 20, "Max number of results")
}
```

### 4. message get 명령어
```go
var messageGetCmd = &cobra.Command{
    Use:   "get <id>",
    Short: "Get message details",
    Args:  cobra.ExactArgs(1),
    RunE:  runMessageGet,
}
```

### 5. message delete 명령어
```go
var messageDeleteCmd = &cobra.Command{
    Use:   "delete <id>",
    Short: "Delete a message",
    Args:  cobra.ExactArgs(1),
    RunE:  runMessageDelete,
}
```

## JSON 출력 형식

### send 성공
```json
{
  "success": true,
  "message_id": 1,
  "tasks_created": 3,
  "report_path": "reports/2026-02-04-message-001.md"
}
```

### list 출력
```json
{
  "success": true,
  "messages": [...],
  "total": 5
}
```

## 테스트
- `clari message send "로그인 페이지에 소셜 로그인 추가해줘"`
- `clari message send --feature 1 "API 에러 핸들링 개선"`
- `clari message list`
- `clari message list --status completed`
- `clari message get 1`
- `clari message delete 1`
