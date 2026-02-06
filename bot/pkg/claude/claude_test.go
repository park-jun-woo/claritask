package claude

import (
	"sync"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	// Reset manager for test
	Init(Config{
		Timeout: 30 * time.Second,
		Max:     3,
	})

	result, err := Run(Options{
		SystemPrompt: "You are a helpful assistant. Keep responses very brief.",
		UserPrompt:   "Say 'Hello Claribot' and nothing else.",
	})

	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	t.Logf("Exit code: %d", result.ExitCode)
	t.Logf("Output:\n%s", result.Output)

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		result   *Result
		expected bool
	}{
		{
			name:     "nil result",
			result:   nil,
			expected: false,
		},
		{
			name:     "exit code 0",
			result:   &Result{Output: "does not have access", ExitCode: 0},
			expected: false,
		},
		{
			name:     "does not have access",
			result:   &Result{Output: "Error: This account does not have access to Claude", ExitCode: 1},
			expected: true,
		},
		{
			name:     "please login again",
			result:   &Result{Output: "Session expired. Please login again.", ExitCode: 1},
			expected: true,
		},
		{
			name:     "authentication failed",
			result:   &Result{Output: "Authentication failed for this request", ExitCode: 1},
			expected: true,
		},
		{
			name:     "unauthorized",
			result:   &Result{Output: "401 Unauthorized", ExitCode: 1},
			expected: true,
		},
		{
			name:     "token expired",
			result:   &Result{Output: "Your token expired, please re-authenticate", ExitCode: 1},
			expected: true,
		},
		{
			name:     "invalid credentials",
			result:   &Result{Output: "invalid API credentials provided", ExitCode: 1},
			expected: true,
		},
		{
			name:     "API key invalid",
			result:   &Result{Output: "API key is invalid", ExitCode: 1},
			expected: true,
		},
		{
			name:     "API key expired",
			result:   &Result{Output: "API key expired", ExitCode: 1},
			expected: true,
		},
		{
			name:     "not authenticated",
			result:   &Result{Output: "Error: not authenticated", ExitCode: 1},
			expected: true,
		},
		{
			name:     "session expired",
			result:   &Result{Output: "Error: session expired", ExitCode: 1},
			expected: true,
		},
		{
			name:     "access denied",
			result:   &Result{Output: "access denied to this resource", ExitCode: 1},
			expected: true,
		},
		{
			name:     "normal compile error",
			result:   &Result{Output: "main.go:10: undefined: foo", ExitCode: 1},
			expected: false,
		},
		{
			name:     "normal runtime error",
			result:   &Result{Output: "panic: runtime error: index out of range", ExitCode: 2},
			expected: false,
		},
		{
			name:     "generic error message",
			result:   &Result{Output: "Error: something went wrong", ExitCode: 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAuthError(tt.result)
			if got != tt.expected {
				t.Errorf("IsAuthError() = %v, want %v (output: %q)", got, tt.expected, tt.result)
			}
		})
	}
}

func TestConcurrencyLimit(t *testing.T) {
	// Initialize with max 2
	globalManager = nil
	managerOnce = sync.Once{}
	Init(Config{
		Timeout: 30 * time.Second,
		Max:     2,
	})

	mgr := GetManager()

	// Check initial state
	if mgr.Available() != 2 {
		t.Errorf("Expected 2 available, got %d", mgr.Available())
	}

	t.Logf("Max: 2, Available: %d", mgr.Available())
}
