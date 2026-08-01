import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useCart } from "../context/CartContext";
import { api } from "../api/client";
import type { Part } from "../api/types";
import { formatCents } from "../utils/money";

export function CartPage() {
  const { cart, removeItem, loading } = useCart();
  const [parts, setParts] = useState<Record<string, Part>>({});
  const navigate = useNavigate();

  useEffect(() => {
    if (!cart) return;
    Promise.all(cart.items.map((item) => api.getPart(item.partId)))
      .then((fetched) => {
        setParts(Object.fromEntries(fetched.map((p) => [p.id, p])));
      })
      .catch(() => {
        // best-effort - if a part lookup fails we just show the raw id
      });
  }, [cart]);

  if (!cart || cart.items.length === 0) {
    return (
      <div>
        <h1>Your Cart</h1>
        <p data-testid="empty-cart-message">Your cart is empty.</p>
        <Link to="/">Continue shopping</Link>
      </div>
    );
  }

  const totalCents = cart.items.reduce((sum, item) => {
    const part = parts[item.partId];
    return sum + (part ? part.priceCents * item.quantity : 0);
  }, 0);

  return (
    <div>
      <h1>Your Cart</h1>
      <ul className="cart-items" data-testid="cart-items">
        {cart.items.map((item) => {
          const part = parts[item.partId];
          return (
            <li key={item.partId} data-testid={`cart-item-${item.partId}`}>
              <span>{part ? part.name : item.partId}</span>
              <span>Qty: {item.quantity}</span>
              <span>{part ? formatCents(part.priceCents * item.quantity) : ""}</span>
              <button
                data-testid={`remove-item-${item.partId}`}
                onClick={() => removeItem(item.partId)}
                disabled={loading}
              >
                Remove
              </button>
            </li>
          );
        })}
      </ul>
      <p className="cart-total" data-testid="cart-total">
        Total: {formatCents(totalCents)}
      </p>
      <button data-testid="checkout-btn" onClick={() => navigate("/checkout")}>
        Proceed to Checkout
      </button>
    </div>
  );
}
