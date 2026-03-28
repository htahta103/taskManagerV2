package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/htahta103/taskmanagerv2/internal/config"
	"github.com/htahta103/taskmanagerv2/internal/httpapi"
	"github.com/htahta103/taskmanagerv2/internal/jsonresp"
	"github.com/htahta103/taskmanagerv2/internal/tasks"
)

func main() {
	cfg := config.Load()
	taskStore := tasks.NewStore()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		jsonresp.Write(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/v1", func(w http.ResponseWriter, _ *http.Request) {
		jsonresp.Write(w, http.StatusOK, map[string]string{
			"name":    "taskmanagerv2-api",
			"version": "v1",
		})
	})
	mux.HandleFunc("DELETE /functions/v1/tasks/clear/done", httpapi.HandleClearDone(taskStore))
	mux.HandleFunc("DELETE /functions/v1/tasks/{id}", httpapi.HandleTaskDelete(taskStore))

	addr := ":" + strconv.Itoa(cfg.Port)
	log.Printf("listening on %s (env=%s)", addr, cfg.Environment)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
