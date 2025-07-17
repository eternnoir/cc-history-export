package exporter

import (
	"fmt"
	"io"

	"github.com/eternnoir/cc-history-export/internal/models"
)

// Format represents the export format
type Format string

const (
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
	FormatHTML     Format = "html"
)

// ExportType represents what to export
type ExportType string

const (
	ExportTypeSession  ExportType = "session"
	ExportTypeProject  ExportType = "project"
	ExportTypeProjects ExportType = "projects"
)

// Exporter is the interface for exporting data
type Exporter interface {
	// Export writes the exported data to the writer
	Export(writer io.Writer, data interface{}, exportType ExportType) error
	
	// ExportToFile exports data to a file
	ExportToFile(filename string, data interface{}, exportType ExportType) error
	
	// GetFormat returns the format of this exporter
	GetFormat() Format
}

// ExportOptions provides common export options
type ExportOptions struct {
	// Format to export to
	Format Format
	
	// Include metadata in export
	IncludeMetadata bool
	
	// Include statistics
	IncludeStats bool
	
	// Custom options for specific formats
	FormatOptions interface{}
}

// Validate validates the export options
func (o *ExportOptions) Validate() error {
	switch o.Format {
	case FormatJSON, FormatMarkdown, FormatHTML:
		// Valid formats
	default:
		return fmt.Errorf("unsupported format: %s", o.Format)
	}
	return nil
}

// ExportResult contains information about an export operation
type ExportResult struct {
	// Number of items exported
	ItemsExported int
	
	// Total size of exported data
	BytesWritten int64
	
	// Any warnings during export
	Warnings []string
	
	// Export format used
	Format Format
}

// BaseExporter provides common functionality for exporters
type BaseExporter struct {
	format  Format
	options *ExportOptions
}

// NewBaseExporter creates a new base exporter
func NewBaseExporter(format Format, options *ExportOptions) *BaseExporter {
	if options == nil {
		options = &ExportOptions{
			Format:          format,
			IncludeMetadata: true,
			IncludeStats:    true,
		}
	}
	return &BaseExporter{
		format:  format,
		options: options,
	}
}

// GetFormat returns the export format
func (e *BaseExporter) GetFormat() Format {
	return e.format
}

// GetOptions returns the export options
func (e *BaseExporter) GetOptions() *ExportOptions {
	return e.options
}

// ValidateData validates that the data is of the expected type
func ValidateData(data interface{}, exportType ExportType) error {
	switch exportType {
	case ExportTypeSession:
		if _, ok := data.(*models.Session); !ok {
			return fmt.Errorf("expected *models.Session for export type %s", exportType)
		}
	case ExportTypeProject:
		if _, ok := data.(*models.Project); !ok {
			return fmt.Errorf("expected *models.Project for export type %s", exportType)
		}
	case ExportTypeProjects:
		if _, ok := data.([]*models.Project); !ok {
			return fmt.Errorf("expected []*models.Project for export type %s", exportType)
		}
	default:
		return fmt.Errorf("unsupported export type: %s", exportType)
	}
	return nil
}

// CountingWriter wraps an io.Writer and counts bytes written
type CountingWriter struct {
	writer       io.Writer
	bytesWritten int64
}

// NewCountingWriter creates a new counting writer
func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{writer: w}
}

// Write implements io.Writer
func (cw *CountingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.writer.Write(p)
	cw.bytesWritten += int64(n)
	return n, err
}

// BytesWritten returns the total bytes written
func (cw *CountingWriter) BytesWritten() int64 {
	return cw.bytesWritten
}