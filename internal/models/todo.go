package models

// TodoStatus represents the status of a todo item
type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
)

// TodoPriority represents the priority of a todo item
type TodoPriority string

const (
	TodoPriorityHigh   TodoPriority = "high"
	TodoPriorityMedium TodoPriority = "medium"
	TodoPriorityLow    TodoPriority = "low"
)

// Todo represents a single todo item
type Todo struct {
	ID       string       `json:"id"`
	Content  string       `json:"content"`
	Status   TodoStatus   `json:"status"`
	Priority TodoPriority `json:"priority"`
}

// TodoList represents a collection of todos for a session
type TodoList struct {
	SessionID string  `json:"session_id"`
	AgentID   string  `json:"agent_id"`
	Todos     []*Todo `json:"todos"`
}

// GetTodosByStatus returns todos filtered by status
func (tl *TodoList) GetTodosByStatus(status TodoStatus) []*Todo {
	var filtered []*Todo
	for _, todo := range tl.Todos {
		if todo.Status == status {
			filtered = append(filtered, todo)
		}
	}
	return filtered
}

// GetTodosByPriority returns todos filtered by priority
func (tl *TodoList) GetTodosByPriority(priority TodoPriority) []*Todo {
	var filtered []*Todo
	for _, todo := range tl.Todos {
		if todo.Priority == priority {
			filtered = append(filtered, todo)
		}
	}
	return filtered
}

// GetCompletionRate returns the percentage of completed todos
func (tl *TodoList) GetCompletionRate() float64 {
	if len(tl.Todos) == 0 {
		return 0
	}
	
	completed := 0
	for _, todo := range tl.Todos {
		if todo.Status == TodoStatusCompleted {
			completed++
		}
	}
	
	return float64(completed) / float64(len(tl.Todos)) * 100
}