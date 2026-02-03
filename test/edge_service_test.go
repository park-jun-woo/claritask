package test

import (
	"fmt"
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

// Helper to convert int64 to string for edge functions
func taskIDStr(id int64) string {
	return fmt.Sprintf("%d", id)
}

func TestAddTaskEdge(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})

	// Add edge: task2 depends on task1
	err := service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))
	if err != nil {
		t.Fatalf("AddTaskEdge failed: %v", err)
	}
}

func TestRemoveTaskEdge(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})

	service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))

	err := service.RemoveTaskEdge(database, taskIDStr(task2), taskIDStr(task1))
	if err != nil {
		t.Fatalf("RemoveTaskEdge failed: %v", err)
	}
}

func TestGetTaskEdges(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})
	task3, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 3"})

	service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))
	service.AddTaskEdge(database, taskIDStr(task3), taskIDStr(task2))

	edges, err := service.GetTaskEdges(database)
	if err != nil {
		t.Fatalf("GetTaskEdges failed: %v", err)
	}

	if len(edges) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(edges))
	}
}

func TestGetTaskDependencies(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})

	service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))

	deps, err := service.GetTaskDependencies(database, taskIDStr(task2))
	if err != nil {
		t.Fatalf("GetTaskDependencies failed: %v", err)
	}

	if len(deps) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(deps))
	}

	if deps[0].ID != taskIDStr(task1) {
		t.Errorf("Expected dependency on task1")
	}
}

func TestGetTaskDependents(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})

	service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))

	dependents, err := service.GetTaskDependents(database, taskIDStr(task1))
	if err != nil {
		t.Fatalf("GetTaskDependents failed: %v", err)
	}

	if len(dependents) != 1 {
		t.Errorf("Expected 1 dependent, got %d", len(dependents))
	}
}

func TestCheckTaskCycle(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})
	task3, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 3"})

	// Create chain: task3 -> task2 -> task1
	service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))
	service.AddTaskEdge(database, taskIDStr(task3), taskIDStr(task2))

	// Check if adding task1 -> task3 would create cycle
	hasCycle, _, err := service.CheckTaskCycle(database, taskIDStr(task1), taskIDStr(task3))
	if err != nil {
		t.Fatalf("CheckTaskCycle failed: %v", err)
	}

	if !hasCycle {
		t.Error("Expected cycle detection")
	}
}

func TestTopologicalSortTasks(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})
	task3, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 3"})

	// Create chain: task3 -> task2 -> task1
	service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))
	service.AddTaskEdge(database, taskIDStr(task3), taskIDStr(task2))

	sorted, err := service.TopologicalSortTasks(database, phaseID)
	if err != nil {
		t.Fatalf("TopologicalSortTasks failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(sorted))
	}

	// Task1 should come before Task2, Task2 before Task3
	task1Pos, task2Pos, task3Pos := -1, -1, -1
	for i, task := range sorted {
		switch task.ID {
		case taskIDStr(task1):
			task1Pos = i
		case taskIDStr(task2):
			task2Pos = i
		case taskIDStr(task3):
			task3Pos = i
		}
	}

	if task1Pos > task2Pos || task2Pos > task3Pos {
		t.Error("Tasks not in correct topological order")
	}
}

func TestGetExecutableTasks(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})

	// Task2 depends on Task1
	service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))

	executable, err := service.GetExecutableTasks(database)
	if err != nil {
		t.Fatalf("GetExecutableTasks failed: %v", err)
	}

	// Only Task1 should be executable (Task2 is blocked)
	if len(executable) != 1 {
		t.Errorf("Expected 1 executable task, got %d", len(executable))
	}

	if executable[0].ID != taskIDStr(task1) {
		t.Error("Expected Task1 to be executable")
	}
}

func TestIsTaskExecutable(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})

	// Task2 depends on Task1
	service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))

	// Task1 should be executable
	executable, _, err := service.IsTaskExecutable(database, taskIDStr(task1))
	if err != nil {
		t.Fatalf("IsTaskExecutable failed: %v", err)
	}
	if !executable {
		t.Error("Task1 should be executable")
	}

	// Task2 should not be executable
	executable, blocking, err := service.IsTaskExecutable(database, taskIDStr(task2))
	if err != nil {
		t.Fatalf("IsTaskExecutable failed: %v", err)
	}
	if executable {
		t.Error("Task2 should not be executable")
	}
	if len(blocking) != 1 {
		t.Errorf("Expected 1 blocking task, got %d", len(blocking))
	}
}

func TestListAllEdges(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})

	// Create feature edges
	feature1, _ := service.CreateFeature(database, "test-project", "Feature1", "")
	feature2, _ := service.CreateFeature(database, "test-project", "Feature2", "")
	service.AddFeatureEdge(database, feature2, feature1)

	// Create task edges
	task1, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{PhaseID: phaseID, Title: "Task 2"})
	service.AddTaskEdge(database, taskIDStr(task2), taskIDStr(task1))

	result, err := service.ListAllEdges(database)
	if err != nil {
		t.Fatalf("ListAllEdges failed: %v", err)
	}

	if result.TotalFeature != 1 {
		t.Errorf("Expected 1 feature edge, got %d", result.TotalFeature)
	}
	if result.TotalTask != 1 {
		t.Errorf("Expected 1 task edge, got %d", result.TotalTask)
	}
}
