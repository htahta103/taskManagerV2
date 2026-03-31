package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/htahta103/taskmanagerv2/internal/config"
)

func TestBuildHandler_servesFunctionsV1TasksCreate(t *testing.T) {
	h, cleanup, err := buildHandler(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/functions/v1/tasks", "application/json", strings.NewReader(`{"title":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d", res.StatusCode)
	}

	var got map[string]any
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["id"] == "" || got["title"] != "x" {
		t.Fatalf("unexpected body: %+v", got)
	}
}

