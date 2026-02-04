package pagination

// PageRequest represents pagination request parameters
type PageRequest struct {
	Page     int // 현재 페이지 (1-based)
	PageSize int // 페이지당 항목 수
}

// PageResponse represents paginated response
type PageResponse[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 10

// NewPageRequest creates a new PageRequest with defaults
func NewPageRequest(page, pageSize int) PageRequest {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	return PageRequest{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset returns the SQL OFFSET value
func (p PageRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the SQL LIMIT value
func (p PageRequest) Limit() int {
	return p.PageSize
}

// NewPageResponse creates a new PageResponse from items and total count
func NewPageResponse[T any](items []T, page, pageSize, totalItems int) PageResponse[T] {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}

	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return PageResponse[T]{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
