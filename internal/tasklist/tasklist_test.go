package tasklist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListHandler_all(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()
	ListHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var p page
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != len(allTasks) {
		t.Fatalf("want %d items, got %d", len(allTasks), len(p.Items))
	}
}

func TestListHandler_statusTodo(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?status=todo", nil)
	rec := httptest.NewRecorder()
	ListHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var p page
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	for _, it := range p.Items {
		if it.Status != "todo" {
			t.Fatalf("non-todo item: %+v", it)
		}
	}
}

func TestListHandler_queryTitle(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?q=dashboard", nil)
	rec := httptest.NewRecorder()
	ListHandler(rec, req)
	var p page
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 1 || p.Items[0].Title != "Plan dashboard filters" {
		t.Fatalf("unexpected items: %+v", p.Items)
	}
}
