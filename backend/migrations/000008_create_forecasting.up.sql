CREATE TABLE forecast_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku_id UUID NOT NULL REFERENCES skus (id) ON DELETE CASCADE,
    reference_month DATE NOT NULL,
    forecast_quantity INTEGER NOT NULL,
    safety_stock_quantity INTEGER NOT NULL DEFAULT 0,
    suggested_purchase_quantity INTEGER NOT NULL DEFAULT 0,
    confidence_level VARCHAR(32),
    method VARCHAR(64) NOT NULL,
    parameters JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_forecast_snapshots_sku_month ON forecast_snapshots (sku_id, reference_month DESC);
