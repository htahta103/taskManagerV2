package httpapi

import (
	"net/http"

	"github.com/htahta103/taskmanagerv2/internal/jsonresp"
	"github.com/htahta103/taskmanagerv2/internal/tasks"
)

// TaskDeleteResponse is returned for successful DELETE /functions/v1/tasks/{id}.
type TaskDeleteResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

// HandleTaskDelete handles DELETE /functions/v1/tasks/{id}.
func HandleTaskDelete(store *tasks.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			jsonresp.Error(w, http.StatusMethodNotAllowed, "method not allowed", "method_not_allowed")
			return
		}
		id := r.PathValue("id")
		if id == "" {
			jsonresp.Error(w, http.StatusBadRequest, "missing task id", "bad_request")
			return
		}
		if !store.Delete(id) {
			jsonresp.Error(w, http.StatusNotFound, "task not found", "not_found")
			return
		}
		jsonresp.Write(w, http.StatusOK, TaskDeleteResponse{
			Message: "deleted",
			ID:      id,
		})
	}
}
