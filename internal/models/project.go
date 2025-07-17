package models

import (
	"path/filepath"
	"strings"
	"time"
)

// Project represents a Claude Code project
type Project struct {
	ID          string       `json:"id"`
	Path        string       `json:"path"`        // Original project path
	EncodedPath string       `json:"encoded_path"` // Path as stored in .claude directory
	Sessions    []*Session   `json:"sessions"`
	TodoLists   []*TodoList  `json:"todo_lists,omitempty"`
}

// NewProject creates a new project from an encoded path
func NewProject(encodedPath string) *Project {
	// Decode the path by replacing - with /
	decodedPath := strings.ReplaceAll(encodedPath, "-", "/")
	
	return &Project{
		ID:          encodedPath,
		Path:        decodedPath,
		EncodedPath: encodedPath,
		Sessions:    make([]*Session, 0),
		TodoLists:   make([]*TodoList, 0),
	}
}

// AddSession adds a session to the project
func (p *Project) AddSession(session *Session) {
	session.ProjectID = p.ID
	p.Sessions = append(p.Sessions, session)
}

// AddTodoList adds a todo list to the project
func (p *Project) AddTodoList(todoList *TodoList) {
	p.TodoLists = append(p.TodoLists, todoList)
}

// GetSessionCount returns the total number of sessions
func (p *Project) GetSessionCount() int {
	return len(p.Sessions)
}

// GetTotalMessages returns the total number of messages across all sessions
func (p *Project) GetTotalMessages() int {
	total := 0
	for _, session := range p.Sessions {
		total += session.GetMessageCount()
	}
	return total
}

// GetTimeRange returns the earliest and latest timestamps in the project
func (p *Project) GetTimeRange() (start, end time.Time) {
	if len(p.Sessions) == 0 {
		return
	}
	
	for _, session := range p.Sessions {
		if start.IsZero() || session.StartTime.Before(start) {
			start = session.StartTime
		}
		if session.EndTime.After(end) {
			end = session.EndTime
		}
	}
	return
}

// GetProjectName returns a human-readable project name
func (p *Project) GetProjectName() string {
	// Extract the last component of the path as the project name
	return filepath.Base(p.Path)
}

// GetTotalTokenUsage calculates total token usage across all sessions
func (p *Project) GetTotalTokenUsage() (input int, output int) {
	for _, session := range p.Sessions {
		sessionInput, sessionOutput := session.GetTokenUsage()
		input += sessionInput
		output += sessionOutput
	}
	return
}