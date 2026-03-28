package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/htahta103/taskmanagerv2/internal/config"
	"github.com/htahta103/taskmanagerv2/internal/task"
)

func main() {
	cfg := config.Load()
	mux := http.NewServeMux()

	store := task.NewMemoryStore()
	taskHandler := &task.Handler{Store: store}
	mux.HandleFunc("POST /functions/v1/tasks", taskHandler.Create)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/v1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"name":    "taskmanagerv2-api",
			"version": "v1",
		})
	})

	addr := ":" + strconv.Itoa(cfg.Port)
	log.Printf("listening on %s (env=%s)", addr, cfg.Environment)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
