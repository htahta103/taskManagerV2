package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/htahta103/taskmanagerv2/internal/config"
	"github.com/htahta103/taskmanagerv2/internal/httpserver"
	"github.com/htahta103/taskmanagerv2/internal/task"
)

func main() {
	cfg := config.Load()
	mux := httpserver.NewMux()
	store := task.NewMemoryStore()
	taskHandler := &task.Handler{Store: store}
	mux.HandleFunc("POST /functions/v1/tasks", taskHandler.Create)

	addr := ":" + strconv.Itoa(cfg.Port)
	log.Printf("listening on %s (env=%s)", addr, cfg.Environment)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
