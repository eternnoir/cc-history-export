package models

import (
	"testing"
	"time"
)

func TestNewProject(t *testing.T) {
	tests := []struct {
		encodedPath string
		wantPath    string
		wantName    string
	}{
		{
			encodedPath: "-Users-test_user-src-myproject",
			wantPath:    "/Users/test_user/src/myproject",
			wantName:    "myproject",
		},
		{
			encodedPath: "-Users-test_user-Downloads-OUTPUT",
			wantPath:    "/Users/test_user/Downloads/OUTPUT",
			wantName:    "OUTPUT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.encodedPath, func(t *testing.T) {
			project := NewProject(tt.encodedPath)
			
			if project.Path != tt.wantPath {
				t.Errorf("Path = %v, want %v", project.Path, tt.wantPath)
			}
			
			if project.EncodedPath != tt.encodedPath {
				t.Errorf("EncodedPath = %v, want %v", project.EncodedPath, tt.encodedPath)
			}
			
			if name := project.GetProjectName(); name != tt.wantName {
				t.Errorf("GetProjectName() = %v, want %v", name, tt.wantName)
			}
		})
	}
}

func TestProjectOperations(t *testing.T) {
	project := NewProject("-Users-test-project")
	
	// Create test sessions
	session1 := &Session{
		ID:        "session1",
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now().Add(-90 * time.Minute),
		Messages: []*Message{
			{UUID: "msg1", Type: MessageTypeUser},
			{UUID: "msg2", Type: MessageTypeAssistant},
		},
	}
	
	session2 := &Session{
		ID:        "session2",
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(-30 * time.Minute),
		Messages: []*Message{
			{UUID: "msg3", Type: MessageTypeUser},
			{UUID: "msg4", Type: MessageTypeAssistant},
			{UUID: "msg5", Type: MessageTypeUser},
		},
	}
	
	// Add sessions
	project.AddSession(session1)
	project.AddSession(session2)
	
	// Test session count
	if count := project.GetSessionCount(); count != 2 {
		t.Errorf("GetSessionCount() = %v, want 2", count)
	}
	
	// Test total messages
	if total := project.GetTotalMessages(); total != 5 {
		t.Errorf("GetTotalMessages() = %v, want 5", total)
	}
	
	// Test time range
	start, end := project.GetTimeRange()
	if !start.Equal(session1.StartTime) {
		t.Errorf("Start time = %v, want %v", start, session1.StartTime)
	}
	if !end.Equal(session2.EndTime) {
		t.Errorf("End time = %v, want %v", end, session2.EndTime)
	}
	
	// Test project ID assignment
	if session1.ProjectID != project.ID {
		t.Errorf("Session1 ProjectID = %v, want %v", session1.ProjectID, project.ID)
	}
}

func TestProjectWithTodos(t *testing.T) {
	project := NewProject("-Users-test-todos")
	
	todoList1 := &TodoList{
		SessionID: "session1",
		AgentID:   "agent1",
		Todos: []*Todo{
			{ID: "1", Content: "Task 1", Status: TodoStatusPending, Priority: TodoPriorityHigh},
		},
	}
	
	todoList2 := &TodoList{
		SessionID: "session2",
		AgentID:   "agent2",
		Todos: []*Todo{
			{ID: "2", Content: "Task 2", Status: TodoStatusCompleted, Priority: TodoPriorityMedium},
		},
	}
	
	project.AddTodoList(todoList1)
	project.AddTodoList(todoList2)
	
	if len(project.TodoLists) != 2 {
		t.Errorf("TodoLists length = %v, want 2", len(project.TodoLists))
	}
}