import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";
import { api } from "../api/client";
import type { Cart } from "../api/types";

const CUSTOMER_ID_KEY = "autoparts.customerId";
const CART_ID_KEY = "autoparts.cartId";

// There's no auth in this app (see plan assumptions), so we mint a random
// customer id on first visit and keep it in localStorage - good enough to
// demonstrate per-customer carts/orders without building real accounts.
function getOrCreateCustomerId(): string {
  let id = localStorage.getItem(CUSTOMER_ID_KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(CUSTOMER_ID_KEY, id);
  }
  return id;
}

interface CartContextValue {
  customerId: string;
  cart: Cart | null;
  loading: boolean;
  error: string | null;
  addItem: (partId: string, quantity: number) => Promise<void>;
  removeItem: (partId: string) => Promise<void>;
  clearCartAfterCheckout: () => void;
}

const CartContext = createContext<CartContextValue | null>(null);

export function CartProvider({ children }: { children: ReactNode }) {
  const [customerId] = useState(getOrCreateCustomerId);
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const cartId = localStorage.getItem(CART_ID_KEY);
    if (!cartId) return;
    api
      .getCart(cartId)
      .then(setCart)
      .catch(() => localStorage.removeItem(CART_ID_KEY));
  }, []);

  const ensureCart = useCallback(async (): Promise<string> => {
    if (cart) return cart.id;
    const existingId = localStorage.getItem(CART_ID_KEY);
    if (existingId) {
      try {
        const existing = await api.getCart(existingId);
        setCart(existing);
        return existing.id;
      } catch {
        localStorage.removeItem(CART_ID_KEY);
      }
    }
    const created = await api.createCart(customerId);
    localStorage.setItem(CART_ID_KEY, created.id);
    setCart(created);
    return created.id;
  }, [cart, customerId]);

  const addItem = useCallback(
    async (partId: string, quantity: number) => {
      setLoading(true);
      setError(null);
      try {
        const cartId = await ensureCart();
        const updated = await api.addCartItem(cartId, partId, quantity);
        setCart(updated);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to add item to cart");
      } finally {
        setLoading(false);
      }
    },
    [ensureCart],
  );

  const removeItem = useCallback(
    async (partId: string) => {
      if (!cart) return;
      setLoading(true);
      setError(null);
      try {
        const updated = await api.removeCartItem(cart.id, partId);
        setCart(updated);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to remove item from cart");
      } finally {
        setLoading(false);
      }
    },
    [cart],
  );

  const clearCartAfterCheckout = useCallback(() => {
    localStorage.removeItem(CART_ID_KEY);
    setCart(null);
  }, []);

  return (
    <CartContext.Provider
      value={{ customerId, cart, loading, error, addItem, removeItem, clearCartAfterCheckout }}
    >
      {children}
    </CartContext.Provider>
  );
}

export function useCart(): CartContextValue {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error("useCart must be used within a CartProvider");
  return ctx;
}
