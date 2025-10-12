package postgres

import (
	"cis-engine/internal/storage"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type DB struct {
	pool *pgxpool.Pool
	q    *Queries
}

var _ storage.Storer = (*DB)(nil)

func New(ctx context.Context, connString string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &DB{pool: pool, q: NewQuerier(pool)}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) StorePage(ctx context.Context, page *storage.Page) (int64, error) {
	page.CrawledAt = time.Now()
	return db.q.StorePage(ctx, page)
}

func (db *DB) GetNextPageToIndex(ctx context.Context) (*storage.Page, error) {
	return db.q.GetNextPageToIndex(ctx)
}

func (db *DB) UpdatePageVector(ctx context.Context, page *storage.Page) error {
	return db.q.UpdatePageVector(ctx, page.ID)
}

func (db *DB) SearchPages(ctx context.Context, query string) ([]*storage.Page, error) {
	return db.q.SearchPages(ctx, query)
}

func (db *DB) GetMetrics(ctx context.Context) (*storage.Metrics, error) {
	return db.q.GetMetrics(ctx)
}