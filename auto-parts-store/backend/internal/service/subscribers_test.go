package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/dl9346/auto-parts-store/backend/internal/pubsub"
	"github.com/dl9346/auto-parts-store/backend/internal/service"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
	"github.com/dl9346/auto-parts-store/backend/internal/testutil"
)

func TestInventoryHandler_Handle_DecrementsStockPerLineItem(t *testing.T) {
	fake := testutil.NewFakeStore()
	fake.SeedPart(domain.Part{ID: "p1", Name: "Brake Pad", PriceCents: 4899, StockQty: 10})
	fake.SeedPart(domain.Part{ID: "p2", Name: "Oil Filter", PriceCents: 899, StockQty: 5})

	handler := service.NewInventoryHandler(fake)
	event := pubsub.OrderPlaced{
		OrderID:    "order-1",
		CustomerID: "cust-1",
		Items: []pubsub.OrderItem{
			{PartID: "p1", Quantity: 3},
			{PartID: "p2", Quantity: 2},
		},
	}

	require.NoError(t, handler.Handle(context.Background(), event))

	p1, err := fake.GetPart(context.Background(), "p1")
	require.NoError(t, err)
	assert.Equal(t, 7, p1.StockQty)

	p2, err := fake.GetPart(context.Background(), "p2")
	require.NoError(t, err)
	assert.Equal(t, 3, p2.StockQty)
}

func TestInventoryHandler_Handle_InsufficientStock(t *testing.T) {
	fake := testutil.NewFakeStore()
	fake.SeedPart(domain.Part{ID: "p1", Name: "Brake Pad", PriceCents: 4899, StockQty: 1})

	handler := service.NewInventoryHandler(fake)
	event := pubsub.OrderPlaced{
		Items: []pubsub.OrderItem{{PartID: "p1", Quantity: 5}},
	}

	err := handler.Handle(context.Background(), event)
	assert.ErrorIs(t, err, store.ErrInsufficientStock)
}

func TestNotificationHandler_Handle_NeverErrors(t *testing.T) {
	handler := service.NewNotificationHandler()
	event := pubsub.OrderPlaced{OrderID: "order-1", CustomerID: "cust-1"}

	assert.NoError(t, handler.Handle(context.Background(), event))
}
