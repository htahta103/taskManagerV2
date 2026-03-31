package httpserver

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/htahta103/taskmanagerv2/internal/auth"
	"github.com/htahta103/taskmanagerv2/internal/store"
)

type registerBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *server) issueAuthResponse(w http.ResponseWriter, r *http.Request, status int, u store.User) {
	ctx := r.Context()
	access, err := auth.MintAccessToken(u.ID, s.cfg.JWTSecret, s.cfg.AccessTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token", "token_error")
		return
	}
	rawRefresh, err := randomRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue refresh token", "token_error")
		return
	}
	exp := time.Now().UTC().Add(s.cfg.RefreshTTL)
	if _, err := s.store.InsertRefreshToken(ctx, u.ID, rawRefresh, exp); err != nil {
		writeError(w, http.StatusInternalServerError, "could not persist session", "token_error")
		return
	}
	writeJSON(w, status, map[string]any{
		"access_token":  access,
		"refresh_token": rawRefresh,
		"token_type":    "Bearer",
		"expires_in":    int(s.cfg.AccessTTL.Seconds()),
		"user":          formatUser(u),
	})
}

func (s *server) postRegister(w http.ResponseWriter, r *http.Request) {
	var body registerBody
	if err := readJSON(r, &body); err != nil {
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON or unknown fields"})
		return
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))
	nameTrim := strings.TrimSpace(body.Name)
	fields := map[string]string{}
	if email == "" {
		fields["email"] = "required"
	}
	if len(body.Password) < 8 {
		fields["password"] = "must be at least 8 characters"
	}
	if nameTrim == "" {
		fields["name"] = "required"
	} else if len(nameTrim) > 120 {
		fields["name"] = "must be at most 120 characters"
	}
	if len(fields) > 0 {
		writeValidation(w, "validation failed", fields)
		return
	}
	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not hash password", "internal")
		return
	}
	u, err := s.store.CreateUser(r.Context(), email, hash, nameTrim)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "email already registered", "duplicate")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create user", "internal")
		return
	}
	s.issueAuthResponse(w, r, http.StatusCreated, u)
}

func (s *server) postLogin(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := readJSON(r, &body); err != nil {
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON or unknown fields"})
		return
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))
	fields := map[string]string{}
	if email == "" {
		fields["email"] = "required"
	}
	if body.Password == "" {
		fields["password"] = "required"
	}
	if len(fields) > 0 {
		writeValidation(w, "validation failed", fields)
		return
	}
	row, err := s.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid credentials", "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load user", "internal")
		return
	}
	if !auth.CheckPassword(row.PasswordHash, body.Password) {
		writeError(w, http.StatusUnauthorized, "invalid credentials", "unauthorized")
		return
	}
	s.issueAuthResponse(w, r, http.StatusOK, row.User)
}

func (s *server) postLogout(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if err := s.store.RevokeAllRefreshTokensForUser(r.Context(), uid); err != nil {
		writeError(w, http.StatusInternalServerError, "could not sign out", "internal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) postRefresh(w http.ResponseWriter, r *http.Request) {
	var body refreshBody
	if err := readJSON(r, &body); err != nil {
		writeValidation(w, "invalid JSON body", map[string]string{"body": "malformed JSON or unknown fields"})
		return
	}
	raw := strings.TrimSpace(body.RefreshToken)
	if raw == "" {
		writeValidation(w, "validation failed", map[string]string{"refresh_token": "required"})
		return
	}
	ctx := r.Context()
	uid, err := s.store.UserIDForValidRefresh(ctx, raw)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid or expired refresh token", "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not validate refresh token", "internal")
		return
	}
	if err := s.store.RevokeRefreshToken(ctx, raw); err != nil {
		writeError(w, http.StatusInternalServerError, "could not rotate session", "internal")
		return
	}
	u, err := s.store.GetUser(ctx, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user", "internal")
		return
	}
	access, err := auth.MintAccessToken(u.ID, s.cfg.JWTSecret, s.cfg.AccessTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token", "token_error")
		return
	}
	newRaw, err := randomRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue refresh token", "token_error")
		return
	}
	exp := time.Now().UTC().Add(s.cfg.RefreshTTL)
	if _, err := s.store.InsertRefreshToken(ctx, u.ID, newRaw, exp); err != nil {
		writeError(w, http.StatusInternalServerError, "could not persist session", "token_error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  access,
		"refresh_token": newRaw,
		"token_type":    "Bearer",
		"expires_in":    int(s.cfg.AccessTTL.Seconds()),
	})
}
