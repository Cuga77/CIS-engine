package search

import (
	"context"
	"errors"
	"testing"

	"cis-engine/internal/storage"

	"github.com/stretchr/testify/require"
)

type mockStorer struct {
	searchPagesFunc func(ctx context.Context, query string) ([]*storage.Page, error)
	getMetricsFunc  func(ctx context.Context) (*storage.Metrics, error)
}

func (m *mockStorer) SearchPages(ctx context.Context, query string) ([]*storage.Page, error) {
	if m.searchPagesFunc != nil {
		return m.searchPagesFunc(ctx, query)
	}
	return nil, errors.New("searchPagesFunc не был определен")
}

func (m *mockStorer) GetMetrics(ctx context.Context) (*storage.Metrics, error) {
	if m.getMetricsFunc != nil {
		return m.getMetricsFunc(ctx)
	}
	return nil, errors.New("getMetricsFunc не был определен")
}

func (m *mockStorer) StorePage(ctx context.Context, page *storage.Page) (int64, error) { return 0, nil }
func (m *mockStorer) GetNextPageToIndex(ctx context.Context) (*storage.Page, error)    { return nil, nil }
func (m *mockStorer) UpdatePageVector(ctx context.Context, page *storage.Page) error   { return nil }
func (m *mockStorer) Close()                                                           {}

func TestSearchService(t *testing.T) {
	ctx := context.Background()

	t.Run("Успешный поиск", func(t *testing.T) {
		mockStorage := &mockStorer{
			searchPagesFunc: func(ctx context.Context, query string) ([]*storage.Page, error) {
				require.Equal(t, "go", query)
				return []*storage.Page{
					{URL: "https://golang.org", Title: "The Go Language"},
					{URL: "https://go.dev", Title: "Official Go Website"},
				}, nil
			},
		}
		service := NewService(mockStorage)

		results, err := service.Search(ctx, "go")

		require.NoError(t, err)
		require.Len(t, results, 2)
		require.Equal(t, "https://golang.org", results[0].URL)
	})

	t.Run("Поиск не дал результатов", func(t *testing.T) {
		mockStorage := &mockStorer{
			searchPagesFunc: func(ctx context.Context, query string) ([]*storage.Page, error) {
				return []*storage.Page{}, nil
			},
		}
		service := NewService(mockStorage)

		results, err := service.Search(ctx, "nonexistent")

		require.NoError(t, err)
		require.Len(t, results, 0)
	})

	t.Run("Ошибка от хранилища", func(t *testing.T) {
		mockStorage := &mockStorer{
			searchPagesFunc: func(ctx context.Context, query string) ([]*storage.Page, error) {
				return nil, errors.New("DB connection failed")
			},
		}
		service := NewService(mockStorage)

		_, err := service.Search(ctx, "any query")

		require.Error(t, err)
		require.Equal(t, "DB connection failed", err.Error())
	})
}
func TestGetStats(t *testing.T) {
	ctx := context.Background()

	t.Run("Успешное получение статистики", func(t *testing.T) {
		mockStorage := &mockStorer{
			getMetricsFunc: func(ctx context.Context) (*storage.Metrics, error) {
				return &storage.Metrics{PagesCount: 42}, nil
			},
		}
		service := NewService(mockStorage)
		stats, err := service.GetStats(ctx)

		require.NoError(t, err)
		require.NotNil(t, stats)
		require.Equal(t, int64(42), stats.PagesCount)
	})

	t.Run("Ошибка от хранилища при получении статистики", func(t *testing.T) {
		mockStorage := &mockStorer{
			getMetricsFunc: func(ctx context.Context) (*storage.Metrics, error) {
				return nil, errors.New("failed to query metrics")
			},
		}
		service := NewService(mockStorage)
		stats, err := service.GetStats(ctx)

		require.Error(t, err)
		require.Nil(t, stats)
		require.Equal(t, "failed to query metrics", err.Error())
	})
}
