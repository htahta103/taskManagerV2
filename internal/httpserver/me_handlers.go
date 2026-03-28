package httpserver

import (
	"errors"
	"net/http"
	"strings"

	"github.com/htahta103/taskmanagerv2/internal/store"
)

type mePatchBody struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func (s *server) getMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	u, err := s.store.GetUser(r.Context(), uid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "user not found", "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load profile", "internal")
		return
	}
	writeJSON(w, http.StatusOK, formatUser(u))
}

func (s *server) patchMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var body mePatchBody
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "validation")
		return
	}
	var emailPtr *string
	var namePtr *string
	if body.Email != nil {
		v := strings.TrimSpace(strings.ToLower(*body.Email))
		if v == "" {
			writeError(w, http.StatusUnprocessableEntity, "email cannot be empty", "validation")
			return
		}
		emailPtr = &v
	}
	if body.Name != nil {
		v := strings.TrimSpace(*body.Name)
		if v == "" || len(v) > 120 {
			writeError(w, http.StatusUnprocessableEntity, "name must be 1-120 characters", "validation")
			return
		}
		namePtr = &v
	}
	if emailPtr == nil && namePtr == nil {
		writeError(w, http.StatusUnprocessableEntity, "no fields to update", "validation")
		return
	}
	u, err := s.store.UpdateUser(r.Context(), uid, emailPtr, namePtr)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "email already in use", "duplicate")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "user not found", "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update profile", "internal")
		return
	}
	writeJSON(w, http.StatusOK, formatUser(u))
}
