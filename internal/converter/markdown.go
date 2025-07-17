package converter

import (
	"fmt"
	"strings"
	"time"

	"github.com/eternnoir/cc-history-export/internal/models"
)

// MarkdownConverter converts sessions and projects to Markdown format
type MarkdownConverter struct {
	options MarkdownOptions
}

// MarkdownOptions provides options for Markdown conversion
type MarkdownOptions struct {
	// Include timestamps for each message
	ShowTimestamps bool
	// Include token usage statistics
	ShowTokenUsage bool
	// Include thinking/internal content
	ShowThinking bool
	// Include message UUIDs
	ShowUUIDs bool
}

// NewMarkdownConverter creates a new Markdown converter
func NewMarkdownConverter(options *MarkdownOptions) *MarkdownConverter {
	if options == nil {
		options = &MarkdownOptions{
			ShowTimestamps: true,
			ShowTokenUsage: true,
		}
	}
	return &MarkdownConverter{
		options: *options,
	}
}

// ConvertSession converts a session to Markdown format
func (c *MarkdownConverter) ConvertSession(session *models.Session) string {
	var sb strings.Builder

	// Session header
	sb.WriteString(fmt.Sprintf("# Session: %s\n\n", session.ID))
	
	if !session.StartTime.IsZero() {
		sb.WriteString(fmt.Sprintf("**Started:** %s  \n", session.StartTime.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("**Ended:** %s  \n", session.EndTime.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("**Duration:** %s  \n", session.GetDuration()))
	}
	
	sb.WriteString(fmt.Sprintf("**Messages:** %d  \n", session.GetMessageCount()))
	
	if c.options.ShowTokenUsage {
		inputTokens, outputTokens := session.GetTokenUsage()
		if inputTokens > 0 || outputTokens > 0 {
			sb.WriteString(fmt.Sprintf("**Token Usage:** Input: %d, Output: %d  \n", inputTokens, outputTokens))
		}
	}
	
	sb.WriteString("\n---\n\n")

	// Convert each message
	for i, msg := range session.Messages {
		if i > 0 {
			sb.WriteString("\n---\n\n")
		}
		sb.WriteString(c.ConvertMessage(msg))
	}

	return sb.String()
}

// ConvertMessage converts a single message to Markdown format
func (c *MarkdownConverter) ConvertMessage(msg *models.Message) string {
	var sb strings.Builder

	// Message header
	switch msg.Type {
	case models.MessageTypeUser:
		sb.WriteString("### 👤 User\n\n")
	case models.MessageTypeAssistant:
		sb.WriteString("### 🤖 Assistant\n\n")
	default:
		sb.WriteString(fmt.Sprintf("### %s\n\n", msg.Type))
	}

	// Metadata
	if c.options.ShowTimestamps && !msg.Timestamp.IsZero() {
		sb.WriteString(fmt.Sprintf("*%s*  \n", msg.Timestamp.Format("2006-01-02 15:04:05")))
	}
	if c.options.ShowUUIDs && msg.UUID != "" {
		sb.WriteString(fmt.Sprintf("*UUID: %s*  \n", msg.UUID))
	}
	if msg.CWD != "" {
		sb.WriteString(fmt.Sprintf("*Working Directory: `%s`*  \n", msg.CWD))
	}
	if sb.Len() > 0 {
		sb.WriteString("\n")
	}

	// Message content
	switch msg.Type {
	case models.MessageTypeUser:
		if userMsg, ok := msg.Content.(*models.UserMessage); ok {
			sb.WriteString(userMsg.Content)
			sb.WriteString("\n")
		} else if toolResults, ok := msg.Content.([]models.ToolResult); ok {
			sb.WriteString("**Tool Results:**\n\n")
			for _, result := range toolResults {
				sb.WriteString(fmt.Sprintf("- Tool: `%s`\n", result.ToolUseID))
				sb.WriteString(fmt.Sprintf("  - Type: %s\n", result.Type))
				sb.WriteString(fmt.Sprintf("  - Content: %s\n", string(result.Content)))
			}
		}
		
	case models.MessageTypeAssistant:
		if assistantMsg, ok := msg.Content.(*models.AssistantMessage); ok {
			// Model info
			if assistantMsg.Model != "" {
				sb.WriteString(fmt.Sprintf("*Model: %s*\n\n", assistantMsg.Model))
			}
			
			// Content blocks
			for _, content := range assistantMsg.Content {
				switch content.Type {
				case "text":
					sb.WriteString(content.Text)
					sb.WriteString("\n\n")
					
				case "thinking":
					if c.options.ShowThinking {
						sb.WriteString("<details>\n<summary>💭 Thinking</summary>\n\n")
						sb.WriteString(content.Thinking)
						sb.WriteString("\n\n</details>\n\n")
					}
					
				case "tool_use":
					sb.WriteString(fmt.Sprintf("**🔧 Tool Use:** `%s`\n\n", content.Name))
					if content.ID != "" {
						sb.WriteString(fmt.Sprintf("*ID: %s*\n\n", content.ID))
					}
					sb.WriteString("```json\n")
					sb.WriteString(string(content.Input))
					sb.WriteString("\n```\n\n")
					
				default:
					sb.WriteString(fmt.Sprintf("**%s:**\n\n", content.Type))
					if content.Text != "" {
						sb.WriteString(content.Text)
						sb.WriteString("\n\n")
					}
				}
			}
			
			// Token usage
			if c.options.ShowTokenUsage && assistantMsg.Usage != nil {
				sb.WriteString(fmt.Sprintf("\n*Tokens - Input: %d, Output: %d*\n", 
					assistantMsg.Usage.InputTokens+assistantMsg.Usage.CacheReadInputTokens,
					assistantMsg.Usage.OutputTokens))
			}
		}
	}

	return sb.String()
}

// ConvertProject converts an entire project to Markdown format
func (c *MarkdownConverter) ConvertProject(project *models.Project) string {
	var sb strings.Builder

	// Project header
	sb.WriteString(fmt.Sprintf("# Project: %s\n\n", project.GetProjectName()))
	sb.WriteString(fmt.Sprintf("**Path:** `%s`  \n", project.Path))
	sb.WriteString(fmt.Sprintf("**Sessions:** %d  \n", project.GetSessionCount()))
	sb.WriteString(fmt.Sprintf("**Total Messages:** %d  \n", project.GetTotalMessages()))
	
	if c.options.ShowTokenUsage {
		inputTokens, outputTokens := project.GetTotalTokenUsage()
		if inputTokens > 0 || outputTokens > 0 {
			sb.WriteString(fmt.Sprintf("**Total Token Usage:** Input: %d, Output: %d  \n", inputTokens, outputTokens))
		}
	}
	
	start, end := project.GetTimeRange()
	if !start.IsZero() {
		sb.WriteString(fmt.Sprintf("**Date Range:** %s to %s  \n", 
			start.Format("2006-01-02"), 
			end.Format("2006-01-02")))
	}
	
	// Todo lists summary
	if len(project.TodoLists) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Todo Lists (%d)\n\n", len(project.TodoLists)))
		for _, todoList := range project.TodoLists {
			sb.WriteString(c.ConvertTodoList(todoList))
			sb.WriteString("\n")
		}
	}
	
	// Sessions
	sb.WriteString("\n## Sessions\n\n")
	for i, session := range project.Sessions {
		if i > 0 {
			sb.WriteString("\n\n---\n\n")
		}
		sb.WriteString(c.ConvertSession(session))
	}

	return sb.String()
}

// ConvertTodoList converts a todo list to Markdown format
func (c *MarkdownConverter) ConvertTodoList(todoList *models.TodoList) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### Todo List - Session: %s\n\n", todoList.SessionID))
	
	if todoList.AgentID != "" {
		sb.WriteString(fmt.Sprintf("*Agent: %s*  \n", todoList.AgentID))
	}
	
	completionRate := todoList.GetCompletionRate()
	sb.WriteString(fmt.Sprintf("*Completion: %.0f%%*\n\n", completionRate))

	// Group todos by status
	pending := todoList.GetTodosByStatus(models.TodoStatusPending)
	inProgress := todoList.GetTodosByStatus(models.TodoStatusInProgress)
	completed := todoList.GetTodosByStatus(models.TodoStatusCompleted)

	if len(pending) > 0 {
		sb.WriteString("#### ⏳ Pending\n\n")
		for _, todo := range pending {
			sb.WriteString(fmt.Sprintf("- [ ] %s (%s)\n", todo.Content, todo.Priority))
		}
		sb.WriteString("\n")
	}

	if len(inProgress) > 0 {
		sb.WriteString("#### 🔄 In Progress\n\n")
		for _, todo := range inProgress {
			sb.WriteString(fmt.Sprintf("- [ ] %s (%s)\n", todo.Content, todo.Priority))
		}
		sb.WriteString("\n")
	}

	if len(completed) > 0 {
		sb.WriteString("#### ✅ Completed\n\n")
		for _, todo := range completed {
			sb.WriteString(fmt.Sprintf("- [x] %s (%s)\n", todo.Content, todo.Priority))
		}
	}

	return sb.String()
}