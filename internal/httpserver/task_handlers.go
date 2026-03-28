package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
		writeError(w, http.StatusUnprocessableEntity, "query parameter q is required", "validation")
		return
	}
	var qPtr *string
	if q != "" {
		qPtr = &q
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	var cur *store.PageCursor
	if cs := r.URL.Query().Get("cursor"); cs != "" {
		c, err := store.DecodePageCursor(cs)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "invalid cursor", "validation")
			return
		}
		cur = &c
	}
	var projectID *uuid.UUID
	if v := r.URL.Query().Get("project_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "invalid project_id", "validation")
			return
		}
		projectID = &id
	}
	var status *string
	if v := r.URL.Query().Get("status"); v != "" {
		status = &v
	}
	var view *string
	if v := r.URL.Query().Get("view"); v != "" {
		view = &v
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
			writeError(w, http.StatusUnprocessableEntity, "invalid list parameters", "validation")
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
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "validation")
		return
	}
	due, err := parseDueDate(body.DueDate)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid due_date (use YYYY-MM-DD)", "validation")
		return
	}
	var projectID *uuid.UUID
	if body.ProjectID != nil && strings.TrimSpace(*body.ProjectID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*body.ProjectID))
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "invalid project_id", "validation")
			return
		}
		projectID = &id
	}
	var assigneeID *uuid.UUID
	if body.AssigneeID != nil && strings.TrimSpace(*body.AssigneeID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*body.AssigneeID))
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "invalid assignee_id", "validation")
			return
		}
		assigneeID = &id
	}
	tagIDs := make([]uuid.UUID, 0, len(body.TagIDs))
	for _, raw := range body.TagIDs {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "invalid tag id", "validation")
			return
		}
		tagIDs = append(tagIDs, id)
	}
	t, err := s.store.CreateTask(r.Context(), uid, store.TaskCreateInput{
		Title:       body.Title,
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
			writeError(w, http.StatusUnprocessableEntity, "invalid task payload", "validation")
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
		writeError(w, http.StatusUnprocessableEntity, "invalid task id", "validation")
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
		writeError(w, http.StatusUnprocessableEntity, "invalid task id", "validation")
		return
	}
	var m map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "validation")
		return
	}
	if len(m) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "no fields to update", "validation")
		return
	}
	var patch store.TaskUpdateValues
	for k, raw := range m {
		switch k {
		case "title":
			patch.SetTitle = true
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeError(w, http.StatusUnprocessableEntity, "invalid title", "validation")
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
				writeError(w, http.StatusUnprocessableEntity, "invalid description", "validation")
				return
			}
			patch.Description = &v
		case "status":
			patch.SetStatus = true
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeError(w, http.StatusUnprocessableEntity, "invalid status", "validation")
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
				writeError(w, http.StatusUnprocessableEntity, "invalid priority", "validation")
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
				writeError(w, http.StatusUnprocessableEntity, "invalid due_date", "validation")
				return
			}
			d, err := time.Parse("2006-01-02", strings.TrimSpace(v))
			if err != nil {
				writeError(w, http.StatusUnprocessableEntity, "invalid due_date (use YYYY-MM-DD)", "validation")
				return
			}
			utc := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
			patch.DueDate = &utc
		case "focus_bucket":
			patch.SetFocusBucket = true
			var v string
			if err := json.Unmarshal(raw, &v); err != nil {
				writeError(w, http.StatusUnprocessableEntity, "invalid focus_bucket", "validation")
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
				writeError(w, http.StatusUnprocessableEntity, "invalid project_id", "validation")
				return
			}
			id, err := uuid.Parse(strings.TrimSpace(v))
			if err != nil {
				writeError(w, http.StatusUnprocessableEntity, "invalid project_id", "validation")
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
				writeError(w, http.StatusUnprocessableEntity, "invalid assignee_id", "validation")
				return
			}
			id, err := uuid.Parse(strings.TrimSpace(v))
			if err != nil {
				writeError(w, http.StatusUnprocessableEntity, "invalid assignee_id", "validation")
				return
			}
			patch.AssigneeID = &id
		default:
			writeError(w, http.StatusUnprocessableEntity, "unknown field: "+k, "validation")
			return
		}
	}
	t, err := s.store.UpdateTask(r.Context(), uid, tid, patch)
	if err != nil {
		if errors.Is(err, store.ErrInvalidInput) {
			writeError(w, http.StatusUnprocessableEntity, "invalid task update", "validation")
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
		writeError(w, http.StatusUnprocessableEntity, "invalid task id", "validation")
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
