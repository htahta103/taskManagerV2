package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// InsertStore persists new tasks.
type InsertStore interface {
	Insert(ctx context.Context, t *Task) error
}

// Handler serves task HTTP endpoints.
type Handler struct {
	Store InsertStore
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// Create handles POST /functions/v1/tasks.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeJSONError(w, http.StatusBadRequest, "request body too large", "body_too_large")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "could not read request body", "read_error")
		return
	}
	if len(raw) == 0 {
		writeJSONError(w, http.StatusBadRequest, "request body is required", "body_required")
		return
	}

	var payload CreatePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, jsonErrorMessage(err), "invalid_json")
		return
	}
	if err := ValidateCreatePayload(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error(), "validation")
		return
	}

	t := NewTaskFromPayload(&payload)
	if err := h.Store.Insert(r.Context(), t); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create task", "internal")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

// HandleGetOne serves GET /functions/v1/tasks/{id}.
func HandleGetOne(getter Getter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if getter == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "database not configured", "db_unavailable")
			return
		}
		idStr := strings.TrimSpace(r.PathValue("id"))
		if _, err := uuid.Parse(idStr); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid task id", "invalid_uuid")
			return
		}
		t, err := getter.GetTask(r.Context(), idStr)
		if errors.Is(err, ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "task not found", "not_found")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal server error", "internal")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(t)
	}
}

func writeJSONError(w http.ResponseWriter, status int, msg string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: msg, Code: code})
}

func jsonErrorMessage(err error) string {
	var se *json.SyntaxError
	if errors.As(err, &se) {
		return "malformed JSON body"
	}
	var te *json.UnmarshalTypeError
	if errors.As(err, &te) {
		if te.Field == "" {
			return fmt.Sprintf("invalid type for JSON value at offset %d", te.Offset)
		}
		return fmt.Sprintf("invalid type for field %q", te.Field)
	}
	return "invalid JSON body"
}
