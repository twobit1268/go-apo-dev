import { useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { useCart } from "../context/CartContext";
import { api, ApiError } from "../api/client";

export function CheckoutPage() {
  const { cart, customerId, clearCartAfterCheckout } = useCart();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // Once an order is placed, clearCartAfterCheckout() nulls the cart, which
  // would otherwise make the "empty cart" redirect below fire again and
  // race the navigate() to the confirmation page - this flag keeps the
  // guard from re-triggering after a successful checkout.
  const [orderPlaced, setOrderPlaced] = useState(false);
  const navigate = useNavigate();

  if (!orderPlaced && (!cart || cart.items.length === 0)) {
    return <Navigate to="/cart" replace />;
  }

  const handlePlaceOrder = async () => {
    if (!cart) return;
    setSubmitting(true);
    setError(null);
    try {
      const order = await api.checkout(cart.id, customerId);
      setOrderPlaced(true);
      clearCartAfterCheckout();
      navigate(`/orders/${order.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to place order");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <h1 className="page-title">Checkout</h1>
      <p>{cart?.items.length ?? 0} item(s) in your cart.</p>
      {error && <p role="alert">{error}</p>}
      <button
        className="btn-large"
        data-testid="place-order-btn"
        onClick={handlePlaceOrder}
        disabled={submitting}
      >
        {submitting ? "Placing order…" : "Place Order"}
      </button>
    </div>
  );
}
