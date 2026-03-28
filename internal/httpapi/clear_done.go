package httpapi

import (
	"net/http"

	"github.com/htahta103/taskmanagerv2/internal/jsonresp"
	"github.com/htahta103/taskmanagerv2/internal/tasks"
)

// ClearDoneResponse is returned for successful DELETE /functions/v1/tasks/clear/done.
type ClearDoneResponse struct {
	DeletedCount int `json:"deleted_count"`
}

// HandleClearDone handles DELETE /functions/v1/tasks/clear/done.
func HandleClearDone(store *tasks.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			jsonresp.Error(w, http.StatusMethodNotAllowed, "method not allowed", "method_not_allowed")
			return
		}
		n := store.ClearDone()
		jsonresp.Write(w, http.StatusOK, ClearDoneResponse{DeletedCount: n})
	}
}
