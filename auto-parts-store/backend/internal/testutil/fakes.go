// Package testutil provides in-memory fakes for the store and pubsub
// interfaces, so the service layer's unit tests run with no Docker, no
// network, and no Postgres/Pub/Sub emulator.
package testutil

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dl9346/auto-parts-store/backend/internal/domain"
	"github.com/dl9346/auto-parts-store/backend/internal/pubsub"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
)

// FakeStore implements store.CategoryStore, store.PartStore, store.CartStore,
// and store.OrderStore entirely in memory.
type FakeStore struct {
	mu         sync.Mutex
	categories []domain.Category
	parts      map[string]domain.Part
	carts      map[string]domain.Cart
	orders     map[string]domain.Order
	seq        int
}

func NewFakeStore() *FakeStore {
	return &FakeStore{
		parts:  make(map[string]domain.Part),
		carts:  make(map[string]domain.Cart),
		orders: make(map[string]domain.Order),
	}
}

func (f *FakeStore) nextID(prefix string) string {
	f.seq++
	return fmt.Sprintf("%s-%d", prefix, f.seq)
}

// SeedCategory and SeedPart let tests populate fixture data directly,
// bypassing the store interfaces (which have no "create category/part"
// methods - those are out of scope, seeded via migration in the real app).
func (f *FakeStore) SeedCategory(c domain.Category) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.categories = append(f.categories, c)
}

func (f *FakeStore) SeedPart(p domain.Part) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.parts[p.ID] = p
}

func (f *FakeStore) ListCategories(ctx context.Context) ([]domain.Category, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domain.Category, len(f.categories))
	copy(out, f.categories)
	return out, nil
}

func (f *FakeStore) categorySlugToID(slug string) (string, bool) {
	for _, c := range f.categories {
		if c.Slug == slug {
			return c.ID, true
		}
	}
	return "", false
}

func (f *FakeStore) ListParts(ctx context.Context, filter store.PartFilter) ([]domain.Part, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var categoryID string
	if filter.CategorySlug != "" {
		id, ok := f.categorySlugToID(filter.CategorySlug)
		if !ok {
			return []domain.Part{}, nil
		}
		categoryID = id
	}

	out := []domain.Part{} // never nil - matches Postgres.ListParts' never-null contract
	for _, p := range f.parts {
		if categoryID != "" && p.CategoryID != categoryID {
			continue
		}
		if filter.Query != "" && !containsFold(p.Name, filter.Query) {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func (f *FakeStore) GetPart(ctx context.Context, id string) (domain.Part, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.parts[id]
	if !ok {
		return domain.Part{}, store.ErrNotFound
	}
	return p, nil
}

func (f *FakeStore) DecrementStock(ctx context.Context, partID string, qty int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.parts[partID]
	if !ok {
		return store.ErrNotFound
	}
	if p.StockQty < qty {
		return store.ErrInsufficientStock
	}
	p.StockQty -= qty
	f.parts[partID] = p
	return nil
}

func (f *FakeStore) CreateCart(ctx context.Context, customerID string) (domain.Cart, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cart := domain.Cart{
		ID:         f.nextID("cart"),
		CustomerID: customerID,
		Items:      []domain.CartItem{},
		CreatedAt:  time.Now(),
	}
	f.carts[cart.ID] = cart
	return cart, nil
}

func (f *FakeStore) GetCart(ctx context.Context, id string) (domain.Cart, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.carts[id]
	if !ok {
		return domain.Cart{}, store.ErrNotFound
	}
	itemsCopy := make([]domain.CartItem, len(c.Items))
	copy(itemsCopy, c.Items)
	c.Items = itemsCopy
	return c, nil
}

func (f *FakeStore) AddItem(ctx context.Context, cartID, partID string, quantity int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.carts[cartID]
	if !ok {
		return store.ErrNotFound
	}
	for i, item := range c.Items {
		if item.PartID == partID {
			c.Items[i].Quantity += quantity
			f.carts[cartID] = c
			return nil
		}
	}
	c.Items = append(c.Items, domain.CartItem{PartID: partID, Quantity: quantity})
	f.carts[cartID] = c
	return nil
}

func (f *FakeStore) RemoveItem(ctx context.Context, cartID, partID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.carts[cartID]
	if !ok {
		return store.ErrNotFound
	}
	for i, item := range c.Items {
		if item.PartID == partID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			break
		}
	}
	f.carts[cartID] = c
	return nil
}

func (f *FakeStore) CreateOrderFromCart(ctx context.Context, cartID, customerID string) (domain.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	cart, ok := f.carts[cartID]
	if !ok {
		return domain.Order{}, store.ErrNotFound
	}
	if len(cart.Items) == 0 {
		return domain.Order{}, store.ErrEmptyCart
	}

	var total int64
	items := make([]domain.OrderItem, 0, len(cart.Items))
	for _, ci := range cart.Items {
		part, ok := f.parts[ci.PartID]
		if !ok {
			return domain.Order{}, store.ErrNotFound
		}
		items = append(items, domain.OrderItem{
			PartID:         ci.PartID,
			Quantity:       ci.Quantity,
			UnitPriceCents: part.PriceCents,
		})
		total += part.PriceCents * int64(ci.Quantity)
	}

	order := domain.Order{
		ID:         f.nextID("order"),
		CustomerID: customerID,
		Status:     domain.OrderStatusPlaced,
		TotalCents: total,
		Items:      items,
		CreatedAt:  time.Now(),
	}
	f.orders[order.ID] = order

	cart.Items = []domain.CartItem{}
	f.carts[cartID] = cart

	return order, nil
}

func (f *FakeStore) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	o, ok := f.orders[id]
	if !ok {
		return domain.Order{}, store.ErrNotFound
	}
	return o, nil
}

func (f *FakeStore) ListOrdersByCustomer(ctx context.Context, customerID string) ([]domain.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []domain.Order{} // never nil - matches Postgres.ListOrdersByCustomer's never-null contract
	for _, o := range f.orders {
		if o.CustomerID == customerID {
			out = append(out, o)
		}
	}
	return out, nil
}

func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// FakePublisher implements pubsub.Publisher, recording every published
// event so tests can assert on it, and optionally returning a canned error.
type FakePublisher struct {
	mu     sync.Mutex
	Events []pubsub.OrderPlaced
	Err    error
}

func (f *FakePublisher) PublishOrderPlaced(ctx context.Context, event pubsub.OrderPlaced) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return f.Err
	}
	f.Events = append(f.Events, event)
	return nil
}

func (f *FakePublisher) PublishedEvents() []pubsub.OrderPlaced {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]pubsub.OrderPlaced, len(f.Events))
	copy(out, f.Events)
	return out
}
