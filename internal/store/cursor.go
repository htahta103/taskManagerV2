package store

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type pageCursorPayload struct {
	U string `json:"u"`
	I string `json:"i"`
}

// PageCursor is a keyset cursor for task list pagination (updated_at DESC, id DESC).
type PageCursor struct {
	UpdatedAt time.Time
	ID        uuid.UUID
}

// EncodePageCursor returns an opaque cursor string.
func EncodePageCursor(c PageCursor) (string, error) {
	p := pageCursorPayload{
		U: c.UpdatedAt.UTC().Format(time.RFC3339Nano),
		I: c.ID.String(),
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// DecodePageCursor parses a cursor from ListTasks.
func DecodePageCursor(s string) (PageCursor, error) {
	if s == "" {
		return PageCursor{}, errors.New("empty cursor")
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return PageCursor{}, err
	}
	var p pageCursorPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return PageCursor{}, err
	}
	ts, err := time.Parse(time.RFC3339Nano, p.U)
	if err != nil {
		return PageCursor{}, err
	}
	id, err := uuid.Parse(p.I)
	if err != nil {
		return PageCursor{}, err
	}
	return PageCursor{UpdatedAt: ts, ID: id}, nil
}
