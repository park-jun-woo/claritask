package task

// MaxDepth is the maximum depth for task subdivision
const MaxDepth = 5

// Task represents a task
type Task struct {
	ID        int
	ParentID  *int
	Title     string
	Spec      string // 요구사항 명세서 (불변)
	Plan      string // 계획서 (1회차 순회에서 생성, leaf만)
	Report    string // 완료 보고서 (2회차 순회 후 생성)
	Status    string // spec_ready → subdivided/plan_ready → done
	Error     string
	IsLeaf    bool // true: 실행 대상, false: 분할됨
	Depth     int  // 트리 깊이 (root=0)
	CreatedAt string
	UpdatedAt string
}

// Stats represents task statistics
type Stats struct {
	Total      int // 전체 task 수
	Leaf       int // 실행 대상 (is_leaf=1)
	SpecReady  int // 명세 작성됨 (plan 대기)
	PlanReady  int // plan 작성됨 (실행 대기)
	Done       int // 완료
	Failed     int // 실패
	InProgress int // 현재 실행 중 (Claude 점유)
}
