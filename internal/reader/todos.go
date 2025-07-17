package reader

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/eternnoir/cc-history-export/internal/models"
)

// TodoReader reads todo JSON files
type TodoReader struct {
	filePath string
}

// NewTodoReader creates a new todo reader
func NewTodoReader(filePath string) *TodoReader {
	return &TodoReader{
		filePath: filePath,
	}
}

// Read reads and parses a todo JSON file
func (r *TodoReader) Read() ([]*models.Todo, error) {
	content, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read todo file: %w", err)
	}

	// The todo files contain an array of todo objects
	var todos []*models.Todo
	if err := json.Unmarshal(content, &todos); err != nil {
		return nil, fmt.Errorf("failed to parse todo JSON: %w", err)
	}

	return todos, nil
}

// ReadTodoList reads a todo file and returns it as a TodoList
func (r *TodoReader) ReadTodoList(sessionID, agentID string) (*models.TodoList, error) {
	todos, err := r.Read()
	if err != nil {
		return nil, err
	}

	return &models.TodoList{
		SessionID: sessionID,
		AgentID:   agentID,
		Todos:     todos,
	}, nil
}