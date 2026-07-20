ALTER TABLE stock_movements
    ADD COLUMN unit_cost_cents BIGINT,
    ADD CONSTRAINT stock_movements_unit_cost_check CHECK (unit_cost_cents IS NULL OR unit_cost_cents >= 0);

CREATE INDEX idx_stock_movements_product_created ON stock_movements (sku_id, created_at DESC)
    WHERE movement_type IN ('entry', 'initial_stock');

COMMENT ON COLUMN stock_movements.unit_cost_cents IS 'Custo unitário de aquisição (centavos), preenchido em entradas manuais';
