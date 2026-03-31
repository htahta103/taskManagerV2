package db

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	repodb "github.com/htahta103/taskmanagerv2/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrate applies SQL files from db/migrations in lexicographic order (001_, 002_, …).
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("db: nil pool")
	}
	entries, err := fs.ReadDir(repodb.MigrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("db migrate: list migrations: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		p := path.Join("migrations", name)
		body, err := fs.ReadFile(repodb.MigrationFiles, p)
		if err != nil {
			return fmt.Errorf("db migrate: read %s: %w", p, err)
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("db migrate: exec %s: %w", name, err)
		}
	}
	return nil
}
