package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eternnoir/cc-history-export/internal/converter"
	"github.com/eternnoir/cc-history-export/internal/exporter"
	"github.com/eternnoir/cc-history-export/internal/models"
	"github.com/eternnoir/cc-history-export/internal/reader"
)

const version = "1.0.0"

type config struct {
	// Input options
	sourcePath   string
	projectPaths []string
	startTime    string
	endTime      string
	
	// Output options
	outputPath   string
	format       string
	batchExport  bool
	
	// Format-specific options
	prettyJSON   bool
	showThinking bool
	includeRaw   bool
	includeTodos bool
	
	// Other options
	maxSessions int
	verbose     bool
	version     bool
}

// parseDateTime parses various datetime formats
func parseDateTime(s string) (time.Time, error) {
	// Supported formats in order of precedence
	formats := []string{
		"2006-01-02 15:04:05",      // YYYY-MM-DD HH:MM:SS (local time)
		"2006-01-02 15:04",         // YYYY-MM-DD HH:MM (local time)
		"2006-01-02",               // YYYY-MM-DD (start of day in local time)
		time.RFC3339,               // Full RFC3339 with timezone
		"2006-01-02T15:04:05Z07:00", // ISO 8601 with timezone
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			// For formats without timezone info, use local time
			if format == "2006-01-02" || format == "2006-01-02 15:04:05" || format == "2006-01-02 15:04" {
				// Get local timezone
				loc := time.Local
				// Create time in local timezone
				t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
			}
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unsupported datetime format")
}

// isDateOnly checks if the input string is in date-only format
func isDateOnly(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func main() {
	cfg := parseFlags()
	
	if cfg.version {
		fmt.Printf("cc-export version %s\n", version)
		os.Exit(0)
	}
	
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *config {
	cfg := &config{}
	
	// Define flags
	flag.StringVar(&cfg.sourcePath, "source", "", "Path to .claude directory (defaults to ~/.claude)")
	flag.StringVar(&cfg.outputPath, "output", "", "Output file path (required)")
	flag.StringVar(&cfg.format, "format", "markdown", "Export format: json, markdown, html")
	
	// Filter flags
	projectsStr := flag.String("projects", "", "Comma-separated project paths to filter")
	flag.StringVar(&cfg.startTime, "start-time", "", "Start date/time (YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)")
	flag.StringVar(&cfg.endTime, "end-time", "", "End date/time (YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)")
	flag.IntVar(&cfg.maxSessions, "max-sessions", 0, "Maximum number of sessions to export (0 = unlimited)")
	
	// Format options
	flag.BoolVar(&cfg.prettyJSON, "pretty", true, "Pretty print JSON output")
	flag.BoolVar(&cfg.showThinking, "show-thinking", false, "Include thinking content in Markdown")
	flag.BoolVar(&cfg.includeRaw, "include-raw", false, "Include raw message data in JSON")
	flag.BoolVar(&cfg.includeTodos, "include-todos", true, "Include todo lists")
	
	// Export options
	flag.BoolVar(&cfg.batchExport, "batch", false, "Export each project/session to separate files")
	
	// Other flags
	flag.BoolVar(&cfg.verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&cfg.version, "version", false, "Show version")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Claude Code History Export Tool v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Export all data to Markdown (default)\n")
		fmt.Fprintf(os.Stderr, "  cc-export --output conversations.md\n\n")
		fmt.Fprintf(os.Stderr, "  # Export specific project to JSON\n")
		fmt.Fprintf(os.Stderr, "  cc-export --projects /Users/myproject --format json --output project.json\n\n")
		fmt.Fprintf(os.Stderr, "  # Export date range with batch output\n")
		fmt.Fprintf(os.Stderr, "  cc-export --start-time 2024-01-01 --end-time 2024-12-31 --batch --output exports/\n\n")
		fmt.Fprintf(os.Stderr, "  # Export with specific time range (use quotes for spaces)\n")
		fmt.Fprintf(os.Stderr, "  cc-export --start-time \"2024-01-01 09:00:00\" --end-time \"2024-01-31 18:00:00\" --output january.md\n\n")
	}
	
	flag.Parse()
	
	// Parse project paths
	if *projectsStr != "" {
		cfg.projectPaths = strings.Split(*projectsStr, ",")
		for i := range cfg.projectPaths {
			cfg.projectPaths[i] = strings.TrimSpace(cfg.projectPaths[i])
		}
	}
	
	// Default source path
	if cfg.sourcePath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.sourcePath = filepath.Join(home, ".claude")
		}
	}
	
	return cfg
}

func validateConfig(cfg *config) error {
	if cfg.outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	
	if cfg.sourcePath == "" {
		return fmt.Errorf("could not determine .claude directory path")
	}
	
	// Check if source directory exists
	if _, err := os.Stat(cfg.sourcePath); os.IsNotExist(err) {
		return fmt.Errorf(".claude directory not found at %s", cfg.sourcePath)
	}
	
	// Validate format
	switch cfg.format {
	case "json", "markdown":
		// Valid formats
	case "html":
		return fmt.Errorf("HTML format not yet implemented")
	default:
		return fmt.Errorf("unsupported format: %s", cfg.format)
	}
	
	// Validate dates
	if cfg.startTime != "" {
		if _, err := parseDateTime(cfg.startTime); err != nil {
			return fmt.Errorf("invalid start time format: %s (use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)", cfg.startTime)
		}
	}
	
	if cfg.endTime != "" {
		if _, err := parseDateTime(cfg.endTime); err != nil {
			return fmt.Errorf("invalid end time format: %s (use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)", cfg.endTime)
		}
	}
	
	return nil
}

func run(cfg *config) error {
	if cfg.verbose {
		fmt.Printf("Scanning %s...\n", cfg.sourcePath)
	}
	
	// Create scanner options
	scanOpts := &reader.ScanOptions{
		ProjectPaths: cfg.projectPaths,
		IncludeTodos: cfg.includeTodos,
		MaxSessions:  cfg.maxSessions,
	}
	
	// Parse dates
	if cfg.startTime != "" {
		t, _ := parseDateTime(cfg.startTime)
		scanOpts.StartDate = &t
	}
	if cfg.endTime != "" {
		t, _ := parseDateTime(cfg.endTime)
		// For date-only input, add 1 day to include the entire end date
		if isDateOnly(cfg.endTime) {
			t = t.Add(24 * time.Hour)
		}
		scanOpts.EndDate = &t
	}
	
	// Scan projects
	scanner := reader.NewScanner(cfg.sourcePath, scanOpts)
	projects, err := scanner.ScanProjects()
	if err != nil {
		return fmt.Errorf("failed to scan projects: %w", err)
	}
	
	if len(projects) == 0 {
		fmt.Println("No projects found matching the criteria")
		return nil
	}
	
	if cfg.verbose {
		fmt.Printf("Found %d projects\n", len(projects))
		totalSessions := 0
		totalMessages := 0
		for _, p := range projects {
			totalSessions += p.GetSessionCount()
			totalMessages += p.GetTotalMessages()
		}
		fmt.Printf("Total sessions: %d\n", totalSessions)
		fmt.Printf("Total messages: %d\n", totalMessages)
	}
	
	// Create exporter
	exportOpts := &exporter.ExportOptions{
		Format:          exporter.Format(cfg.format),
		IncludeMetadata: true,
		IncludeStats:    true,
	}
	
	// Set format-specific options
	switch cfg.format {
	case "json":
		exportOpts.FormatOptions = &converter.JSONOptions{
			PrettyPrint:        cfg.prettyJSON,
			IncludeRawMessages: cfg.includeRaw,
			OmitEmpty:          true,
		}
	case "markdown":
		exportOpts.FormatOptions = &converter.MarkdownOptions{
			ShowTimestamps: true,
			ShowTokenUsage: true,
			ShowThinking:   cfg.showThinking,
			ShowUUIDs:      false,
		}
	}
	
	fileExporter, err := exporter.NewFileExporter(exportOpts)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}
	
	// Export data
	if cfg.batchExport {
		return batchExport(fileExporter, projects, cfg)
	} else {
		return singleExport(fileExporter, projects, cfg)
	}
}

func singleExport(exp *exporter.FileExporter, projects []*models.Project, cfg *config) error {
	if cfg.verbose {
		fmt.Printf("Exporting to %s...\n", cfg.outputPath)
	}
	
	// Export based on number of projects
	var err error
	if len(projects) == 1 {
		err = exp.ExportToFile(cfg.outputPath, projects[0], exporter.ExportTypeProject)
	} else {
		err = exp.ExportToFile(cfg.outputPath, projects, exporter.ExportTypeProjects)
	}
	
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}
	
	fmt.Printf("Successfully exported to %s\n", cfg.outputPath)
	return nil
}

func batchExport(exp *exporter.FileExporter, projects []*models.Project, cfg *config) error {
	// Ensure output directory exists
	if err := os.MkdirAll(cfg.outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Determine file extension
	ext := ".md"
	if cfg.format == "json" {
		ext = ".json"
	} else if cfg.format == "html" {
		ext = ".html"
	}
	
	// Create batch exporter
	nameFormat := "project_%s" + ext
	batchExp := exporter.NewBatchExporter(exp, cfg.outputPath, nameFormat)
	
	if cfg.verbose {
		fmt.Printf("Batch exporting %d projects to %s...\n", len(projects), cfg.outputPath)
	}
	
	// Export projects
	result, err := batchExp.ExportProjects(projects)
	if err != nil {
		return fmt.Errorf("batch export failed: %w", err)
	}
	
	// Print results
	fmt.Println(result.Summary())
	
	if result.HasErrors() {
		fmt.Fprintf(os.Stderr, "\nErrors occurred:\n")
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", e.Item, e.Error)
		}
	}
	
	if cfg.verbose && len(result.Files) > 0 {
		fmt.Println("\nExported files:")
		for _, f := range result.Files {
			fmt.Printf("  - %s\n", f)
		}
	}
	
	return nil
}