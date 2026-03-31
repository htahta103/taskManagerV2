package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/htahta103/taskmanagerv2/internal/config"
	"github.com/htahta103/taskmanagerv2/internal/db"
	"github.com/htahta103/taskmanagerv2/internal/httpserver"
)

func TestIntegration_APIAuthProjectsTasks(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; start postgres (e.g. docker compose) to run integration tests")
	}
	if os.Getenv("JWT_SECRET") == "" {
		t.Setenv("JWT_SECRET", "integration-test-jwt-secret-32-characters!!")
	}
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		t.Fatal("JWT_SECRET required for integration test")
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("db pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	srv := httptest.NewServer(httpserver.NewHandler(cfg, pool))
	t.Cleanup(srv.Close)

	email := "integration-" + uuid.NewString() + "@example.com"
	regBody := map[string]any{
		"email":    email,
		"password": "longpassword1",
		"name":     "Integration",
	}
	rb, _ := json.Marshal(regBody)
	resp, err := http.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(rb))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("register status %d: %s", resp.StatusCode, b)
	}
	var regOut map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&regOut); err != nil {
		t.Fatalf("register json: %v", err)
	}
	token, _ := regOut["access_token"].(string)
	if token == "" {
		t.Fatal("missing access_token")
	}

	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	meResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	defer meResp.Body.Close()
	if meResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(meResp.Body)
		t.Fatalf("me status %d: %s", meResp.StatusCode, b)
	}

	pb, _ := json.Marshal(map[string]any{"name": "Proj A"})
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/projects", bytes.NewReader(pb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	pr, err := client.Do(req)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	defer pr.Body.Close()
	if pr.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(pr.Body)
		t.Fatalf("project status %d: %s", pr.StatusCode, b)
	}
	var proj map[string]any
	if err := json.NewDecoder(pr.Body).Decode(&proj); err != nil {
		t.Fatalf("project json: %v", err)
	}
	pid, _ := proj["id"].(string)

	tb, _ := json.Marshal(map[string]any{
		"title":      "Hello task",
		"project_id": pid,
	})
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/tasks", bytes.NewReader(tb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	tr, err := client.Do(req)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	defer tr.Body.Close()
	if tr.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(tr.Body)
		t.Fatalf("task status %d: %s", tr.StatusCode, b)
	}

	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	lr, err := client.Do(req)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	defer lr.Body.Close()
	if lr.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(lr.Body)
		t.Fatalf("list tasks status %d: %s", lr.StatusCode, b)
	}

	rz, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatalf("readyz: %v", err)
	}
	defer rz.Body.Close()
	if rz.StatusCode != http.StatusOK {
		t.Fatalf("readyz status %d", rz.StatusCode)
	}
}

func TestIntegration_APIValidation422Envelope(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; start postgres (e.g. docker compose) to run integration tests")
	}
	if os.Getenv("JWT_SECRET") == "" {
		t.Setenv("JWT_SECRET", "integration-test-jwt-secret-32-characters!!")
	}
	t.Setenv("APP_ENV", "development")
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("db pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	srv := httptest.NewServer(httpserver.NewHandler(cfg, pool))
	t.Cleanup(srv.Close)

	regBody := map[string]any{
		"email":    "badpw-" + t.Name() + "@example.com",
		"password": "short",
		"name":     "x",
	}
	rb, _ := json.Marshal(regBody)
	resp, err := http.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(rb))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("register status %d want 422: %s", resp.StatusCode, b)
	}
	var env map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("json: %v", err)
	}
	if env["code"] != "validation" {
		t.Fatalf("code = %v want validation", env["code"])
	}
	details, ok := env["details"].(map[string]any)
	if !ok {
		t.Fatalf("missing details: %#v", env)
	}
	fields, ok := details["fields"].(map[string]any)
	if !ok || fields["password"] == nil {
		t.Fatalf("missing fields.password in %#v", details)
	}
}
