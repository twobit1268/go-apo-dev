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

func TestPostgres_ListParts_FiltersBySeedData(t *testing.T) {
	pg := testutil.NewTestPostgres(t)
	ctx := context.Background()

	parts, err := pg.ListParts(ctx, store.PartFilter{CategorySlug: "brakes"})
	require.NoError(t, err)
	assert.NotEmpty(t, parts)
	for _, p := range parts {
		// seed data (0002_seed.up.sql) SKUs the brakes category as BRK-* -
		// part names in that category vary ("Brake Pad", "Drilled & Slotted
		// Rotor"), so the SKU prefix is the reliable category signal.
		assert.Contains(t, p.SKU, "BRK-")
	}

	parts, err = pg.ListParts(ctx, store.PartFilter{Query: "oil"})
	require.NoError(t, err)
	require.Len(t, parts, 1)
	assert.Equal(t, testutil.SeedPartOilFilterID, parts[0].ID)
}

func TestPostgres_GetPart_NotFound(t *testing.T) {
	pg := testutil.NewTestPostgres(t)

	_, err := pg.GetPart(context.Background(), "00000000-0000-0000-0000-000000000000")
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func TestPostgres_DecrementStock(t *testing.T) {
	pg := testutil.NewTestPostgres(t)
	ctx := context.Background()

	before, err := pg.GetPart(ctx, testutil.SeedPartOilFilterID)
	require.NoError(t, err)

	require.NoError(t, pg.DecrementStock(ctx, testutil.SeedPartOilFilterID, 5))

	after, err := pg.GetPart(ctx, testutil.SeedPartOilFilterID)
	require.NoError(t, err)
	assert.Equal(t, before.StockQty-5, after.StockQty)

	// decrementing more than is in stock must fail without changing stock
	err = pg.DecrementStock(ctx, testutil.SeedPartOilFilterID, before.StockQty*1000)
	assert.ErrorIs(t, err, store.ErrInsufficientStock)

	unchanged, err := pg.GetPart(ctx, testutil.SeedPartOilFilterID)
	require.NoError(t, err)
	assert.Equal(t, after.StockQty, unchanged.StockQty)
}
