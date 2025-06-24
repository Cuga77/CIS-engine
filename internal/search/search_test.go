// internal/search/search_test.go

package search

import (
	"context"
	"errors"
	"testing"

	"cis-engine/internal/storage"

	"github.com/stretchr/testify/require"
)

// mockStorer - это фальшивая реализация интерфейса storage.Storer для тестов.
type mockStorer struct {
	// Мы можем "запрограммировать" мок на возврат определенных данных или ошибок.
	searchPagesFunc func(ctx context.Context, query string) ([]*storage.Page, error)
}

// Реализуем методы интерфейса storage.Storer для нашего мока.
func (m *mockStorer) SearchPages(ctx context.Context, query string) ([]*storage.Page, error) {
	if m.searchPagesFunc != nil {
		return m.searchPagesFunc(ctx, query)
	}
	return nil, errors.New("searchPagesFunc не был определен")
}

// Методы-заглушки для других частей интерфейса, они нам в этом тесте не понадобятся.
func (m *mockStorer) StorePage(ctx context.Context, page *storage.Page) (int64, error) { return 0, nil }
func (m *mockStorer) GetNextPageToIndex(ctx context.Context) (*storage.Page, error)    { return nil, nil }
func (m *mockStorer) UpdatePageVector(ctx context.Context, page *storage.Page) error   { return nil }
func (m *mockStorer) Close()                                                           {}

func TestSearchService(t *testing.T) {
	ctx := context.Background()

	t.Run("Успешный поиск", func(t *testing.T) {
		// 1. Настройка мока: говорим ему вернуть две страницы при вызове SearchPages.
		mockStorage := &mockStorer{
			searchPagesFunc: func(ctx context.Context, query string) ([]*storage.Page, error) {
				require.Equal(t, "go", query) // Проверяем, что сервис вызвал метод с правильным запросом
				return []*storage.Page{
					{URL: "https://golang.org", Title: "The Go Language"},
					{URL: "https://go.dev", Title: "Official Go Website"},
				}, nil
			},
		}
		service := NewService(mockStorage)

		// 2. Вызов тестируемого метода
		results, err := service.Search(ctx, "go")

		// 3. Проверка результата
		require.NoError(t, err)
		require.Len(t, results, 2)
		require.Equal(t, "https://golang.org", results[0].URL)
	})

	t.Run("Поиск не дал результатов", func(t *testing.T) {
		// 1. Настройка мока: говорим ему вернуть пустой срез.
		mockStorage := &mockStorer{
			searchPagesFunc: func(ctx context.Context, query string) ([]*storage.Page, error) {
				return []*storage.Page{}, nil
			},
		}
		service := NewService(mockStorage)

		// 2. Вызов
		results, err := service.Search(ctx, "nonexistent")

		// 3. Проверка
		require.NoError(t, err)
		require.Len(t, results, 0)
	})

	t.Run("Ошибка от хранилища", func(t *testing.T) {
		// 1. Настройка мока: говорим ему вернуть ошибку.
		mockStorage := &mockStorer{
			searchPagesFunc: func(ctx context.Context, query string) ([]*storage.Page, error) {
				return nil, errors.New("DB connection failed")
			},
		}
		service := NewService(mockStorage)

		// 2. Вызов
		_, err := service.Search(ctx, "any query")

		// 3. Проверка
		require.Error(t, err)
		require.Equal(t, "DB connection failed", err.Error())
	})
}
