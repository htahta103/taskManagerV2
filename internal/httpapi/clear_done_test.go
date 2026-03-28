package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/htahta103/taskmanagerv2/internal/tasks"
)

func TestHandleClearDone_wrongMethod(t *testing.T) {
	store := tasks.NewStore()
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /functions/v1/tasks/clear/done", HandleClearDone(store))

	req := httptest.NewRequest(http.MethodGet, "/functions/v1/tasks/clear/done", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleClearDone_empty(t *testing.T) {
	store := tasks.NewStore()
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /functions/v1/tasks/clear/done", HandleClearDone(store))

	req := httptest.NewRequest(http.MethodDelete, "/functions/v1/tasks/clear/done", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}
	want := "{\"deleted_count\":0}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("body: got %q want %q", got, want)
	}
}

func TestHandleClearDone_deletesOnlyDone(t *testing.T) {
	const (
		doneA = "550e8400-e29b-41d4-a716-446655440001"
		doneB = "550e8400-e29b-41d4-a716-446655440002"
		todoC = "550e8400-e29b-41d4-a716-446655440003"
	)
	store := tasks.NewStore()
	store.Put(doneA)
	store.Put(doneB)
	store.Put(todoC)
	_ = store.SetStatus(doneA, tasks.StatusDone)
	_ = store.SetStatus(doneB, tasks.StatusDone)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /functions/v1/tasks/clear/done", HandleClearDone(store))

	req := httptest.NewRequest(http.MethodDelete, "/functions/v1/tasks/clear/done", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}
	want := "{\"deleted_count\":2}\n"
	if got := rec.Body.String(); got != want {
		t.Fatalf("body: got %q want %q", got, want)
	}

	if !store.SetStatus(todoC, tasks.StatusDoing) {
		t.Fatal("todo task should still exist")
	}
}
