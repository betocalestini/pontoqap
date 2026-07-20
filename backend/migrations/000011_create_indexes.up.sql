-- Additional indexes (some created inline; this migration adds supplementary ones)

CREATE INDEX IF NOT EXISTS idx_skus_product_active ON skus (product_id, active);
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items (cart_id);
CREATE INDEX IF NOT EXISTS idx_billing_entries_period ON billing_entries (billing_period_id);
