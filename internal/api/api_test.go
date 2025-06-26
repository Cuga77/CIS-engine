package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cis-engine/internal/search"
	"cis-engine/internal/storage"

	"github.com/stretchr/testify/require"
)

type mockSearchService struct {
	searchFunc        func(ctx context.Context, query string) ([]search.Result, error)
	scheduleCrawlFunc func(ctx context.Context, url string) error
	getStatsFunc      func(ctx context.Context) (*storage.Metrics, error)
}

func (m *mockSearchService) Search(ctx context.Context, query string) ([]search.Result, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query)
	}
	return nil, errors.New("searchFunc не был определен")
}

func (m *mockSearchService) ScheduleCrawl(ctx context.Context, url string) error {
	if m.scheduleCrawlFunc != nil {
		return m.scheduleCrawlFunc(ctx, url)
	}
	return errors.New("scheduleCrawlFunc не был определен")
}

func (m *mockSearchService) GetStats(ctx context.Context) (*storage.Metrics, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx)
	}
	return nil, errors.New("getStatsFunc не был определен")
}

func TestSearchHandler(t *testing.T) {
	t.Run("Успешный запрос", func(t *testing.T) {
		mockService := &mockSearchService{
			searchFunc: func(ctx context.Context, query string) ([]search.Result, error) {
				require.Equal(t, "test", query)
				return []search.Result{{URL: "test.com", Title: "Test"}}, nil
			},
		}
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=test", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var responseBody map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		require.Equal(t, "test", responseBody["query"])
	})

	t.Run("Запрос без параметра q", func(t *testing.T) {
		mockService := &mockSearchService{}
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Сервис возвращает ошибку", func(t *testing.T) {
		mockService := &mockSearchService{
			searchFunc: func(ctx context.Context, query string) ([]search.Result, error) {
				return nil, errors.New("internal error")
			},
		}
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=error", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestCrawlHandler(t *testing.T) {
	t.Run("Успешный запрос на сканирование", func(t *testing.T) {
		crawlURL := "https://example.com/to-crawl"
		scheduleCrawlCalled := false
		mockService := &mockSearchService{
			scheduleCrawlFunc: func(ctx context.Context, url string) error {
				require.Equal(t, crawlURL, url)
				scheduleCrawlCalled = true
				return nil
			},
		}
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		requestBody := `{"url": "https://example.com/to-crawl"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawl", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusAccepted, rec.Code)
		require.True(t, scheduleCrawlCalled, "Метод ScheduleCrawl должен был быть вызван")
	})

	t.Run("Запрос с неверным JSON", func(t *testing.T) {
		mockService := &mockSearchService{}
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		requestBody := `{"invalid_field": "test"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawl", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestStatusHandler(t *testing.T) {
	t.Run("Успешное получение статуса", func(t *testing.T) {
		mockService := &mockSearchService{
			getStatsFunc: func(ctx context.Context) (*storage.Metrics, error) {
				return &storage.Metrics{PagesCount: 123}, nil
			},
		}
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var stats storage.Metrics
		err := json.Unmarshal(rec.Body.Bytes(), &stats)
		require.NoError(t, err)
		require.Equal(t, int64(123), stats.PagesCount)
	})

	t.Run("Сервис возвращает ошибку при получении статуса", func(t *testing.T) {
		mockService := &mockSearchService{
			getStatsFunc: func(ctx context.Context) (*storage.Metrics, error) {
				return nil, errors.New("db is down")
			},
		}
		handler := NewHandler(mockService)
		router := NewRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
