package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Project is a task container owned by a user.
type Project struct {
	ID          uuid.UUID
	Name        string
	Archived    bool
	ArchivedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *Store) ListProjects(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]Project, error) {
	var q string
	if includeArchived {
		q = `SELECT id, name, archived, archived_at, created_at, updated_at
FROM projects WHERE user_id = $1 ORDER BY updated_at DESC`
	} else {
		q = `SELECT id, name, archived, archived_at, created_at, updated_at
FROM projects WHERE user_id = $1 AND archived = FALSE ORDER BY updated_at DESC`
	}
	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Archived, &p.ArchivedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Store) CreateProject(ctx context.Context, userID uuid.UUID, name string) (Project, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 {
		return Project{}, ErrInvalidInput
	}
	const q = `
INSERT INTO projects (user_id, name)
VALUES ($1, $2)
RETURNING id, name, archived, archived_at, created_at, updated_at`
	var p Project
	err := s.pool.QueryRow(ctx, q, userID, name).Scan(
		&p.ID, &p.Name, &p.Archived, &p.ArchivedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return Project{}, err
	}
	return p, nil
}

func (s *Store) GetProject(ctx context.Context, userID, projectID uuid.UUID) (Project, error) {
	const q = `
SELECT id, name, archived, archived_at, created_at, updated_at
FROM projects WHERE id = $1 AND user_id = $2`
	var p Project
	err := s.pool.QueryRow(ctx, q, projectID, userID).Scan(
		&p.ID, &p.Name, &p.Archived, &p.ArchivedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if err != nil {
		return Project{}, err
	}
	return p, nil
}

func (s *Store) UpdateProject(ctx context.Context, userID, projectID uuid.UUID, name *string, archived *bool) (Project, error) {
	p, err := s.GetProject(ctx, userID, projectID)
	if err != nil {
		return Project{}, err
	}
	newName := p.Name
	newArchived := p.Archived
	var archAt *time.Time = p.ArchivedAt
	if name != nil {
		v := strings.TrimSpace(*name)
		if v != "" && len(v) <= 200 {
			newName = v
		}
	}
	if archived != nil {
		newArchived = *archived
		if newArchived {
			now := time.Now().UTC()
			archAt = &now
		} else {
			archAt = nil
		}
	}
	const q = `
UPDATE projects SET name = $3, archived = $4, archived_at = $5, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING id, name, archived, archived_at, created_at, updated_at`
	var out Project
	err = s.pool.QueryRow(ctx, q, projectID, userID, newName, newArchived, archAt).Scan(
		&out.ID, &out.Name, &out.Archived, &out.ArchivedAt, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return Project{}, err
	}
	return out, nil
}

func (s *Store) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1 AND user_id = $2`, projectID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ProjectOwnedBy(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND user_id = $2)`, projectID, userID).Scan(&exists)
	return exists, err
}
