package httpserver_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/htahta103/taskmanagerv2/internal/httpserver"
)

func TestIntegration_Healthz(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(httpserver.NewMux())
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]string
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("json: %v", err)
	}
	if out["status"] != "ok" {
		t.Fatalf("status field = %q, want ok", out["status"])
	}
}

func TestIntegration_APIv1Root(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(httpserver.NewMux())
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/v1")
	if err != nil {
		t.Fatalf("GET /api/v1: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var out map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("json: %v", err)
	}
	if out["name"] != "taskmanagerv2-api" || out["version"] != "v1" {
		t.Fatalf("body = %#v, want name=taskmanagerv2-api version=v1", out)
	}
}
