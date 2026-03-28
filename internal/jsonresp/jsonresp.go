// Package jsonresp writes consistent JSON bodies for API responses.
package jsonresp

import (
	"encoding/json"
	"net/http"
)

// Write sends a JSON response with the given HTTP status.
func Write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ErrorBody is the standard error JSON shape (see docs/api/openapi.yaml components/schemas/Error).
type ErrorBody struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// Error writes a JSON error response.
func Error(w http.ResponseWriter, status int, message, code string) {
	Write(w, status, ErrorBody{Error: message, Code: code})
}
