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

func TestCatalogService_ListParts_FiltersByCategoryAndQuery(t *testing.T) {
	fake := testutil.NewFakeStore()
	fake.SeedCategory(domain.Category{ID: "cat-brakes", Name: "Brakes", Slug: "brakes"})
	fake.SeedCategory(domain.Category{ID: "cat-filters", Name: "Filters", Slug: "filters"})
	fake.SeedPart(domain.Part{ID: "p1", Name: "Ceramic Brake Pad Set", CategoryID: "cat-brakes", PriceCents: 4899, StockQty: 10})
	fake.SeedPart(domain.Part{ID: "p2", Name: "Drilled Rotor", CategoryID: "cat-brakes", PriceCents: 12999, StockQty: 5})
	fake.SeedPart(domain.Part{ID: "p3", Name: "Oil Filter", CategoryID: "cat-filters", PriceCents: 899, StockQty: 100})

	svc := service.NewCatalogService(fake, fake)
	ctx := context.Background()

	tests := []struct {
		name   string
		filter store.PartFilter
		want   []string // expected part IDs, order-independent
	}{
		{"no filter returns all", store.PartFilter{}, []string{"p1", "p2", "p3"}},
		{"filter by category", store.PartFilter{CategorySlug: "brakes"}, []string{"p1", "p2"}},
		{"filter by query", store.PartFilter{Query: "oil"}, []string{"p3"}},
		{"filter by category and query", store.PartFilter{CategorySlug: "brakes", Query: "rotor"}, []string{"p2"}},
		{"unknown category returns none", store.PartFilter{CategorySlug: "nope"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := svc.ListParts(ctx, tt.filter)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, partIDs(parts))
		})
	}
}

func TestCatalogService_GetPart_NotFound(t *testing.T) {
	fake := testutil.NewFakeStore()
	svc := service.NewCatalogService(fake, fake)

	_, err := svc.GetPart(context.Background(), "missing")
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func partIDs(parts []domain.Part) []string {
	ids := make([]string, len(parts))
	for i, p := range parts {
		ids[i] = p.ID
	}
	return ids
}
