package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dl9346/auto-parts-store/backend/internal/store"
)

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.catalog.ListCategories(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (s *Server) handleListParts(w http.ResponseWriter, r *http.Request) {
	filter := store.PartFilter{
		CategorySlug: r.URL.Query().Get("category"),
		Query:        r.URL.Query().Get("q"),
	}
	parts, err := s.catalog.ListParts(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, parts)
}

func (s *Server) handleGetPart(w http.ResponseWriter, r *http.Request) {
	part, err := s.catalog.GetPart(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, part)
}

func (s *Server) handleCreateCart(w http.ResponseWriter, r *http.Request) {
	var req createCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	cart, err := s.cart.CreateCart(r.Context(), req.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, cart)
}

func (s *Server) handleGetCart(w http.ResponseWriter, r *http.Request) {
	cart, err := s.cart.GetCart(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cart)
}

func (s *Server) handleAddCartItem(w http.ResponseWriter, r *http.Request) {
	var req addCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	cartID := chi.URLParam(r, "id")
	if err := s.cart.AddItem(r.Context(), cartID, req.PartID, req.Quantity); err != nil {
		writeError(w, err)
		return
	}
	cart, err := s.cart.GetCart(r.Context(), cartID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cart)
}

func (s *Server) handleRemoveCartItem(w http.ResponseWriter, r *http.Request) {
	cartID := chi.URLParam(r, "id")
	if err := s.cart.RemoveItem(r.Context(), cartID, chi.URLParam(r, "partId")); err != nil {
		writeError(w, err)
		return
	}
	cart, err := s.cart.GetCart(r.Context(), cartID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cart)
}

func (s *Server) handleCheckout(w http.ResponseWriter, r *http.Request) {
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	order, err := s.checkout.PlaceOrder(r.Context(), req.CartID, req.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	order, err := s.checkout.GetOrder(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) handleListCustomerOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := s.checkout.ListOrdersByCustomer(r.Context(), chi.URLParam(r, "customerId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}
