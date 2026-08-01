package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/dl9346/auto-parts-store/backend/internal/pubsub"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
)

var ErrCartCustomerMismatch = errors.New("cart does not belong to customer")

type CheckoutService struct {
	carts     store.CartStore
	orders    store.OrderStore
	publisher pubsub.Publisher
}

func NewCheckoutService(carts store.CartStore, orders store.OrderStore, publisher pubsub.Publisher) *CheckoutService {
	return &CheckoutService{carts: carts, orders: orders, publisher: publisher}
}

// PlaceOrder turns a cart into an order and publishes an OrderPlaced event
// so the inventory and notification subscribers can react asynchronously.
// A publish failure is logged but does not fail the request or roll back
// the purchase - the order already committed, and the caller shouldn't be
// told their order failed when it didn't.
func (s *CheckoutService) PlaceOrder(ctx context.Context, cartID, customerID string) (domain.Order, error) {
	cart, err := s.carts.GetCart(ctx, cartID)
	if err != nil {
		return domain.Order{}, err
	}
	if cart.CustomerID != customerID {
		return domain.Order{}, ErrCartCustomerMismatch
	}

	order, err := s.orders.CreateOrderFromCart(ctx, cartID, customerID)
	if err != nil {
		return domain.Order{}, err
	}

	event := pubsub.OrderPlaced{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		Items:      make([]pubsub.OrderItem, len(order.Items)),
	}
	for i, item := range order.Items {
		event.Items[i] = pubsub.OrderItem{PartID: item.PartID, Quantity: item.Quantity}
	}
	if err := s.publisher.PublishOrderPlaced(ctx, event); err != nil {
		slog.Error("failed to publish OrderPlaced event", "orderId", order.ID, "error", err)
	}

	return order, nil
}

func (s *CheckoutService) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	return s.orders.GetOrder(ctx, id)
}

func (s *CheckoutService) ListOrdersByCustomer(ctx context.Context, customerID string) ([]domain.Order, error) {
	return s.orders.ListOrdersByCustomer(ctx, customerID)
}
