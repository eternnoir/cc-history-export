package converter

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/eternnoir/cc-history-export/internal/models"
)

func TestJSONConverter(t *testing.T) {
	// Create test session
	session := &models.Session{
		ID:        "test-session",
		ProjectID: "test-project",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
	}

	// Add messages
	userMsg := &models.Message{
		UUID:      "msg1",
		Type:      models.MessageTypeUser,
		UserType:  "external",
		Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		CWD:       "/test/dir",
		Message:   json.RawMessage(`{"role":"user","content":"Hello"}`),
	}
	userMsg.ParseContent()
	session.AddMessage(userMsg)

	assistantMsg := &models.Message{
		UUID:      "msg2",
		Type:      models.MessageTypeAssistant,
		Timestamp: time.Date(2024, 1, 1, 10, 0, 5, 0, time.UTC),
		Message: json.RawMessage(`{
			"id": "asst1",
			"type": "message",
			"role": "assistant",
			"model": "claude-3",
			"content": [{"type": "text", "text": "Hi there!"}],
			"usage": {"input_tokens": 10, "output_tokens": 20}
		}`),
	}
	assistantMsg.ParseContent()
	session.AddMessage(assistantMsg)

	// Test basic conversion
	converter := NewJSONConverter(&JSONOptions{
		PrettyPrint:        true,
		IncludeRawMessages: false,
		OmitEmpty:          true,
	})

	data, err := converter.ConvertSession(session)
	if err != nil {
		t.Fatalf("ConvertSession() error = %v", err)
	}

	// Parse the result
	var result JSONSession
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Verify fields
	if result.ID != "test-session" {
		t.Errorf("Session ID = %v, want test-session", result.ID)
	}

	if result.MessageCount != 2 {
		t.Errorf("MessageCount = %v, want 2", result.MessageCount)
	}

	if result.UserMessages != 1 {
		t.Errorf("UserMessages = %v, want 1", result.UserMessages)
	}

	if result.AssistantMessages != 1 {
		t.Errorf("AssistantMessages = %v, want 1", result.AssistantMessages)
	}

	if result.TokenUsage == nil {
		t.Error("TokenUsage should not be nil")
	} else {
		if result.TokenUsage.Input != 10 {
			t.Errorf("Input tokens = %v, want 10", result.TokenUsage.Input)
		}
		if result.TokenUsage.Output != 20 {
			t.Errorf("Output tokens = %v, want 20", result.TokenUsage.Output)
		}
		if result.TokenUsage.Total != 30 {
			t.Errorf("Total tokens = %v, want 30", result.TokenUsage.Total)
		}
	}

	if len(result.Messages) != 2 {
		t.Errorf("Messages count = %v, want 2", len(result.Messages))
	}
}

func TestJSONConverterProject(t *testing.T) {
	project := models.NewProject("-Users-test-project")
	
	// Add session
	session := &models.Session{
		ID:        "session1",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
	}
	
	msg := &models.Message{
		UUID:      "msg1",
		Type:      models.MessageTypeUser,
		Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Message:   json.RawMessage(`{"role":"user","content":"Test"}`),
	}
	msg.ParseContent()
	session.AddMessage(msg)
	
	project.AddSession(session)
	
	// Add todo list
	todoList := &models.TodoList{
		SessionID: "session1",
		AgentID:   "agent1",
		Todos: []*models.Todo{
			{ID: "1", Content: "Task 1", Status: models.TodoStatusPending, Priority: models.TodoPriorityHigh},
			{ID: "2", Content: "Task 2", Status: models.TodoStatusCompleted, Priority: models.TodoPriorityMedium},
		},
	}
	project.AddTodoList(todoList)
	
	converter := NewJSONConverter(nil)
	data, err := converter.ConvertProject(project)
	if err != nil {
		t.Fatalf("ConvertProject() error = %v", err)
	}
	
	// Parse the result
	var result JSONProject
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	
	// Verify project fields
	if result.Name != "project" {
		t.Errorf("Project name = %v, want project", result.Name)
	}
	
	if result.Path != "/Users/test/project" {
		t.Errorf("Project path = %v, want /Users/test/project", result.Path)
	}
	
	if result.SessionCount != 1 {
		t.Errorf("SessionCount = %v, want 1", result.SessionCount)
	}
	
	if len(result.TodoLists) != 1 {
		t.Errorf("TodoLists count = %v, want 1", len(result.TodoLists))
	}
	
	if result.TodoLists[0].CompletionRate != 50.0 {
		t.Errorf("CompletionRate = %v, want 50.0", result.TodoLists[0].CompletionRate)
	}
}

func TestJSONConverterMultipleProjects(t *testing.T) {
	projects := []*models.Project{
		models.NewProject("-Users-project1"),
		models.NewProject("-Users-project2"),
	}
	
	converter := NewJSONConverter(nil)
	data, err := converter.ConvertProjects(projects)
	if err != nil {
		t.Fatalf("ConvertProjects() error = %v", err)
	}
	
	// Parse the result
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	
	if result["project_count"].(float64) != 2 {
		t.Errorf("project_count = %v, want 2", result["project_count"])
	}
	
	projectList, ok := result["projects"].([]interface{})
	if !ok {
		t.Fatal("projects field is not an array")
	}
	
	if len(projectList) != 2 {
		t.Errorf("projects array length = %v, want 2", len(projectList))
	}
}

func TestJSONConverterRawMessages(t *testing.T) {
	session := &models.Session{
		ID: "test-session",
	}
	
	msg := &models.Message{
		UUID:      "msg1",
		Type:      models.MessageTypeUser,
		Timestamp: time.Now(),
		Message:   json.RawMessage(`{"role":"user","content":"Test"}`),
	}
	msg.ParseContent()
	session.AddMessage(msg)
	
	// Test with raw messages included
	converter := NewJSONConverter(&JSONOptions{
		PrettyPrint:        false,
		IncludeRawMessages: true,
	})
	
	data, err := converter.ConvertSession(session)
	if err != nil {
		t.Fatalf("ConvertSession() error = %v", err)
	}
	
	var result JSONSession
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	
	if result.Messages[0].RawMessage == nil {
		t.Error("RawMessage should be included when IncludeRawMessages is true")
	}
}

func TestJSONConverterCompact(t *testing.T) {
	session := &models.Session{
		ID:        "test-session",
		ProjectID: "test-project",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
	}
	
	// Add a message with token usage
	msg := &models.Message{
		UUID:      "msg1",
		Type:      models.MessageTypeAssistant,
		Timestamp: time.Now(),
		Message: json.RawMessage(`{
			"id": "asst1",
			"type": "message",
			"role": "assistant",
			"model": "claude-3",
			"content": [{"type": "text", "text": "Test"}],
			"usage": {"input_tokens": 5, "output_tokens": 10}
		}`),
	}
	msg.ParseContent()
	session.AddMessage(msg)
	
	converter := NewJSONConverter(nil)
	data, err := converter.ConvertSessionToCompactJSON(session)
	if err != nil {
		t.Fatalf("ConvertSessionToCompactJSON() error = %v", err)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal compact JSON: %v", err)
	}
	
	if result["id"] != "test-session" {
		t.Errorf("id = %v, want test-session", result["id"])
	}
	
	if result["messages"].(float64) != 1 {
		t.Errorf("messages = %v, want 1", result["messages"])
	}
	
	tokens, ok := result["tokens"].(map[string]interface{})
	if !ok {
		t.Fatal("tokens field is not a map")
	}
	
	if tokens["in"].(float64) != 5 {
		t.Errorf("input tokens = %v, want 5", tokens["in"])
	}
}

func TestJSONConverterValidation(t *testing.T) {
	converter := NewJSONConverter(nil)
	
	// Test valid data
	validData := map[string]interface{}{
		"test": "data",
		"number": 123,
		"nested": map[string]interface{}{
			"field": "value",
		},
	}
	
	if err := converter.ValidateJSON(validData); err != nil {
		t.Errorf("ValidateJSON() error for valid data = %v", err)
	}
	
	// Test invalid data (circular reference)
	type CircularStruct struct {
		Name string
		Self *CircularStruct
	}
	
	circular := &CircularStruct{Name: "test"}
	circular.Self = circular
	
	if err := converter.ValidateJSON(circular); err == nil {
		t.Error("ValidateJSON() should error for circular reference")
	}
}