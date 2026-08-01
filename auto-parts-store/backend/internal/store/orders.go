package store

import (
	"context"
	"errors"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

var ErrEmptyCart = errors.New("cart is empty")

func (p *Postgres) CreateOrderFromCart(ctx context.Context, cartID, customerID string) (domain.Order, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT ci.part_id, ci.quantity, p.price_cents
		FROM cart_items ci
		JOIN parts p ON p.id = ci.part_id
		WHERE ci.cart_id = $1`, cartID)
	if err != nil {
		return domain.Order{}, err
	}

	var items []domain.OrderItem
	var total int64
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.PartID, &item.Quantity, &item.UnitPriceCents); err != nil {
			rows.Close()
			return domain.Order{}, err
		}
		total += item.UnitPriceCents * int64(item.Quantity)
		items = append(items, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return domain.Order{}, err
	}
	if len(items) == 0 {
		return domain.Order{}, ErrEmptyCart
	}

	var order domain.Order
	order.Status = domain.OrderStatusPlaced
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (customer_id, status, total_cents) VALUES ($1, $2, $3)
		RETURNING id, customer_id, status, total_cents, created_at`,
		customerID, order.Status, total,
	).Scan(&order.ID, &order.CustomerID, &order.Status, &order.TotalCents, &order.CreatedAt)
	if err != nil {
		return domain.Order{}, err
	}

	for _, item := range items {
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_items (order_id, part_id, quantity, unit_price_cents)
			VALUES ($1, $2, $3, $4)`,
			order.ID, item.PartID, item.Quantity, item.UnitPriceCents); err != nil {
			return domain.Order{}, err
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID); err != nil {
		return domain.Order{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, err
	}

	order.Items = items
	return order, nil
}

func (p *Postgres) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	var order domain.Order
	err := p.pool.QueryRow(ctx, `
		SELECT id, customer_id, status, total_cents, created_at FROM orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.CustomerID, &order.Status, &order.TotalCents, &order.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, ErrNotFound
	}
	if err != nil {
		return domain.Order{}, err
	}

	items, err := p.orderItems(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	order.Items = items
	return order, nil
}

func (p *Postgres) ListOrdersByCustomer(ctx context.Context, customerID string) ([]domain.Order, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, customer_id, status, total_cents, created_at
		FROM orders WHERE customer_id = $1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []domain.Order{} // never nil - encodes as [] not null over the API
	for rows.Next() {
		var order domain.Order
		if err := rows.Scan(&order.ID, &order.CustomerID, &order.Status,
			&order.TotalCents, &order.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range orders {
		items, err := p.orderItems(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}
	return orders, nil
}

func (p *Postgres) orderItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT part_id, quantity, unit_price_cents FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.OrderItem{}
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.PartID, &item.Quantity, &item.UnitPriceCents); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
