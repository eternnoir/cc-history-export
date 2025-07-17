package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessageParsing(t *testing.T) {
	tests := []struct {
		name    string
		message *Message
		wantErr bool
	}{
		{
			name: "parse user message",
			message: &Message{
				UUID:      "test-uuid",
				SessionID: "test-session",
				Type:      MessageTypeUser,
				UserType:  "external",
				Timestamp: time.Now(),
				Message:   json.RawMessage(`{"role":"user","content":"Hello, world!"}`),
			},
			wantErr: false,
		},
		{
			name: "parse assistant message",
			message: &Message{
				UUID:      "test-uuid-2",
				SessionID: "test-session",
				Type:      MessageTypeAssistant,
				Timestamp: time.Now(),
				Message: json.RawMessage(`{
					"id": "msg_123",
					"type": "message",
					"role": "assistant",
					"model": "claude-3",
					"content": [
						{"type": "text", "text": "Hello! How can I help you?"}
					],
					"usage": {
						"input_tokens": 10,
						"output_tokens": 20
					}
				}`),
			},
			wantErr: false,
		},
		{
			name: "parse tool result",
			message: &Message{
				UUID:      "test-uuid-3",
				SessionID: "test-session",
				Type:      MessageTypeUser,
				UserType:  "external",
				Timestamp: time.Now(),
				Message: json.RawMessage(`{
					"role": "user",
					"content": [
						{
							"tool_use_id": "tool_123",
							"type": "tool_result",
							"content": {"result": "success"}
						}
					]
				}`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.ParseContent()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContent() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.message.Content == nil {
				t.Error("ParseContent() did not set Content field")
			}
		})
	}
}

func TestUserMessageParsing(t *testing.T) {
	msg := &Message{
		Type:     MessageTypeUser,
		UserType: "external",
		Message:  json.RawMessage(`{"role":"user","content":"Test message"}`),
	}

	err := msg.ParseContent()
	if err != nil {
		t.Fatalf("ParseContent() error = %v", err)
	}

	userMsg, ok := msg.Content.(*UserMessage)
	if !ok {
		t.Fatal("Content is not *UserMessage")
	}

	if userMsg.Role != "user" {
		t.Errorf("Role = %v, want user", userMsg.Role)
	}

	if userMsg.Content != "Test message" {
		t.Errorf("Content = %v, want Test message", userMsg.Content)
	}
}

func TestAssistantMessageParsing(t *testing.T) {
	msg := &Message{
		Type: MessageTypeAssistant,
		Message: json.RawMessage(`{
			"id": "msg_test",
			"type": "message",
			"role": "assistant",
			"model": "claude-3",
			"content": [
				{"type": "text", "text": "Response text"},
				{"type": "thinking", "thinking": "Internal thoughts"}
			],
			"usage": {
				"input_tokens": 100,
				"output_tokens": 50,
				"cache_read_input_tokens": 25
			}
		}`),
	}

	err := msg.ParseContent()
	if err != nil {
		t.Fatalf("ParseContent() error = %v", err)
	}

	assistantMsg, ok := msg.Content.(*AssistantMessage)
	if !ok {
		t.Fatal("Content is not *AssistantMessage")
	}

	if assistantMsg.Model != "claude-3" {
		t.Errorf("Model = %v, want claude-3", assistantMsg.Model)
	}

	if len(assistantMsg.Content) != 2 {
		t.Errorf("Content length = %v, want 2", len(assistantMsg.Content))
	}

	if assistantMsg.Usage.InputTokens != 100 {
		t.Errorf("InputTokens = %v, want 100", assistantMsg.Usage.InputTokens)
	}
}