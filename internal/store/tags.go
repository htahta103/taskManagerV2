package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Tag is a user-scoped label.
type Tag struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

func (s *Store) CreateTag(ctx context.Context, userID uuid.UUID, name string) (Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 64 {
		return Tag{}, ErrInvalidInput
	}
	const q = `
INSERT INTO tags (user_id, name)
VALUES ($1, $2)
RETURNING id, name, created_at`
	var t Tag
	err := s.pool.QueryRow(ctx, q, userID, name).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return Tag{}, ErrDuplicate
		}
		return Tag{}, err
	}
	return t, nil
}

func (s *Store) ListTags(ctx context.Context, userID uuid.UUID, limit int, after *PageCursor) ([]Tag, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	args := []any{userID}
	where := "user_id = $1"
	if after != nil {
		where += " AND (created_at < $2 OR (created_at = $2 AND id < $3))"
		args = append(args, after.UpdatedAt, after.ID)
	}
	limitArg := len(args) + 1
	args = append(args, limit)
	q := fmt.Sprintf(
		`SELECT id, name, created_at FROM tags WHERE %s ORDER BY created_at DESC, id DESC LIMIT $%d`,
		where, limitArg,
	)
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

// EncodeTagCursor builds a pagination cursor from a tag row (uses created_at + id).
func EncodeTagCursor(createdAt time.Time, id uuid.UUID) (string, error) {
	return EncodePageCursor(PageCursor{UpdatedAt: createdAt, ID: id})
}

// TagsForTaskIDs returns tags grouped by task id for tasks owned by userID.
func (s *Store) TagsForTaskIDs(ctx context.Context, userID uuid.UUID, taskIDs []uuid.UUID) (map[uuid.UUID][]Tag, error) {
	out := make(map[uuid.UUID][]Tag)
	if len(taskIDs) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `
SELECT tt.task_id, tg.id, tg.name, tg.created_at
FROM task_tags tt
JOIN tags tg ON tg.id = tt.tag_id
JOIN tasks t ON t.id = tt.task_id
WHERE t.user_id = $1 AND tt.task_id = ANY($2::uuid[])`, userID, taskIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tid uuid.UUID
		var tag Tag
		if err := rows.Scan(&tid, &tag.ID, &tag.Name, &tag.CreatedAt); err != nil {
			return nil, err
		}
		out[tid] = append(out[tid], tag)
	}
	return out, rows.Err()
}

// AddTaskTags attaches tags to a task (idempotent per tag).
func (s *Store) AddTaskTags(ctx context.Context, userID, taskID uuid.UUID, tagIDs []uuid.UUID) error {
	_, err := s.GetTask(ctx, userID, taskID)
	if err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	var n int
	err = s.pool.QueryRow(ctx, `
SELECT COUNT(*) FROM tags WHERE user_id = $1 AND id = ANY($2::uuid[])`, userID, tagIDs).Scan(&n)
	if err != nil {
		return err
	}
	if n != len(tagIDs) {
		return ErrNotFound
	}
	for _, tid := range tagIDs {
		if _, err := s.pool.Exec(ctx, `
INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING`, taskID, tid); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) RemoveTaskTag(ctx context.Context, userID, taskID, tagID uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `
DELETE FROM task_tags tt USING tasks t
WHERE tt.task_id = $1 AND tt.tag_id = $2 AND t.id = tt.task_id AND t.user_id = $3`,
		taskID, tagID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
