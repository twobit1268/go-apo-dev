package service

import (
	"context"
	"log/slog"

	"github.com/dl9346/auto-parts-store/backend/internal/pubsub"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
)

// InventoryHandler decrements stock for each line item of a placed order.
// Wired up as the handler for pubsub.InventorySubscription in cmd/worker.
type InventoryHandler struct {
	parts store.PartStore
}

func NewInventoryHandler(parts store.PartStore) *InventoryHandler {
	return &InventoryHandler{parts: parts}
}

func (h *InventoryHandler) Handle(ctx context.Context, event pubsub.OrderPlaced) error {
	for _, item := range event.Items {
		if err := h.parts.DecrementStock(ctx, item.PartID, item.Quantity); err != nil {
			return err
		}
	}
	return nil
}

// NotificationHandler stands in for a real email/SMS provider: it just logs
// that a confirmation would have been sent. Wired up as the handler for
// pubsub.NotificationsSubscription in cmd/worker.
type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

func (h *NotificationHandler) Handle(_ context.Context, event pubsub.OrderPlaced) error {
	slog.Info("order confirmation sent", "orderId", event.OrderID, "customerId", event.CustomerID)
	return nil
}
