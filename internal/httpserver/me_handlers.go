package httpserver

import (
	"errors"
	"net/http"
	"strings"

	"github.com/htahta103/taskmanagerv2/internal/store"
)

func validateEmailFormat(email string) bool {
	if len(email) < 3 || len(email) > 254 || !strings.Contains(email, "@") {
		return false
	}
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return false
	}
	return true
}

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
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON or unknown fields"})
		return
	}
	var emailPtr *string
	var namePtr *string
	if body.Email != nil {
		v := strings.TrimSpace(strings.ToLower(*body.Email))
		if v == "" {
			writeValidation(w, "validation failed", map[string]string{"email": "cannot be empty"})
			return
		}
		if !validateEmailFormat(v) {
			writeValidation(w, "validation failed", map[string]string{"email": "must be a valid email address"})
			return
		}
		emailPtr = &v
	}
	if body.Name != nil {
		v := strings.TrimSpace(*body.Name)
		if v == "" || len(v) > 120 {
			writeValidation(w, "validation failed", map[string]string{"name": "must be 1-120 characters"})
			return
		}
		namePtr = &v
	}
	if emailPtr == nil && namePtr == nil {
		writeValidation(w, "no fields to update", map[string]string{"body": "at least one of name or email is required"})
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
