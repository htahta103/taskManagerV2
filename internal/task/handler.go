package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
			writeJSONError(w, http.StatusBadRequest, "request body too large")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "could not read request body")
		return
	}
	if len(raw) == 0 {
		writeJSONError(w, http.StatusBadRequest, "request body is required")
		return
	}

	var payload CreatePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, jsonErrorMessage(err))
		return
	}
	if err := ValidateCreatePayload(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	t := NewTaskFromPayload(&payload)
	if err := h.Store.Insert(r.Context(), t); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: msg})
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
