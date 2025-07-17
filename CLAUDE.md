# Claude Code History Export - Development Guidelines

## Technology Stack
- Language: Go
- Minimum Go version: 1.21

## Development Standards

### Go Development Guidelines
- Follow modern Go idioms and best practices
- Use Go modules for dependency management
- Adhere to effective Go patterns (https://go.dev/doc/effective_go)
- Follow standard Go project layout conventions
- Use meaningful variable and function names
- Prefer composition over inheritance
- Handle errors explicitly - no silent failures
- Write tests for core functionality

### Code Style
- Use `gofmt` for code formatting
- Follow Go naming conventions (exported vs unexported identifiers)
- Keep functions small and focused
- Document exported types, functions, and packages
- Use context for cancellation and timeouts where appropriate

### Project Structure
```
/cmd         - Main applications
/internal    - Private application code
/examples    - Example usage
```