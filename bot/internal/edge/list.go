package edge

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// EdgeWithTitles includes task titles for display
type EdgeWithTitles struct {
	Edge
	FromTitle string
	ToTitle   string
}

// List lists edges with pagination (all or filtered by taskID)
func List(projectPath, taskID string, req pagination.PageRequest) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var countQuery string
	var countArgs []interface{}
	var listCmd string

	if taskID == "" {
		countQuery = `SELECT COUNT(*) FROM task_edges`
		listCmd = "edge list"
	} else {
		tid, err := strconv.Atoi(taskID)
		if err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("잘못된 task_id: %s", taskID),
			}
		}
		countQuery = `SELECT COUNT(*) FROM task_edges WHERE from_task_id = ? OR to_task_id = ?`
		countArgs = []interface{}{tid, tid}
		listCmd = fmt.Sprintf("edge list %s", taskID)
	}

	// Count total
	var total int
	if err := localDB.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("카운트 실패: %v", err),
		}
	}

	if total == 0 {
		msg := "의존성이 없습니다."
		if taskID != "" {
			msg = fmt.Sprintf("작업 #%s에 대한 의존성이 없습니다.", taskID)
		}
		return types.Result{
			Success: true,
			Message: msg + "\n[추가:edge add]",
		}
	}

	var query string
	var args []interface{}

	if taskID == "" {
		query = `
			SELECT e.from_task_id, e.to_task_id, e.created_at,
			       t1.title as from_title, t2.title as to_title
			FROM task_edges e
			JOIN tasks t1 ON e.from_task_id = t1.id
			JOIN tasks t2 ON e.to_task_id = t2.id
			ORDER BY e.from_task_id, e.to_task_id
			LIMIT ? OFFSET ?
		`
		args = []interface{}{req.Limit(), req.Offset()}
	} else {
		tid, _ := strconv.Atoi(taskID)
		query = `
			SELECT e.from_task_id, e.to_task_id, e.created_at,
			       t1.title as from_title, t2.title as to_title
			FROM task_edges e
			JOIN tasks t1 ON e.from_task_id = t1.id
			JOIN tasks t2 ON e.to_task_id = t2.id
			WHERE e.from_task_id = ? OR e.to_task_id = ?
			ORDER BY e.from_task_id, e.to_task_id
			LIMIT ? OFFSET ?
		`
		args = []interface{}{tid, tid, req.Limit(), req.Offset()}
	}

	rows, err := localDB.Query(query, args...)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}
	defer rows.Close()

	var edges []EdgeWithTitles
	for rows.Next() {
		var e EdgeWithTitles
		if err := rows.Scan(&e.FromTaskID, &e.ToTaskID, &e.CreatedAt, &e.FromTitle, &e.ToTitle); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("스캔 실패: %v", err),
			}
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("행 순회 오류: %v", err),
		}
	}

	pageResp := pagination.NewPageResponse(edges, req.Page, req.PageSize, total)

	var sb strings.Builder
	if taskID == "" {
		sb.WriteString(fmt.Sprintf("📋 의존성 (%d/%d 페이지, 총 %d개)\n", pageResp.Page, pageResp.TotalPages, total))
	} else {
		sb.WriteString(fmt.Sprintf("📋 작업 #%s 의존성 (%d/%d 페이지, 총 %d개)\n", taskID, pageResp.Page, pageResp.TotalPages, total))
	}

	for _, e := range edges {
		sb.WriteString(fmt.Sprintf("  #%d(%s) → #%d(%s) [삭제:edge delete %d %d]\n",
			e.FromTaskID, truncate(e.FromTitle, 15),
			e.ToTaskID, truncate(e.ToTitle, 15),
			e.FromTaskID, e.ToTaskID))
	}

	// Add pagination buttons
	if pageResp.HasPrev || pageResp.HasNext {
		sb.WriteString("\n")
		if pageResp.HasPrev {
			sb.WriteString(fmt.Sprintf("[◀ 이전:%s -p %d]", listCmd, pageResp.Page-1))
		}
		if pageResp.HasNext {
			sb.WriteString(fmt.Sprintf("[다음 ▶:%s -p %d]", listCmd, pageResp.Page+1))
		}
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    pageResp,
	}
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max-2]) + ".."
}
