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
func (s *PostgresStore) GetTask(ctx context.Context, id string) (Task, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return Task{}, fmt.Errorf("parse id: %w", err)
	}
	const q = `
SELECT id, title, description, status, priority, due_date, focus_bucket,
       created_at, updated_at, project_id, assignee_id
FROM tasks
WHERE id = $1`

	var t Task
	var desc sql.NullString
	var prio sql.NullString
	var due sql.NullTime
	var tid uuid.UUID
	var focus string
	var projectID, assigneeID *uuid.UUID

	err = s.pool.QueryRow(ctx, q, uid).Scan(
		&tid,
		&t.Title,
		&desc,
		&t.Status,
		&prio,
		&due,
		&focus,
		&t.CreatedAt,
		&t.UpdatedAt,
		&projectID,
		&assigneeID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("query task: %w", err)
	}

	t.ID = tid.String()
	if desc.Valid {
		s := desc.String
		t.Description = &s
	}
	if prio.Valid {
		p := Priority(prio.String)
		t.Priority = &p
	}
	if due.Valid {
		s := due.Time.UTC().Format("2006-01-02")
		t.DueDate = &s
	}
	t.FocusBucket = FocusBucket(focus)
	if projectID != nil {
		s := projectID.String()
		t.ProjectID = &s
	}
	if assigneeID != nil {
		s := assigneeID.String()
		t.AssigneeID = &s
	}
	t.Tags = []Tag{}
	return t, nil
}
