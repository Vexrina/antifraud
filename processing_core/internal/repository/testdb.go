package repository

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func setupSchema(
	ctx context.Context,
	t *testing.T,
	pool *pgxpool.Pool,
) string {
	t.Helper()

	schema := "test_" + strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err := pool.Exec(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
	})

	return schema
}

func NewTestPool(
	ctx context.Context,
	t *testing.T,
	connStr string,
	schema string,
) *pgxpool.Pool {
	t.Helper()

	cfg, err := pgxpool.ParseConfig(connStr)
	require.NoError(t, err)

	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO "+schema)
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func findModuleRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found (cannot locate module root)")
		}
		wd = parent
	}
}
func migrationsURL(t *testing.T) string {
	t.Helper()

	root := findModuleRoot(t)
	path := filepath.Join(root, "migrations")

	_, err := os.Stat(path)
	require.NoError(t, err)

	return "file://" + path
}

func runMigrations(t *testing.T, connStr, schema string) {
	t.Helper()

	dbURL := connStr + "&search_path=" + schema

	m, err := migrate.New(
		migrationsURL(t),
		dbURL,
	)
	require.NoError(t, err)

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
}

func NewTestDB(
	ctx context.Context,
	t *testing.T,
) *pgDb {
	t.Helper()

	connStr := os.Getenv("TEST_DB_CONNSTR")
	// pool без search_path — только для CREATE SCHEMA
	basePool, err := NewPool(ctx, connStr)
	require.NoError(t, err)

	schema := setupSchema(ctx, t, basePool)

	pool := NewTestPool(ctx, t, connStr, schema)

	runMigrations(t, connStr, schema)

	return &pgDb{conn: pool}
}
