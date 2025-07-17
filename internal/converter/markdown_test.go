package converter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/eternnoir/cc-history-export/internal/models"
)

func TestMarkdownConverter(t *testing.T) {
	// Create test session
	session := &models.Session{
		ID:        "test-session",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
	}

	// Add user message
	userMsg := &models.Message{
		UUID:      "msg1",
		Type:      models.MessageTypeUser,
		UserType:  "external",
		Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		CWD:       "/test/dir",
		Message:   json.RawMessage(`{"role":"user","content":"Hello, can you help me?"}`),
	}
	userMsg.ParseContent()
	session.AddMessage(userMsg)

	// Add assistant message
	assistantMsg := &models.Message{
		UUID:      "msg2",
		Type:      models.MessageTypeAssistant,
		Timestamp: time.Date(2024, 1, 1, 10, 0, 5, 0, time.UTC),
		Message: json.RawMessage(`{
			"id": "asst1",
			"type": "message",
			"role": "assistant",
			"model": "claude-3",
			"content": [
				{"type": "text", "text": "Of course! I'd be happy to help."},
				{"type": "thinking", "thinking": "The user is asking for help."},
				{
					"type": "tool_use",
					"id": "tool1",
					"name": "search",
					"input": {"query": "help topics"}
				}
			],
			"usage": {"input_tokens": 10, "output_tokens": 20}
		}`),
	}
	assistantMsg.ParseContent()
	session.AddMessage(assistantMsg)

	// Test basic conversion
	converter := NewMarkdownConverter(&MarkdownOptions{
		ShowTimestamps: true,
		ShowTokenUsage: true,
		ShowThinking:   false,
		ShowUUIDs:      false,
	})

	markdown := converter.ConvertSession(session)

	// Verify content
	if !strings.Contains(markdown, "# Session: test-session") {
		t.Error("Missing session header")
	}

	if !strings.Contains(markdown, "**Started:** 2024-01-01") {
		t.Error("Missing start time")
	}

	if !strings.Contains(markdown, "**Duration:** 30m0s") {
		t.Error("Missing duration")
	}

	if !strings.Contains(markdown, "**Messages:** 2") {
		t.Error("Missing message count")
	}

	if !strings.Contains(markdown, "👤 User") {
		t.Error("Missing user header")
	}

	if !strings.Contains(markdown, "Hello, can you help me?") {
		t.Error("Missing user content")
	}

	if !strings.Contains(markdown, "🤖 Assistant") {
		t.Error("Missing assistant header")
	}

	if !strings.Contains(markdown, "*Model: claude-3*") {
		t.Error("Missing model info")
	}

	if !strings.Contains(markdown, "Of course! I'd be happy to help.") {
		t.Error("Missing assistant text")
	}

	if !strings.Contains(markdown, "**🔧 Tool Use:** `search`") {
		t.Error("Missing tool use")
	}

	// Thinking should not be included
	if strings.Contains(markdown, "The user is asking for help") {
		t.Error("Thinking should not be included when ShowThinking is false")
	}

	// Test with thinking enabled
	converter.options.ShowThinking = true
	markdown = converter.ConvertSession(session)

	if !strings.Contains(markdown, "💭 Thinking") {
		t.Error("Missing thinking section when ShowThinking is true")
	}
}

func TestMarkdownConverterProject(t *testing.T) {
	project := models.NewProject("-Users-test-project")
	
	// Add sessions
	session1 := &models.Session{
		ID:        "session1",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
	}
	
	msg := &models.Message{
		UUID:      "msg1",
		Type:      models.MessageTypeUser,
		Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Message:   json.RawMessage(`{"role":"user","content":"Test message"}`),
	}
	msg.ParseContent()
	session1.AddMessage(msg)
	
	project.AddSession(session1)
	
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
	
	converter := NewMarkdownConverter(nil)
	markdown := converter.ConvertProject(project)
	
	// Verify project content
	if !strings.Contains(markdown, "# Project: project") {
		t.Error("Missing project header")
	}
	
	if !strings.Contains(markdown, "**Path:** `/Users/test/project`") {
		t.Error("Missing project path")
	}
	
	if !strings.Contains(markdown, "## Todo Lists (1)") {
		t.Error("Missing todo lists section")
	}
	
	if !strings.Contains(markdown, "- [ ] Task 1 (high)") {
		t.Error("Missing pending todo")
	}
	
	if !strings.Contains(markdown, "- [x] Task 2 (medium)") {
		t.Error("Missing completed todo")
	}
	
	if !strings.Contains(markdown, "*Completion: 50%*") {
		t.Error("Missing completion rate")
	}
}

func TestMarkdownConverterOptions(t *testing.T) {
	msg := &models.Message{
		UUID:      "test-uuid",
		Type:      models.MessageTypeUser,
		Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Message:   json.RawMessage(`{"role":"user","content":"Test"}`),
	}
	msg.ParseContent()
	
	// Test with UUIDs enabled
	converter := NewMarkdownConverter(&MarkdownOptions{
		ShowUUIDs:      true,
		ShowTimestamps: false,
	})
	
	markdown := converter.ConvertMessage(msg)
	
	if !strings.Contains(markdown, "UUID: test-uuid") {
		t.Error("Missing UUID when ShowUUIDs is true")
	}
	
	// Test with timestamps disabled
	if strings.Contains(markdown, "2024-01-01") {
		t.Error("Timestamp should not be shown when ShowTimestamps is false")
	}
}

func TestMarkdownConverterToolResults(t *testing.T) {
	msg := &models.Message{
		UUID:     "msg1",
		Type:     models.MessageTypeUser,
		UserType: "external",
		Message: json.RawMessage(`{
			"role": "user",
			"content": [
				{
					"tool_use_id": "tool_123",
					"type": "tool_result",
					"content": {"result": "success", "data": "test data"}
				}
			]
		}`),
	}
	msg.ParseContent()
	
	converter := NewMarkdownConverter(nil)
	markdown := converter.ConvertMessage(msg)
	
	if !strings.Contains(markdown, "**Tool Results:**") {
		t.Error("Missing tool results header")
	}
	
	if !strings.Contains(markdown, "Tool: `tool_123`") {
		t.Error("Missing tool ID")
	}
}