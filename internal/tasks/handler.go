package tasks

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// Handler serves task HTTP endpoints.
type Handler struct {
	Store *MemoryStore
}

// NewHandler returns a Handler using the given store.
func NewHandler(s *MemoryStore) *Handler {
	return &Handler{Store: s}
}

// Patch handles PATCH /.../tasks/{id} (partial update).
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.PathValue("id")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "missing task id", "bad_request")
		return
	}
	taskID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id", "bad_request")
		return
	}

	base, ok := h.Store.Get(taskID)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found", "not_found")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read body", "bad_request")
		return
	}
	_ = r.Body.Close()

	updated, err := ApplyPatch(base, body)
	if err != nil {
		var perr *PatchError
		if errors.As(err, &perr) {
			writeError(w, http.StatusBadRequest, perr.Msg, perr.Code)
			return
		}
		if errors.Is(err, ErrInvalidJSON) {
			writeError(w, http.StatusBadRequest, "invalid json", "bad_request")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	h.Store.Replace(updated)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(updated)
}

func writeError(w http.ResponseWriter, status int, msg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": msg,
		"code":  code,
	})
}
