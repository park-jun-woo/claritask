package model

import "time"

// Project - 프로젝트 테이블
type Project struct {
	ID          string
	Name        string
	Description string
	Status      string // active, archived
	CreatedAt   time.Time
}

// Phase - 단계 테이블
type Phase struct {
	ID          string
	ProjectID   string
	Name        string
	Description string
	OrderNum    int
	Status      string // pending, active, done
	CreatedAt   time.Time
}

// Task - 작업 테이블
type Task struct {
	ID          string
	PhaseID     string
	ParentID    *string // nullable
	Status      string  // pending, doing, done, failed
	Title       string
	Level       string // "", "node", "leaf"
	Skill       string
	References  []string // JSON array
	Content     string
	Result      string
	Error       string
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	FailedAt    *time.Time
}

// Context - 프로젝트 컨텍스트 (싱글톤)
type Context struct {
	ID        int // always 1
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Tech - 기술 스택 (싱글톤)
type Tech struct {
	ID        int // always 1
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Design - 설계 결정 (싱글톤)
type Design struct {
	ID        int // always 1
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// State - 현재 상태 (key-value)
type State struct {
	Key   string
	Value string
}

// Memo - 메모
type Memo struct {
	Scope     string // project, phase, task
	ScopeID   string
	Key       string
	Data      string // JSON
	Priority  int    // 1, 2, 3
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Response - 기본 응답
type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// MemoData - 메모 데이터 (JSON 출력용)
type MemoData struct {
	Scope    string                 `json:"scope"`
	ScopeID  string                 `json:"scope_id"`
	Key      string                 `json:"key"`
	Data     map[string]interface{} `json:"data"`
	Priority int                    `json:"priority"`
}

// TaskPopResponse - task pop 응답
type TaskPopResponse struct {
	Task     *Task    `json:"task"`
	Manifest Manifest `json:"manifest"`
}

// Manifest - task pop 시 반환되는 컨텍스트
type Manifest struct {
	Context map[string]interface{} `json:"context"`
	Tech    map[string]interface{} `json:"tech"`
	Design  map[string]interface{} `json:"design"`
	State   map[string]string      `json:"state"`
	Memos   []MemoData             `json:"memos"`
}
