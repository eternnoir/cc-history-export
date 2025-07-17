package exporter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eternnoir/cc-history-export/internal/converter"
	"github.com/eternnoir/cc-history-export/internal/models"
)

func createTestSession() *models.Session {
	session := &models.Session{
		ID:        "test-session",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
	}

	msg := &models.Message{
		UUID:      "msg1",
		Type:      models.MessageTypeUser,
		UserType:  "external",
		Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Message:   json.RawMessage(`{"role":"user","content":"Test message"}`),
	}
	msg.ParseContent()
	session.AddMessage(msg)

	return session
}

func createTestProject() *models.Project {
	project := models.NewProject("-Users-test-project")
	project.AddSession(createTestSession())
	return project
}

func TestFileExporterJSON(t *testing.T) {
	exporter, err := NewFileExporter(&ExportOptions{
		Format: FormatJSON,
		FormatOptions: &converter.JSONOptions{
			PrettyPrint: true,
		},
	})
	if err != nil {
		t.Fatalf("NewFileExporter() error = %v", err)
	}

	// Test session export
	session := createTestSession()
	var buf bytes.Buffer
	
	err = exporter.Export(&buf, session, ExportTypeSession)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Verify JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if result["id"] != "test-session" {
		t.Errorf("Session ID = %v, want test-session", result["id"])
	}
}

func TestFileExporterMarkdown(t *testing.T) {
	exporter, err := NewFileExporter(&ExportOptions{
		Format: FormatMarkdown,
		FormatOptions: &converter.MarkdownOptions{
			ShowTimestamps: true,
		},
	})
	if err != nil {
		t.Fatalf("NewFileExporter() error = %v", err)
	}

	// Test project export
	project := createTestProject()
	var buf bytes.Buffer
	
	err = exporter.Export(&buf, project, ExportTypeProject)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Verify Markdown output
	output := buf.String()
	if !strings.Contains(output, "# Project: project") {
		t.Errorf("Missing project header in Markdown. Output:\n%s", output)
	}
	if !strings.Contains(output, "Test message") {
		t.Errorf("Missing message content in Markdown. Output:\n%s", output)
	}
}

func TestFileExporterToFile(t *testing.T) {
	tmpDir := t.TempDir()
	
	exporter, err := NewFileExporter(&ExportOptions{
		Format: FormatJSON,
	})
	if err != nil {
		t.Fatalf("NewFileExporter() error = %v", err)
	}

	// Test export to file
	session := createTestSession()
	filename := filepath.Join(tmpDir, "test-export.json")
	
	err = exporter.ExportToFile(filename, session, ExportTypeSession)
	if err != nil {
		t.Fatalf("ExportToFile() error = %v", err)
	}

	// Verify file exists and contains valid JSON
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse exported JSON: %v", err)
	}

	if result["id"] != "test-session" {
		t.Errorf("Session ID in file = %v, want test-session", result["id"])
	}
}

func TestFileExporterProjects(t *testing.T) {
	exporter, err := NewFileExporter(&ExportOptions{
		Format: FormatJSON,
	})
	if err != nil {
		t.Fatalf("NewFileExporter() error = %v", err)
	}

	// Test multiple projects export
	projects := []*models.Project{
		createTestProject(),
		createTestProject(),
	}
	
	var buf bytes.Buffer
	err = exporter.Export(&buf, projects, ExportTypeProjects)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Verify JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if result["project_count"].(float64) != 2 {
		t.Errorf("project_count = %v, want 2", result["project_count"])
	}
}

func TestExportOptionsValidation(t *testing.T) {
	// Test valid options
	opts := &ExportOptions{
		Format: FormatJSON,
	}
	if err := opts.Validate(); err != nil {
		t.Errorf("Validate() error for valid options = %v", err)
	}

	// Test invalid format
	opts = &ExportOptions{
		Format: "invalid",
	}
	if err := opts.Validate(); err == nil {
		t.Error("Validate() should error for invalid format")
	}
}

func TestValidateData(t *testing.T) {
	session := createTestSession()
	project := createTestProject()
	projects := []*models.Project{project}

	// Test valid data
	if err := ValidateData(session, ExportTypeSession); err != nil {
		t.Errorf("ValidateData() error for valid session = %v", err)
	}

	if err := ValidateData(project, ExportTypeProject); err != nil {
		t.Errorf("ValidateData() error for valid project = %v", err)
	}

	if err := ValidateData(projects, ExportTypeProjects); err != nil {
		t.Errorf("ValidateData() error for valid projects = %v", err)
	}

	// Test invalid data
	if err := ValidateData(session, ExportTypeProject); err == nil {
		t.Error("ValidateData() should error for mismatched type")
	}
}

func TestBatchExporter(t *testing.T) {
	tmpDir := t.TempDir()
	
	fileExporter, err := NewFileExporter(&ExportOptions{
		Format: FormatJSON,
	})
	if err != nil {
		t.Fatalf("NewFileExporter() error = %v", err)
	}

	batchExporter := NewBatchExporter(fileExporter, tmpDir, "session_%s.json")

	// Create test sessions
	sessions := []*models.Session{
		createTestSession(),
		createTestSession(),
	}
	sessions[1].ID = "test-session-2"

	// Export sessions
	result, err := batchExporter.ExportSessions(sessions)
	if err != nil {
		t.Fatalf("ExportSessions() error = %v", err)
	}

	// Verify results
	if result.TotalItems != 2 {
		t.Errorf("TotalItems = %v, want 2", result.TotalItems)
	}

	if result.SuccessCount != 2 {
		t.Errorf("SuccessCount = %v, want 2", result.SuccessCount)
	}

	if len(result.Files) != 2 {
		t.Errorf("Files count = %v, want 2", len(result.Files))
	}

	if result.HasErrors() {
		t.Error("Unexpected errors in batch export")
	}

	// Verify files exist
	for _, file := range result.Files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", file)
		}
	}

	// Test summary
	summary := result.Summary()
	if !strings.Contains(summary, "2/2") {
		t.Errorf("Summary does not contain expected count: %s", summary)
	}
}

func TestCountingWriter(t *testing.T) {
	var buf bytes.Buffer
	cw := NewCountingWriter(&buf)

	data := []byte("Hello, World!")
	n, err := cw.Write(data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if n != len(data) {
		t.Errorf("Write() returned %d, want %d", n, len(data))
	}

	if cw.BytesWritten() != int64(len(data)) {
		t.Errorf("BytesWritten() = %d, want %d", cw.BytesWritten(), len(data))
	}

	// Write more data
	moreData := []byte(" More data")
	cw.Write(moreData)

	expectedTotal := int64(len(data) + len(moreData))
	if cw.BytesWritten() != expectedTotal {
		t.Errorf("BytesWritten() after second write = %d, want %d", cw.BytesWritten(), expectedTotal)
	}
}