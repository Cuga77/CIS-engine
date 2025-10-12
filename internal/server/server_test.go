package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cis-engine/internal/search"
	"cis-engine/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSearcher struct {
	mock.Mock
}

func (m *MockSearcher) Search(ctx context.Context, query string) ([]search.Result, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]search.Result), args.Error(1)
}

func (m *MockSearcher) ScheduleCrawl(ctx context.Context, url string) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockSearcher) GetStats(ctx context.Context) (*storage.Metrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*storage.Metrics), args.Error(1)
}

func TestSearchHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful search", func(t *testing.T) {
		mockSearcher := new(MockSearcher)
		handler := NewHandler(mockSearcher)
		router := NewRouter(handler)

		rr := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/search?q=test", nil)

		expectedResults := []search.Result{{URL: "http://example.com"}}
		mockSearcher.On("Search", mock.Anything, "test").Return(expectedResults, nil)

		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		mockSearcher.AssertExpectations(t)
	})

	t.Run("empty query", func(t *testing.T) {
		mockSearcher := new(MockSearcher)
		handler := NewHandler(mockSearcher)
		router := NewRouter(handler)

		rr := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/search", nil)

		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestCrawlHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful crawl", func(t *testing.T) {
		mockSearcher := new(MockSearcher)
		handler := NewHandler(mockSearcher)
		router := NewRouter(handler)

		rr := httptest.NewRecorder()
		body, _ := json.Marshal(gin.H{"url": "http://example.com"})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/crawl", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		mockSearcher.On("ScheduleCrawl", mock.Anything, "http://example.com").Return(nil)

		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)
		mockSearcher.AssertExpectations(t)
	})

	t.Run("invalid url", func(t *testing.T) {
		mockSearcher := new(MockSearcher)
		handler := NewHandler(mockSearcher)
		router := NewRouter(handler)

		rr := httptest.NewRecorder()
		body, _ := json.Marshal(gin.H{"url": "not a url"})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/crawl", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestStatusHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful status", func(t *testing.T) {
		mockSearcher := new(MockSearcher)
		handler := NewHandler(mockSearcher)
		router := NewRouter(handler)

		rr := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/status", nil)

		expectedMetrics := &storage.Metrics{PagesCount: 10}
		mockSearcher.On("GetStats", mock.Anything).Return(expectedMetrics, nil)

		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		mockSearcher.AssertExpectations(t)
	})
}