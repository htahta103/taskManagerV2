package tasks

import "time"

// TaskStatus matches OpenAPI TaskStatus enum.
type TaskStatus string

const (
	StatusTodo  TaskStatus = "todo"
	StatusDoing TaskStatus = "doing"
	StatusDone  TaskStatus = "done"
)

// FocusBucket matches OpenAPI FocusBucket enum.
type FocusBucket string

const (
	FocusNone  FocusBucket = "none"
	FocusToday FocusBucket = "today"
	FocusNext  FocusBucket = "next"
	FocusLater FocusBucket = "later"
)

// TaskPriority matches OpenAPI TaskPriority enum.
type TaskPriority string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
)

// Task is the API task model (OpenAPI Task schema).
type Task struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description *string       `json:"description,omitempty"`
	Status      TaskStatus    `json:"status"`
	Priority    *TaskPriority `json:"priority,omitempty"`
	DueDate     *string       `json:"due_date,omitempty"`
	FocusBucket FocusBucket   `json:"focus_bucket"`
	ProjectID   *string       `json:"project_id,omitempty"`
	AssigneeID  *string       `json:"assignee_id,omitempty"`
	Tags        []Tag         `json:"tags,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// Tag is embedded on Task responses (OpenAPI Tag).
type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
