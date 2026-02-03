# 2026-02-04 Messages 기능 및 Config 구현

## 1. Messages 기능 구현

### 스펙 문서 업데이트
- `specs/CLI/15-Message.md` - Message 명령어 상세 스펙 (이전 세션에서 작성)
- `specs/DB/02-C-Content.md` - messages, message_tasks 테이블 스키마
- `specs/CLI/01-Overview.md` - message 명령어 추가 (v0.0.10)

### DB 마이그레이션 (TASK-DEV-108)
- `cli/internal/db/db.go`
  - LatestVersion: 6 → 7
  - messages 테이블 추가
  - message_tasks 테이블 추가 (Message-Task 연결)
  - 인덱스: idx_messages_project, idx_messages_status, idx_message_tasks_message

### Message 모델 (TASK-DEV-109)
- `cli/internal/model/models.go`
```go
type Message struct {
    ID, ProjectID, FeatureID, Content, Response, Status, Error, CreatedAt, CompletedAt
}
type MessageTask struct { MessageID, TaskID, CreatedAt }
type MessageListItem struct { ID, Content, Status, FeatureID, TasksCount, CreatedAt }
type MessageDetail struct { Message + Tasks []TaskListItem }
type TaskListItem struct { ID, Title, Status }
```

### Message Service (TASK-DEV-110)
- `cli/internal/service/message_service.go` (신규)
  - CreateMessage, GetMessage, ListMessages
  - UpdateMessageStatus, DeleteMessage
  - LinkMessageTask, CountMessageTasks
  - MessageAnalysisSystemPrompt, RunMessageAnalysisWithTTY
  - SaveMessageReport

### Message 명령어 (TASK-DEV-111)
- `cli/internal/cmd/message.go` (신규)
  - `clari message send <content>` - 수정 요청 전송
  - `clari message list` - 메시지 목록
  - `clari message get <id>` - 상세 조회
  - `clari message delete <id>` - 삭제

### 테스트 (TASK-DEV-112)
- `cli/test/message_service_test.go` - 12개 테스트 케이스 모두 통과

---

## 2. Config 및 TTY 세션 관리 구현

### 배경 (토론 모드)
- VSCode에서 clari CLI 실행 시 터미널 관리 방안 논의
- **결론**:
  - clari CLI: 무제한 실행 가능
  - Claude Code 세션: max_parallel_sessions 설정에 따라 제한 (기본값: 3)
  - 먼저 실행된 세션이 우선권 (FIFO)
  - 설정은 `.claritask/config.yaml`로 분리

### 스펙 문서
- `specs/CLI/16-Config.md` (신규) - Config 설정 파일 스펙
- `specs/VSCode/14-CLICompatibility.md` (v0.0.7) - TTY 세션 관리 섹션 추가

### Config 설정 파일 형식
```yaml
# .claritask/config.yaml
tty:
  max_parallel_sessions: 3  # Claude Code 동시 실행 제한 (1-10)
  terminal_close_delay: 1   # 완료 후 터미널 종료 대기 (초, -1: 안닫음)
vscode:
  sync_interval: 1000       # DB 동기화 간격 (ms)
  watch_feature_files: true # FDL 파일 감시
```

### CLI 구현 (TASK-DEV-113, 114)
- `cli/internal/service/config_service.go` (신규)
```go
type Config struct {
    TTY    TTYConfig    `yaml:"tty"`
    VSCode VSCodeConfig `yaml:"vscode"`
}
func LoadConfig() (*Config, error)
func DefaultConfig() *Config
```

- `cli/internal/service/tty_service.go` (수정)
```go
var (
    sessionMutex   sync.Mutex
    activeSessions int
    sessionCond    *sync.Cond
)
func acquireSession(maxSessions int) // 대기 가능
func releaseSession()
func GetSessionStatus() (active, max int)
```

### VSCode Extension 구현 (TASK-EXT-037, 038, 039)

#### configService.ts (TASK-EXT-037)
```typescript
interface ClaritaskConfig {
  tty: { max_parallel_sessions, terminal_close_delay };
  vscode: { sync_interval, watch_feature_files };
}
function loadConfig(workspacePath: string): ClaritaskConfig
```

#### ttySessionManager.ts (TASK-EXT-038)
```typescript
class TTYSessionManager {
  private activeSessions: Map<string, Terminal>
  private waitingQueue: PendingSession[]

  async startSession(id, command): Promise<Terminal>
  getStatus(): { active, waiting, max }
  onStatusChange: Event<SessionStatus>
}
```

#### extension.ts 수정 (TASK-EXT-039)
- Session StatusBar 추가: `$(terminal) Claude: 2/3 (1 대기)`
- `claritask.showSessionStatus` 명령어 등록
- ttySessionManager 전역 export

#### CltEditorProvider.ts 수정
- handleCreateFeature에서 ttySessionManager 사용

### 버전 업데이트
- CLI: 테스트 통과
- VSCode Extension: v0.0.7 → v0.0.8
- yaml 패키지 의존성 추가

---

## 3. 완료된 TASK 파일

### Messages 관련
- TASK-DEV-108-messages-db-migration.md
- TASK-DEV-109-message-models.md
- TASK-DEV-110-message-service.md
- TASK-DEV-111-message-cmd.md
- TASK-DEV-112-message-service-test.md

### Config/TTY 관련
- TASK-DEV-113-config-service.md
- TASK-DEV-114-tty-session-limit.md
- TASK-EXT-037-config-service.md
- TASK-EXT-038-tty-session-manager.md
- TASK-EXT-039-session-statusbar.md

---

## 4. 참고: 이전 세션에서 완료된 작업

(컨텍스트 요약에서 가져옴)
- FDL 문서 Go embed 구현 (`cli/internal/docs/fdl_spec.md`, `embed.go`)
- `.claritask/complete` 파일 감지 로직 (`watchCompleteFile`)
- `--dangerously-skip-permissions` 옵션 추가
- VSCode Feature 추가 시 WSL 경로 변환 (`windowsToWslPath`)
- VSCode Delete 버튼 수정 (ConfirmModal, task_edges 삭제)
