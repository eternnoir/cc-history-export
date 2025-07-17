package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSessionOperations(t *testing.T) {
	session := &Session{
		ID: "test-session",
	}

	// Test adding messages
	msg1 := &Message{
		UUID:      "msg1",
		Type:      MessageTypeUser,
		Timestamp: time.Now().Add(-10 * time.Minute),
	}
	
	msg2 := &Message{
		UUID:      "msg2",
		Type:      MessageTypeAssistant,
		Timestamp: time.Now().Add(-5 * time.Minute),
		Message: json.RawMessage(`{
			"id": "test",
			"type": "message",
			"role": "assistant",
			"model": "claude-3",
			"content": [{"type": "text", "text": "Response"}],
			"usage": {"input_tokens": 10, "output_tokens": 20}
		}`),
	}
	msg2.ParseContent()

	msg3 := &Message{
		UUID:      "msg3",
		Type:      MessageTypeUser,
		Timestamp: time.Now(),
	}

	session.AddMessage(msg1)
	session.AddMessage(msg2)
	session.AddMessage(msg3)

	// Test message counts
	if count := session.GetMessageCount(); count != 3 {
		t.Errorf("GetMessageCount() = %v, want 3", count)
	}

	if count := session.GetUserMessageCount(); count != 2 {
		t.Errorf("GetUserMessageCount() = %v, want 2", count)
	}

	if count := session.GetAssistantMessageCount(); count != 1 {
		t.Errorf("GetAssistantMessageCount() = %v, want 1", count)
	}

	// Test timestamps
	if session.StartTime != msg1.Timestamp {
		t.Error("StartTime not set correctly")
	}

	if session.EndTime != msg3.Timestamp {
		t.Error("EndTime not set correctly")
	}

	// Test duration
	duration := session.GetDuration()
	expectedDuration := 10 * time.Minute
	if duration < expectedDuration-time.Second || duration > expectedDuration+time.Second {
		t.Errorf("GetDuration() = %v, want approximately %v", duration, expectedDuration)
	}

	// Test token usage
	inputTokens, outputTokens := session.GetTokenUsage()
	if inputTokens != 10 {
		t.Errorf("Input tokens = %v, want 10", inputTokens)
	}
	if outputTokens != 20 {
		t.Errorf("Output tokens = %v, want 20", outputTokens)
	}
}

func TestEmptySession(t *testing.T) {
	session := &Session{
		ID: "empty-session",
	}

	if count := session.GetMessageCount(); count != 0 {
		t.Errorf("GetMessageCount() = %v, want 0", count)
	}

	if duration := session.GetDuration(); duration != 0 {
		t.Errorf("GetDuration() = %v, want 0", duration)
	}

	inputTokens, outputTokens := session.GetTokenUsage()
	if inputTokens != 0 || outputTokens != 0 {
		t.Errorf("GetTokenUsage() = (%v, %v), want (0, 0)", inputTokens, outputTokens)
	}
}