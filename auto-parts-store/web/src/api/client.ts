import type { Cart, Category, Order, Part } from "./types";

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error ?? res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export const api = {
  listCategories: () => request<Category[]>("/categories"),

  listParts: (params: { category?: string; q?: string } = {}) => {
    const search = new URLSearchParams();
    if (params.category) search.set("category", params.category);
    if (params.q) search.set("q", params.q);
    const qs = search.toString();
    return request<Part[]>(`/parts${qs ? `?${qs}` : ""}`);
  },

  getPart: (id: string) => request<Part>(`/parts/${id}`),

  createCart: (customerId: string) =>
    request<Cart>("/carts", { method: "POST", body: JSON.stringify({ customerId }) }),

  getCart: (id: string) => request<Cart>(`/carts/${id}`),

  addCartItem: (cartId: string, partId: string, quantity: number) =>
    request<Cart>(`/carts/${cartId}/items`, {
      method: "POST",
      body: JSON.stringify({ partId, quantity }),
    }),

  removeCartItem: (cartId: string, partId: string) =>
    request<Cart>(`/carts/${cartId}/items/${partId}`, { method: "DELETE" }),

  checkout: (cartId: string, customerId: string) =>
    request<Order>("/checkout", {
      method: "POST",
      body: JSON.stringify({ cartId, customerId }),
    }),

  getOrder: (id: string) => request<Order>(`/orders/${id}`),

  listCustomerOrders: (customerId: string) =>
    request<Order[]>(`/customers/${customerId}/orders`),
};
