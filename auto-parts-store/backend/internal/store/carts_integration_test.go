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

func TestPostgres_CartLifecycle(t *testing.T) {
	pg := testutil.NewTestPostgres(t)
	ctx := context.Background()

	cart, err := pg.CreateCart(ctx, "cust-integration-1")
	require.NoError(t, err)
	assert.Equal(t, "cust-integration-1", cart.CustomerID)
	assert.Empty(t, cart.Items)

	require.NoError(t, pg.AddItem(ctx, cart.ID, testutil.SeedPartOilFilterID, 2))
	require.NoError(t, pg.AddItem(ctx, cart.ID, testutil.SeedPartOilFilterID, 3)) // accumulates

	got, err := pg.GetCart(ctx, cart.ID)
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, 5, got.Items[0].Quantity)

	require.NoError(t, pg.RemoveItem(ctx, cart.ID, testutil.SeedPartOilFilterID))

	got, err = pg.GetCart(ctx, cart.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Items)
}

func TestPostgres_GetCart_NotFound(t *testing.T) {
	pg := testutil.NewTestPostgres(t)

	_, err := pg.GetCart(context.Background(), "00000000-0000-0000-0000-000000000000")
	assert.ErrorIs(t, err, store.ErrNotFound)
}
