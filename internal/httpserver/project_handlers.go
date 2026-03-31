package httpserver

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/htahta103/taskmanagerv2/internal/store"
)

type projectCreateBody struct {
	Name string `json:"name"`
}

type projectPatchBody struct {
	Name     *string `json:"name"`
	Archived *bool   `json:"archived"`
}

func (s *server) listProjects(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	include := r.URL.Query().Get("include_archived") == "true"
	list, err := s.store.ListProjects(r.Context(), uid, include)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list projects", "internal")
		return
	}
	items := make([]any, 0, len(list))
	for _, p := range list {
		items = append(items, formatProject(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *server) createProject(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var body projectCreateBody
	if err := readJSON(r, &body); err != nil {
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON or unknown fields"})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeValidation(w, "validation failed", map[string]string{"name": "required"})
		return
	}
	if len(name) > 200 {
		writeValidation(w, "validation failed", map[string]string{"name": "must be at most 200 characters"})
		return
	}
	p, err := s.store.CreateProject(r.Context(), uid, name)
	if err != nil {
		if errors.Is(err, store.ErrInvalidInput) {
			writeValidation(w, "invalid project name", map[string]string{"name": "failed validation"})
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create project", "internal")
		return
	}
	writeJSON(w, http.StatusCreated, formatProject(p))
}

func (s *server) getProject(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	pid, err := uuid.Parse(r.PathValue("projectId"))
	if err != nil {
		writeValidation(w, "invalid path", map[string]string{"projectId": "must be a UUID"})
		return
	}
	p, err := s.store.GetProject(r.Context(), uid, pid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "project not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load project", "internal")
		return
	}
	writeJSON(w, http.StatusOK, formatProject(p))
}

func (s *server) patchProject(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	pid, err := uuid.Parse(r.PathValue("projectId"))
	if err != nil {
		writeValidation(w, "invalid path", map[string]string{"projectId": "must be a UUID"})
		return
	}
	var body projectPatchBody
	if err := readJSON(r, &body); err != nil {
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON or unknown fields"})
		return
	}
	if body.Name == nil && body.Archived == nil {
		writeValidation(w, "no fields to update", map[string]string{"body": "at least one of name or archived is required"})
		return
	}
	var namePtr *string
	if body.Name != nil {
		v := strings.TrimSpace(*body.Name)
		if v == "" {
			writeValidation(w, "validation failed", map[string]string{"name": "cannot be empty"})
			return
		}
		if len(v) > 200 {
			writeValidation(w, "validation failed", map[string]string{"name": "must be at most 200 characters"})
			return
		}
		namePtr = &v
	}
	p, err := s.store.UpdateProject(r.Context(), uid, pid, namePtr, body.Archived)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "project not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update project", "internal")
		return
	}
	writeJSON(w, http.StatusOK, formatProject(p))
}

func (s *server) deleteProject(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	pid, err := uuid.Parse(r.PathValue("projectId"))
	if err != nil {
		writeValidation(w, "invalid path", map[string]string{"projectId": "must be a UUID"})
		return
	}
	if err := s.store.DeleteProject(r.Context(), uid, pid); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "project not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete project", "internal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
