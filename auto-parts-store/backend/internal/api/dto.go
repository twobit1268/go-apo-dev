package api

type createCartRequest struct {
	CustomerID string `json:"customerId"`
}

type addCartItemRequest struct {
	PartID   string `json:"partId"`
	Quantity int    `json:"quantity"`
}

type checkoutRequest struct {
	CartID     string `json:"cartId"`
	CustomerID string `json:"customerId"`
}

type errorResponse struct {
	Error string `json:"error"`
}
