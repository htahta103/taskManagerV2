package httpserver

import (
	"log"
	"net/http"

	"github.com/htahta103/taskmanagerv2/internal/config"
	"github.com/htahta103/taskmanagerv2/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	cfg    config.Config
	pool   *pgxpool.Pool
	store  *store.Store
	authRL *authMinuteLimiter
}

// NewHandler returns the full API HTTP handler (routes, CORS, database-backed handlers).
func NewHandler(cfg config.Config, pool *pgxpool.Pool) http.Handler {
	s := &server{
		cfg:    cfg,
		pool:   pool,
		store:  store.New(pool),
		authRL: newAuthMinuteLimiter(cfg.AuthRateLimitPerMinute),
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.getHealthz)
	mux.HandleFunc("GET /readyz", s.getReadyz)

	mux.HandleFunc("GET /api/v1", s.getAPIv1Root)

	mux.HandleFunc("POST /api/v1/auth/register", s.withAuthRateLimit(s.postRegister))
	mux.HandleFunc("POST /api/v1/auth/login", s.withAuthRateLimit(s.postLogin))
	mux.HandleFunc("POST /api/v1/auth/logout", s.requireAuth(s.postLogout))
	mux.HandleFunc("POST /api/v1/auth/refresh", s.withAuthRateLimit(s.postRefresh))

	mux.HandleFunc("GET /api/v1/me", s.requireAuth(s.getMe))
	mux.HandleFunc("PATCH /api/v1/me", s.requireAuth(s.patchMe))

	mux.HandleFunc("GET /api/v1/projects", s.requireAuth(s.listProjects))
	mux.HandleFunc("POST /api/v1/projects", s.requireAuth(s.createProject))
	mux.HandleFunc("GET /api/v1/projects/{projectId}", s.requireAuth(s.getProject))
	mux.HandleFunc("PATCH /api/v1/projects/{projectId}", s.requireAuth(s.patchProject))
	mux.HandleFunc("DELETE /api/v1/projects/{projectId}", s.requireAuth(s.deleteProject))

	mux.HandleFunc("GET /api/v1/tasks", s.requireAuth(s.listTasks))
	mux.HandleFunc("POST /api/v1/tasks", s.requireAuth(s.createTask))
	mux.HandleFunc("GET /api/v1/tasks/{taskId}", s.requireAuth(s.getTask))
	mux.HandleFunc("PATCH /api/v1/tasks/{taskId}", s.requireAuth(s.patchTask))
	mux.HandleFunc("DELETE /api/v1/tasks/{taskId}", s.requireAuth(s.deleteTask))

	mux.HandleFunc("POST /api/v1/tasks/{taskId}/tags", s.requireAuth(s.postTaskTags))
	mux.HandleFunc("DELETE /api/v1/tasks/{taskId}/tags/{tagId}", s.requireAuth(s.deleteTaskTag))

	mux.HandleFunc("GET /api/v1/tags", s.requireAuth(s.listTags))
	mux.HandleFunc("POST /api/v1/tags", s.requireAuth(s.createTag))

	mux.HandleFunc("GET /api/v1/search", s.requireAuth(s.getSearch))

	return s.cors(mux)
}

func (s *server) getHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) getReadyz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := s.pool.Ping(ctx); err != nil {
		log.Printf("readyz: ping failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "database unreachable", "not_ready")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) getAPIv1Root(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name":    "taskmanagerv2-api",
		"version": "v1",
	})
}

// NewMux is kept for backward compatibility in tests that only need the legacy stub routes.
// Prefer NewHandler for production; it requires a live database pool.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"name":    "taskmanagerv2-api",
			"version": "v1",
		})
	})
	return mux
}
