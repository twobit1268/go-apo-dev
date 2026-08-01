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

func TestAPI_ListCategories(t *testing.T) {
	ts := newTestServer(t)

	var categories []domain.Category
	resp := doJSON(t, http.MethodGet, ts.URL+"/categories", nil, &categories)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, categories)
}

func TestAPI_ListParts_FilterByCategoryAndSearch(t *testing.T) {
	ts := newTestServer(t)

	var brakes []domain.Part
	resp := doJSON(t, http.MethodGet, ts.URL+"/parts?category=brakes", nil, &brakes)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, brakes)

	var oilFilters []domain.Part
	resp = doJSON(t, http.MethodGet, ts.URL+"/parts?q=oil", nil, &oilFilters)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, oilFilters, 1)
	assert.Equal(t, testutil.SeedPartOilFilterID, oilFilters[0].ID)
}

func TestAPI_GetPart(t *testing.T) {
	ts := newTestServer(t)

	var part domain.Part
	resp := doJSON(t, http.MethodGet, ts.URL+"/parts/"+testutil.SeedPartOilFilterID, nil, &part)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, testutil.SeedPartOilFilterID, part.ID)
}

func TestAPI_GetPart_NotFound(t *testing.T) {
	ts := newTestServer(t)

	resp := doJSON(t, http.MethodGet, ts.URL+"/parts/00000000-0000-0000-0000-000000000000", nil, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
