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

func TestIndexingAndSearchWorkflow(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	pagesToStore := []*storage.Page{
		{URL: "https://golang.org", Title: "Go Language", Body: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software."},
		{URL: "https://vuejs.org", Title: "Vue.js", Body: "Vue.js is a progressive, incrementally-adoptable JavaScript framework for building UI on the web."},
		{URL: "https://gobyexample.com", Title: "Go by Example", Body: "A hands-on introduction to Go using annotated example programs. A great resource for learning Go."},
	}
	for _, p := range pagesToStore {
		_, err := db.StorePage(ctx, p)
		require.NoError(t, err)
	}

	for {
		page, err := db.GetNextPageToIndex(ctx)
		require.NoError(t, err)
		if page == nil {
			break
		}
		err = db.UpdatePageVector(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Поиск по уникальному слову 'framework'", func(t *testing.T) {
		results, err := db.SearchPages(ctx, "framework")
		require.NoError(t, err)
		require.Len(t, results, 1)
		require.Equal(t, "https://vuejs.org", results[0].URL)
	})

	t.Run("Поиск по общему слову 'go'", func(t *testing.T) {
		results, err := db.SearchPages(ctx, "go")
		require.NoError(t, err)
		require.Len(t, results, 2)
		foundURLs := []string{results[0].URL, results[1].URL}
		require.Contains(t, foundURLs, "https://golang.org")
		require.Contains(t, foundURLs, "https://gobyexample.com")
	})

	t.Run("Поиск по несуществующему слову", func(t *testing.T) {
		results, err := db.SearchPages(ctx, "nonexistentword")
		require.NoError(t, err)
		require.Len(t, results, 0)
	})
}
