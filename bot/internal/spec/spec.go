package spec

// Spec represents a requirements specification document
type Spec struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Status    string `json:"status"`   // draft, review, approved, deprecated
	Priority  int    `json:"priority"` // 우선순위
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
