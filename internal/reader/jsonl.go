package reader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/eternnoir/cc-history-export/internal/models"
)

// JSONLReader reads and parses JSONL conversation files
type JSONLReader struct {
	filePath string
}

// NewJSONLReader creates a new JSONL reader for the given file
func NewJSONLReader(filePath string) *JSONLReader {
	return &JSONLReader{
		filePath: filePath,
	}
}

// ReadSession reads all messages from the JSONL file and returns a session
func (r *JSONLReader) ReadSession() (*models.Session, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	session := &models.Session{}
	scanner := bufio.NewScanner(file)
	
	// Increase buffer size for large lines
	const maxCapacity = 1024 * 1024 * 10 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		
		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		var msg models.Message
		if err := json.Unmarshal(line, &msg); err != nil {
			// Log error but continue processing
			fmt.Fprintf(os.Stderr, "Warning: failed to parse line %d: %v\n", lineNum, err)
			continue
		}

		// Set session ID from first message
		if session.ID == "" && msg.SessionID != "" {
			session.ID = msg.SessionID
		}

		// Parse message content
		if err := msg.ParseContent(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse content for message %s: %v\n", msg.UUID, err)
		}

		session.AddMessage(&msg)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if len(session.Messages) == 0 {
		return nil, fmt.Errorf("no messages found in file")
	}

	return session, nil
}

// StreamMessages reads messages one by one using a callback function
func (r *JSONLReader) StreamMessages(callback func(*models.Message) error) error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return StreamJSONLMessages(file, callback)
}

// StreamJSONLMessages streams messages from any io.Reader
func StreamJSONLMessages(reader io.Reader, callback func(*models.Message) error) error {
	scanner := bufio.NewScanner(reader)
	
	// Increase buffer size for large lines
	const maxCapacity = 1024 * 1024 * 10 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		
		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		var msg models.Message
		if err := json.Unmarshal(line, &msg); err != nil {
			// Log error but continue processing
			fmt.Fprintf(os.Stderr, "Warning: failed to parse line %d: %v\n", lineNum, err)
			continue
		}

		// Parse message content
		if err := msg.ParseContent(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse content for message %s: %v\n", msg.UUID, err)
		}

		// Call the callback function
		if err := callback(&msg); err != nil {
			return fmt.Errorf("callback error at line %d: %w", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading: %w", err)
	}

	return nil
}