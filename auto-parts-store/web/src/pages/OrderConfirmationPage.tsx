import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api } from "../api/client";
import type { Order } from "../api/types";
import { formatCents } from "../utils/money";

export function OrderConfirmationPage() {
  const { id } = useParams<{ id: string }>();
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    api
      .getOrder(id)
      .then(setOrder)
      .catch((err) => setError(err instanceof Error ? err.message : "Order not found"));
  }, [id]);

  if (error) return <p role="alert">{error}</p>;
  if (!order) return <p>Loading…</p>;

  return (
    <div data-testid="order-confirmation">
      <div className="order-summary">
        <div className="order-summary-header">
          <span className="order-check" aria-hidden="true">
            ✓
          </span>
          <h1>Order Confirmed</h1>
        </div>
        <p data-testid="order-id">Order #{order.id}</p>
        <p data-testid="order-status">
          Status: <span className="status-badge">{order.status}</span>
        </p>
        <p className="order-total" data-testid="order-total">
          Total: {formatCents(order.totalCents)}
        </p>
        <ul className="order-items">
          {order.items.map((item) => (
            <li key={item.partId}>
              {item.quantity} × {item.partId} @ {formatCents(item.unitPriceCents)}
            </li>
          ))}
        </ul>
      </div>
      <p style={{ marginTop: "1.25rem" }}>
        <Link to="/">Continue shopping</Link>
      </p>
    </div>
  );
}
