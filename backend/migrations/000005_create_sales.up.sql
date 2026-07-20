CREATE TABLE carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers (id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT carts_status_check CHECK (status IN ('active', 'checked_out', 'abandoned'))
);

CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    sku_id UUID NOT NULL REFERENCES skus (id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT cart_items_quantity_check CHECK (quantity > 0),
    CONSTRAINT cart_items_cart_sku_unique UNIQUE (cart_id, sku_id)
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(32) NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers (id),
    status VARCHAR(32) NOT NULL DEFAULT 'confirmed',
    subtotal_cents BIGINT NOT NULL,
    discount_cents BIGINT NOT NULL DEFAULT 0,
    total_cents BIGINT NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    confirmed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT orders_order_number_unique UNIQUE (order_number),
    CONSTRAINT orders_idempotency_key_unique UNIQUE (idempotency_key),
    CONSTRAINT orders_subtotal_check CHECK (subtotal_cents >= 0),
    CONSTRAINT orders_discount_check CHECK (discount_cents >= 0),
    CONSTRAINT orders_total_check CHECK (total_cents >= 0)
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    sku_id UUID NOT NULL REFERENCES skus (id),
    product_name_snapshot VARCHAR(255) NOT NULL,
    sku_code_snapshot VARCHAR(64) NOT NULL,
    unit_price_cents BIGINT NOT NULL,
    quantity INTEGER NOT NULL,
    total_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT order_items_quantity_check CHECK (quantity > 0),
    CONSTRAINT order_items_unit_price_check CHECK (unit_price_cents >= 0),
    CONSTRAINT order_items_total_check CHECK (total_cents >= 0)
);

CREATE TABLE order_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders (id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    reason TEXT,
    created_by UUID REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE order_return_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_return_id UUID NOT NULL REFERENCES order_returns (id) ON DELETE CASCADE,
    order_item_id UUID NOT NULL REFERENCES order_items (id),
    quantity INTEGER NOT NULL,
    credit_amount_cents BIGINT NOT NULL,
    return_to_stock BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT order_return_items_quantity_check CHECK (quantity > 0)
);

CREATE INDEX idx_orders_customer_created ON orders (customer_id, created_at DESC);
CREATE INDEX idx_orders_status_created ON orders (status, created_at DESC);
