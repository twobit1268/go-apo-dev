//go:build api

// Package api_test contains black-box HTTP tests: they boot the real chi
// router wired to a real Postgres (docker-compose) and a real Pub/Sub
// emulator, then drive it purely over HTTP - no direct calls into the
// service or store layers. This exercises the full request path: router ->
// middleware -> handler -> service -> DB/pubsub.
package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dl9346/auto-parts-store/backend/internal/api"
	"github.com/dl9346/auto-parts-store/backend/internal/service"
	"github.com/dl9346/auto-parts-store/backend/internal/testutil"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	pg := testutil.NewTestPostgres(t)
	psClient := testutil.NewTestPubSubClient(t)

	catalogSvc := service.NewCatalogService(pg, pg)
	cartSvc := service.NewCartService(pg, pg)
	checkoutSvc := service.NewCheckoutService(pg, pg, psClient)

	srv := api.NewServer(catalogSvc, cartSvc, checkoutSvc)
	ts := httptest.NewServer(srv.Router())
	t.Cleanup(ts.Close)
	return ts
}

func doJSON(t *testing.T, method, url string, body any, out any) *http.Response {
	t.Helper()

	var reqBody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })

	if out != nil {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(out))
	}
	return resp
}
