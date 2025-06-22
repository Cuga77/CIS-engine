package storage

import (
	"context"
	"time"
)

type Page struct {
	ID        int64
	URL       string
	Title     string
	Body      string
	CrawledAt time.Time
}

type Storer interface {
	StorePage(ctx context.Context, page *Page) (int64, error)
	GetNextPageToIndex(ctx context.Context) (*Page, error)
	UpdatePageVector(ctx context.Context, page *Page) error
	SearchPages(ctx context.Context, query string) ([]*Page, error)
	Close()
}
