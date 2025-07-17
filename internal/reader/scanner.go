package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eternnoir/cc-history-export/internal/models"
)

// ScanOptions provides options for scanning the Claude directory
type ScanOptions struct {
	// Filter by date range
	StartDate *time.Time
	EndDate   *time.Time
	
	// Filter by project paths
	ProjectPaths []string
	
	// Include todo lists
	IncludeTodos bool
	
	// Include shell snapshots
	IncludeShellSnapshots bool
	
	// Maximum number of sessions to process (0 = unlimited)
	MaxSessions int
}

// Scanner scans the Claude directory structure
type Scanner struct {
	basePath string
	options  *ScanOptions
}

// NewScanner creates a new scanner for the given Claude directory
func NewScanner(basePath string, options *ScanOptions) *Scanner {
	if options == nil {
		options = &ScanOptions{}
	}
	return &Scanner{
		basePath: basePath,
		options:  options,
	}
}

// ScanProjects scans all projects in the Claude directory
func (s *Scanner) ScanProjects() ([]*models.Project, error) {
	projectsPath := filepath.Join(s.basePath, "projects")
	
	// Check if projects directory exists
	if _, err := os.Stat(projectsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("projects directory not found: %s", projectsPath)
	}

	// Read all project directories
	entries, err := os.ReadDir(projectsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	var projects []*models.Project
	sessionCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if we should process this project
		if !s.shouldProcessProject(entry.Name()) {
			continue
		}

		project := models.NewProject(entry.Name())
		projectPath := filepath.Join(projectsPath, entry.Name())

		// Scan sessions in the project
		sessions, err := s.scanProjectSessions(projectPath, project.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to scan sessions for project %s: %v\n", entry.Name(), err)
			continue
		}

		// Apply date filters and session limit
		for _, session := range sessions {
			if s.shouldIncludeSession(session) {
				project.AddSession(session)
				sessionCount++
				
				// Check session limit
				if s.options.MaxSessions > 0 && sessionCount >= s.options.MaxSessions {
					projects = append(projects, project)
					return projects, nil
				}
			}
		}

		// Scan todos if requested
		if s.options.IncludeTodos {
			todos, err := s.scanProjectTodos(project.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to scan todos for project %s: %v\n", entry.Name(), err)
			} else {
				for _, todo := range todos {
					project.AddTodoList(todo)
				}
			}
		}

		if len(project.Sessions) > 0 {
			projects = append(projects, project)
		}
	}

	return projects, nil
}

// scanProjectSessions scans all JSONL files in a project directory
func (s *Scanner) scanProjectSessions(projectPath, projectID string) ([]*models.Session, error) {
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil, err
	}

	var sessions []*models.Session

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		filePath := filepath.Join(projectPath, entry.Name())
		reader := NewJSONLReader(filePath)
		
		session, err := reader.ReadSession()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read session file %s: %v\n", filePath, err)
			continue
		}

		session.ProjectID = projectID
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// scanProjectTodos scans all todo JSON files for a project
func (s *Scanner) scanProjectTodos(projectID string) ([]*models.TodoList, error) {
	todosPath := filepath.Join(s.basePath, "todos")
	
	if _, err := os.Stat(todosPath); os.IsNotExist(err) {
		return nil, nil // Todos directory might not exist
	}

	entries, err := os.ReadDir(todosPath)
	if err != nil {
		return nil, err
	}

	var todoLists []*models.TodoList
	
	// Look for files that match the project pattern
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Extract session ID from filename (format: sessionID-agent-agentID.json)
		parts := strings.Split(entry.Name(), "-agent-")
		if len(parts) != 2 {
			continue
		}
		
		sessionID := parts[0]
		
		// Check if this todo belongs to one of our project's sessions
		// This is a simplified check - in a real implementation, we'd need to
		// verify the session ID belongs to the project
		
		filePath := filepath.Join(todosPath, entry.Name())
		todoReader := NewTodoReader(filePath)
		
		todos, err := todoReader.Read()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read todo file %s: %v\n", filePath, err)
			continue
		}

		if len(todos) > 0 {
			agentID := strings.TrimSuffix(parts[1], ".json")
			todoList := &models.TodoList{
				SessionID: sessionID,
				AgentID:   agentID,
				Todos:     todos,
			}
			todoLists = append(todoLists, todoList)
		}
	}

	return todoLists, nil
}

// shouldProcessProject checks if a project should be processed based on filters
func (s *Scanner) shouldProcessProject(encodedPath string) bool {
	if len(s.options.ProjectPaths) == 0 {
		return true
	}

	// Decode the path
	decodedPath := strings.ReplaceAll(encodedPath, "-", "/")
	
	for _, filterPath := range s.options.ProjectPaths {
		if strings.Contains(decodedPath, filterPath) {
			return true
		}
	}

	return false
}

// shouldIncludeSession checks if a session should be included based on date filters
func (s *Scanner) shouldIncludeSession(session *models.Session) bool {
	if s.options.StartDate != nil && session.EndTime.Before(*s.options.StartDate) {
		return false
	}
	
	if s.options.EndDate != nil && session.EndTime.After(*s.options.EndDate) {
		return false
	}
	
	return true
}

// ScanClaudeConfig reads the CLAUDE.md configuration file
func (s *Scanner) ScanClaudeConfig() (string, error) {
	configPath := filepath.Join(s.basePath, "CLAUDE.md")
	
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Config file is optional
		}
		return "", fmt.Errorf("failed to read CLAUDE.md: %w", err)
	}
	
	return string(content), nil
}