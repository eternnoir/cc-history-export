package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestCLIIntegration(t *testing.T) {
	// Create test data directory
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	projectsDir := filepath.Join(claudeDir, "projects")
	projectDir := filepath.Join(projectsDir, "-Users-test-project")
	
	// Create directories
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}
	
	// Create test session file
	sessionContent := `{"uuid":"msg1","sessionId":"session1","type":"user","userType":"external","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":"Hello"}}
{"uuid":"msg2","sessionId":"session1","type":"assistant","timestamp":"2024-01-01T10:00:05Z","message":{"id":"asst1","type":"message","role":"assistant","model":"claude-3","content":[{"type":"text","text":"Hi there!"}],"usage":{"input_tokens":5,"output_tokens":10}}}`
	
	sessionFile := filepath.Join(projectDir, "session1.jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}
	
	// Test basic configuration
	cfg := &config{
		sourcePath:   claudeDir,
		outputPath:   filepath.Join(tmpDir, "export.json"),
		format:       "json",
		prettyJSON:   true,
		includeTodos: false,
		verbose:      false,
	}
	
	// Run the export
	if err := run(cfg); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	
	// Verify output file exists
	if _, err := os.Stat(cfg.outputPath); os.IsNotExist(err) {
		t.Error("Expected output file does not exist")
	}
	
	// Test Markdown export
	cfg.outputPath = filepath.Join(tmpDir, "export.md")
	cfg.format = "markdown"
	
	if err := run(cfg); err != nil {
		t.Fatalf("run() with markdown format error = %v", err)
	}
	
	// Verify Markdown file exists
	if _, err := os.Stat(cfg.outputPath); os.IsNotExist(err) {
		t.Error("Expected markdown file does not exist")
	}
}

func TestValidateConfig(t *testing.T) {
	// Test valid config
	cfg := &config{
		sourcePath: "/tmp/.claude",
		outputPath: "/tmp/output.json",
		format:     "json",
		startDate:  "2024-01-01",
		endDate:    "2024-12-31",
	}
	
	// Create source directory for validation
	os.MkdirAll(cfg.sourcePath, 0755)
	defer os.RemoveAll(cfg.sourcePath)
	
	if err := validateConfig(cfg); err != nil {
		t.Errorf("validateConfig() error for valid config = %v", err)
	}
	
	// Test missing output path
	cfg.outputPath = ""
	if err := validateConfig(cfg); err == nil {
		t.Error("validateConfig() should error for missing output path")
	}
	
	// Test invalid date format
	cfg.outputPath = "/tmp/output.json"
	cfg.startDate = "01-01-2024"
	if err := validateConfig(cfg); err == nil {
		t.Error("validateConfig() should error for invalid date format")
	}
	
	// Test unsupported format
	cfg.startDate = "2024-01-01"
	cfg.format = "xml"
	if err := validateConfig(cfg); err == nil {
		t.Error("validateConfig() should error for unsupported format")
	}
}

func TestParseFlags(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	// Test with arguments
	os.Args = []string{
		"cc-export",
		"--output", "test.json",
		"--projects", "proj1,proj2",
		"--format", "markdown",
		"--verbose",
	}
	
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	cfg := parseFlags()
	
	if cfg.outputPath != "test.json" {
		t.Errorf("outputPath = %v, want test.json", cfg.outputPath)
	}
	
	if cfg.format != "markdown" {
		t.Errorf("format = %v, want markdown", cfg.format)
	}
	
	if !cfg.verbose {
		t.Error("verbose should be true")
	}
	
	if len(cfg.projectPaths) != 2 {
		t.Errorf("projectPaths length = %v, want 2", len(cfg.projectPaths))
	}
}