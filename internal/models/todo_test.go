package models

import (
	"testing"
)

func TestTodoListOperations(t *testing.T) {
	todoList := &TodoList{
		SessionID: "test-session",
		AgentID:   "test-agent",
		Todos: []*Todo{
			{
				ID:       "1",
				Content:  "High priority pending task",
				Status:   TodoStatusPending,
				Priority: TodoPriorityHigh,
			},
			{
				ID:       "2",
				Content:  "Medium priority in progress task",
				Status:   TodoStatusInProgress,
				Priority: TodoPriorityMedium,
			},
			{
				ID:       "3",
				Content:  "High priority completed task",
				Status:   TodoStatusCompleted,
				Priority: TodoPriorityHigh,
			},
			{
				ID:       "4",
				Content:  "Low priority completed task",
				Status:   TodoStatusCompleted,
				Priority: TodoPriorityLow,
			},
		},
	}

	// Test filtering by status
	pending := todoList.GetTodosByStatus(TodoStatusPending)
	if len(pending) != 1 {
		t.Errorf("GetTodosByStatus(pending) returned %d items, want 1", len(pending))
	}

	completed := todoList.GetTodosByStatus(TodoStatusCompleted)
	if len(completed) != 2 {
		t.Errorf("GetTodosByStatus(completed) returned %d items, want 2", len(completed))
	}

	// Test filtering by priority
	highPriority := todoList.GetTodosByPriority(TodoPriorityHigh)
	if len(highPriority) != 2 {
		t.Errorf("GetTodosByPriority(high) returned %d items, want 2", len(highPriority))
	}

	// Test completion rate
	rate := todoList.GetCompletionRate()
	expectedRate := 50.0 // 2 out of 4 completed
	if rate != expectedRate {
		t.Errorf("GetCompletionRate() = %v, want %v", rate, expectedRate)
	}
}

func TestEmptyTodoList(t *testing.T) {
	todoList := &TodoList{
		SessionID: "empty-session",
		AgentID:   "empty-agent",
		Todos:     []*Todo{},
	}

	if rate := todoList.GetCompletionRate(); rate != 0 {
		t.Errorf("GetCompletionRate() for empty list = %v, want 0", rate)
	}

	if todos := todoList.GetTodosByStatus(TodoStatusPending); len(todos) != 0 {
		t.Errorf("GetTodosByStatus() for empty list returned %d items, want 0", len(todos))
	}
}

func TestAllCompletedTodoList(t *testing.T) {
	todoList := &TodoList{
		SessionID: "completed-session",
		AgentID:   "completed-agent",
		Todos: []*Todo{
			{ID: "1", Content: "Task 1", Status: TodoStatusCompleted, Priority: TodoPriorityHigh},
			{ID: "2", Content: "Task 2", Status: TodoStatusCompleted, Priority: TodoPriorityMedium},
			{ID: "3", Content: "Task 3", Status: TodoStatusCompleted, Priority: TodoPriorityLow},
		},
	}

	if rate := todoList.GetCompletionRate(); rate != 100.0 {
		t.Errorf("GetCompletionRate() for all completed = %v, want 100.0", rate)
	}
}