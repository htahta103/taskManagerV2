package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/htahta103/taskmanagerv2/internal/store"
)

type taskCreateBody struct {
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	Status      *string  `json:"status"`
	Priority    *string  `json:"priority"`
	DueDate     *string  `json:"due_date"`
	FocusBucket *string  `json:"focus_bucket"`
	ProjectID   *string  `json:"project_id"`
	AssigneeID  *string  `json:"assignee_id"`
	TagIDs      []string `json:"tag_ids"`
}

func parseDueDate(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, err
	}
	utc := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return &utc, nil
}

func (s *server) listTasks(w http.ResponseWriter, r *http.Request) {
	s.listTasksQuery(w, r, false)
}

func (s *server) listTasksQuery(w http.ResponseWriter, r *http.Request, search bool) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if search && q == "" {
		writeValidation(w, "query parameter q is required", map[string]string{"q": "required"})
		return
	}
	var qPtr *string
	if q != "" {
		qPtr = &q
	}
	limit, okL, limFields := parseLimit(r.URL.Query().Get("limit"), 50)
	if !okL {
		writeLimitError(w, limFields)
		return
	}
	var cur *store.PageCursor
	if cs := r.URL.Query().Get("cursor"); cs != "" {
		c, err := store.DecodePageCursor(cs)
		if err != nil {
			writeValidation(w, "invalid cursor", map[string]string{"cursor": "invalid or expired"})
			return
		}
		cur = &c
	}
	var projectID *uuid.UUID
	if v := r.URL.Query().Get("project_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeValidation(w, "invalid query parameters", map[string]string{"project_id": "must be a UUID"})
			return
		}
		projectID = &id
	}
	var status *string
	if v := r.URL.Query().Get("status"); v != "" {
		st, ok := normalizeTaskStatus(v)
		if !ok {
			writeValidation(w, "invalid query parameters", map[string]string{"status": "must be todo, doing, or done"})
			return
		}
		status = &st
	}
	var view *string
	if v := r.URL.Query().Get("view"); v != "" {
		vi, ok := normalizeTaskView(v)
		if !ok {
			writeValidation(w, "invalid query parameters", map[string]string{"view": "must be inbox, today, next, or later"})
			return
		}
		view = &vi
	}
	tasks, next, err := s.store.ListTasks(r.Context(), store.ListTasksParams{
		UserID:    uid,
		ProjectID: projectID,
		Status:    status,
		View:      view,
		Q:         qPtr,
		Limit:     limit,
		Cursor:    cur,
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidInput) {
			writeValidation(w, "invalid list parameters", map[string]string{"_": "one or more filters are invalid"})
			return
		}
		writeError(w, http.StatusInternalServerError, "could not list tasks", "internal")
		return
	}
	ids := make([]uuid.UUID, len(tasks))
	for i := range tasks {
		ids[i] = tasks[i].ID
	}
	tagMap, err := s.store.TagsForTaskIDs(r.Context(), uid, ids)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load task tags", "internal")
		return
	}
	items := make([]any, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, formatTask(t, tagMap[t.ID]))
	}
	body := map[string]any{"items": items}
	if next != nil {
		body["next_cursor"] = *next
	} else {
		body["next_cursor"] = nil
	}
	writeJSON(w, http.StatusOK, body)
}

func (s *server) createTask(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var body taskCreateBody
	if err := readJSON(r, &body); err != nil {
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON or unknown fields"})
		return
	}
	if fields := validateDescriptionPtr(body.Description); fields != nil {
		writeValidation(w, "validation failed", fields)
		return
	}
	title, fields := validateTaskTitle(body.Title)
	if fields != nil {
		writeValidation(w, "validation failed", fields)
		return
	}
	if body.Status != nil {
		st, ok := normalizeTaskStatus(*body.Status)
		if !ok {
			writeValidation(w, "validation failed", map[string]string{"status": "must be todo, doing, or done"})
			return
		}
		body.Status = &st
	}
	prio, ok := normalizePriorityPtr(body.Priority)
	if !ok {
		writeValidation(w, "validation failed", map[string]string{"priority": "must be low, medium, or high"})
		return
	}
	body.Priority = prio
	fb, ok := normalizeFocusBucketPtr(body.FocusBucket)
	if !ok {
		writeValidation(w, "validation failed", map[string]string{"focus_bucket": "must be none, today, next, or later"})
		return
	}
	body.FocusBucket = fb
	due, err := parseDueDate(body.DueDate)
	if err != nil {
		writeValidation(w, "invalid due_date", map[string]string{"due_date": "use ISO date YYYY-MM-DD"})
		return
	}
	var projectID *uuid.UUID
	if body.ProjectID != nil && strings.TrimSpace(*body.ProjectID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*body.ProjectID))
		if err != nil {
			writeValidation(w, "validation failed", map[string]string{"project_id": "must be a UUID"})
			return
		}
		projectID = &id
	}
	var assigneeID *uuid.UUID
	if body.AssigneeID != nil && strings.TrimSpace(*body.AssigneeID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*body.AssigneeID))
		if err != nil {
			writeValidation(w, "validation failed", map[string]string{"assignee_id": "must be a UUID"})
			return
		}
		assigneeID = &id
	}
	tagIDs := make([]uuid.UUID, 0, len(body.TagIDs))
	for _, raw := range body.TagIDs {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			writeValidation(w, "validation failed", map[string]string{"tag_ids": "each tag id must be a UUID"})
			return
		}
		tagIDs = append(tagIDs, id)
	}
	t, err := s.store.CreateTask(r.Context(), uid, store.TaskCreateInput{
		Title:       title,
		Description: body.Description,
		Status:      body.Status,
		Priority:    body.Priority,
		DueDate:     due,
		FocusBucket: body.FocusBucket,
		ProjectID:   projectID,
		AssigneeID:  assigneeID,
		TagIDs:      tagIDs,
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidInput) {
			writeValidation(w, "invalid task payload", map[string]string{"_": "one or more fields failed validation"})
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "project or tag not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create task", "internal")
		return
	}
	tags, err := s.store.TagsForTaskIDs(r.Context(), uid, []uuid.UUID{t.ID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load task tags", "internal")
		return
	}
	writeJSON(w, http.StatusCreated, formatTask(t, tags[t.ID]))
}

func (s *server) getTask(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	tid, err := uuid.Parse(r.PathValue("taskId"))
	if err != nil {
		writeValidation(w, "invalid path", map[string]string{"taskId": "must be a UUID"})
		return
	}
	t, err := s.store.GetTask(r.Context(), uid, tid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load task", "internal")
		return
	}
	tags, err := s.store.TagsForTaskIDs(r.Context(), uid, []uuid.UUID{t.ID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load task tags", "internal")
		return
	}
	writeJSON(w, http.StatusOK, formatTask(t, tags[t.ID]))
}

func (s *server) patchTask(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	tid, err := uuid.Parse(r.PathValue("taskId"))
	if err != nil {
		writeValidation(w, "invalid path", map[string]string{"taskId": "must be a UUID"})
		return
	}
	var m map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON"})
		return
	}
	if len(m) == 0 {
		writeValidation(w, "no fields to update", map[string]string{"body": "at least one field is required"})
		return
	}
	var patch store.TaskUpdateValues
	for k, raw := range m {
		switch k {
		case "title":
			patch.SetTitle = true
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeValidation(w, "validation failed", map[string]string{"title": "must be a string"})
				return
			}
			patch.Title = v
		case "description":
			patch.SetDescription = true
			if string(raw) == "null" {
				patch.Description = nil
				break
			}
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeValidation(w, "validation failed", map[string]string{"description": "must be a string or null"})
				return
			}
			patch.Description = &v
		case "status":
			patch.SetStatus = true
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeValidation(w, "validation failed", map[string]string{"status": "must be a string"})
				return
			}
			patch.Status = v
		case "priority":
			patch.SetPriority = true
			if string(raw) == "null" {
				patch.Priority = nil
				break
			}
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeValidation(w, "validation failed", map[string]string{"priority": "must be a string or null"})
				return
			}
			patch.Priority = &v
		case "due_date":
			patch.SetDueDate = true
			if string(raw) == "null" {
				patch.DueDate = nil
				break
			}
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeValidation(w, "validation failed", map[string]string{"due_date": "must be a string or null"})
				return
			}
			d, err := time.Parse("2006-01-02", strings.TrimSpace(v))
			if err != nil {
				writeValidation(w, "invalid due_date", map[string]string{"due_date": "use ISO date YYYY-MM-DD"})
				return
			}
			utc := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
			patch.DueDate = &utc
		case "focus_bucket":
			patch.SetFocusBucket = true
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeValidation(w, "validation failed", map[string]string{"focus_bucket": "must be a string"})
				return
			}
			patch.FocusBucket = v
		case "project_id":
			patch.SetProjectID = true
			if string(raw) == "null" {
				patch.ProjectID = nil
				break
			}
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeValidation(w, "validation failed", map[string]string{"project_id": "must be a string UUID or null"})
				return
			}
			id, err := uuid.Parse(strings.TrimSpace(v))
			if err != nil {
				writeValidation(w, "validation failed", map[string]string{"project_id": "must be a UUID"})
				return
			}
			patch.ProjectID = &id
		case "assignee_id":
			patch.SetAssigneeID = true
			if string(raw) == "null" {
				patch.AssigneeID = nil
				break
			}
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeValidation(w, "validation failed", map[string]string{"assignee_id": "must be a string UUID or null"})
				return
			}
			id, err := uuid.Parse(strings.TrimSpace(v))
			if err != nil {
				writeValidation(w, "validation failed", map[string]string{"assignee_id": "must be a UUID"})
				return
			}
			patch.AssigneeID = &id
		default:
			writeValidation(w, "unknown field in request body", map[string]string{k: "not supported on PATCH"})
			return
		}
	}
	if patch.SetTitle {
		tt, fields := validateTaskTitle(patch.Title)
		if fields != nil {
			writeValidation(w, "validation failed", fields)
			return
		}
		patch.Title = tt
	}
	if patch.SetDescription && patch.Description != nil {
		if fields := validateDescriptionPtr(patch.Description); fields != nil {
			writeValidation(w, "validation failed", fields)
			return
		}
	}
	if patch.SetStatus {
		st, ok := normalizeTaskStatus(patch.Status)
		if !ok {
			writeValidation(w, "validation failed", map[string]string{"status": "must be todo, doing, or done"})
			return
		}
		patch.Status = st
	}
	if patch.SetPriority {
		pr, ok := normalizePriorityPtr(patch.Priority)
		if !ok {
			writeValidation(w, "validation failed", map[string]string{"priority": "must be low, medium, or high"})
			return
		}
		patch.Priority = pr
	}
	if patch.SetFocusBucket {
		fbk := patch.FocusBucket
		fb, ok := normalizeFocusBucketPtr(&fbk)
		if !ok {
			writeValidation(w, "validation failed", map[string]string{"focus_bucket": "must be none, today, next, or later"})
			return
		}
		patch.FocusBucket = *fb
	}
	t, err := s.store.UpdateTask(r.Context(), uid, tid, patch)
	if err != nil {
		if errors.Is(err, store.ErrInvalidInput) {
			writeValidation(w, "invalid task update", map[string]string{"_": "one or more fields failed validation"})
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task or project not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update task", "internal")
		return
	}
	tags, err := s.store.TagsForTaskIDs(r.Context(), uid, []uuid.UUID{t.ID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load task tags", "internal")
		return
	}
	writeJSON(w, http.StatusOK, formatTask(t, tags[t.ID]))
}

func (s *server) deleteTask(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	tid, err := uuid.Parse(r.PathValue("taskId"))
	if err != nil {
		writeValidation(w, "invalid path", map[string]string{"taskId": "must be a UUID"})
		return
	}
	if err := s.store.DeleteTask(r.Context(), uid, tid); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete task", "internal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
