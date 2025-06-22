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
		INSERT INTO pages (url, html_content, title, last_crawled_at, is_indexed)
		VALUES ($1, $2, $3, $4, FALSE)
		ON CONFLICT (url) DO UPDATE
		SET html_content = EXCLUDED.html_content,
			title = EXCLUDED.title,
		    last_crawled_at = EXCLUDED.last_crawled_at,
		    is_indexed = FALSE
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка при получении страницы для индексации: %w", err)
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

func (db *DB) FindDocumentsByTerm(ctx context.Context, term string) ([]int64, error) {
	query := `
		SELECT p.page_id
		FROM postings p
		JOIN terms t ON p.term_id = t.id
		WHERE t.term = $1
	`
	rows, err := db.pool.Query(ctx, query, term)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске документов по термину '%s': %w", term, err)
	}
	defer rows.Close()

	var pageIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании ID страницы: %w", err)
		}
		pageIDs = append(pageIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по строкам: %w", err)
	}

	return pageIDs, nil
}

func (db *DB) GetPagesByIDs(ctx context.Context, ids []int64) ([]*storage.Page, error) {
	if len(ids) == 0 {
		return []*storage.Page{}, nil
	}

	query := `
		SELECT id, url, title, last_crawled_at
		FROM pages
		WHERE id = ANY($1)
	`
	rows, err := db.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении страниц по IDs: %w", err)
	}
	defer rows.Close()

	pages := make([]*storage.Page, 0, len(ids))
	for rows.Next() {
		var p storage.Page
		// Обратите внимание, мы не запрашиваем html_content,
		// чтобы не передавать большие объемы текста.
		err := rows.Scan(&p.ID, &p.URL, &p.Title, &p.CrawledAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании страницы: %w", err)
		}
		pages = append(pages, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по строкам: %w", err)
	}

	return pages, nil
}
