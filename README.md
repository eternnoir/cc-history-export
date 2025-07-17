# Claude Code History Export

A command-line tool to export conversation history from Claude Code (.claude directory) to various formats including JSON and Markdown.

## Features

- Export entire Claude Code conversation history
- Filter by project paths and date ranges
- Multiple export formats: JSON, Markdown
- Batch export to separate files per project
- Include todo lists and session metadata
- Token usage statistics
- Flexible output options

## Installation

### From Source

Requires Go 1.21 or later:

```bash
git clone https://github.com/frankwang/cc-history-export.git
cd cc-history-export
go build -o cc-export ./cmd/cc-export
```

### Pre-built Binaries

Download the latest release from the [Releases](https://github.com/frankwang/cc-history-export/releases) page.

## Usage

### Basic Export

Export all conversations to Markdown (default):
```bash
cc-export --output conversations.md
```

Export to JSON format:
```bash
cc-export --format json --output export.json
```

### Filtering Options

Filter by project path:
```bash
cc-export --projects "/Users/myproject" --output myproject.json
```

Filter by date/time range:
```bash
# Date only (includes entire day)
cc-export --start-time 2024-01-01 --end-time 2024-12-31 --output 2024-export.json

# With specific time (use quotes for spaces)
cc-export --start-time "2024-01-01 09:00:00" --end-time "2024-01-31 18:00:00" --output january-work-hours.json
```

**Note on Time Zones:**
- Date/time values without timezone info are interpreted in your local timezone
- Sessions are filtered based on their last activity time (EndTime)
- To ensure consistent results across different timezones, consider using specific times rather than just dates

### Batch Export

Export each project to a separate file:
```bash
cc-export --batch --output exports/
```

This will create files like:
- `exports/project_myproject1.json`
- `exports/project_myproject2.json`

### Advanced Options

Include raw message data in JSON export:
```bash
cc-export --include-raw --output detailed-export.json
```

Include thinking content in Markdown:
```bash
cc-export --format markdown --show-thinking --output with-thinking.md
```

Limit number of sessions:
```bash
cc-export --max-sessions 100 --output limited-export.json
```

### Command-Line Options

```
  -batch
        Export each project/session to separate files
  -end-time string
        End date/time (YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)
  -format string
        Export format: json, markdown, html (default "markdown")
  -include-raw
        Include raw message data in JSON
  -include-todos
        Include todo lists (default true)
  -max-sessions int
        Maximum number of sessions to export (0 = unlimited)
  -output string
        Output file path (required)
  -pretty
        Pretty print JSON output (default true)
  -projects string
        Comma-separated project paths to filter
  -show-thinking
        Include thinking content in Markdown
  -source string
        Path to .claude directory (defaults to ~/.claude)
  -start-time string
        Start date/time (YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)
  -verbose
        Verbose output
  -version
        Show version
```

## Date/Time Filtering

The tool supports flexible date/time filtering options:

### Time Formats
- **Date only**: `YYYY-MM-DD` (e.g., `2024-01-01`)
  - Start time: Beginning of day (00:00:00) in local timezone
  - End time: End of day (23:59:59) in local timezone
- **Date and time**: `YYYY-MM-DD HH:MM:SS` (e.g., `2024-01-01 14:30:00`)
  - Must use quotes in command line due to space

### Filtering Logic
- Sessions are filtered based on their **last activity time** (EndTime)
- A session is included if its EndTime falls within the specified range
- This means long-running sessions that extend past the start time will be included

### Examples
```bash
# Include all sessions active on January 15, 2024
cc-export --start-time 2024-01-15 --end-time 2024-01-15 --output jan15.json

# Include sessions that were active between specific times
cc-export --start-time "2024-01-15 09:00:00" --end-time "2024-01-15 17:30:00" --output workday.json

# Include sessions from multiple days that ended within the range
cc-export --start-time 2024-01-14 --end-time 2024-01-16 --output three-days.json
```

## Export Formats

### JSON Format

The JSON export includes structured data with:
- Session metadata (ID, timestamps, duration)
- Message content with parsed structure
- Token usage statistics
- Todo lists (if present)

Example structure:
```json
{
  "projects": [
    {
      "id": "project-id",
      "name": "Project Name",
      "path": "/path/to/project",
      "session_count": 5,
      "message_count": 150,
      "token_usage": {
        "input": 10000,
        "output": 20000,
        "total": 30000
      },
      "sessions": [...]
    }
  ]
}
```

### Markdown Format

The Markdown export creates human-readable documents with:
- Project and session headers
- Formatted conversation threads
- Code blocks with syntax highlighting
- Todo lists with completion status
- Token usage summaries

## Development

### Project Structure

```
/cmd/cc-export         - CLI application
/internal/models       - Data models
/internal/reader       - File readers (JSONL, JSON)
/internal/converter    - Format converters (JSON, Markdown)
/internal/exporter     - Export logic
```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o cc-export ./cmd/cc-export
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.