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
	ID             string
	PhaseID        string
	ParentID       *string // nullable
	Status         string  // pending, doing, done, failed
	Title          string
	Level          string // "", "node", "leaf"
	Skill          string
	References     []string // JSON array
	Content        string
	Result         string
	Error          string
	FeatureID      *int64  // Feature ID (nullable)
	SkeletonID     *int64  // Skeleton ID (nullable)
	TargetFile     string  // 구현 대상 파일 경로
	TargetLine     *int    // 구현 대상 라인 번호
	TargetFunction string  // 구현 대상 함수명
	CreatedAt      time.Time
	StartedAt      *time.Time
	CompletedAt    *time.Time
	FailedAt       *time.Time
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

// Feature - Feature 테이블
type Feature struct {
	ID                int64
	ProjectID         string
	Name              string
	Description       string
	Spec              string // Feature 상세 스펙
	FDL               string // FDL YAML 원문
	FDLHash           string // FDL 변경 감지용 해시
	SkeletonGenerated bool   // 스켈레톤 생성 완료 여부
	Status            string // pending, active, done
	CreatedAt         time.Time
}

// FeatureEdge - Feature 간 의존성
type FeatureEdge struct {
	FromFeatureID int64
	ToFeatureID   int64
	CreatedAt     time.Time
}

// TaskEdge - Task 간 의존성
type TaskEdge struct {
	FromTaskID string
	ToTaskID   string
	CreatedAt  time.Time
}

// Skeleton - 생성된 스켈레톤 파일 추적
type Skeleton struct {
	ID        int64
	FeatureID int64
	FilePath  string // 생성된 파일 경로
	Layer     string // model, service, api, ui
	Checksum  string // 파일 변경 감지용
	CreatedAt time.Time
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

// FDLInfo - FDL 정보 (task pop 시 반환)
type FDLInfo struct {
	Feature string                 `json:"feature"`
	Service map[string]interface{} `json:"service,omitempty"`
	Models  map[string]interface{} `json:"models,omitempty"`
	API     map[string]interface{} `json:"api,omitempty"`
	UI      map[string]interface{} `json:"ui,omitempty"`
}

// SkeletonInfo - 스켈레톤 정보 (task pop 시 반환)
type SkeletonInfo struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"current_content"`
}

// Dependency - 의존 Task 정보
type Dependency struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Result string `json:"result"`
	File   string `json:"file,omitempty"`
}

// TaskPopResponse - task pop 응답
type TaskPopResponse struct {
	Task         *Task         `json:"task"`
	FDL          *FDLInfo      `json:"fdl,omitempty"`
	Skeleton     *SkeletonInfo `json:"skeleton,omitempty"`
	Dependencies []Dependency  `json:"dependencies,omitempty"`
	Manifest     Manifest      `json:"manifest"`
}

// Manifest - task pop 시 반환되는 컨텍스트
type Manifest struct {
	Context map[string]interface{} `json:"context"`
	Tech    map[string]interface{} `json:"tech"`
	Design  map[string]interface{} `json:"design"`
	Feature map[string]interface{} `json:"feature,omitempty"`
	State   map[string]string      `json:"state"`
	Memos   []MemoData             `json:"memos"`
}
