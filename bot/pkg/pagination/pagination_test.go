package pagination

import (
	"testing"
)

func TestNewPageRequest(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{"valid values", 2, 20, 2, 20},
		{"page zero becomes 1", 0, 10, 1, 10},
		{"negative page becomes 1", -5, 10, 1, 10},
		{"pageSize zero becomes default", 1, 0, 1, DefaultPageSize},
		{"negative pageSize becomes default", 1, -1, 1, DefaultPageSize},
		{"both invalid", 0, 0, 1, DefaultPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewPageRequest(tt.page, tt.pageSize)
			if req.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", req.Page, tt.wantPage)
			}
			if req.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %d, want %d", req.PageSize, tt.wantPageSize)
			}
		})
	}
}

func TestPageRequestOffset(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		want     int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 10, 20},
		{1, 5, 0},
		{2, 5, 5},
		{5, 20, 80},
	}

	for _, tt := range tests {
		req := NewPageRequest(tt.page, tt.pageSize)
		got := req.Offset()
		if got != tt.want {
			t.Errorf("Offset() for page=%d, size=%d = %d, want %d", tt.page, tt.pageSize, got, tt.want)
		}
	}
}

func TestPageRequestLimit(t *testing.T) {
	req := NewPageRequest(1, 15)
	if req.Limit() != 15 {
		t.Errorf("Limit() = %d, want 15", req.Limit())
	}
}

func TestNewPageResponse(t *testing.T) {
	tests := []struct {
		name       string
		items      []string
		page       int
		pageSize   int
		totalItems int
		wantPages  int
		wantNext   bool
		wantPrev   bool
	}{
		{
			name:       "first of multiple pages",
			items:      []string{"a", "b"},
			page:       1,
			pageSize:   2,
			totalItems: 5,
			wantPages:  3,
			wantNext:   true,
			wantPrev:   false,
		},
		{
			name:       "middle page",
			items:      []string{"c", "d"},
			page:       2,
			pageSize:   2,
			totalItems: 5,
			wantPages:  3,
			wantNext:   true,
			wantPrev:   true,
		},
		{
			name:       "last page",
			items:      []string{"e"},
			page:       3,
			pageSize:   2,
			totalItems: 5,
			wantPages:  3,
			wantNext:   false,
			wantPrev:   true,
		},
		{
			name:       "single page",
			items:      []string{"a"},
			page:       1,
			pageSize:   10,
			totalItems: 1,
			wantPages:  1,
			wantNext:   false,
			wantPrev:   false,
		},
		{
			name:       "empty items",
			items:      []string{},
			page:       1,
			pageSize:   10,
			totalItems: 0,
			wantPages:  1,
			wantNext:   false,
			wantPrev:   false,
		},
		{
			name:       "invalid page becomes 1",
			items:      []string{"a"},
			page:       0,
			pageSize:   10,
			totalItems: 1,
			wantPages:  1,
			wantNext:   false,
			wantPrev:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewPageResponse(tt.items, tt.page, tt.pageSize, tt.totalItems)

			if resp.TotalPages != tt.wantPages {
				t.Errorf("TotalPages = %d, want %d", resp.TotalPages, tt.wantPages)
			}
			if resp.HasNext != tt.wantNext {
				t.Errorf("HasNext = %v, want %v", resp.HasNext, tt.wantNext)
			}
			if resp.HasPrev != tt.wantPrev {
				t.Errorf("HasPrev = %v, want %v", resp.HasPrev, tt.wantPrev)
			}
			if resp.TotalItems != tt.totalItems {
				t.Errorf("TotalItems = %d, want %d", resp.TotalItems, tt.totalItems)
			}
			if len(resp.Items) != len(tt.items) {
				t.Errorf("Items len = %d, want %d", len(resp.Items), len(tt.items))
			}
		})
	}
}

func TestPageResponseCalculation(t *testing.T) {
	// Edge case: exactly divisible
	resp := NewPageResponse([]int{1, 2, 3, 4, 5}, 1, 5, 15)
	if resp.TotalPages != 3 {
		t.Errorf("TotalPages for 15 items / 5 per page = %d, want 3", resp.TotalPages)
	}

	// Edge case: one extra
	resp = NewPageResponse([]int{1, 2, 3, 4, 5}, 1, 5, 16)
	if resp.TotalPages != 4 {
		t.Errorf("TotalPages for 16 items / 5 per page = %d, want 4", resp.TotalPages)
	}

	// Edge case: one less
	resp = NewPageResponse([]int{1, 2, 3, 4, 5}, 1, 5, 14)
	if resp.TotalPages != 3 {
		t.Errorf("TotalPages for 14 items / 5 per page = %d, want 3", resp.TotalPages)
	}
}
