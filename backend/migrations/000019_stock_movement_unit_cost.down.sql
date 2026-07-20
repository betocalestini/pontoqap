DROP INDEX IF EXISTS idx_stock_movements_product_created;
ALTER TABLE stock_movements DROP CONSTRAINT IF EXISTS stock_movements_unit_cost_check;
ALTER TABLE stock_movements DROP COLUMN IF EXISTS unit_cost_cents;
