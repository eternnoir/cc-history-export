package models

import (
	"time"
)

// Session represents a conversation session
type Session struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"project_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	Messages  []*Message `json:"messages"`
}

// AddMessage adds a message to the session and updates timestamps
func (s *Session) AddMessage(msg *Message) {
	s.Messages = append(s.Messages, msg)
	
	// Update session timestamps
	if s.StartTime.IsZero() || msg.Timestamp.Before(s.StartTime) {
		s.StartTime = msg.Timestamp
	}
	if msg.Timestamp.After(s.EndTime) {
		s.EndTime = msg.Timestamp
	}
}

// GetMessageCount returns the total number of messages
func (s *Session) GetMessageCount() int {
	return len(s.Messages)
}

// GetUserMessageCount returns the number of user messages
func (s *Session) GetUserMessageCount() int {
	count := 0
	for _, msg := range s.Messages {
		if msg.Type == MessageTypeUser {
			count++
		}
	}
	return count
}

// GetAssistantMessageCount returns the number of assistant messages
func (s *Session) GetAssistantMessageCount() int {
	count := 0
	for _, msg := range s.Messages {
		if msg.Type == MessageTypeAssistant {
			count++
		}
	}
	return count
}

// GetDuration returns the duration of the session
func (s *Session) GetDuration() time.Duration {
	if s.StartTime.IsZero() || s.EndTime.IsZero() {
		return 0
	}
	return s.EndTime.Sub(s.StartTime)
}

// GetTokenUsage calculates total token usage for the session
func (s *Session) GetTokenUsage() (input int, output int) {
	for _, msg := range s.Messages {
		if msg.Type == MessageTypeAssistant && msg.Content != nil {
			if assistantMsg, ok := msg.Content.(*AssistantMessage); ok && assistantMsg.Usage != nil {
				input += assistantMsg.Usage.InputTokens + assistantMsg.Usage.CacheReadInputTokens
				output += assistantMsg.Usage.OutputTokens
			}
		}
	}
	return
}