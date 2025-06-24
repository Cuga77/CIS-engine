// internal/api/api_test.go

package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"cis-engine/internal/search"

	"github.com/stretchr/testify/require"
)

// searcher - это интерфейс, который удовлетворяет наш search.Service.
// Он нужен для создания мока.
type searcher interface {
	Search(ctx context.Context, query string) ([]search.Result, error)
}

// mockSearchService - это фальшивая реализация нашего сервиса поиска.
type mockSearchService struct {
	searchFunc func(ctx context.Context, query string) ([]search.Result, error)
}

func (m *mockSearchService) Search(ctx context.Context, query string) ([]search.Result, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query)
	}
	return nil, errors.New("searchFunc не был определен")
}

func TestSearchHandler(t *testing.T) {
	t.Run("Успешный запрос", func(t *testing.T) {
		// 1. Настройка мока
		mockService := &mockSearchService{
			searchFunc: func(ctx context.Context, query string) ([]search.Result, error) {
				require.Equal(t, "test", query)
				return []search.Result{{URL: "test.com", Title: "Test"}}, nil
			},
		}
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		// 2. Создание тестового HTTP запроса и рекордера для записи ответа
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=test", nil)
		rec := httptest.NewRecorder()

		// 3. Выполнение запроса
		router.ServeHTTP(rec, req)

		// 4. Проверка результата
		require.Equal(t, http.StatusOK, rec.Code)

		var responseBody map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		require.Equal(t, "test", responseBody["query"])
	})

	t.Run("Запрос без параметра q", func(t *testing.T) {
		// 1. Настройка
		mockService := &mockSearchService{} // Мок не будет вызван
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		// 2. Запрос
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
		rec := httptest.NewRecorder()

		// 3. Выполнение
		router.ServeHTTP(rec, req)

		// 4. Проверка
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Сервис возвращает ошибку", func(t *testing.T) {
		// 1. Настройка мока на возврат ошибки
		mockService := &mockSearchService{
			searchFunc: func(ctx context.Context, query string) ([]search.Result, error) {
				return nil, errors.New("internal error")
			},
		}
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		// 2. Запрос
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=error", nil)
		rec := httptest.NewRecorder()

		// 3. Выполнение
		router.ServeHTTP(rec, req)

		// 4. Проверка
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
