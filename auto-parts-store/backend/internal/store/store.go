// Package store defines the repository interfaces the service layer depends
// on, plus their Postgres implementations. Interfaces live here (not in
// service) so unit tests can supply in-memory fakes without importing pgx.
package store

import (
	"context"
	"errors"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
)

// ErrNotFound is returned by any store method when the requested row
// doesn't exist. Callers translate it to an HTTP 404.
var ErrNotFound = errors.New("not found")

type PartFilter struct {
	CategorySlug string
	Query        string
}

type CategoryStore interface {
	ListCategories(ctx context.Context) ([]domain.Category, error)
}

type PartStore interface {
	ListParts(ctx context.Context, filter PartFilter) ([]domain.Part, error)
	GetPart(ctx context.Context, id string) (domain.Part, error)
	// DecrementStock reduces stock_qty by qty, failing with ErrInsufficientStock
	// if that would take it negative.
	DecrementStock(ctx context.Context, partID string, qty int) error
}

var ErrInsufficientStock = errors.New("insufficient stock")

type CartStore interface {
	CreateCart(ctx context.Context, customerID string) (domain.Cart, error)
	GetCart(ctx context.Context, id string) (domain.Cart, error)
	AddItem(ctx context.Context, cartID, partID string, quantity int) error
	RemoveItem(ctx context.Context, cartID, partID string) error
}

type OrderStore interface {
	// CreateOrderFromCart atomically reads the cart's items, snapshots part
	// prices, inserts the order + order_items, and returns the created order.
	// It does NOT touch stock_qty - that's the inventory subscriber's job,
	// triggered by the OrderPlaced event this creates.
	CreateOrderFromCart(ctx context.Context, cartID, customerID string) (domain.Order, error)
	GetOrder(ctx context.Context, id string) (domain.Order, error)
	ListOrdersByCustomer(ctx context.Context, customerID string) ([]domain.Order, error)
}
