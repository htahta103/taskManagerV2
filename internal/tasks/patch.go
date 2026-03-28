package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("task not found")
	ErrInvalidJSON   = errors.New("invalid json")
	ErrInvalidTaskID = errors.New("invalid task id")
)

// PatchError carries HTTP-level validation errors.
type PatchError struct {
	Msg  string
	Code string
}

func (e *PatchError) Error() string { return e.Msg }

// ApplyPatch merges a JSON body into an existing task using partial-field semantics:
// omitted keys are unchanged; null sets nullable fields to cleared.
func ApplyPatch(base Task, body []byte) (Task, error) {
	if len(strings.TrimSpace(string(body))) == 0 {
		return Task{}, &PatchError{Msg: "request body required", Code: "bad_request"}
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return Task{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	if len(raw) == 0 {
		return base, nil
	}

	out := base
	out.UpdatedAt = time.Now().UTC()

	if v, ok := raw["title"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return Task{}, &PatchError{Msg: "title must be a string", Code: "validation_error"}
		}
		if utf8.RuneCountInString(s) > 200 {
			return Task{}, &PatchError{Msg: "title exceeds max length", Code: "validation_error"}
		}
		out.Title = s
	}

	if v, ok := raw["description"]; ok {
		if string(v) == "null" {
			out.Description = nil
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return Task{}, &PatchError{Msg: "description must be a string or null", Code: "validation_error"}
			}
			if utf8.RuneCountInString(s) > 10000 {
				return Task{}, &PatchError{Msg: "description exceeds max length", Code: "validation_error"}
			}
			out.Description = &s
		}
	}

	if v, ok := raw["status"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return Task{}, &PatchError{Msg: "status must be a string", Code: "validation_error"}
		}
		st := TaskStatus(s)
		switch st {
		case StatusTodo, StatusDoing, StatusDone:
			out.Status = st
		default:
			return Task{}, &PatchError{Msg: "status must be one of todo, doing, done", Code: "validation_error"}
		}
	}

	if v, ok := raw["priority"]; ok {
		if string(v) == "null" {
			out.Priority = nil
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return Task{}, &PatchError{Msg: "priority must be a string or null", Code: "validation_error"}
			}
			p := TaskPriority(s)
			switch p {
			case PriorityLow, PriorityMedium, PriorityHigh:
				out.Priority = &p
			default:
				return Task{}, &PatchError{Msg: "priority must be one of low, medium, high", Code: "validation_error"}
			}
		}
	}

	if v, ok := raw["due_date"]; ok {
		if string(v) == "null" {
			out.DueDate = nil
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return Task{}, &PatchError{Msg: "due_date must be a date string or null", Code: "validation_error"}
			}
			if !isISODate(s) {
				return Task{}, &PatchError{Msg: "due_date must be YYYY-MM-DD", Code: "validation_error"}
			}
			out.DueDate = &s
		}
	}

	if v, ok := raw["focus_bucket"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return Task{}, &PatchError{Msg: "focus_bucket must be a string", Code: "validation_error"}
		}
		fb := FocusBucket(s)
		switch fb {
		case FocusNone, FocusToday, FocusNext, FocusLater:
			out.FocusBucket = fb
		default:
			return Task{}, &PatchError{Msg: "focus_bucket must be one of none, today, next, later", Code: "validation_error"}
		}
	}

	if v, ok := raw["project_id"]; ok {
		if string(v) == "null" {
			out.ProjectID = nil
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return Task{}, &PatchError{Msg: "project_id must be a uuid string or null", Code: "validation_error"}
			}
			if _, err := uuid.Parse(s); err != nil {
				return Task{}, &PatchError{Msg: "project_id must be a valid uuid", Code: "validation_error"}
			}
			out.ProjectID = &s
		}
	}

	if v, ok := raw["assignee_id"]; ok {
		if string(v) == "null" {
			out.AssigneeID = nil
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return Task{}, &PatchError{Msg: "assignee_id must be a uuid string or null", Code: "validation_error"}
			}
			if _, err := uuid.Parse(s); err != nil {
				return Task{}, &PatchError{Msg: "assignee_id must be a valid uuid", Code: "validation_error"}
			}
			out.AssigneeID = &s
		}
	}

	return out, nil
}

func isISODate(s string) bool {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return false
	}
	return t.Format("2006-01-02") == s
}
