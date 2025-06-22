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
	UpdatePageIndexed(ctx context.Context, pageID int64) error
	FindDocumentsByTerm(ctx context.Context, term string) ([]int64, error)
	GetPagesByIDs(ctx context.Context, ids []int64) ([]*Page, error)
	Close()
}
