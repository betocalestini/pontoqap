CREATE TABLE store_settings (
    key VARCHAR(128) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO store_settings (key, value) VALUES ('default_margin_percent', '30');

ALTER TABLE products
    ADD COLUMN margin_percent NUMERIC(5, 2) NOT NULL DEFAULT 30.00;

CREATE TABLE inventory_lots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES inventory_locations (id),
    sku_id UUID NOT NULL REFERENCES skus (id) ON DELETE RESTRICT,
    quantity_remaining INTEGER NOT NULL,
    unit_cost_cents BIGINT NOT NULL DEFAULT 0,
    source_movement_id UUID REFERENCES stock_movements (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inventory_lots_qty_check CHECK (quantity_remaining >= 0),
    CONSTRAINT inventory_lots_cost_check CHECK (unit_cost_cents >= 0)
);

CREATE INDEX idx_inventory_lots_sku_fifo ON inventory_lots (sku_id, created_at);

-- Lotes sintéticos para saldo já existente
INSERT INTO inventory_lots (id, location_id, sku_id, quantity_remaining, unit_cost_cents, created_at)
SELECT gen_random_uuid(), ib.location_id, ib.sku_id, ib.available_quantity,
       COALESCE(s.cost_price_cents, 0), NOW()
FROM inventory_balances ib
JOIN skus s ON s.id = ib.sku_id
WHERE ib.available_quantity > 0;
