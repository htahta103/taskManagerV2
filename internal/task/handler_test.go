package task

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockGetter struct {
	task Task
	err  error
}

func (m *mockGetter) GetTask(_ context.Context, _ uuid.UUID) (Task, error) {
	return m.task, m.err
}

func TestHandleGetOne_invalidUUID(t *testing.T) {
	m := &mockGetter{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := r.Clone(context.Background())
		r2.SetPathValue("id", "not-a-uuid")
		HandleGetOne(m).ServeHTTP(w, r2)
	}))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/x")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
	var body ErrorBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != "invalid_uuid" {
		t.Fatalf("code = %q", body.Code)
	}
}

func TestHandleGetOne_notFound(t *testing.T) {
	m := &mockGetter{err: ErrNotFound}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := r.Clone(context.Background())
		r2.SetPathValue("id", uuid.NewString())
		HandleGetOne(m).ServeHTTP(w, r2)
	}))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/x")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
}

func TestHandleGetOne_ok(t *testing.T) {
	id := uuid.New()
	ts := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	m := &mockGetter{
		task: Task{
			ID:          id,
			Title:       "hello",
			Status:      "todo",
			FocusBucket: "none",
			Tags:        []string{},
			CreatedAt:   ts,
			UpdatedAt:   ts,
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := r.Clone(context.Background())
		r2.SetPathValue("id", id.String())
		HandleGetOne(m).ServeHTTP(w, r2)
	}))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/x")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	var got Task
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID != id || got.Title != "hello" {
		t.Fatalf("got %+v", got)
	}
}

func TestHandleGetOne_noDatabase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := r.Clone(context.Background())
		r2.SetPathValue("id", uuid.NewString())
		HandleGetOne(nil).ServeHTTP(w, r2)
	}))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/x")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", res.StatusCode)
	}
}
