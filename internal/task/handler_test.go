package task

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type mockGetter struct {
	task Task
	err  error
}

func (m *mockGetter) GetTask(_ context.Context, _ string) (Task, error) {
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
	var body errorResponse
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
			ID:          id.String(),
			Title:       "hello",
			Status:      StatusTodo,
			FocusBucket: FocusNone,
			Tags:        []Tag{},
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
	if got.ID != id.String() || got.Title != "hello" {
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

func TestCreateTask_success(t *testing.T) {
	store := NewMemoryStore()
	h := &Handler{Store: store}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.Create(w, r)
	}))
	t.Cleanup(srv.Close)

	body := `{"title":"  Buy milk  "}`
	res, err := http.Post(srv.URL, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d want %d", res.StatusCode, http.StatusCreated)
	}
	var got Task
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Title != "Buy milk" {
		t.Fatalf("title %q want trimmed", got.Title)
	}
	if got.ID == "" || got.Status != StatusTodo || got.FocusBucket != FocusNone {
		t.Fatalf("defaults: %+v", got)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatal("timestamps missing")
	}
}

func TestCreateTask_titleRequired(t *testing.T) {
	h := &Handler{Store: NewMemoryStore()}
	req := httptest.NewRequest(http.MethodPost, "/functions/v1/tasks", strings.NewReader(`{"title":""}`))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code %d", rec.Code)
	}
	var er errorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &er)
	if !strings.Contains(er.Error, "title") {
		t.Fatalf("error %q", er.Error)
	}
}

func TestCreateTask_titleTooLong(t *testing.T) {
	long := strings.Repeat("é", 256) // 256 runes, multi-byte UTF-8
	if utf8.RuneCountInString(long) != 256 {
		t.Fatal("test setup: rune count")
	}
	payload, _ := json.Marshal(map[string]string{"title": long})
	h := &Handler{Store: NewMemoryStore()}
	req := httptest.NewRequest(http.MethodPost, "/functions/v1/tasks", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code %d body %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTask_invalidJSON(t *testing.T) {
	h := &Handler{Store: NewMemoryStore()}
	req := httptest.NewRequest(http.MethodPost, "/functions/v1/tasks", strings.NewReader(`{`))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code %d", rec.Code)
	}
}

func TestCreateTask_invalidStatus(t *testing.T) {
	h := &Handler{Store: NewMemoryStore()}
	body := `{"title":"x","status":"nope"}`
	req := httptest.NewRequest(http.MethodPost, "/functions/v1/tasks", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code %d", rec.Code)
	}
}

func TestValidateTitle_maxLengthBoundary(t *testing.T) {
	ok := strings.Repeat("a", 255)
	if err := ValidateTitle(ok); err != nil {
		t.Fatal(err)
	}
	bad := ok + "b"
	if err := ValidateTitle(bad); err == nil {
		t.Fatal("expected error")
	}
}
