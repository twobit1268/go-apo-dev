package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/dl9346/auto-parts-store/backend/internal/service"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
	"github.com/dl9346/auto-parts-store/backend/internal/testutil"
)

func newCheckoutFixture() (*testutil.FakeStore, *testutil.FakePublisher, *service.CheckoutService) {
	fake := testutil.NewFakeStore()
	fake.SeedPart(domain.Part{ID: "p1", Name: "Brake Pad", PriceCents: 4899, StockQty: 10})
	fake.SeedPart(domain.Part{ID: "p2", Name: "Oil Filter", PriceCents: 899, StockQty: 100})
	pub := &testutil.FakePublisher{}
	svc := service.NewCheckoutService(fake, fake, pub)
	return fake, pub, svc
}

func TestCheckoutService_PlaceOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("creates order, computes total, publishes event, empties cart", func(t *testing.T) {
		fake, pub, checkoutSvc := newCheckoutFixture()
		cartSvc := service.NewCartService(fake, fake)

		cart, err := cartSvc.CreateCart(ctx, "cust-1")
		require.NoError(t, err)
		require.NoError(t, cartSvc.AddItem(ctx, cart.ID, "p1", 2)) // 2 * 4899 = 9798
		require.NoError(t, cartSvc.AddItem(ctx, cart.ID, "p2", 3)) // 3 * 899  = 2697

		order, err := checkoutSvc.PlaceOrder(ctx, cart.ID, "cust-1")
		require.NoError(t, err)

		assert.Equal(t, "cust-1", order.CustomerID)
		assert.Equal(t, domain.OrderStatusPlaced, order.Status)
		assert.EqualValues(t, 9798+2697, order.TotalCents)
		assert.Len(t, order.Items, 2)

		// cart should be emptied after checkout
		gotCart, err := cartSvc.GetCart(ctx, cart.ID)
		require.NoError(t, err)
		assert.Empty(t, gotCart.Items)

		// exactly one OrderPlaced event, matching the order
		events := pub.PublishedEvents()
		require.Len(t, events, 1)
		assert.Equal(t, order.ID, events[0].OrderID)
		assert.Equal(t, "cust-1", events[0].CustomerID)
		assert.Len(t, events[0].Items, 2)
	})

	t.Run("rejects checkout for empty cart", func(t *testing.T) {
		fake, _, checkoutSvc := newCheckoutFixture()
		cartSvc := service.NewCartService(fake, fake)

		cart, err := cartSvc.CreateCart(ctx, "cust-1")
		require.NoError(t, err)

		_, err = checkoutSvc.PlaceOrder(ctx, cart.ID, "cust-1")
		assert.ErrorIs(t, err, store.ErrEmptyCart)
	})

	t.Run("rejects checkout when customer does not own the cart", func(t *testing.T) {
		fake, _, checkoutSvc := newCheckoutFixture()
		cartSvc := service.NewCartService(fake, fake)

		cart, err := cartSvc.CreateCart(ctx, "cust-1")
		require.NoError(t, err)
		require.NoError(t, cartSvc.AddItem(ctx, cart.ID, "p1", 1))

		_, err = checkoutSvc.PlaceOrder(ctx, cart.ID, "someone-else")
		assert.ErrorIs(t, err, service.ErrCartCustomerMismatch)
	})

	t.Run("rejects checkout for unknown cart", func(t *testing.T) {
		_, _, checkoutSvc := newCheckoutFixture()

		_, err := checkoutSvc.PlaceOrder(ctx, "missing-cart", "cust-1")
		assert.ErrorIs(t, err, store.ErrNotFound)
	})

	t.Run("order still succeeds even if publishing the event fails", func(t *testing.T) {
		fake, pub, checkoutSvc := newCheckoutFixture()
		cartSvc := service.NewCartService(fake, fake)
		pub.Err = errors.New("pubsub unavailable")

		cart, err := cartSvc.CreateCart(ctx, "cust-1")
		require.NoError(t, err)
		require.NoError(t, cartSvc.AddItem(ctx, cart.ID, "p1", 1))

		order, err := checkoutSvc.PlaceOrder(ctx, cart.ID, "cust-1")
		require.NoError(t, err, "a downstream pubsub outage should not fail an already-committed order")
		assert.NotEmpty(t, order.ID)
	})
}

func TestCheckoutService_GetOrder_NotFound(t *testing.T) {
	_, _, checkoutSvc := newCheckoutFixture()

	_, err := checkoutSvc.GetOrder(context.Background(), "missing")
	assert.ErrorIs(t, err, store.ErrNotFound)
}
