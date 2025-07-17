package models

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
)

// Message represents a single message in a conversation
type Message struct {
	UUID       string          `json:"uuid"`
	ParentUUID *string         `json:"parentUuid"`
	SessionID  string          `json:"sessionId"`
	Type       MessageType     `json:"type"`
	UserType   string          `json:"userType,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
	RequestID  string          `json:"requestId,omitempty"`
	Version    string          `json:"version,omitempty"`
	CWD        string          `json:"cwd,omitempty"`
	Message    json.RawMessage `json:"message"`
	
	// Parsed message content
	Content interface{} `json:"-"`
}

// UserMessage represents a user's message
type UserMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AssistantMessage represents an assistant's response
type AssistantMessage struct {
	ID      string           `json:"id"`
	Type    string           `json:"type"`
	Role    string           `json:"role"`
	Model   string           `json:"model"`
	Content []MessageContent `json:"content"`
	Usage   *Usage           `json:"usage,omitempty"`
}

// MessageContent represents content within an assistant message
type MessageContent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	Thinking string          `json:"thinking,omitempty"`
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name,omitempty"`
	Input    json.RawMessage `json:"input,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens               int    `json:"input_tokens"`
	OutputTokens              int    `json:"output_tokens"`
	CacheCreationInputTokens  int    `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens      int    `json:"cache_read_input_tokens,omitempty"`
	ServiceTier              string `json:"service_tier,omitempty"`
}

// ToolResult represents the result of a tool use
type ToolResult struct {
	ToolUseID string          `json:"tool_use_id"`
	Type      string          `json:"type"`
	Content   json.RawMessage `json:"content"`
}

// ParseContent parses the raw message content based on message type
func (m *Message) ParseContent() error {
	switch m.Type {
	case MessageTypeUser:
		if m.UserType == "external" {
			var msg struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			}
			if err := json.Unmarshal(m.Message, &msg); err != nil {
				return err
			}
			
			// Content can be string or array of tool results
			var content string
			if err := json.Unmarshal(msg.Content, &content); err == nil {
				m.Content = &UserMessage{
					Role:    msg.Role,
					Content: content,
				}
			} else {
				// Try parsing as tool result array
				var toolResults []ToolResult
				if err := json.Unmarshal(msg.Content, &toolResults); err == nil {
					m.Content = toolResults
				}
			}
		}
	case MessageTypeAssistant:
		var msg AssistantMessage
		if err := json.Unmarshal(m.Message, &msg); err != nil {
			return err
		}
		m.Content = &msg
	}
	return nil
}