package service

import (
	"context"
	"errors"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
)

var ErrInvalidQuantity = errors.New("quantity must be greater than zero")

type CartService struct {
	carts store.CartStore
	parts store.PartStore
}

func NewCartService(carts store.CartStore, parts store.PartStore) *CartService {
	return &CartService{carts: carts, parts: parts}
}

func (s *CartService) CreateCart(ctx context.Context, customerID string) (domain.Cart, error) {
	return s.carts.CreateCart(ctx, customerID)
}

func (s *CartService) GetCart(ctx context.Context, id string) (domain.Cart, error) {
	return s.carts.GetCart(ctx, id)
}

// AddItem validates the part exists and the quantity is sane before adding
// it to the cart. Stock isn't reserved here - that's checked/decremented at
// checkout, so two carts may both list stock they'll race for at purchase.
func (s *CartService) AddItem(ctx context.Context, cartID, partID string, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	if _, err := s.carts.GetCart(ctx, cartID); err != nil {
		return err
	}
	if _, err := s.parts.GetPart(ctx, partID); err != nil {
		return err
	}
	return s.carts.AddItem(ctx, cartID, partID, quantity)
}

func (s *CartService) RemoveItem(ctx context.Context, cartID, partID string) error {
	if _, err := s.carts.GetCart(ctx, cartID); err != nil {
		return err
	}
	return s.carts.RemoveItem(ctx, cartID, partID)
}
