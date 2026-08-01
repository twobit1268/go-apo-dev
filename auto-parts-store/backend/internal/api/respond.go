package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dl9346/auto-parts-store/backend/internal/service"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// writeError maps a domain/store error to an HTTP status code. Anything it
// doesn't recognize is treated as a 500 with a generic message so internal
// details never leak to the client.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
	case errors.Is(err, service.ErrInvalidQuantity),
		errors.Is(err, service.ErrCartCustomerMismatch),
		errors.Is(err, store.ErrEmptyCart),
		errors.Is(err, store.ErrInsufficientStock):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
