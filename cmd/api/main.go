package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/htahta103/taskmanagerv2/internal/config"
	"github.com/htahta103/taskmanagerv2/internal/db"
	"github.com/htahta103/taskmanagerv2/internal/httpserver"
	"github.com/htahta103/taskmanagerv2/internal/task"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required (use a long random secret in production)")
	}

	h, cleanup, err := buildHandler(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	addr := ":" + strconv.Itoa(cfg.Port)
	log.Printf("listening on %s (env=%s)", addr, cfg.Environment)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatal(err)
	}
}

func buildHandler(cfg config.Config) (handler http.Handler, cleanup func(), err error) {
	// Always serve the Supabase Edge-style create route in-memory for the MVP.
	taskHandler := &task.Handler{Store: task.NewMemoryStore()}

	var base http.Handler
	if cfg.DatabaseURL == "" {
		log.Printf("DATABASE_URL not set; starting in stub mode (no DB-backed /api/v1 routes)")
		base = httpserver.NewMux()
	} else {
		ctx := context.Background()
		pool, err := db.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, nil, err
		}
		if err := db.Migrate(ctx, pool); err != nil {
			pool.Close()
			return nil, nil, err
		}
		base = httpserver.NewHandler(cfg, pool)
		cleanup = pool.Close
	}

	// Route multiplexer to host both the app API and the Edge-function-style endpoint.
	mux := http.NewServeMux()
	mux.Handle("/", base)
	mux.HandleFunc("POST /functions/v1/tasks", taskHandler.Create)
	return mux, cleanup, nil
}
