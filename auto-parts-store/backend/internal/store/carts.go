package store

import (
	"context"
	"errors"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) CreateCart(ctx context.Context, customerID string) (domain.Cart, error) {
	var cart domain.Cart
	err := p.pool.QueryRow(ctx, `
		INSERT INTO carts (customer_id) VALUES ($1)
		RETURNING id, customer_id, created_at`, customerID,
	).Scan(&cart.ID, &cart.CustomerID, &cart.CreatedAt)
	if err != nil {
		return domain.Cart{}, err
	}
	cart.Items = []domain.CartItem{}
	return cart, nil
}

func (p *Postgres) GetCart(ctx context.Context, id string) (domain.Cart, error) {
	var cart domain.Cart
	err := p.pool.QueryRow(ctx, `
		SELECT id, customer_id, created_at FROM carts WHERE id = $1`, id,
	).Scan(&cart.ID, &cart.CustomerID, &cart.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Cart{}, ErrNotFound
	}
	if err != nil {
		return domain.Cart{}, err
	}

	rows, err := p.pool.Query(ctx, `
		SELECT part_id, quantity FROM cart_items WHERE cart_id = $1 ORDER BY part_id`, id)
	if err != nil {
		return domain.Cart{}, err
	}
	defer rows.Close()

	cart.Items = []domain.CartItem{}
	for rows.Next() {
		var item domain.CartItem
		if err := rows.Scan(&item.PartID, &item.Quantity); err != nil {
			return domain.Cart{}, err
		}
		cart.Items = append(cart.Items, item)
	}
	return cart, rows.Err()
}

func (p *Postgres) AddItem(ctx context.Context, cartID, partID string, quantity int) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO cart_items (cart_id, part_id, quantity) VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, part_id)
		DO UPDATE SET quantity = cart_items.quantity + excluded.quantity`,
		cartID, partID, quantity)
	return err
}

func (p *Postgres) RemoveItem(ctx context.Context, cartID, partID string) error {
	_, err := p.pool.Exec(ctx, `
		DELETE FROM cart_items WHERE cart_id = $1 AND part_id = $2`, cartID, partID)
	return err
}
