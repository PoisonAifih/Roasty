package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	candidates := []string{
		"schema.sql",
		filepath.Join("backend", "schema.sql"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "schema.sql"),
			filepath.Join(wd, "backend", "schema.sql"),
		)
	}
	var sqlBytes []byte
	var readErr error
	for _, c := range candidates {
		sqlBytes, readErr = os.ReadFile(c)
		if readErr == nil {
			break
		}
	}
	if readErr != nil {
		return fmt.Errorf("read schema.sql: %w", readErr)
	}
	_, err := pool.Exec(ctx, string(sqlBytes))
	return err
}
