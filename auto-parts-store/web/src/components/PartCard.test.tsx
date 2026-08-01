import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { PartCard } from "./PartCard";
import type { Part } from "../api/types";
import { useCart } from "../context/CartContext";

vi.mock("../context/CartContext", () => ({
  useCart: vi.fn(),
}));

const part: Part = {
  id: "p1",
  sku: "BRK-1001",
  name: "Ceramic Brake Pad Set",
  description: "Low-dust ceramic pads",
  categoryId: "cat-brakes",
  priceCents: 4899,
  stockQty: 10,
};

function renderPartCard() {
  return render(
    <MemoryRouter>
      <PartCard part={part} />
    </MemoryRouter>,
  );
}

describe("PartCard", () => {
  it("renders the part's name, price, and stock", () => {
    vi.mocked(useCart).mockReturnValue({
      customerId: "cust-1",
      cart: null,
      loading: false,
      error: null,
      addItem: vi.fn(),
      removeItem: vi.fn(),
      clearCartAfterCheckout: vi.fn(),
    });

    renderPartCard();

    expect(screen.getByText("Ceramic Brake Pad Set")).toBeInTheDocument();
    expect(screen.getByText("$48.99")).toBeInTheDocument();
    expect(screen.getByText("10 in stock")).toBeInTheDocument();
  });

  it("calls addItem with the selected quantity when Add to Cart is clicked", async () => {
    const addItem = vi.fn();
    vi.mocked(useCart).mockReturnValue({
      customerId: "cust-1",
      cart: null,
      loading: false,
      error: null,
      addItem,
      removeItem: vi.fn(),
      clearCartAfterCheckout: vi.fn(),
    });

    const user = userEvent.setup();
    renderPartCard();

    const qtyInput = screen.getByTestId("part-qty-p1");
    await user.clear(qtyInput);
    await user.type(qtyInput, "3");
    await user.click(screen.getByTestId("add-to-cart-p1"));

    expect(addItem).toHaveBeenCalledWith("p1", 3);
  });

  it("disables Add to Cart when the part is out of stock", () => {
    vi.mocked(useCart).mockReturnValue({
      customerId: "cust-1",
      cart: null,
      loading: false,
      error: null,
      addItem: vi.fn(),
      removeItem: vi.fn(),
      clearCartAfterCheckout: vi.fn(),
    });

    render(
      <MemoryRouter>
        <PartCard part={{ ...part, stockQty: 0 }} />
      </MemoryRouter>,
    );

    expect(screen.getByTestId("add-to-cart-p1")).toBeDisabled();
    expect(screen.getByText("Out of stock")).toBeInTheDocument();
  });
});
