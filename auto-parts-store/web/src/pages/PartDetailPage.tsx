import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { api } from "../api/client";
import type { Part } from "../api/types";
import { formatCents } from "../utils/money";
import { useCart } from "../context/CartContext";

export function PartDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { addItem } = useCart();
  const [part, setPart] = useState<Part | null>(null);
  const [quantityInput, setQuantityInput] = useState("1");
  const [error, setError] = useState<string | null>(null);
  const [added, setAdded] = useState(false);

  useEffect(() => {
    if (!id) return;
    api
      .getPart(id)
      .then(setPart)
      .catch((err) => setError(err instanceof Error ? err.message : "Part not found"));
  }, [id]);

  if (error) return <p role="alert">{error}</p>;
  if (!part) return <p>Loading…</p>;

  return (
    <div data-testid="part-detail">
      <Link to="/">&larr; Back to catalog</Link>
      <h1 className="page-title" data-testid="part-detail-name">
        {part.name}
      </h1>
      <p className="part-sku">SKU: {part.sku}</p>
      <p>{part.description}</p>
      <p className="part-price" data-testid="part-detail-price">
        {formatCents(part.priceCents)}
      </p>
      <p className={`part-stock ${part.stockQty > 0 ? "in-stock" : "out-of-stock"}`}>
        {part.stockQty > 0 ? `${part.stockQty} in stock` : "Out of stock"}
      </p>

      <div className="part-detail-actions">
        <input
          type="number"
          min={1}
          value={quantityInput}
          data-testid="part-detail-qty"
          onChange={(e) => setQuantityInput(e.target.value)}
        />
        <button
          data-testid="part-detail-add-to-cart"
          disabled={part.stockQty === 0}
          onClick={async () => {
            await addItem(part.id, Math.max(1, parseInt(quantityInput, 10) || 1));
            setAdded(true);
          }}
        >
          Add to Cart
        </button>
        {added && (
          <span className="added-confirmation" data-testid="part-detail-added-confirmation">
            {" "}
            Added!
          </span>
        )}
      </div>
    </div>
  );
}
