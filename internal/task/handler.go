package task

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// ErrorBody matches the OpenAPI error envelope (minimal fields).
type ErrorBody struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// HandleGetOne serves GET /functions/v1/tasks/{id}.
func HandleGetOne(getter Getter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if getter == nil {
			writeErr(w, http.StatusServiceUnavailable, "database not configured", "db_unavailable")
			return
		}
		idStr := r.PathValue("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid task id", "invalid_uuid")
			return
		}

		t, err := getter.GetTask(r.Context(), id)
		if errors.Is(err, ErrNotFound) {
			writeErr(w, http.StatusNotFound, "task not found", "not_found")
			return
		}
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal server error", "internal")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(t)
	}
}

func writeErr(w http.ResponseWriter, status int, msg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorBody{Error: msg, Code: code})
}
