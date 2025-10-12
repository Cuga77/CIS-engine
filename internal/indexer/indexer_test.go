package indexer

import (
	"context"
	"testing"
	"time"

	"cis-engine/internal/storage"

	"github.com/stretchr/testify/mock"
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
	// Handle the case where the first return value might be nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

func TestIndexer(t *testing.T) {
	mockStorer := new(MockStorer)
	indexer := NewIndexer(mockStorer, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	page := &storage.Page{ID: 1, URL: "http://example.com"}

	// Expect GetNextPageToIndex to be called, return a page, and then expect UpdatePageVector to be called.
	mockStorer.On("GetNextPageToIndex", ctx).Return(page, nil).Once()
	mockStorer.On("UpdatePageVector", ctx, page).Return(nil).Once()

	// Then, expect GetNextPageToIndex to be called again, but this time return nil to stop the loop.
	mockStorer.On("GetNextPageToIndex", ctx).Return(nil, nil).Once()

	go indexer.Start(ctx)

	// Allow the indexer to run for a short period.
	time.Sleep(50 * time.Millisecond)
	cancel() // Stop the indexer.

	mockStorer.AssertExpectations(t)
}