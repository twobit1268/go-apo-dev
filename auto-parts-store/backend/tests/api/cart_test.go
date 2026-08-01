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

func TestAPI_CartLifecycle(t *testing.T) {
	ts := newTestServer(t)

	var cart domain.Cart
	resp := doJSON(t, http.MethodPost, ts.URL+"/carts", map[string]string{"customerId": "cust-api-1"}, &cart)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "cust-api-1", cart.CustomerID)

	addReq := map[string]any{"partId": testutil.SeedPartOilFilterID, "quantity": 3}
	resp = doJSON(t, http.MethodPost, ts.URL+"/carts/"+cart.ID+"/items", addReq, &cart)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, cart.Items, 1)
	assert.Equal(t, 3, cart.Items[0].Quantity)

	resp = doJSON(t, http.MethodGet, ts.URL+"/carts/"+cart.ID, nil, &cart)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, cart.Items, 1)

	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/carts/"+cart.ID+"/items/"+testutil.SeedPartOilFilterID, nil)
	require.NoError(t, err)
	delResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer delResp.Body.Close()
	assert.Equal(t, http.StatusOK, delResp.StatusCode)
}

func TestAPI_AddCartItem_InvalidQuantity(t *testing.T) {
	ts := newTestServer(t)

	var cart domain.Cart
	doJSON(t, http.MethodPost, ts.URL+"/carts", map[string]string{"customerId": "cust-api-2"}, &cart)

	addReq := map[string]any{"partId": testutil.SeedPartOilFilterID, "quantity": 0}
	resp := doJSON(t, http.MethodPost, ts.URL+"/carts/"+cart.ID+"/items", addReq, nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAPI_AddCartItem_UnknownPart(t *testing.T) {
	ts := newTestServer(t)

	var cart domain.Cart
	doJSON(t, http.MethodPost, ts.URL+"/carts", map[string]string{"customerId": "cust-api-3"}, &cart)

	addReq := map[string]any{"partId": "00000000-0000-0000-0000-000000000000", "quantity": 1}
	resp := doJSON(t, http.MethodPost, ts.URL+"/carts/"+cart.ID+"/items", addReq, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAPI_GetCart_NotFound(t *testing.T) {
	ts := newTestServer(t)

	resp := doJSON(t, http.MethodGet, ts.URL+"/carts/00000000-0000-0000-0000-000000000000", nil, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
