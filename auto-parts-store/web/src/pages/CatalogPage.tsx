import { useEffect, useState } from "react";
import { api } from "../api/client";
import type { Category, Part } from "../api/types";
import { PartCard } from "../components/PartCard";

export function CatalogPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [parts, setParts] = useState<Part[]>([]);
  const [category, setCategory] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.listCategories().then(setCategories).catch(() => setCategories([]));
  }, []);

  useEffect(() => {
    setLoading(true);
    setError(null);
    api
      .listParts({ category: category || undefined, q: query || undefined })
      .then(setParts)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load parts"))
      .finally(() => setLoading(false));
  }, [category, query]);

  return (
    <div>
      <h1 className="page-title">Shop Parts</h1>
      <form
        className="catalog-filters"
        data-testid="search-form"
        onSubmit={(e) => e.preventDefault()}
      >
        <input
          type="search"
          placeholder="Search parts…"
          value={query}
          data-testid="search-input"
          onChange={(e) => setQuery(e.target.value)}
        />
        <select
          value={category}
          data-testid="category-filter"
          onChange={(e) => setCategory(e.target.value)}
        >
          <option value="">All categories</option>
          {categories.map((c) => (
            <option key={c.id} value={c.slug}>
              {c.name}
            </option>
          ))}
        </select>
      </form>

      {loading && <p>Loading…</p>}
      {error && <p role="alert">{error}</p>}
      {!loading && !error && parts.length === 0 && (
        <p data-testid="no-results">No parts match your search.</p>
      )}

      <div className="part-list" data-testid="part-list">
        {parts.map((part) => (
          <PartCard key={part.id} part={part} />
        ))}
      </div>
    </div>
  );
}
