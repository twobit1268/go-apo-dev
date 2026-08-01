package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres implements CategoryStore, PartStore, CartStore, and OrderStore
// against a real Postgres database via pgx.
type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

// ResetTransactionalData truncates carts/orders (and their line items) but
// leaves categories/parts alone, since those are seeded reference data that
// tests key off of by ID. Used between test cases to isolate them.
func (p *Postgres) ResetTransactionalData(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, `TRUNCATE cart_items, carts, order_items, orders RESTART IDENTITY CASCADE`)
	return err
}
