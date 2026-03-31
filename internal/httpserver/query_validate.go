package httpserver

import (
	"net/http"
	"strconv"
	"strings"
)

// parseLimit parses OpenAPI Limit (1–100, default def).
func parseLimit(q string, def int) (int, bool, map[string]string) {
	if strings.TrimSpace(q) == "" {
		return def, true, nil
	}
	n, err := strconv.Atoi(q)
	if err != nil || n < 1 || n > 100 {
		return 0, false, map[string]string{"limit": "must be an integer between 1 and 100"}
	}
	return n, true, nil
}

func writeLimitError(w http.ResponseWriter, fields map[string]string) {
	writeValidation(w, "invalid query parameters", fields)
}

const (
	taskStatusTodo  = "todo"
	taskStatusDoing = "doing"
	taskStatusDone  = "done"
)

func normalizeTaskStatus(raw string) (string, bool) {
	s := strings.TrimSpace(strings.ToLower(raw))
	switch s {
	case taskStatusTodo, taskStatusDoing, taskStatusDone:
		return s, true
	default:
		return "", false
	}
}

func normalizeTaskView(raw string) (string, bool) {
	v := strings.TrimSpace(strings.ToLower(raw))
	switch v {
	case "inbox", "today", "next", "later":
		return v, true
	default:
		return "", false
	}
}

func normalizePriorityPtr(p *string) (*string, bool) {
	if p == nil {
		return nil, true
	}
	s := strings.TrimSpace(strings.ToLower(*p))
	switch s {
	case "low", "medium", "high":
		out := s
		return &out, true
	default:
		return nil, false
	}
}

func normalizeFocusBucketPtr(f *string) (*string, bool) {
	if f == nil {
		return nil, true
	}
	s := strings.TrimSpace(strings.ToLower(*f))
	switch s {
	case "none", "today", "next", "later":
		out := s
		return &out, true
	default:
		return nil, false
	}
}

func validateTaskTitle(title string) (trimmed string, fields map[string]string) {
	t := strings.TrimSpace(title)
	if t == "" {
		return "", map[string]string{"title": "required"}
	}
	if len(t) > 200 {
		return "", map[string]string{"title": "must be at most 200 characters"}
	}
	return t, nil
}

func validateDescriptionPtr(p *string) map[string]string {
	if p == nil {
		return nil
	}
	if len(*p) > 10000 {
		return map[string]string{"description": "must be at most 10000 characters"}
	}
	return nil
}
