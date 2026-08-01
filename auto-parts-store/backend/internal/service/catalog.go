// Package service holds business logic that sits between the HTTP handlers
// and the store/pubsub layers. Each service depends only on the narrow
// interfaces it needs, so unit tests can supply in-memory fakes.
package service

import (
	"context"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
)

type CatalogService struct {
	categories store.CategoryStore
	parts      store.PartStore
}

func NewCatalogService(categories store.CategoryStore, parts store.PartStore) *CatalogService {
	return &CatalogService{categories: categories, parts: parts}
}

func (s *CatalogService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.ListCategories(ctx)
}

func (s *CatalogService) ListParts(ctx context.Context, filter store.PartFilter) ([]domain.Part, error) {
	return s.parts.ListParts(ctx, filter)
}

func (s *CatalogService) GetPart(ctx context.Context, id string) (domain.Part, error) {
	return s.parts.GetPart(ctx, id)
}
