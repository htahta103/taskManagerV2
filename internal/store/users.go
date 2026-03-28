package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// User is a persisted account row (no password hash exposed).
type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Store) CreateUser(ctx context.Context, email, passwordHash, name string) (User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)
	if email == "" || name == "" {
		return User{}, ErrInvalidInput
	}
	const q = `
INSERT INTO users (email, password_hash, name)
VALUES ($1, $2, $3)
RETURNING id, email, name, created_at, updated_at`
	var u User
	err := s.pool.QueryRow(ctx, q, email, passwordHash, name).Scan(
		&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrDuplicate
		}
		return User{}, err
	}
	return u, nil
}

// UserWithHash is used for login verification.
type UserWithHash struct {
	User
	PasswordHash string
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (UserWithHash, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	const q = `
SELECT id, email, password_hash, name, created_at, updated_at
FROM users WHERE lower(email) = lower($1)`
	var u UserWithHash
	err := s.pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserWithHash{}, ErrNotFound
	}
	if err != nil {
		return UserWithHash{}, err
	}
	return u, nil
}

func (s *Store) GetUser(ctx context.Context, id uuid.UUID) (User, error) {
	const q = `SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`
	var u User
	err := s.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Store) UpdateUser(ctx context.Context, id uuid.UUID, email, name *string) (User, error) {
	u, err := s.GetUser(ctx, id)
	if err != nil {
		return User{}, err
	}
	newEmail := u.Email
	newName := u.Name
	if email != nil {
		v := strings.TrimSpace(strings.ToLower(*email))
		if v != "" {
			newEmail = v
		}
	}
	if name != nil {
		newName = strings.TrimSpace(*name)
	}
	const q = `
UPDATE users SET email = $2, name = $3, updated_at = now()
WHERE id = $1
RETURNING id, email, name, created_at, updated_at`
	var out User
	err = s.pool.QueryRow(ctx, q, id, newEmail, newName).Scan(
		&out.ID, &out.Email, &out.Name, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrDuplicate
		}
		return User{}, err
	}
	return out, nil
}
