// Package api wires the REST endpoints to the service layer using chi.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/dl9346/auto-parts-store/backend/internal/service"
)

type Server struct {
	catalog  *service.CatalogService
	cart     *service.CartService
	checkout *service.CheckoutService
}

func NewServer(catalog *service.CatalogService, cart *service.CartService, checkout *service.CheckoutService) *Server {
	return &Server{catalog: catalog, cart: cart, checkout: checkout}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Get("/categories", s.handleListCategories)

	r.Get("/parts", s.handleListParts)
	r.Get("/parts/{id}", s.handleGetPart)

	r.Post("/carts", s.handleCreateCart)
	r.Get("/carts/{id}", s.handleGetCart)
	r.Post("/carts/{id}/items", s.handleAddCartItem)
	r.Delete("/carts/{id}/items/{partId}", s.handleRemoveCartItem)

	r.Post("/checkout", s.handleCheckout)

	r.Get("/orders/{id}", s.handleGetOrder)
	r.Get("/customers/{customerId}/orders", s.handleListCustomerOrders)

	return r
}

// corsMiddleware allows the SPA (served from a different origin in dev,
// e.g. Vite on :5173 vs the API on :8080) to call the API directly.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
