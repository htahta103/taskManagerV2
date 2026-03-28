package task

import "time"

// Status is a task workflow status (OpenAPI TaskStatus).
type Status string

const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

// Priority is optional task priority.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// FocusBucket supports Today / Next / Later workflow.
type FocusBucket string

const (
	FocusNone  FocusBucket = "none"
	FocusToday FocusBucket = "today"
	FocusNext  FocusBucket = "next"
	FocusLater FocusBucket = "later"
)

// Task is the persisted task shape returned by the API.
type Task struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description *string      `json:"description"`
	Status      Status       `json:"status"`
	Priority    *Priority    `json:"priority"`
	DueDate     *string      `json:"due_date"`
	FocusBucket FocusBucket  `json:"focus_bucket"`
	ProjectID   *string      `json:"project_id"`
	AssigneeID  *string      `json:"assignee_id"`
	Tags        []Tag        `json:"tags"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Tag is embedded on tasks when expanded (create response uses empty list until tag attach exists).
type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreatePayload is the JSON body for POST /functions/v1/tasks.
type CreatePayload struct {
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	Status       *Status    `json:"status,omitempty"`
	Priority     *Priority  `json:"priority,omitempty"`
	DueDate      *string    `json:"due_date,omitempty"`
	FocusBucket  *FocusBucket `json:"focus_bucket,omitempty"`
	ProjectID    *string    `json:"project_id,omitempty"`
	AssigneeID   *string    `json:"assignee_id,omitempty"`
	TagIDs       []string   `json:"tag_ids,omitempty"`
}
