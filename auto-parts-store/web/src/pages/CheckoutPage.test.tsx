import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { CheckoutPage } from "./CheckoutPage";
import { api, ApiError } from "../api/client";
import type { Cart, Order } from "../api/types";
import { useCart } from "../context/CartContext";

vi.mock("../api/client", async () => {
  const actual = await vi.importActual<typeof import("../api/client")>("../api/client");
  return {
    ApiError: actual.ApiError,
    api: { checkout: vi.fn() },
  };
});

vi.mock("../context/CartContext", () => ({
  useCart: vi.fn(),
}));

const cartWithItems: Cart = {
  id: "cart-1",
  customerId: "cust-1",
  items: [{ partId: "p1", quantity: 2 }],
  createdAt: new Date().toISOString(),
};

function renderCheckout() {
  return render(
    <MemoryRouter initialEntries={["/checkout"]}>
      <Routes>
        <Route path="/checkout" element={<CheckoutPage />} />
        <Route path="/cart" element={<div>Cart Page</div>} />
        <Route path="/orders/:id" element={<div>Order Confirmation</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("CheckoutPage", () => {
  it("redirects to /cart when the cart is empty", () => {
    vi.mocked(useCart).mockReturnValue({
      customerId: "cust-1",
      cart: null,
      loading: false,
      error: null,
      addItem: vi.fn(),
      removeItem: vi.fn(),
      clearCartAfterCheckout: vi.fn(),
    });

    renderCheckout();

    expect(screen.getByText("Cart Page")).toBeInTheDocument();
  });

  it("places the order and navigates to the confirmation page", async () => {
    const clearCartAfterCheckout = vi.fn();
    vi.mocked(useCart).mockReturnValue({
      customerId: "cust-1",
      cart: cartWithItems,
      loading: false,
      error: null,
      addItem: vi.fn(),
      removeItem: vi.fn(),
      clearCartAfterCheckout,
    });
    const order: Order = {
      id: "order-1",
      customerId: "cust-1",
      status: "placed",
      totalCents: 9798,
      items: [],
      createdAt: new Date().toISOString(),
    };
    vi.mocked(api.checkout).mockResolvedValue(order);

    const user = userEvent.setup();
    renderCheckout();
    await user.click(screen.getByTestId("place-order-btn"));

    await waitFor(() => {
      expect(screen.getByText("Order Confirmation")).toBeInTheDocument();
    });
    expect(api.checkout).toHaveBeenCalledWith("cart-1", "cust-1");
    expect(clearCartAfterCheckout).toHaveBeenCalled();
  });

  it("shows an error message when checkout fails", async () => {
    vi.mocked(useCart).mockReturnValue({
      customerId: "cust-1",
      cart: cartWithItems,
      loading: false,
      error: null,
      addItem: vi.fn(),
      removeItem: vi.fn(),
      clearCartAfterCheckout: vi.fn(),
    });
    vi.mocked(api.checkout).mockRejectedValue(new ApiError(400, "cart is empty"));

    const user = userEvent.setup();
    renderCheckout();
    await user.click(screen.getByTestId("place-order-btn"));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent("cart is empty");
    });
  });
});
