package exporter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/eternnoir/cc-history-export/internal/converter"
	"github.com/eternnoir/cc-history-export/internal/models"
)

// FileExporter exports data to files
type FileExporter struct {
	*BaseExporter
	jsonConverter     *converter.JSONConverter
	markdownConverter *converter.MarkdownConverter
}

// NewFileExporter creates a new file exporter
func NewFileExporter(options *ExportOptions) (*FileExporter, error) {
	if options == nil {
		options = &ExportOptions{
			Format:          FormatJSON,
			IncludeMetadata: true,
			IncludeStats:    true,
		}
	}

	if err := options.Validate(); err != nil {
		return nil, err
	}

	exporter := &FileExporter{
		BaseExporter: NewBaseExporter(options.Format, options),
	}

	// Initialize converters based on format
	switch options.Format {
	case FormatJSON:
		jsonOpts := &converter.JSONOptions{
			PrettyPrint: true,
			OmitEmpty:   true,
		}
		if opts, ok := options.FormatOptions.(*converter.JSONOptions); ok {
			jsonOpts = opts
		}
		exporter.jsonConverter = converter.NewJSONConverter(jsonOpts)

	case FormatMarkdown:
		mdOpts := &converter.MarkdownOptions{
			ShowTimestamps: true,
			ShowTokenUsage: true,
		}
		if opts, ok := options.FormatOptions.(*converter.MarkdownOptions); ok {
			mdOpts = opts
		}
		exporter.markdownConverter = converter.NewMarkdownConverter(mdOpts)

	case FormatHTML:
		return nil, fmt.Errorf("HTML format not yet implemented")
	}

	return exporter, nil
}

// Export writes the exported data to the writer
func (e *FileExporter) Export(writer io.Writer, data interface{}, exportType ExportType) error {
	if err := ValidateData(data, exportType); err != nil {
		return err
	}

	countingWriter := NewCountingWriter(writer)

	switch e.format {
	case FormatJSON:
		return e.exportJSON(countingWriter, data, exportType)
	case FormatMarkdown:
		return e.exportMarkdown(countingWriter, data, exportType)
	default:
		return fmt.Errorf("unsupported format: %s", e.format)
	}
}

// ExportToFile exports data to a file
func (e *FileExporter) ExportToFile(filename string, data interface{}, exportType ExportType) error {
	// If filename is empty or "-", write to stdout
	if filename == "" || filename == "-" {
		return e.Export(os.Stdout, data, exportType)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file for writing
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Export to file
	if err := e.Export(file, data, exportType); err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	return nil
}

// exportJSON exports data as JSON
func (e *FileExporter) exportJSON(writer io.Writer, data interface{}, exportType ExportType) error {
	var jsonData []byte
	var err error

	switch exportType {
	case ExportTypeSession:
		session := data.(*models.Session)
		jsonData, err = e.jsonConverter.ConvertSession(session)
		
	case ExportTypeProject:
		project := data.(*models.Project)
		jsonData, err = e.jsonConverter.ConvertProject(project)
		
	case ExportTypeProjects:
		projects := data.([]*models.Project)
		jsonData, err = e.jsonConverter.ConvertProjects(projects)
		
	default:
		return fmt.Errorf("unsupported export type: %s", exportType)
	}

	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	_, err = writer.Write(jsonData)
	return err
}

// exportMarkdown exports data as Markdown
func (e *FileExporter) exportMarkdown(writer io.Writer, data interface{}, exportType ExportType) error {
	var markdown string

	switch exportType {
	case ExportTypeSession:
		session := data.(*models.Session)
		markdown = e.markdownConverter.ConvertSession(session)
		
	case ExportTypeProject:
		project := data.(*models.Project)
		markdown = e.markdownConverter.ConvertProject(project)
		
	case ExportTypeProjects:
		projects := data.([]*models.Project)
		// Convert each project and combine
		for i, project := range projects {
			if i > 0 {
				markdown += "\n\n---\n\n"
			}
			markdown += e.markdownConverter.ConvertProject(project)
		}
		
	default:
		return fmt.Errorf("unsupported export type: %s", exportType)
	}

	_, err := writer.Write([]byte(markdown))
	return err
}

// BatchExporter exports multiple items to separate files
type BatchExporter struct {
	exporter   *FileExporter
	outputDir  string
	nameFormat string // e.g., "session_%s.json"
}

// NewBatchExporter creates a new batch exporter
func NewBatchExporter(exporter *FileExporter, outputDir string, nameFormat string) *BatchExporter {
	return &BatchExporter{
		exporter:   exporter,
		outputDir:  outputDir,
		nameFormat: nameFormat,
	}
}

// ExportSessions exports multiple sessions to separate files
func (b *BatchExporter) ExportSessions(sessions []*models.Session) (*BatchExportResult, error) {
	result := &BatchExportResult{
		TotalItems: len(sessions),
		Format:     b.exporter.GetFormat(),
	}

	for _, session := range sessions {
		filename := filepath.Join(b.outputDir, fmt.Sprintf(b.nameFormat, session.ID))
		
		if err := b.exporter.ExportToFile(filename, session, ExportTypeSession); err != nil {
			result.Errors = append(result.Errors, ExportError{
				Item:  session.ID,
				Error: err.Error(),
			})
		} else {
			result.SuccessCount++
			result.Files = append(result.Files, filename)
		}
	}

	return result, nil
}

// ExportProjects exports multiple projects to separate files
func (b *BatchExporter) ExportProjects(projects []*models.Project) (*BatchExportResult, error) {
	result := &BatchExportResult{
		TotalItems: len(projects),
		Format:     b.exporter.GetFormat(),
	}

	for _, project := range projects {
		filename := filepath.Join(b.outputDir, fmt.Sprintf(b.nameFormat, project.GetProjectName()))
		
		if err := b.exporter.ExportToFile(filename, project, ExportTypeProject); err != nil {
			result.Errors = append(result.Errors, ExportError{
				Item:  project.ID,
				Error: err.Error(),
			})
		} else {
			result.SuccessCount++
			result.Files = append(result.Files, filename)
		}
	}

	return result, nil
}

// BatchExportResult contains results from batch export
type BatchExportResult struct {
	TotalItems   int
	SuccessCount int
	Files        []string
	Errors       []ExportError
	Format       Format
}

// ExportError represents an error during batch export
type ExportError struct {
	Item  string
	Error string
}

// HasErrors returns true if there were any errors
func (r *BatchExportResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// Summary returns a summary of the batch export
func (r *BatchExportResult) Summary() string {
	return fmt.Sprintf("Exported %d/%d items successfully in %s format", 
		r.SuccessCount, r.TotalItems, r.Format)
}