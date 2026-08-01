import { Link } from "react-router-dom";
import { useState } from "react";
import type { Part } from "../api/types";
import { formatCents } from "../utils/money";
import { useCart } from "../context/CartContext";

export function PartCard({ part }: { part: Part }) {
  const { addItem } = useCart();
  // Kept as raw text (not a clamped number) so the user can clear the
  // field and type a new value - clamping on every keystroke made it
  // impossible to backspace past "1" before typing a new digit.
  const [quantityInput, setQuantityInput] = useState("1");

  return (
    <div className="part-card" data-testid={`part-card-${part.id}`}>
      <Link to={`/parts/${part.id}`} data-testid={`part-link-${part.id}`}>
        <h3>{part.name}</h3>
      </Link>
      <p className="part-sku">SKU: {part.sku}</p>
      <p className="part-price">{formatCents(part.priceCents)}</p>
      <p className="part-stock">
        {part.stockQty > 0 ? `${part.stockQty} in stock` : "Out of stock"}
      </p>
      <div className="part-card-actions">
        <input
          type="number"
          min={1}
          value={quantityInput}
          data-testid={`part-qty-${part.id}`}
          onChange={(e) => setQuantityInput(e.target.value)}
        />
        <button
          data-testid={`add-to-cart-${part.id}`}
          disabled={part.stockQty === 0}
          onClick={() => addItem(part.id, Math.max(1, parseInt(quantityInput, 10) || 1))}
        >
          Add to Cart
        </button>
      </div>
    </div>
  );
}
