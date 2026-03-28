package httpserver

import (
	"errors"
	"net/http"

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
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "validation")
		return
	}
	p, err := s.store.CreateProject(r.Context(), uid, body.Name)
	if err != nil {
		if errors.Is(err, store.ErrInvalidInput) {
			writeError(w, http.StatusUnprocessableEntity, "invalid project name", "validation")
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
		writeError(w, http.StatusUnprocessableEntity, "invalid project id", "validation")
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
		writeError(w, http.StatusUnprocessableEntity, "invalid project id", "validation")
		return
	}
	var body projectPatchBody
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "validation")
		return
	}
	if body.Name == nil && body.Archived == nil {
		writeError(w, http.StatusUnprocessableEntity, "no fields to update", "validation")
		return
	}
	p, err := s.store.UpdateProject(r.Context(), uid, pid, body.Name, body.Archived)
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
		writeError(w, http.StatusUnprocessableEntity, "invalid project id", "validation")
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
