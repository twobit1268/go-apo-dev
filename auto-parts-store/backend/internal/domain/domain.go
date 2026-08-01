// Package domain holds the plain data types shared across the store,
// service, and api layers. No behavior lives here beyond simple helpers.
package domain

import "time"

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Part struct {
	ID          string `json:"id"`
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  string `json:"categoryId"`
	PriceCents  int64  `json:"priceCents"`
	StockQty    int    `json:"stockQty"`
}

type Cart struct {
	ID         string     `json:"id"`
	CustomerID string     `json:"customerId"`
	Items      []CartItem `json:"items"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type CartItem struct {
	PartID   string `json:"partId"`
	Quantity int    `json:"quantity"`
}

type OrderStatus string

const (
	OrderStatusPlaced OrderStatus = "placed"
)

type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customerId"`
	Status     OrderStatus `json:"status"`
	TotalCents int64       `json:"totalCents"`
	Items      []OrderItem `json:"items"`
	CreatedAt  time.Time   `json:"createdAt"`
}

type OrderItem struct {
	PartID         string `json:"partId"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unitPriceCents"`
}
