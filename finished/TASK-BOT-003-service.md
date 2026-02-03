# TASK-BOT-003: Service 모듈

## 목표
CLI의 DB/Service 연동 - 기존 CLI 패키지 재사용

## 파일
`bot/internal/service/service.go`

## 작업 내용

### 접근 방식
CLI의 internal 패키지를 직접 import하여 재사용

```go
import (
    clidb "claritask/cli/internal/db"
    "claritask/cli/internal/model"
)
```

### Service 구조체
```go
type Service struct {
    db *clidb.DB
}

func New(dbPath string) (*Service, error)
func (s *Service) Close() error
```

### 프로젝트 관련
```go
func (s *Service) GetCurrentProject() (*model.Project, error)
func (s *Service) ListProjects() ([]model.Project, error)
func (s *Service) GetProjectStatus(projectID string) (*ProjectStatus, error)
```

### 태스크 관련
```go
func (s *Service) ListTasks(projectID string, status string) ([]model.Task, error)
func (s *Service) GetTask(id int) (*model.Task, error)
func (s *Service) StartTask(id int) error
func (s *Service) CompleteTask(id int) error
func (s *Service) FailTask(id int, reason string) error
```

### 메시지 관련
```go
func (s *Service) ListMessages(projectID string, limit int) ([]model.Message, error)
func (s *Service) SendMessage(projectID, title, content, recipient string) error
```

### Expert 관련
```go
func (s *Service) ListExperts(projectID string) ([]model.Expert, error)
```

### ProjectStatus 구조체
```go
type ProjectStatus struct {
    TotalTasks     int
    CompletedTasks int
    InProgressTasks int
    PendingTasks   int
    Progress       float64
}
```

## 완료 조건
- [ ] CLI 패키지 import 설정
- [ ] Service 구조체 구현
- [ ] 프로젝트/태스크/메시지 함수 구현
