// internal/storage/postgres/db_test.go

package postgres

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"cis-engine/internal/storage"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *DB {
	ctx := context.Background()

	schemaPath, err := filepath.Abs("../schema.sql")
	require.NoError(t, err)

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:14-alpine"),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.WithInitScripts(schemaPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := New(ctx, connStr)
	require.NoError(t, err)

	return db
}

func TestStorePage(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	testPage := &storage.Page{
		URL:   "https://example.com",
		Title: "Example Domain",
		Body:  "This is the body of the example page.",
	}

	id, err := db.StorePage(ctx, testPage)

	require.NoError(t, err)
	require.Greater(t, id, int64(0))
}

// TODO: Добавить тесты для других методов, например, GetNextPageToIndex и SearchPages.
