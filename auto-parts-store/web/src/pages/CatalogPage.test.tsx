import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { CatalogPage } from "./CatalogPage";
import { api } from "../api/client";
import type { Category, Part } from "../api/types";
import { useCart } from "../context/CartContext";

vi.mock("../api/client", () => ({
  api: {
    listCategories: vi.fn(),
    listParts: vi.fn(),
  },
}));

vi.mock("../context/CartContext", () => ({
  useCart: vi.fn(),
}));

const categories: Category[] = [
  { id: "cat-brakes", name: "Brakes", slug: "brakes" },
  { id: "cat-filters", name: "Filters", slug: "filters" },
];

const parts: Part[] = [
  {
    id: "p1",
    sku: "BRK-1001",
    name: "Ceramic Brake Pad Set",
    description: "",
    categoryId: "cat-brakes",
    priceCents: 4899,
    stockQty: 10,
  },
];

function renderCatalog() {
  return render(
    <MemoryRouter>
      <CatalogPage />
    </MemoryRouter>,
  );
}

describe("CatalogPage", () => {
  beforeEach(() => {
    vi.mocked(useCart).mockReturnValue({
      customerId: "cust-1",
      cart: null,
      loading: false,
      error: null,
      addItem: vi.fn(),
      removeItem: vi.fn(),
      clearCartAfterCheckout: vi.fn(),
    });
    vi.mocked(api.listCategories).mockResolvedValue(categories);
    vi.mocked(api.listParts).mockResolvedValue(parts);
  });

  it("lists parts returned by the API", async () => {
    renderCatalog();

    await waitFor(() => {
      expect(screen.getByText("Ceramic Brake Pad Set")).toBeInTheDocument();
    });
    expect(api.listParts).toHaveBeenCalledWith({ category: undefined, q: undefined });
  });

  it("shows a no-results message when the API returns nothing", async () => {
    vi.mocked(api.listParts).mockResolvedValue([]);
    renderCatalog();

    await waitFor(() => {
      expect(screen.getByTestId("no-results")).toBeInTheDocument();
    });
  });

  it("re-queries the API with the selected category", async () => {
    renderCatalog();
    await waitFor(() => expect(api.listParts).toHaveBeenCalled());

    const user = userEvent.setup();
    await user.selectOptions(screen.getByTestId("category-filter"), "brakes");

    await waitFor(() => {
      expect(api.listParts).toHaveBeenCalledWith({ category: "brakes", q: undefined });
    });
  });

  it("re-queries the API with the search term", async () => {
    renderCatalog();
    await waitFor(() => expect(api.listParts).toHaveBeenCalled());

    const user = userEvent.setup();
    await user.type(screen.getByTestId("search-input"), "brake");

    await waitFor(() => {
      expect(api.listParts).toHaveBeenLastCalledWith({ category: undefined, q: "brake" });
    });
  });
});
