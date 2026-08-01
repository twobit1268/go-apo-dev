//go:build integration

package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dl9346/auto-parts-store/backend/internal/store"
	"github.com/dl9346/auto-parts-store/backend/internal/testutil"
)

func TestPostgres_CreateOrderFromCart(t *testing.T) {
	pg := testutil.NewTestPostgres(t)
	ctx := context.Background()

	cart, err := pg.CreateCart(ctx, "cust-integration-2")
	require.NoError(t, err)
	require.NoError(t, pg.AddItem(ctx, cart.ID, testutil.SeedPartOilFilterID, 4))
	require.NoError(t, pg.AddItem(ctx, cart.ID, testutil.SeedPartBrakePadID, 1))

	order, err := pg.CreateOrderFromCart(ctx, cart.ID, "cust-integration-2")
	require.NoError(t, err)

	wantTotal := int64(4*testutil.SeedPartOilFilterPrice + 1*testutil.SeedPartBrakePadPrice)
	assert.Equal(t, wantTotal, order.TotalCents)
	assert.Len(t, order.Items, 2)

	// cart should be emptied by checkout
	emptied, err := pg.GetCart(ctx, cart.ID)
	require.NoError(t, err)
	assert.Empty(t, emptied.Items)

	// order is independently fetchable
	fetched, err := pg.GetOrder(ctx, order.ID)
	require.NoError(t, err)
	assert.Equal(t, order.TotalCents, fetched.TotalCents)
	assert.Len(t, fetched.Items, 2)
}

func TestPostgres_CreateOrderFromCart_EmptyCart(t *testing.T) {
	pg := testutil.NewTestPostgres(t)
	ctx := context.Background()

	cart, err := pg.CreateCart(ctx, "cust-integration-3")
	require.NoError(t, err)

	_, err = pg.CreateOrderFromCart(ctx, cart.ID, "cust-integration-3")
	assert.ErrorIs(t, err, store.ErrEmptyCart)
}

func TestPostgres_ListOrdersByCustomer(t *testing.T) {
	pg := testutil.NewTestPostgres(t)
	ctx := context.Background()

	customerID := "cust-integration-4"
	cart, err := pg.CreateCart(ctx, customerID)
	require.NoError(t, err)
	require.NoError(t, pg.AddItem(ctx, cart.ID, testutil.SeedPartOilFilterID, 1))
	_, err = pg.CreateOrderFromCart(ctx, cart.ID, customerID)
	require.NoError(t, err)

	orders, err := pg.ListOrdersByCustomer(ctx, customerID)
	require.NoError(t, err)
	require.Len(t, orders, 1)
	assert.Equal(t, customerID, orders[0].CustomerID)
}
