package task

import (
	"fmt"
	"time"
)

// MaxDepth is the maximum depth for task subdivision
const MaxDepth = 5

// globalNotifier is the callback for sending notifications (e.g. Telegram)
var globalNotifier func(projectID *string, msg string)

// Init initializes the task package with a notifier callback
func Init(notifier func(projectID *string, msg string)) {
	globalNotifier = notifier
}

// formatDuration formats a duration as "Xm Ys"
func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", m, s)
}

// Task represents a task
type Task struct {
	ID        int    `json:"id"`
	ParentID  *int   `json:"parent_id,omitempty"`
	Title     string `json:"title"`
	Spec      string `json:"spec"`      // 요구사항 명세서 (불변)
	Plan      string `json:"plan"`      // 계획서 (1회차 순회에서 생성, leaf만)
	Report    string `json:"report"`    // 완료 보고서 (2회차 순회 후 생성)
	Status    string `json:"status"`    // todo → split/planned → done
	Error     string `json:"error,omitempty"`
	IsLeaf    bool   `json:"is_leaf"`   // true: 실행 대상, false: 분할됨
	Depth     int    `json:"depth"`     // 트리 깊이 (root=0)
	Priority  int    `json:"priority"`  // 실행 우선순위 (높을수록 먼저 실행)
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Stats represents task statistics
type Stats struct {
	Total      int `json:"total"`       // 전체 task 수
	Leaf       int `json:"leaf"`        // 실행 대상 (is_leaf=1)
	Todo       int `json:"todo"`        // 명세 작성됨 (plan 대기)
	Planned    int `json:"planned"`     // plan 작성됨 (실행 대기)
	Done       int `json:"done"`        // 완료
	Failed     int `json:"failed"`      // 실패
	InProgress int `json:"in_progress"` // 현재 실행 중 (Claude 점유)
}
