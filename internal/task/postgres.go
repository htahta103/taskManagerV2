package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore reads tasks from PostgreSQL.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore returns a store backed by pool.
func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

// GetTask implements Getter.
func (s *PostgresStore) GetTask(ctx context.Context, id uuid.UUID) (Task, error) {
	const q = `
SELECT id, title, description, status::text, priority::text, due_date,
       created_at, updated_at, project_id, assignee_id, tags
FROM tasks
WHERE id = $1`

	var t Task
	var desc sql.NullString
	var prio sql.NullString
	var due sql.NullTime
	var projectID, assigneeID *uuid.UUID
	var tags []string

	err := s.pool.QueryRow(ctx, q, id).Scan(
		&t.ID,
		&t.Title,
		&desc,
		&t.Status,
		&prio,
		&due,
		&t.CreatedAt,
		&t.UpdatedAt,
		&projectID,
		&assigneeID,
		&tags,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("query task: %w", err)
	}

	if desc.Valid {
		s := desc.String
		t.Description = &s
	}
	if prio.Valid {
		s := prio.String
		t.Priority = &s
	}
	if due.Valid {
		s := due.Time.UTC().Format("2006-01-02")
		t.DueDate = &s
	}
	t.FocusBucket = "none"
	t.ProjectID = projectID
	t.AssigneeID = assigneeID
	if tags == nil {
		t.Tags = []string{}
	} else {
		t.Tags = tags
	}
	return t, nil
}
