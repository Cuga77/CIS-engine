package postgres

import (
	"cis-engine/internal/storage"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	storePageQuery = `
		INSERT INTO pages (url, html_content, title, last_crawled_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (url) DO UPDATE
		SET html_content = EXCLUDED.html_content,
			title = EXCLUDED.title,
		    last_crawled_at = EXCLUDED.last_crawled_at,
			content_tsvector = NULL
		RETURNING id
	`
	getNextPageToIndexQuery = `SELECT id, url, title, html_content FROM pages WHERE content_tsvector IS NULL LIMIT 1`
	updatePageVectorQuery   = `
		UPDATE pages
		SET content_tsvector = to_tsvector('russian', coalesce(title, '') || ' ' || coalesce(html_content, ''))
		WHERE id = $1
	`
	searchPagesQuery = `
		SELECT
			id,
			url,
			title,
			ts_rank(content_tsvector, websearch_to_tsquery('russian', $1)) as rank
		FROM pages
		WHERE content_tsvector @@ websearch_to_tsquery('russian', $1)
		ORDER BY rank DESC
		LIMIT 20
	`
	getMetricsQuery = `SELECT COUNT(*) FROM pages`
)

type Querier interface {
	StorePage(ctx context.Context, arg *storage.Page) (int64, error)
	GetNextPageToIndex(ctx context.Context) (*storage.Page, error)
	UpdatePageVector(ctx context.Context, pageID int64) error
	SearchPages(ctx context.Context, query string) ([]*storage.Page, error)
	GetMetrics(ctx context.Context) (*storage.Metrics, error)
}

type Queries struct {
	db DBTX
}

func NewQuerier(db DBTX) *Queries {
	return &Queries{db: db}
}

func (q *Queries) StorePage(ctx context.Context, page *storage.Page) (int64, error) {
	var pageID int64
	err := q.db.QueryRow(ctx, storePageQuery, page.URL, page.Body, page.Title, page.CrawledAt).Scan(&pageID)
	if err != nil {
		return 0, fmt.Errorf("failed to store page %s: %w", page.URL, err)
	}
	return pageID, nil
}

func (q *Queries) GetNextPageToIndex(ctx context.Context) (*storage.Page, error) {
	var p storage.Page
	err := q.db.QueryRow(ctx, getNextPageToIndexQuery).Scan(&p.ID, &p.URL, &p.Title, &p.Body)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get next page to index: %w", err)
	}
	return &p, nil
}

func (q *Queries) UpdatePageVector(ctx context.Context, pageID int64) error {
	_, err := q.db.Exec(ctx, updatePageVectorQuery, pageID)
	if err != nil {
		return fmt.Errorf("failed to update page vector for page %d: %w", pageID, err)
	}
	return nil
}

func (q *Queries) SearchPages(ctx context.Context, query string) ([]*storage.Page, error) {
	rows, err := q.db.Query(ctx, searchPagesQuery, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute full-text search: %w", err)
	}
	defer rows.Close()

	var pages []*storage.Page
	for rows.Next() {
		var p storage.Page
		var rank float32
		if err := rows.Scan(&p.ID, &p.URL, &p.Title, &rank); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		pages = append(pages, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading search results: %w", err)
	}

	return pages, nil
}

func (q *Queries) GetMetrics(ctx context.Context) (*storage.Metrics, error) {
	var count int64
	err := q.db.QueryRow(ctx, getMetricsQuery).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}
	return &storage.Metrics{PagesCount: count}, nil
}