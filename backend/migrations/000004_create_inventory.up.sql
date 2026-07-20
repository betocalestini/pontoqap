CREATE TABLE inventory_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE inventory_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES inventory_locations (id),
    sku_id UUID NOT NULL REFERENCES skus (id) ON DELETE RESTRICT,
    available_quantity INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inventory_balances_location_sku_unique UNIQUE (location_id, sku_id),
    CONSTRAINT inventory_balances_quantity_check CHECK (available_quantity >= 0)
);

CREATE TABLE stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES inventory_locations (id),
    sku_id UUID NOT NULL REFERENCES skus (id) ON DELETE RESTRICT,
    movement_type VARCHAR(32) NOT NULL,
    quantity INTEGER NOT NULL,
    previous_balance INTEGER NOT NULL,
    new_balance INTEGER NOT NULL,
    reference_type VARCHAR(64),
    reference_id UUID,
    reason TEXT,
    created_by UUID REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT stock_movements_quantity_check CHECK (quantity > 0)
);

CREATE INDEX idx_stock_movements_sku_created ON stock_movements (sku_id, created_at DESC);
