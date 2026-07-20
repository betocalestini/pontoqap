CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT categories_slug_unique UNIQUE (slug)
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID REFERENCES categories (id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    visible BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT products_slug_unique UNIQUE (slug)
);

CREATE TABLE skus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL,
    barcode VARCHAR(64),
    unit VARCHAR(32) NOT NULL DEFAULT 'UN',
    sale_price_cents BIGINT NOT NULL DEFAULT 0,
    cost_price_cents BIGINT,
    minimum_stock INTEGER NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT skus_code_unique UNIQUE (code),
    CONSTRAINT skus_barcode_unique UNIQUE (barcode),
    CONSTRAINT skus_sale_price_check CHECK (sale_price_cents >= 0),
    CONSTRAINT skus_cost_price_check CHECK (cost_price_cents IS NULL OR cost_price_cents >= 0),
    CONSTRAINT skus_minimum_stock_check CHECK (minimum_stock >= 0)
);

CREATE TABLE product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    storage_key VARCHAR(512) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    alt_text VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE price_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku_id UUID NOT NULL REFERENCES skus (id) ON DELETE CASCADE,
    previous_price_cents BIGINT NOT NULL,
    new_price_cents BIGINT NOT NULL,
    changed_by UUID NOT NULL REFERENCES users (id),
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_category_active ON products (category_id, active, visible);
