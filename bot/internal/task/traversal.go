package task

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"parkjunwoo.com/claribot/internal/db"
)

// insertTraversal inserts a new traversal record and returns its ID.
func insertTraversal(localDB *db.DB, travType string, targetID *int, trigger string) (int64, error) {
	now := db.TimeNow()
	res, err := localDB.Exec(`
		INSERT INTO traversals (type, target_id, trigger, status, started_at)
		VALUES (?, ?, ?, 'running', ?)
	`, travType, targetID, trigger, now)
	if err != nil {
		return 0, fmt.Errorf("traversal INSERT: %w", err)
	}
	return res.LastInsertId()
}

// finishTraversal updates a traversal record with final status and counts.
func finishTraversal(localDB *db.DB, travID int64, status string, total, success, failed int) {
	now := db.TimeNow()
	_, err := localDB.Exec(`
		UPDATE traversals SET status = ?, total = ?, success = ?, failed = ?, finished_at = ?
		WHERE id = ?
	`, status, total, success, failed, now, travID)
	if err != nil {
		log.Printf("[Task] traversal UPDATE 실패 (id=%d): %v", travID, err)
	}
}

// countFromMessage extracts total/success/failed counts from result message.
// Matches patterns like "성공 3개, 실패 1개"
var reSuccessFailed = regexp.MustCompile(`성공\s*(\d+)개.*실패\s*(\d+)개`)

func countFromMessage(msg string) (total, success, failed int) {
	m := reSuccessFailed.FindStringSubmatch(msg)
	if len(m) == 3 {
		success, _ = strconv.Atoi(m[1])
		failed, _ = strconv.Atoi(m[2])
		total = success + failed
	}
	return
}
