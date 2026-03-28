package tasklist

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Task matches docs/api/openapi.yaml Task schema (subset used by MVP dashboard).
type Task struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
	Status       string  `json:"status"`
	Priority     *string `json:"priority,omitempty"`
	DueDate      *string `json:"due_date,omitempty"`
	FocusBucket  string  `json:"focus_bucket"`
	ProjectID    *string `json:"project_id,omitempty"`
	AssigneeID   *string `json:"assignee_id,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type page struct {
	Items       []Task  `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

func seedTasks() []Task {
	t0 := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	desc := "Kickoff notes"
	pLow := "low"
	pMed := "medium"
	pHigh := "high"
	return []Task{
		{ID: "550e8400-e29b-41d4-a716-446655440001", Title: "Review sprint backlog", Description: &desc, Status: "todo", Priority: &pMed, FocusBucket: "today", CreatedAt: t0.Format(time.RFC3339), UpdatedAt: t0.Add(time.Hour).Format(time.RFC3339)},
		{ID: "550e8400-e29b-41d4-a716-446655440002", Title: "Draft API pagination", Status: "doing", Priority: &pHigh, FocusBucket: "next", CreatedAt: t0.Add(time.Minute).Format(time.RFC3339), UpdatedAt: t0.Add(2 * time.Hour).Format(time.RFC3339)},
		{ID: "550e8400-e29b-41d4-a716-446655440003", Title: "Update dependencies", Status: "todo", Priority: &pLow, FocusBucket: "later", CreatedAt: t0.Add(2 * time.Minute).Format(time.RFC3339), UpdatedAt: t0.Add(3 * time.Hour).Format(time.RFC3339)},
		{ID: "550e8400-e29b-41d4-a716-446655440004", Title: "Plan dashboard filters", Status: "done", Priority: &pMed, FocusBucket: "none", CreatedAt: t0.Add(3 * time.Minute).Format(time.RFC3339), UpdatedAt: t0.Add(4 * time.Hour).Format(time.RFC3339)},
		{ID: "550e8400-e29b-41d4-a716-446655440005", Title: "Email design review", Status: "todo", Priority: &pHigh, FocusBucket: "today", CreatedAt: t0.Add(4 * time.Minute).Format(time.RFC3339), UpdatedAt: t0.Add(5 * time.Hour).Format(time.RFC3339)},
		{ID: "550e8400-e29b-41d4-a716-446655440006", Title: "Fix flaky health check test", Status: "doing", Priority: &pLow, FocusBucket: "next", CreatedAt: t0.Add(5 * time.Minute).Format(time.RFC3339), UpdatedAt: t0.Add(6 * time.Hour).Format(time.RFC3339)},
		{ID: "550e8400-e29b-41d4-a716-446655440007", Title: "Document env vars", Status: "done", Priority: &pLow, FocusBucket: "later", CreatedAt: t0.Add(6 * time.Minute).Format(time.RFC3339), UpdatedAt: t0.Add(7 * time.Hour).Format(time.RFC3339)},
		{ID: "550e8400-e29b-41d4-a716-446655440008", Title: "Schedule 1:1s", Status: "todo", Priority: &pMed, FocusBucket: "none", CreatedAt: t0.Add(7 * time.Minute).Format(time.RFC3339), UpdatedAt: t0.Add(8 * time.Hour).Format(time.RFC3339)},
	}
}

var allTasks = seedTasks()

// ListHandler serves GET /api/v1/tasks with optional query: status, q (title substring).
func ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	qLower := strings.ToLower(q)

	out := make([]Task, 0, len(allTasks))
	for _, t := range allTasks {
		if status != "" && t.Status != status {
			continue
		}
		if qLower != "" && !strings.Contains(strings.ToLower(t.Title), qLower) {
			continue
		}
		out = append(out, t)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(page{Items: out, NextCursor: nil})
}
