// Package db handle koneksi dan migrasi database
package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DSN generate connection string PostgreSQL dari environment variables
func DSN() string {
	host := envOr("DB_HOST", "localhost")
	port := envOr("DB_PORT", "5432")
	user := envOr("DB_USER", "postgres")
	pass := envOr("DB_PASSWORD", "postgres")
	name := envOr("DB_NAME", "tasktracker")
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, name,
	)
}

// Connect inisialisasi koneksi PostgreSQL dengan retry mechanism
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := DSN()
	var (
		pool *pgxpool.Pool
		err  error
	)
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(ctx, dsn)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				log.Printf("✅  Connected to PostgreSQL (attempt %d)", i+1)
				return pool, nil
			}
			pool.Close()
		}
		log.Printf("⏳  Waiting for PostgreSQL... attempt %d/10 (%v)", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to PostgreSQL after 10 attempts: %w", err)
}

// Migrate membuat tabel tasks jika belum ada
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	const ddl = `
	CREATE TABLE IF NOT EXISTS tasks (
		id          UUID        PRIMARY KEY,
		title       VARCHAR(100) NOT NULL,
		description VARCHAR(500) NOT NULL,
		status      VARCHAR(20)  NOT NULL DEFAULT 'pending',
		created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
		updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	);`

	if _, err := pool.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	log.Println("✅  Database schema up-to-date")
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
