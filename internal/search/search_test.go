package search

import (
	"context"
	"testing"

	"cis-engine/internal/storage"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStorer struct {
	mock.Mock
}

func (m *MockStorer) StorePage(ctx context.Context, page *storage.Page) (int64, error) {
	args := m.Called(ctx, page)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorer) GetNextPageToIndex(ctx context.Context) (*storage.Page, error) {
	args := m.Called(ctx)
	return args.Get(0).(*storage.Page), args.Error(1)
}

func (m *MockStorer) UpdatePageVector(ctx context.Context, page *storage.Page) error {
	args := m.Called(ctx, page)
	return args.Error(0)
}

func (m *MockStorer) SearchPages(ctx context.Context, query string) ([]*storage.Page, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*storage.Page), args.Error(1)
}

func (m *MockStorer) GetMetrics(ctx context.Context) (*storage.Metrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*storage.Metrics), args.Error(1)
}

func (m *MockStorer) Close() {}

func TestSearch(t *testing.T) {
	mockStorer := new(MockStorer)
	service := NewService(mockStorer)
	ctx := context.Background()

	t.Run("successful search", func(t *testing.T) {
		expectedPages := []*storage.Page{{URL: "http://example.com", Title: "Example"}}
		mockStorer.On("SearchPages", ctx, "test").Return(expectedPages, nil).Once()

		results, err := service.Search(ctx, "test")
		require.NoError(t, err)
		require.Len(t, results, 1)
		require.Equal(t, "http://example.com", results[0].URL)
		mockStorer.AssertExpectations(t)
	})

	t.Run("empty query", func(t *testing.T) {
		results, err := service.Search(ctx, "")
		require.NoError(t, err)
		require.Empty(t, results)
	})
}

func TestScheduleCrawl(t *testing.T) {
	mockStorer := new(MockStorer)
	service := NewService(mockStorer)
	ctx := context.Background()

	page := &storage.Page{URL: "http://example.com"}
	mockStorer.On("StorePage", ctx, page).Return(int64(1), nil).Once()

	err := service.ScheduleCrawl(ctx, "http://example.com")
	require.NoError(t, err)
	mockStorer.AssertExpectations(t)
}

func TestGetStats(t *testing.T) {
	mockStorer := new(MockStorer)
	service := NewService(mockStorer)
	ctx := context.Background()

	expectedMetrics := &storage.Metrics{PagesCount: 10}
	mockStorer.On("GetMetrics", ctx).Return(expectedMetrics, nil).Once()

	metrics, err := service.GetStats(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedMetrics, metrics)
	mockStorer.AssertExpectations(t)
}