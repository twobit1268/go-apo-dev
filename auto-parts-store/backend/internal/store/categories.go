package store

import (
	"context"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
)

func (p *Postgres) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := p.pool.Query(ctx, `SELECT id, name, slug FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Category{} // never nil - encodes as [] not null over the API
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
