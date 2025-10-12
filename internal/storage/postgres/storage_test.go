package postgres

import (
	"context"
	"testing"

	"cis-engine/internal/storage"

	"github.com/stretchr/testify/require"
)

func TestGetNextPageToIndex(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// 1. Test with no pages in the database
	page, err := db.GetNextPageToIndex(ctx)
	require.NoError(t, err)
	require.Nil(t, page)

	// 2. Test with a page that needs indexing
	pageToStore := &storage.Page{
		URL:   "https://example.com",
		Title: "Example Domain",
		Body:  "This is the body of the example page.",
	}
	_, err = db.StorePage(ctx, pageToStore)
	require.NoError(t, err)

	page, err = db.GetNextPageToIndex(ctx)
	require.NoError(t, err)
	require.NotNil(t, page)
	require.Equal(t, pageToStore.URL, page.URL)

	// 3. Test that the same page is not returned again
	err = db.UpdatePageVector(ctx, page)
	require.NoError(t, err)

	page, err = db.GetNextPageToIndex(ctx)
	require.NoError(t, err)
	require.Nil(t, page)
}