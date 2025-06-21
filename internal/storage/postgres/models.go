package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул соединений: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) StorePage(ctx context.Context, page *storage.Page) (int64, error) {
	query := `
		INSERT INTO pages (url, html_content, last_crawled_at, is_indexed)
		VALUES ($1, $2, $3, FALSE)
		ON CONFLICT (url) DO UPDATE
		SET html_content = EXCLUDED.html_content,
		    last_crawled_at = EXCLUDED.last_crawled_at,
		    is_indexed = FALSE
		RETURNING id
	`
	var pageID int64
	err := db.pool.QueryRow(ctx, query, page.URL, page.Body, time.Now()).Scan(&pageID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при сохранении страницы %s: %w", page.URL, err)
	}
	return pageID, nil
}

func (db *DB) GetNextPageToIndex(ctx context.Context) (*storage.Page, error) {
	query := `
		SELECT id, url, html_content, last_crawled_at
		FROM pages
		WHERE is_indexed = FALSE
		ORDER BY id
		LIMIT 1
	`
	var p storage.Page
	err := db.pool.QueryRow(ctx, query).Scan(&p.ID, &p.URL, &p.Body, &p.CrawledAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) UpdatePageIndexed(ctx context.Context, pageID int64) error {
	query := `UPDATE pages SET is_indexed = TRUE WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, pageID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении статуса индексации для страницы %d: %w", pageID, err)
	}
	return nil
}

// todo: заглушки
func (db *DB) FindDocumentsByTerm(ctx context.Context, term string) ([]int64, error) {
	return nil, fmt.Errorf("метод FindDocumentsByTerm еще не реализован")
}

func (db *DB) GetPagesByIDs(ctx context.Context, ids []int64) ([]*storage.Page, error) {
	return nil, fmt.Errorf("метод GetPagesByIDs еще не реализован")
}
