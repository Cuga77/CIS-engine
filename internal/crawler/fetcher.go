package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Fetcher interface {
	Fetch(ctx context.Context, url string) (io.ReadCloser, error)
}

type HTTPFetcher struct {
	client *http.Client
}

func NewHTTPFetcher(timeout time.Duration) *HTTPFetcher {
	return &HTTPFetcher{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	req.Header.Set("User-Agent", "CIS-Engine-Crawler/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("bad status code for %s: %s", url, resp.Status)
	}

	return resp.Body, nil
}
