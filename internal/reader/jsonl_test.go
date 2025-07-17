package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eternnoir/cc-history-export/internal/models"
)

func TestJSONLReader(t *testing.T) {
	// Create test JSONL content
	testContent := `{"uuid":"msg1","parentUuid":null,"sessionId":"session1","type":"user","userType":"external","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":"Hello"}}
{"uuid":"msg2","parentUuid":"msg1","sessionId":"session1","type":"assistant","timestamp":"2024-01-01T10:00:05Z","message":{"id":"asst1","type":"message","role":"assistant","model":"claude-3","content":[{"type":"text","text":"Hi there!"}],"usage":{"input_tokens":5,"output_tokens":10}}}
{"uuid":"msg3","parentUuid":"msg2","sessionId":"session1","type":"user","userType":"external","timestamp":"2024-01-01T10:00:10Z","message":{"role":"user","content":"How are you?"}}
`

	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test ReadSession
	reader := NewJSONLReader(testFile)
	session, err := reader.ReadSession()
	if err != nil {
		t.Fatalf("ReadSession() error = %v", err)
	}

	// Verify session
	if session.ID != "session1" {
		t.Errorf("Session ID = %v, want session1", session.ID)
	}

	if len(session.Messages) != 3 {
		t.Errorf("Message count = %v, want 3", len(session.Messages))
	}

	// Verify first message
	if session.Messages[0].UUID != "msg1" {
		t.Errorf("First message UUID = %v, want msg1", session.Messages[0].UUID)
	}

	// Verify parsed content
	userMsg, ok := session.Messages[0].Content.(*models.UserMessage)
	if !ok {
		t.Error("First message content is not UserMessage")
	} else if userMsg.Content != "Hello" {
		t.Errorf("First message content = %v, want Hello", userMsg.Content)
	}

	// Verify assistant message
	assistantMsg, ok := session.Messages[1].Content.(*models.AssistantMessage)
	if !ok {
		t.Error("Second message content is not AssistantMessage")
	} else {
		if assistantMsg.Model != "claude-3" {
			t.Errorf("Assistant model = %v, want claude-3", assistantMsg.Model)
		}
		if assistantMsg.Usage.InputTokens != 5 {
			t.Errorf("Input tokens = %v, want 5", assistantMsg.Usage.InputTokens)
		}
	}
}

func TestStreamMessages(t *testing.T) {
	testContent := `{"uuid":"msg1","sessionId":"session1","type":"user","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":"Test"}}
{"uuid":"msg2","sessionId":"session1","type":"assistant","timestamp":"2024-01-01T10:00:05Z","message":{"id":"asst1","type":"message","role":"assistant","model":"claude-3","content":[{"type":"text","text":"Response"}]}}
`

	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "stream.jsonl")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	reader := NewJSONLReader(testFile)
	
	var messages []*models.Message
	err := reader.StreamMessages(func(msg *models.Message) error {
		messages = append(messages, msg)
		return nil
	})

	if err != nil {
		t.Fatalf("StreamMessages() error = %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Message count = %v, want 2", len(messages))
	}
}

func TestStreamJSONLMessages(t *testing.T) {
	testContent := `{"uuid":"msg1","type":"user","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":"Test"}}

{"uuid":"msg2","type":"assistant","timestamp":"2024-01-01T10:00:05Z","message":{"id":"asst1","type":"message","role":"assistant","model":"claude-3","content":[{"type":"text","text":"Response"}]}}
`

	reader := strings.NewReader(testContent)
	
	var count int
	err := StreamJSONLMessages(reader, func(msg *models.Message) error {
		count++
		if msg.UUID == "" {
			t.Error("Message UUID is empty")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("StreamJSONLMessages() error = %v", err)
	}

	if count != 2 {
		t.Errorf("Message count = %v, want 2 (should skip empty line)", count)
	}
}

func TestJSONLReaderErrors(t *testing.T) {
	// Test non-existent file
	reader := NewJSONLReader("/non/existent/file.jsonl")
	_, err := reader.ReadSession()
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test empty file
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.jsonl")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	reader = NewJSONLReader(emptyFile)
	_, err = reader.ReadSession()
	if err == nil {
		t.Error("Expected error for empty file")
	}

	// Test malformed JSON
	malformedFile := filepath.Join(tmpDir, "malformed.jsonl")
	if err := os.WriteFile(malformedFile, []byte("not json\n{\"uuid\":\"msg1\"}"), 0644); err != nil {
		t.Fatalf("Failed to create malformed file: %v", err)
	}

	reader = NewJSONLReader(malformedFile)
	session, err := reader.ReadSession()
	if err != nil {
		t.Errorf("Should handle malformed lines gracefully, got error: %v", err)
	}
	if session == nil || len(session.Messages) != 1 {
		t.Error("Should parse valid lines despite malformed ones")
	}
}

func TestLargeJSONLFile(t *testing.T) {
	// Create a large JSONL file to test buffer handling
	tmpDir := t.TempDir()
	largeFile := filepath.Join(tmpDir, "large.jsonl")
	
	file, err := os.Create(largeFile)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}
	defer file.Close()

	// Write many messages
	for i := 0; i < 1000; i++ {
		timestamp := time.Now().Add(time.Duration(i) * time.Second).Format(time.RFC3339)
		line := fmt.Sprintf(`{"uuid":"msg%d","sessionId":"large-session","type":"user","timestamp":"%s","message":{"role":"user","content":"Message %d with some longer content to test buffer handling in the JSONL reader implementation"}}`, i, timestamp, i)
		if _, err := file.WriteString(line + "\n"); err != nil {
			t.Fatalf("Failed to write to large file: %v", err)
		}
	}
	file.Close()

	// Test reading large file
	reader := NewJSONLReader(largeFile)
	
	count := 0
	err = reader.StreamMessages(func(msg *models.Message) error {
		count++
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to stream large file: %v", err)
	}

	if count != 1000 {
		t.Errorf("Message count = %v, want 1000", count)
	}
}