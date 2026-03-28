package httpserver

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const userIDKey ctxKey = 1

func withUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
