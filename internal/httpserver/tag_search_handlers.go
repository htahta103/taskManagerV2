package httpserver

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/htahta103/taskmanagerv2/internal/store"
)

type taskTagsBody struct {
	TagIDs []string `json:"tag_ids"`
}

type tagCreateBody struct {
	Name string `json:"name"`
}

func (s *server) postTaskTags(w http.ResponseWriter, r *http.Request) {
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
	var body taskTagsBody
	if err := readJSON(r, &body); err != nil {
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON or unknown fields"})
		return
	}
	ids := make([]uuid.UUID, 0, len(body.TagIDs))
	for _, raw := range body.TagIDs {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			writeValidation(w, "validation failed", map[string]string{"tag_ids": "each tag id must be a UUID"})
			return
		}
		ids = append(ids, id)
	}
	if err := s.store.AddTaskTags(r.Context(), uid, tid, ids); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task or tag not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not attach tags", "internal")
		return
	}
	t, err := s.store.GetTask(r.Context(), uid, tid)
	if err != nil {
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

func (s *server) deleteTaskTag(w http.ResponseWriter, r *http.Request) {
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
	tagID, err := uuid.Parse(r.PathValue("tagId"))
	if err != nil {
		writeValidation(w, "invalid path", map[string]string{"tagId": "must be a UUID"})
		return
	}
	if err := s.store.RemoveTaskTag(r.Context(), uid, tid, tagID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "association not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not remove tag", "internal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) listTags(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
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
	list, err := s.store.ListTags(r.Context(), uid, limit, cur)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list tags", "internal")
		return
	}
	items := make([]any, 0, len(list))
	for _, t := range list {
		items = append(items, formatTag(t))
	}
	body := map[string]any{"items": items}
	if len(list) > 0 {
		last := list[len(list)-1]
		next, err := store.EncodeTagCursor(last.CreatedAt, last.ID)
		if err == nil && len(list) == limit {
			body["next_cursor"] = next
		} else {
			body["next_cursor"] = nil
		}
	} else {
		body["next_cursor"] = nil
	}
	writeJSON(w, http.StatusOK, body)
}

func (s *server) createTag(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var body tagCreateBody
	if err := readJSON(r, &body); err != nil {
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON or unknown fields"})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeValidation(w, "validation failed", map[string]string{"name": "required"})
		return
	}
	if len(name) > 64 {
		writeValidation(w, "validation failed", map[string]string{"name": "must be at most 64 characters"})
		return
	}
	t, err := s.store.CreateTag(r.Context(), uid, name)
	if err != nil {
		if errors.Is(err, store.ErrInvalidInput) {
			writeValidation(w, "invalid tag name", map[string]string{"name": "failed validation"})
			return
		}
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "duplicate tag name", "duplicate")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create tag", "internal")
		return
	}
	writeJSON(w, http.StatusCreated, formatTag(t))
}

func (s *server) getSearch(w http.ResponseWriter, r *http.Request) {
	s.listTasksQuery(w, r, true)
}
