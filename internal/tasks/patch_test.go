package tasks

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestApplyPatch_descriptionNullClears(t *testing.T) {
	desc := "x"
	base := Task{
		ID:          uuid.New().String(),
		Title:       "t",
		Description: &desc,
		Status:      StatusTodo,
		FocusBucket: FocusNone,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	out, err := ApplyPatch(base, []byte(`{"description":null}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Description != nil {
		t.Fatalf("expected nil description")
	}
}

func TestApplyPatch_unknownKeysIgnored(t *testing.T) {
	base := Task{
		ID:          uuid.New().String(),
		Title:       "t",
		Status:      StatusTodo,
		FocusBucket: FocusNone,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	out, err := ApplyPatch(base, []byte(`{"extra":true,"status":"done"}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != StatusDone {
		t.Fatalf("status %q", out.Status)
	}
}

func TestApplyPatch_priorityEnum(t *testing.T) {
	base := Task{
		ID:          uuid.New().String(),
		Title:       "t",
		Status:      StatusTodo,
		FocusBucket: FocusNone,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	_, err := ApplyPatch(base, []byte(`{"priority":"urgent"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestApplyPatch_responseJSONShape(t *testing.T) {
	base := Task{
		ID:          uuid.New().String(),
		Title:       "t",
		Status:      StatusTodo,
		FocusBucket: FocusNone,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	out, err := ApplyPatch(base, []byte(`{"title":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"id", "title", "status", "focus_bucket", "created_at", "updated_at"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing key %q in %s", k, string(b))
		}
	}
}
