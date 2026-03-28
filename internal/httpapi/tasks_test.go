package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/htahta103/taskmanagerv2/internal/tasks"
)

func TestHandleTaskDelete_unknownID(t *testing.T) {
	store := tasks.NewStore()
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /functions/v1/tasks/{id}", HandleTaskDelete(store))

	req := httptest.NewRequest(http.MethodDelete, "/functions/v1/tasks/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Body.String(); got != "{\"error\":\"task not found\",\"code\":\"not_found\"}\n" {
		t.Fatalf("body: got %q", got)
	}
}

func TestHandleTaskDelete_success(t *testing.T) {
	const id = "550e8400-e29b-41d4-a716-446655440000"
	store := tasks.NewStore()
	store.Put(id)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /functions/v1/tasks/{id}", HandleTaskDelete(store))

	req := httptest.NewRequest(http.MethodDelete, "/functions/v1/tasks/"+id, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}
	want := "{\"message\":\"deleted\",\"id\":\"" + id + "\"}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("body: got %q want %q", got, want)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type: got %q", ct)
	}
}
