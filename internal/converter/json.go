package converter

import (
	"encoding/json"
	"fmt"

	"github.com/eternnoir/cc-history-export/internal/models"
)

// JSONConverter converts sessions and projects to JSON format
type JSONConverter struct {
	options JSONOptions
}

// JSONOptions provides options for JSON conversion
type JSONOptions struct {
	// Pretty print the JSON output
	PrettyPrint bool
	// Include raw message data
	IncludeRawMessages bool
	// Exclude empty fields
	OmitEmpty bool
}

// NewJSONConverter creates a new JSON converter
func NewJSONConverter(options *JSONOptions) *JSONConverter {
	if options == nil {
		options = &JSONOptions{
			PrettyPrint: true,
			OmitEmpty:   true,
		}
	}
	return &JSONConverter{
		options: *options,
	}
}

// JSONMessage represents a message in the exported JSON format
type JSONMessage struct {
	UUID       string      `json:"uuid"`
	ParentUUID *string     `json:"parent_uuid,omitempty"`
	SessionID  string      `json:"session_id"`
	Type       string      `json:"type"`
	UserType   string      `json:"user_type,omitempty"`
	Timestamp  string      `json:"timestamp"`
	CWD        string      `json:"cwd,omitempty"`
	Content    interface{} `json:"content"`
	RawMessage interface{} `json:"raw_message,omitempty"`
}

// JSONSession represents a session in the exported JSON format
type JSONSession struct {
	ID               string         `json:"id"`
	ProjectID        string         `json:"project_id,omitempty"`
	StartTime        string         `json:"start_time"`
	EndTime          string         `json:"end_time"`
	Duration         string         `json:"duration"`
	MessageCount     int            `json:"message_count"`
	UserMessages     int            `json:"user_messages"`
	AssistantMessages int           `json:"assistant_messages"`
	TokenUsage       *TokenUsage    `json:"token_usage,omitempty"`
	Messages         []*JSONMessage `json:"messages"`
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
	Total  int `json:"total"`
}

// JSONProject represents a project in the exported JSON format
type JSONProject struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Path         string           `json:"path"`
	EncodedPath  string           `json:"encoded_path"`
	SessionCount int              `json:"session_count"`
	MessageCount int              `json:"message_count"`
	DateRange    *DateRange       `json:"date_range,omitempty"`
	TokenUsage   *TokenUsage      `json:"token_usage,omitempty"`
	Sessions     []*JSONSession   `json:"sessions"`
	TodoLists    []*JSONTodoList  `json:"todo_lists,omitempty"`
}

// DateRange represents a date range
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// JSONTodoList represents a todo list in the exported JSON format
type JSONTodoList struct {
	SessionID      string     `json:"session_id"`
	AgentID        string     `json:"agent_id"`
	TodoCount      int        `json:"todo_count"`
	CompletionRate float64    `json:"completion_rate"`
	Todos          []*JSONTodo `json:"todos"`
}

// JSONTodo represents a todo item in the exported JSON format
type JSONTodo struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}

// ConvertSession converts a session to JSON format
func (c *JSONConverter) ConvertSession(session *models.Session) ([]byte, error) {
	jsonSession := c.sessionToJSON(session)
	return c.marshal(jsonSession)
}

// ConvertProject converts a project to JSON format
func (c *JSONConverter) ConvertProject(project *models.Project) ([]byte, error) {
	jsonProject := c.projectToJSON(project)
	return c.marshal(jsonProject)
}

// ConvertProjects converts multiple projects to JSON format
func (c *JSONConverter) ConvertProjects(projects []*models.Project) ([]byte, error) {
	jsonProjects := make([]*JSONProject, len(projects))
	for i, project := range projects {
		jsonProjects[i] = c.projectToJSON(project)
	}
	
	result := map[string]interface{}{
		"projects":      jsonProjects,
		"project_count": len(projects),
	}
	
	return c.marshal(result)
}

// sessionToJSON converts a models.Session to JSONSession
func (c *JSONConverter) sessionToJSON(session *models.Session) *JSONSession {
	inputTokens, outputTokens := session.GetTokenUsage()
	
	jsonSession := &JSONSession{
		ID:                session.ID,
		ProjectID:         session.ProjectID,
		StartTime:         session.StartTime.Format("2006-01-02T15:04:05Z"),
		EndTime:           session.EndTime.Format("2006-01-02T15:04:05Z"),
		Duration:          session.GetDuration().String(),
		MessageCount:      session.GetMessageCount(),
		UserMessages:      session.GetUserMessageCount(),
		AssistantMessages: session.GetAssistantMessageCount(),
		Messages:          make([]*JSONMessage, len(session.Messages)),
	}
	
	if inputTokens > 0 || outputTokens > 0 {
		jsonSession.TokenUsage = &TokenUsage{
			Input:  inputTokens,
			Output: outputTokens,
			Total:  inputTokens + outputTokens,
		}
	}
	
	for i, msg := range session.Messages {
		jsonSession.Messages[i] = c.messageToJSON(msg)
	}
	
	return jsonSession
}

// messageToJSON converts a models.Message to JSONMessage
func (c *JSONConverter) messageToJSON(msg *models.Message) *JSONMessage {
	jsonMsg := &JSONMessage{
		UUID:      msg.UUID,
		SessionID: msg.SessionID,
		Type:      string(msg.Type),
		UserType:  msg.UserType,
		CWD:       msg.CWD,
		Content:   msg.Content,
	}
	
	if msg.ParentUUID != nil {
		jsonMsg.ParentUUID = msg.ParentUUID
	}
	
	if !msg.Timestamp.IsZero() {
		jsonMsg.Timestamp = msg.Timestamp.Format("2006-01-02T15:04:05Z")
	}
	
	if c.options.IncludeRawMessages && len(msg.Message) > 0 {
		var rawData interface{}
		if err := json.Unmarshal(msg.Message, &rawData); err == nil {
			jsonMsg.RawMessage = rawData
		}
	}
	
	return jsonMsg
}

// projectToJSON converts a models.Project to JSONProject
func (c *JSONConverter) projectToJSON(project *models.Project) *JSONProject {
	inputTokens, outputTokens := project.GetTotalTokenUsage()
	
	jsonProject := &JSONProject{
		ID:           project.ID,
		Name:         project.GetProjectName(),
		Path:         project.Path,
		EncodedPath:  project.EncodedPath,
		SessionCount: project.GetSessionCount(),
		MessageCount: project.GetTotalMessages(),
		Sessions:     make([]*JSONSession, len(project.Sessions)),
		TodoLists:    make([]*JSONTodoList, len(project.TodoLists)),
	}
	
	if inputTokens > 0 || outputTokens > 0 {
		jsonProject.TokenUsage = &TokenUsage{
			Input:  inputTokens,
			Output: outputTokens,
			Total:  inputTokens + outputTokens,
		}
	}
	
	start, end := project.GetTimeRange()
	if !start.IsZero() {
		jsonProject.DateRange = &DateRange{
			Start: start.Format("2006-01-02"),
			End:   end.Format("2006-01-02"),
		}
	}
	
	for i, session := range project.Sessions {
		jsonProject.Sessions[i] = c.sessionToJSON(session)
	}
	
	for i, todoList := range project.TodoLists {
		jsonProject.TodoLists[i] = c.todoListToJSON(todoList)
	}
	
	return jsonProject
}

// todoListToJSON converts a models.TodoList to JSONTodoList
func (c *JSONConverter) todoListToJSON(todoList *models.TodoList) *JSONTodoList {
	jsonTodoList := &JSONTodoList{
		SessionID:      todoList.SessionID,
		AgentID:        todoList.AgentID,
		TodoCount:      len(todoList.Todos),
		CompletionRate: todoList.GetCompletionRate(),
		Todos:          make([]*JSONTodo, len(todoList.Todos)),
	}
	
	for i, todo := range todoList.Todos {
		jsonTodoList.Todos[i] = &JSONTodo{
			ID:       todo.ID,
			Content:  todo.Content,
			Status:   string(todo.Status),
			Priority: string(todo.Priority),
		}
	}
	
	return jsonTodoList
}

// marshal handles JSON marshaling with options
func (c *JSONConverter) marshal(v interface{}) ([]byte, error) {
	if c.options.PrettyPrint {
		return json.MarshalIndent(v, "", "  ")
	}
	return json.Marshal(v)
}

// ConvertSessionToCompactJSON converts a session to a compact JSON format
// suitable for streaming or line-by-line processing
func (c *JSONConverter) ConvertSessionToCompactJSON(session *models.Session) ([]byte, error) {
	// Create a simplified representation
	compact := map[string]interface{}{
		"id":        session.ID,
		"project":   session.ProjectID,
		"start":     session.StartTime.Unix(),
		"end":       session.EndTime.Unix(),
		"messages":  session.GetMessageCount(),
	}
	
	// Add token usage if available
	inputTokens, outputTokens := session.GetTokenUsage()
	if inputTokens > 0 || outputTokens > 0 {
		compact["tokens"] = map[string]int{
			"in":  inputTokens,
			"out": outputTokens,
		}
	}
	
	return json.Marshal(compact)
}

// ValidateJSON validates that the data can be properly marshaled to JSON
func (c *JSONConverter) ValidateJSON(v interface{}) error {
	data, err := c.marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	
	// Try to unmarshal back to verify it's valid JSON
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("invalid JSON produced: %w", err)
	}
	
	return nil
}