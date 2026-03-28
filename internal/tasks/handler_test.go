package tasks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHandlerPatch_successPartial(t *testing.T) {
	id := uuid.New()
	created := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	base := Task{
		ID:          id.String(),
		Title:       "Original",
		Status:      StatusTodo,
		FocusBucket: FocusNone,
		CreatedAt:   created,
		UpdatedAt:   created,
		Tags:        []Tag{},
	}
	store := NewMemoryStore()
	if err := store.Put(base); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(store)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+id.String(), bytes.NewBufferString(`{"status":"doing"}`))
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()
	h.Patch(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	var got Task
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusDoing {
		t.Errorf("status = %q, want doing", got.Status)
	}
	if got.Title != "Original" {
		t.Errorf("title = %q, want Original", got.Title)
	}
	if !got.UpdatedAt.After(created) {
		t.Errorf("updated_at should advance")
	}
}

func TestHandlerPatch_notFound(t *testing.T) {
	store := NewMemoryStore()
	h := NewHandler(store)
	id := uuid.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+id.String(), bytes.NewBufferString(`{"status":"done"}`))
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()
	h.Patch(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestHandlerPatch_invalidUUID(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/not-a-uuid", bytes.NewBufferString(`{}`))
	req.SetPathValue("id", "not-a-uuid")
	rec := httptest.NewRecorder()
	h.Patch(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerPatch_invalidStatus(t *testing.T) {
	id := uuid.New()
	base := Task{
		ID:          id.String(),
		Title:       "T",
		Status:      StatusTodo,
		FocusBucket: FocusNone,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	store := NewMemoryStore()
	_ = store.Put(base)
	h := NewHandler(store)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+id.String(), bytes.NewBufferString(`{"status":"invalid"}`))
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()
	h.Patch(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestHandlerPatch_emptyBody(t *testing.T) {
	id := uuid.New()
	base := Task{
		ID:          id.String(),
		Title:       "T",
		Status:      StatusTodo,
		FocusBucket: FocusNone,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	store := NewMemoryStore()
	_ = store.Put(base)
	h := NewHandler(store)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+id.String(), bytes.NewBufferString(``))
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()
	h.Patch(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestHandlerPatch_noopObject(t *testing.T) {
	id := uuid.New()
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	base := Task{
		ID:          id.String(),
		Title:       "T",
		Status:      StatusTodo,
		FocusBucket: FocusNone,
		CreatedAt:   created,
		UpdatedAt:   created,
	}
	store := NewMemoryStore()
	_ = store.Put(base)
	h := NewHandler(store)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+id.String(), bytes.NewBufferString(`{}`))
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()
	h.Patch(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var got Task
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if !got.UpdatedAt.Equal(created) {
		t.Errorf("noop patch should not change updated_at: got %v want %v", got.UpdatedAt, created)
	}
}
