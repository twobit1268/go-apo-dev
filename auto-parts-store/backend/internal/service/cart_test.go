package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/dl9346/auto-parts-store/backend/internal/service"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
	"github.com/dl9346/auto-parts-store/backend/internal/testutil"
)

func TestCartService_AddItem(t *testing.T) {
	ctx := context.Background()

	t.Run("adds a new item", func(t *testing.T) {
		fake := testutil.NewFakeStore()
		fake.SeedPart(domain.Part{ID: "p1", Name: "Brake Pad", PriceCents: 4899, StockQty: 10})
		svc := service.NewCartService(fake, fake)

		cart, err := svc.CreateCart(ctx, "cust-1")
		require.NoError(t, err)

		err = svc.AddItem(ctx, cart.ID, "p1", 2)
		require.NoError(t, err)

		got, err := svc.GetCart(ctx, cart.ID)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, "p1", got.Items[0].PartID)
		assert.Equal(t, 2, got.Items[0].Quantity)
	})

	t.Run("accumulates quantity when adding the same part twice", func(t *testing.T) {
		fake := testutil.NewFakeStore()
		fake.SeedPart(domain.Part{ID: "p1", Name: "Brake Pad", PriceCents: 4899, StockQty: 10})
		svc := service.NewCartService(fake, fake)
		cart, _ := svc.CreateCart(ctx, "cust-1")

		require.NoError(t, svc.AddItem(ctx, cart.ID, "p1", 2))
		require.NoError(t, svc.AddItem(ctx, cart.ID, "p1", 3))

		got, err := svc.GetCart(ctx, cart.ID)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, 5, got.Items[0].Quantity)
	})

	t.Run("rejects zero or negative quantity", func(t *testing.T) {
		fake := testutil.NewFakeStore()
		fake.SeedPart(domain.Part{ID: "p1", Name: "Brake Pad", PriceCents: 4899, StockQty: 10})
		svc := service.NewCartService(fake, fake)
		cart, _ := svc.CreateCart(ctx, "cust-1")

		err := svc.AddItem(ctx, cart.ID, "p1", 0)
		assert.ErrorIs(t, err, service.ErrInvalidQuantity)

		err = svc.AddItem(ctx, cart.ID, "p1", -1)
		assert.ErrorIs(t, err, service.ErrInvalidQuantity)
	})

	t.Run("rejects unknown part", func(t *testing.T) {
		fake := testutil.NewFakeStore()
		svc := service.NewCartService(fake, fake)
		cart, _ := svc.CreateCart(ctx, "cust-1")

		err := svc.AddItem(ctx, cart.ID, "missing-part", 1)
		assert.ErrorIs(t, err, store.ErrNotFound)
	})

	t.Run("rejects unknown cart", func(t *testing.T) {
		fake := testutil.NewFakeStore()
		fake.SeedPart(domain.Part{ID: "p1", Name: "Brake Pad", PriceCents: 4899, StockQty: 10})
		svc := service.NewCartService(fake, fake)

		err := svc.AddItem(ctx, "missing-cart", "p1", 1)
		assert.ErrorIs(t, err, store.ErrNotFound)
	})
}

func TestCartService_RemoveItem(t *testing.T) {
	fake := testutil.NewFakeStore()
	fake.SeedPart(domain.Part{ID: "p1", Name: "Brake Pad", PriceCents: 4899, StockQty: 10})
	svc := service.NewCartService(fake, fake)
	ctx := context.Background()

	cart, _ := svc.CreateCart(ctx, "cust-1")
	require.NoError(t, svc.AddItem(ctx, cart.ID, "p1", 2))

	require.NoError(t, svc.RemoveItem(ctx, cart.ID, "p1"))

	got, err := svc.GetCart(ctx, cart.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Items)
}
