package httpserver

import (
	"net/http"
)

// writeError writes a JSON error body with optional machine-readable code (OpenAPI Error).
func writeError(w http.ResponseWriter, status int, message, code string) {
	writeAPIError(w, status, message, code, nil)
}

// writeValidation sends 422 with a consistent envelope; details typically include a "fields" map (field → message).
func writeValidation(w http.ResponseWriter, message string, fields map[string]string) {
	details := map[string]any{}
	if len(fields) > 0 {
		details["fields"] = fields
	}
	writeAPIError(w, http.StatusUnprocessableEntity, message, "validation", details)
}

func writeAPIError(w http.ResponseWriter, status int, message, code string, details map[string]any) {
	body := map[string]any{"error": message}
	if code != "" {
		body["code"] = code
	}
	if len(details) > 0 {
		body["details"] = details
	}
	writeJSON(w, status, body)
}
