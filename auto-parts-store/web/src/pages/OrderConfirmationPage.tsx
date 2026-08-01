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
      <h1>Order Confirmed</h1>
      <p data-testid="order-id">Order #{order.id}</p>
      <p data-testid="order-status">Status: {order.status}</p>
      <p data-testid="order-total">Total: {formatCents(order.totalCents)}</p>
      <ul>
        {order.items.map((item) => (
          <li key={item.partId}>
            {item.quantity} × {item.partId} @ {formatCents(item.unitPriceCents)}
          </li>
        ))}
      </ul>
      <Link to="/">Continue shopping</Link>
    </div>
  );
}
