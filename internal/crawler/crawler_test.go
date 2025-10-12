package crawler

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"

	"cis-engine/internal/storage"

	"github.com/stretchr/testify/mock"
)

type MockFetcher struct {
	mock.Mock
}

func (m *MockFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

type MockStorer struct {
	mock.Mock
}

func (m *MockStorer) StorePage(ctx context.Context, page *storage.Page) (int64, error) {
	args := m.Called(ctx, page)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorer) GetNextPageToIndex(ctx context.Context) (*storage.Page, error) {
	args := m.Called(ctx)
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

func TestCrawler(t *testing.T) {
	mockFetcher := new(MockFetcher)
	mockStorer := new(MockStorer)
	crawler := NewCrawler(1, 10, mockStorer, mockFetcher)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	seedURL := "http://example.com"
	htmlContent := `
		<html>
			<head><title>Example</title></head>
			<body>
			</body>
		</html>`

	var wg sync.WaitGroup
	wg.Add(1) // Expect one page to be stored

	mockFetcher.On("Fetch", ctx, seedURL).Return(io.NopCloser(strings.NewReader(htmlContent)), nil).Once()
	mockStorer.On("StorePage", ctx, mock.Anything).Run(func(args mock.Arguments) {
		wg.Done()
	}).Return(int64(1), nil).Once()

	go crawler.Start(ctx, []string{seedURL})

	wg.Wait()
	crawler.Stop()

	mockFetcher.AssertExpectations(t)
	mockStorer.AssertExpectations(t)
}