import { Link } from "react-router-dom";
import { useCart } from "../context/CartContext";

export function Header() {
  const { cart } = useCart();
  const itemCount = cart?.items.reduce((sum, item) => sum + item.quantity, 0) ?? 0;

  return (
    <header className="app-header">
      <Link to="/" className="brand" data-testid="nav-home">
        Auto Parts Store
      </Link>
      <nav>
        <Link to="/cart" data-testid="nav-cart">
          Cart ({itemCount})
        </Link>
      </nav>
    </header>
  );
}
