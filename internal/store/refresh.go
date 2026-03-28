package store

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func refreshHash(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

// InsertRefreshToken stores a hashed refresh token and returns its row id.
func (s *Store) InsertRefreshToken(ctx context.Context, userID uuid.UUID, rawToken string, expiresAt time.Time) (uuid.UUID, error) {
	const q = `
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id`
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, q, userID, refreshHash(rawToken), expiresAt).Scan(&id)
	return id, err
}

// UserIDForValidRefresh returns the user id if the token exists, is not revoked, and not expired.
func (s *Store) UserIDForValidRefresh(ctx context.Context, rawToken string) (uuid.UUID, error) {
	const q = `
SELECT user_id FROM refresh_tokens
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()`
	var uid uuid.UUID
	err := s.pool.QueryRow(ctx, q, refreshHash(rawToken)).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}

// RevokeRefreshToken marks a specific refresh token revoked.
func (s *Store) RevokeRefreshToken(ctx context.Context, rawToken string) error {
	ct, err := s.pool.Exec(ctx, `
UPDATE refresh_tokens SET revoked_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL`, refreshHash(rawToken))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// RevokeAllRefreshTokensForUser revokes every refresh token for the user.
func (s *Store) RevokeAllRefreshTokensForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
UPDATE refresh_tokens SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}
