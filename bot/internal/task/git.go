package task

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

var gitMu sync.Mutex     // git 명령어 직렬화
var batchMode atomic.Bool // true: 개별 커밋 억제

// SetBatchMode enables/disables batch mode (suppresses per-task commits).
func SetBatchMode(enabled bool) { batchMode.Store(enabled) }

// IsBatchMode returns whether batch mode is active.
func IsBatchMode() bool { return batchMode.Load() }

var (
	gitAvailableOnce sync.Once
	gitAvailable     bool
)

// isGitAvailable checks if git is installed (cached).
func isGitAvailable() bool {
	gitAvailableOnce.Do(func() {
		_, err := exec.LookPath("git")
		gitAvailable = err == nil
	})
	return gitAvailable
}

// isGitRepo checks if projectPath is inside a git repository.
func isGitRepo(projectPath string) bool {
	if !isGitAvailable() {
		return false
	}
	cmd := exec.Command("git", "-C", projectPath, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// GitAdd stages files in the git repository.
func GitAdd(projectPath string, paths ...string) error {
	if !isGitAvailable() || !isGitRepo(projectPath) {
		return nil
	}
	gitMu.Lock()
	defer gitMu.Unlock()

	args := append([]string{"-C", projectPath, "add"}, paths...)
	cmd := exec.Command("git", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// GitCommit creates a commit in the git repository.
// "nothing to commit" is silently ignored.
func GitCommit(projectPath, message string) error {
	if !isGitAvailable() || !isGitRepo(projectPath) {
		return nil
	}
	gitMu.Lock()
	defer gitMu.Unlock()

	cmd := exec.Command("git", "-C", projectPath, "commit", "-m", message)
	out, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(out))
		// "nothing to commit" is not an error
		if strings.Contains(outStr, "nothing to commit") {
			return nil
		}
		return fmt.Errorf("git commit failed: %s: %w", outStr, err)
	}
	return nil
}

// GitRestore restores a file from HEAD (used for accidental deletion recovery).
func GitRestore(projectPath, filePath string) error {
	if !isGitAvailable() || !isGitRepo(projectPath) {
		return nil
	}
	gitMu.Lock()
	defer gitMu.Unlock()

	cmd := exec.Command("git", "-C", projectPath, "checkout", "HEAD", "--", filePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git restore failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// gitCommitTask adds and commits task-related files for a single task.
// No-op in batch mode.
func gitCommitTask(projectPath string, taskID int, action string) {
	if IsBatchMode() {
		return
	}
	if !isGitAvailable() || !isGitRepo(projectPath) {
		return
	}

	taskDir := filepath.Join(".claribot", taskDirName)
	idStr := fmt.Sprintf("%d", taskID)

	if action == "deleted" {
		// Stage deleted files with git add -u
		paths := []string{
			filepath.Join(taskDir, idStr+".md"),
			filepath.Join(taskDir, idStr+".plan.md"),
			filepath.Join(taskDir, idStr+".report.md"),
			filepath.Join(taskDir, idStr+".error.md"),
		}
		for _, p := range paths {
			if err := GitAdd(projectPath, p); err != nil {
				log.Printf("[Task] git add 실패 (%s): %v", p, err)
			}
		}
	} else {
		// Stage existing task files
		files := []string{
			filepath.Join(taskDir, idStr+".md"),
			filepath.Join(taskDir, idStr+".plan.md"),
			filepath.Join(taskDir, idStr+".report.md"),
			filepath.Join(taskDir, idStr+".error.md"),
		}
		for _, f := range files {
			if err := GitAdd(projectPath, f); err != nil {
				// Ignore errors for files that don't exist
				continue
			}
		}
	}

	msg := fmt.Sprintf("task(#%d): %s", taskID, action)
	if err := GitCommit(projectPath, msg); err != nil {
		log.Printf("[Task] git commit 실패 (task #%d, %s): %v", taskID, action, err)
	}
}

// gitCommitBatch commits all task files in a single batch commit.
func gitCommitBatch(projectPath, message string) {
	if !isGitAvailable() || !isGitRepo(projectPath) {
		return
	}

	taskDir := filepath.Join(".claribot", taskDirName)
	if err := GitAdd(projectPath, taskDir); err != nil {
		log.Printf("[Task] git add 실패 (batch): %v", err)
		return
	}
	if err := GitCommit(projectPath, message); err != nil {
		log.Printf("[Task] git commit 실패 (batch): %v", err)
	}
}

// summarizeResult extracts a summary from result message for commit messages.
// E.g. "📋 Plan 생성 완료: 성공 3개, 실패 1개\n..." → "성공 3개, 실패 1개"
func summarizeResult(msg string) string {
	firstLine := strings.SplitN(msg, "\n", 2)[0]
	// Try to extract after ": "
	if idx := strings.LastIndex(firstLine, ": "); idx >= 0 {
		return firstLine[idx+2:]
	}
	// Truncate if too long
	runes := []rune(firstLine)
	if len(runes) > 60 {
		return string(runes[:60])
	}
	return firstLine
}
