//go:build integration

package pubsub_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dl9346/auto-parts-store/backend/internal/pubsub"
	"github.com/dl9346/auto-parts-store/backend/internal/service"
	"github.com/dl9346/auto-parts-store/backend/internal/testutil"
)

// TestPubSub_OrderPlaced_RoundTrip publishes an OrderPlaced event against
// the real Pub/Sub emulator and asserts the inventory subscriber (wired to
// a real Postgres) actually decrements stock - proving the async plumbing
// end to end, not just each half of it in isolation.
func TestPubSub_OrderPlaced_RoundTrip(t *testing.T) {
	pg := testutil.NewTestPostgres(t)
	client := testutil.NewTestPubSubClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	before, err := pg.GetPart(ctx, testutil.SeedPartOilFilterID)
	require.NoError(t, err)

	inventory := service.NewInventoryHandler(pg)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = client.Subscribe(ctx, pubsub.InventorySubscription, func(hctx context.Context, event pubsub.OrderPlaced) error {
			err := inventory.Handle(hctx, event)
			cancel() // stop the subscriber after handling the one message we expect
			return err
		})
	}()

	event := pubsub.OrderPlaced{
		OrderID:    "order-roundtrip-1",
		CustomerID: "cust-roundtrip-1",
		Items:      []pubsub.OrderItem{{PartID: testutil.SeedPartOilFilterID, Quantity: 2}},
	}
	require.NoError(t, client.PublishOrderPlaced(context.Background(), event))

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("timed out waiting for inventory subscriber to process OrderPlaced event")
	}

	after, err := pg.GetPart(context.Background(), testutil.SeedPartOilFilterID)
	require.NoError(t, err)
	assert.Equal(t, before.StockQty-2, after.StockQty)
}
