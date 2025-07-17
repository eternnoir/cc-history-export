package reader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTodoReader(t *testing.T) {
	// Create test todo file
	tmpDir := t.TempDir()
	todoFile := filepath.Join(tmpDir, "test-todos.json")
	
	todoContent := `[
		{
			"id": "1",
			"content": "First task",
			"status": "pending",
			"priority": "high"
		},
		{
			"id": "2",
			"content": "Second task",
			"status": "completed",
			"priority": "medium"
		},
		{
			"id": "3",
			"content": "Third task",
			"status": "in_progress",
			"priority": "low"
		}
	]`
	
	if err := os.WriteFile(todoFile, []byte(todoContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test reading todos
	reader := NewTodoReader(todoFile)
	todos, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	
	if len(todos) != 3 {
		t.Errorf("Expected 3 todos, got %d", len(todos))
	}
	
	// Verify first todo
	if todos[0].ID != "1" {
		t.Errorf("First todo ID = %v, want 1", todos[0].ID)
	}
	if todos[0].Content != "First task" {
		t.Errorf("First todo content = %v, want 'First task'", todos[0].Content)
	}
	if todos[0].Status != "pending" {
		t.Errorf("First todo status = %v, want pending", todos[0].Status)
	}
	if todos[0].Priority != "high" {
		t.Errorf("First todo priority = %v, want high", todos[0].Priority)
	}
	
	// Test ReadTodoList
	todoList, err := reader.ReadTodoList("session123", "agent456")
	if err != nil {
		t.Fatalf("ReadTodoList() error = %v", err)
	}
	
	if todoList.SessionID != "session123" {
		t.Errorf("TodoList SessionID = %v, want session123", todoList.SessionID)
	}
	if todoList.AgentID != "agent456" {
		t.Errorf("TodoList AgentID = %v, want agent456", todoList.AgentID)
	}
	if len(todoList.Todos) != 3 {
		t.Errorf("TodoList has %d todos, want 3", len(todoList.Todos))
	}
}

func TestTodoReaderEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.json")
	
	// Test empty array
	if err := os.WriteFile(emptyFile, []byte("[]"), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	
	reader := NewTodoReader(emptyFile)
	todos, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error for empty array = %v", err)
	}
	
	if len(todos) != 0 {
		t.Errorf("Expected 0 todos for empty array, got %d", len(todos))
	}
}

func TestTodoReaderErrors(t *testing.T) {
	// Test non-existent file
	reader := NewTodoReader("/non/existent/file.json")
	_, err := reader.Read()
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	
	// Test malformed JSON
	tmpDir := t.TempDir()
	malformedFile := filepath.Join(tmpDir, "malformed.json")
	
	if err := os.WriteFile(malformedFile, []byte("not json"), 0644); err != nil {
		t.Fatalf("Failed to create malformed file: %v", err)
	}
	
	reader = NewTodoReader(malformedFile)
	_, err = reader.Read()
	if err == nil {
		t.Error("Expected error for malformed JSON")
	}
}