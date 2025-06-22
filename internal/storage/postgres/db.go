package postgres

import (
	"cis-engine/internal/storage"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

var _ storage.Storer = (*DB)(nil)

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
		INSERT INTO pages (url, html_content, title, last_crawled_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (url) DO UPDATE
		SET html_content = EXCLUDED.html_content,
			title = EXCLUDED.title,
		    last_crawled_at = EXCLUDED.last_crawled_at,
			-- Сбрасываем tsvector при обновлении, чтобы страница переиндексировалась
			content_tsvector = NULL 
		RETURNING id
	`
	var pageID int64
	err := db.pool.QueryRow(ctx, query, page.URL, page.Body, page.Title, time.Now()).Scan(&pageID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при сохранении страницы %s: %w", page.URL, err)
	}
	return pageID, nil
}

func (db *DB) GetNextPageToIndex(ctx context.Context) (*storage.Page, error) {
	query := `SELECT id, url, title, html_content FROM pages WHERE content_tsvector IS NULL LIMIT 1`
	var p storage.Page
	err := db.pool.QueryRow(ctx, query).Scan(&p.ID, &p.URL, &p.Title, &p.Body)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка при получении страницы для индексации: %w", err)
	}
	return &p, nil
}

func (db *DB) UpdatePageVector(ctx context.Context, page *storage.Page) error {
	query := `
		UPDATE pages
		-- Используем coalesce для объединения заголовка и контента, чтобы искать по обоим
		SET content_tsvector = to_tsvector('russian', coalesce(title, '') || ' ' || coalesce(html_content, ''))
		WHERE id = $1
	`
	_, err := db.pool.Exec(ctx, query, page.ID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении tsvector для страницы %d: %w", page.ID, err)
	}
	return nil
}

func (db *DB) SearchPages(ctx context.Context, query string) ([]*storage.Page, error) {
	sql := `
		SELECT 
			id, 
			url, 
			title,
			-- ts_rank вычисляет релевантность документа запросу
			ts_rank(content_tsvector, websearch_to_tsquery('russian', $1)) as rank
		FROM pages
		WHERE content_tsvector @@ websearch_to_tsquery('russian', $1)
		ORDER BY rank DESC
		LIMIT 20
	`
	rows, err := db.pool.Query(ctx, sql, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении полнотекстового поиска: %w", err)
	}
	defer rows.Close()

	var pages []*storage.Page
	for rows.Next() {
		var p storage.Page
		var rank float32 // ts_rank возвращает float, но мы его пока не используем
		if err := rows.Scan(&p.ID, &p.URL, &p.Title, &rank); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании результата поиска: %w", err)
		}
		pages = append(pages, &p)
	}

	return pages, rows.Err()
}
