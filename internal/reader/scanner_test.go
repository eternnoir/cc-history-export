package reader

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanner(t *testing.T) {
	// Create test directory structure
	tmpDir := t.TempDir()
	
	// Create claude directory structure
	claudeDir := filepath.Join(tmpDir, ".claude")
	projectsDir := filepath.Join(claudeDir, "projects")
	todosDir := filepath.Join(claudeDir, "todos")
	
	// Create directories
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatalf("Failed to create todos dir: %v", err)
	}
	
	// Create test project
	project1Dir := filepath.Join(projectsDir, "-Users-test-project1")
	if err := os.MkdirAll(project1Dir, 0755); err != nil {
		t.Fatalf("Failed to create project1 dir: %v", err)
	}
	
	// Create test session files
	session1Content := `{"uuid":"msg1","sessionId":"session1","type":"user","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":"Hello"}}
{"uuid":"msg2","sessionId":"session1","type":"assistant","timestamp":"2024-01-01T10:00:05Z","message":{"id":"asst1","type":"message","role":"assistant","model":"claude-3","content":[{"type":"text","text":"Hi!"}]}}`
	
	session1File := filepath.Join(project1Dir, "session1.jsonl")
	if err := os.WriteFile(session1File, []byte(session1Content), 0644); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}
	
	// Create test todo file
	todoContent := `[
		{"id":"1","content":"Test todo","status":"pending","priority":"high"},
		{"id":"2","content":"Another todo","status":"completed","priority":"medium"}
	]`
	
	todoFile := filepath.Join(todosDir, "session1-agent-agent1.json")
	if err := os.WriteFile(todoFile, []byte(todoContent), 0644); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}
	
	// Create CLAUDE.md
	configContent := "Test configuration"
	configFile := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Test scanning
	scanner := NewScanner(claudeDir, &ScanOptions{
		IncludeTodos: true,
	})
	
	projects, err := scanner.ScanProjects()
	if err != nil {
		t.Fatalf("ScanProjects() error = %v", err)
	}
	
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}
	
	project := projects[0]
	if project.ID != "-Users-test-project1" {
		t.Errorf("Project ID = %v, want -Users-test-project1", project.ID)
	}
	
	if len(project.Sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(project.Sessions))
	}
	
	if len(project.TodoLists) != 1 {
		t.Errorf("Expected 1 todo list, got %d", len(project.TodoLists))
	}
	
	// Test config reading
	config, err := scanner.ScanClaudeConfig()
	if err != nil {
		t.Fatalf("ScanClaudeConfig() error = %v", err)
	}
	
	if config != configContent {
		t.Errorf("Config content = %v, want %v", config, configContent)
	}
}

func TestScannerWithFilters(t *testing.T) {
	// Create test directory structure
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	projectsDir := filepath.Join(claudeDir, "projects")
	
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}
	
	// Create multiple projects
	projectDirs := []string{
		"-Users-test-project1",
		"-Users-test-project2",
		"-Users-other-project",
	}
	
	for _, proj := range projectDirs {
		projDir := filepath.Join(projectsDir, proj)
		if err := os.MkdirAll(projDir, 0755); err != nil {
			t.Fatalf("Failed to create project dir %s: %v", proj, err)
		}
		
		// Create a session file with different timestamps
		var startTimestamp, endTimestamp string
		switch proj {
		case "-Users-test-project1":
			startTimestamp = "2024-01-01T10:00:00Z"
			endTimestamp = "2024-01-01T10:00:00Z"
		case "-Users-test-project2":
			startTimestamp = "2024-02-01T10:00:00Z"
			endTimestamp = "2024-02-01T10:00:00Z"
		case "-Users-other-project":
			startTimestamp = "2024-03-01T10:00:00Z"
			endTimestamp = "2024-03-01T10:00:00Z"
		}
		
		sessionContent := `{"uuid":"msg1","sessionId":"` + proj + `","type":"user","timestamp":"` + startTimestamp + `","message":{"role":"user","content":"Hello"}}
{"uuid":"msg2","sessionId":"` + proj + `","type":"assistant","timestamp":"` + endTimestamp + `","message":{"role":"assistant","content":"Hi"}}`
		sessionFile := filepath.Join(projDir, "session.jsonl")
		if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
			t.Fatalf("Failed to create session file: %v", err)
		}
	}
	
	// First scan all projects to see what we have
	scanner := NewScanner(claudeDir, &ScanOptions{})
	allProjects, _ := scanner.ScanProjects()
	t.Logf("All projects found: %d", len(allProjects))
	for _, p := range allProjects {
		t.Logf("  Project: %s (decoded path: %s)", p.ID, p.Path)
	}
	
	// Test project path filter
	scanner = NewScanner(claudeDir, &ScanOptions{
		ProjectPaths: []string{"/Users/test/"},
	})
	
	filteredProjects, err := scanner.ScanProjects()
	if err != nil {
		t.Fatalf("ScanProjects() error = %v", err)
	}
	
	if len(filteredProjects) != 2 {
		t.Errorf("Expected 2 projects with 'test-project' in path, got %d", len(filteredProjects))
		for _, p := range filteredProjects {
			t.Logf("Found project: %s (path: %s)", p.ID, p.Path)
		}
	}
	
	// Test date filter
	startDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
	
	scanner = NewScanner(claudeDir, &ScanOptions{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	
	dateFilteredProjects, err := scanner.ScanProjects()
	if err != nil {
		t.Fatalf("ScanProjects() with date filter error = %v", err)
	}
	
	// Should only include project2 (Feb 1, 2024)
	if len(dateFilteredProjects) != 1 {
		t.Errorf("Expected 1 project within date range, got %d", len(dateFilteredProjects))
	}
	
	// Test max sessions limit
	scanner = NewScanner(claudeDir, &ScanOptions{
		MaxSessions: 2,
	})
	
	limitedProjects, err := scanner.ScanProjects()
	if err != nil {
		t.Fatalf("ScanProjects() with max sessions error = %v", err)
	}
	
	totalSessions := 0
	for _, p := range limitedProjects {
		totalSessions += len(p.Sessions)
	}
	
	if totalSessions > 2 {
		t.Errorf("Expected at most 2 sessions, got %d", totalSessions)
	}
}

func TestScannerDateFilterWithEndTime(t *testing.T) {
	// Test the new date filter logic that only checks EndTime
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	projectsDir := filepath.Join(claudeDir, "projects")
	
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}
	
	// Create a project with a long-running session
	projDir := filepath.Join(projectsDir, "-Users-test-longrunning")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}
	
	// Session that starts on 07/14 and ends on 07/16
	sessionContent := `{"uuid":"msg1","sessionId":"longrunning","type":"user","timestamp":"2024-07-14T10:00:00Z","message":{"role":"user","content":"Start"}}
{"uuid":"msg2","sessionId":"longrunning","type":"assistant","timestamp":"2024-07-16T15:00:00Z","message":{"role":"assistant","content":"End"}}`
	
	sessionFile := filepath.Join(projDir, "session.jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}
	
	// Test with date range 07/15 - 07/20
	startDate := time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 7, 20, 0, 0, 0, 0, time.UTC)
	
	scanner := NewScanner(claudeDir, &ScanOptions{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	
	projects, err := scanner.ScanProjects()
	if err != nil {
		t.Fatalf("ScanProjects() error = %v", err)
	}
	
	// Should include the session because EndTime (07/16) is within the range
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}
	
	if len(projects[0].Sessions) != 1 {
		t.Errorf("Expected 1 session to be included, got %d", len(projects[0].Sessions))
	}
	
	// Test with date range that excludes the session
	startDate2 := time.Date(2024, 7, 17, 0, 0, 0, 0, time.UTC)
	endDate2 := time.Date(2024, 7, 20, 0, 0, 0, 0, time.UTC)
	
	scanner2 := NewScanner(claudeDir, &ScanOptions{
		StartDate: &startDate2,
		EndDate:   &endDate2,
	})
	
	projects2, err := scanner2.ScanProjects()
	if err != nil {
		t.Fatalf("ScanProjects() error = %v", err)
	}
	
	// Should not include the session because EndTime (07/16) is before StartDate (07/17)
	if len(projects2) != 0 {
		t.Errorf("Expected 0 projects, got %d", len(projects2))
	}
}

func TestScannerErrors(t *testing.T) {
	// Test non-existent directory
	scanner := NewScanner("/non/existent/path", nil)
	_, err := scanner.ScanProjects()
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
	
	// Test empty directory
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create claude dir: %v", err)
	}
	
	scanner = NewScanner(claudeDir, nil)
	_, err = scanner.ScanProjects()
	if err == nil {
		t.Error("Expected error for missing projects directory")
	}
}