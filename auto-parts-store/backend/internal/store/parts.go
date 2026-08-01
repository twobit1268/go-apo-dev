package store

import (
	"context"
	"errors"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) ListParts(ctx context.Context, filter PartFilter) ([]domain.Part, error) {
	query := `
		SELECT p.id, p.sku, p.name, p.description, p.category_id, p.price_cents, p.stock_qty
		FROM parts p
		JOIN categories c ON c.id = p.category_id
		WHERE ($1 = '' OR c.slug = $1)
		  AND ($2 = '' OR p.name ILIKE '%' || $2 || '%')
		ORDER BY p.name`

	rows, err := p.pool.Query(ctx, query, filter.CategorySlug, filter.Query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Part{} // never nil - encodes as [] not null over the API
	for rows.Next() {
		var part domain.Part
		if err := rows.Scan(&part.ID, &part.SKU, &part.Name, &part.Description,
			&part.CategoryID, &part.PriceCents, &part.StockQty); err != nil {
			return nil, err
		}
		out = append(out, part)
	}
	return out, rows.Err()
}

func (p *Postgres) GetPart(ctx context.Context, id string) (domain.Part, error) {
	var part domain.Part
	err := p.pool.QueryRow(ctx, `
		SELECT id, sku, name, description, category_id, price_cents, stock_qty
		FROM parts WHERE id = $1`, id,
	).Scan(&part.ID, &part.SKU, &part.Name, &part.Description,
		&part.CategoryID, &part.PriceCents, &part.StockQty)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Part{}, ErrNotFound
	}
	if err != nil {
		return domain.Part{}, err
	}
	return part, nil
}

func (p *Postgres) DecrementStock(ctx context.Context, partID string, qty int) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE parts SET stock_qty = stock_qty - $1
		WHERE id = $2 AND stock_qty >= $1`, qty, partID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		// Either the part doesn't exist, or the CHECK on stock_qty>=0 would
		// have been violated - either way we couldn't fulfill the decrement.
		if _, err := p.GetPart(ctx, partID); errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return ErrInsufficientStock
	}
	return nil
}
