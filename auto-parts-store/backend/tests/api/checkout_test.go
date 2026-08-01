//go:build api

package api_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/dl9346/auto-parts-store/backend/internal/testutil"
)

func TestAPI_CheckoutFlow(t *testing.T) {
	ts := newTestServer(t)

	var cart domain.Cart
	resp := doJSON(t, http.MethodPost, ts.URL+"/carts", map[string]string{"customerId": "cust-api-checkout"}, &cart)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	addReq := map[string]any{"partId": testutil.SeedPartOilFilterID, "quantity": 2}
	resp = doJSON(t, http.MethodPost, ts.URL+"/carts/"+cart.ID+"/items", addReq, &cart)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var order domain.Order
	checkoutReq := map[string]string{"cartId": cart.ID, "customerId": "cust-api-checkout"}
	resp = doJSON(t, http.MethodPost, ts.URL+"/checkout", checkoutReq, &order)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.EqualValues(t, 2*testutil.SeedPartOilFilterPrice, order.TotalCents)
	assert.Equal(t, domain.OrderStatusPlaced, order.Status)

	var fetched domain.Order
	resp = doJSON(t, http.MethodGet, ts.URL+"/orders/"+order.ID, nil, &fetched)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, order.ID, fetched.ID)

	var customerOrders []domain.Order
	resp = doJSON(t, http.MethodGet, ts.URL+"/customers/cust-api-checkout/orders", nil, &customerOrders)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, customerOrders, 1)
	assert.Equal(t, order.ID, customerOrders[0].ID)

	// the cart that was checked out should now be empty
	var emptiedCart domain.Cart
	resp = doJSON(t, http.MethodGet, ts.URL+"/carts/"+cart.ID, nil, &emptiedCart)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, emptiedCart.Items)
}

func TestAPI_Checkout_EmptyCart(t *testing.T) {
	ts := newTestServer(t)

	var cart domain.Cart
	doJSON(t, http.MethodPost, ts.URL+"/carts", map[string]string{"customerId": "cust-api-empty"}, &cart)

	checkoutReq := map[string]string{"cartId": cart.ID, "customerId": "cust-api-empty"}
	resp := doJSON(t, http.MethodPost, ts.URL+"/checkout", checkoutReq, nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAPI_Checkout_CustomerMismatch(t *testing.T) {
	ts := newTestServer(t)

	var cart domain.Cart
	doJSON(t, http.MethodPost, ts.URL+"/carts", map[string]string{"customerId": "cust-api-owner"}, &cart)
	addReq := map[string]any{"partId": testutil.SeedPartOilFilterID, "quantity": 1}
	doJSON(t, http.MethodPost, ts.URL+"/carts/"+cart.ID+"/items", addReq, &cart)

	checkoutReq := map[string]string{"cartId": cart.ID, "customerId": "someone-else"}
	resp := doJSON(t, http.MethodPost, ts.URL+"/checkout", checkoutReq, nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAPI_GetOrder_NotFound(t *testing.T) {
	ts := newTestServer(t)

	resp := doJSON(t, http.MethodGet, ts.URL+"/orders/00000000-0000-0000-0000-000000000000", nil, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
