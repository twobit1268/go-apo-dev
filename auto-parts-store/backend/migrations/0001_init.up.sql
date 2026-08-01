CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE categories (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE
);

CREATE TABLE parts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku         TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category_id UUID NOT NULL REFERENCES categories(id),
    price_cents BIGINT NOT NULL CHECK (price_cents >= 0),
    stock_qty   INTEGER NOT NULL DEFAULT 0 CHECK (stock_qty >= 0)
);

CREATE INDEX parts_category_id_idx ON parts(category_id);
CREATE INDEX parts_name_idx ON parts USING gin (to_tsvector('english', name));

CREATE TABLE carts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE cart_items (
    cart_id  UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    part_id  UUID NOT NULL REFERENCES parts(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (cart_id, part_id)
);

CREATE TABLE orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id TEXT NOT NULL,
    status      TEXT NOT NULL,
    total_cents BIGINT NOT NULL CHECK (total_cents >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    order_id         UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    part_id          UUID NOT NULL REFERENCES parts(id),
    quantity         INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents >= 0),
    PRIMARY KEY (order_id, part_id)
);

CREATE INDEX orders_customer_id_idx ON orders(customer_id);
