package store

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store performs data access against PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

// New returns a Store backed by pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}
