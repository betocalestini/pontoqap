ALTER TABLE products
    DROP CONSTRAINT IF EXISTS products_promo_margin_when_active,
    DROP CONSTRAINT IF EXISTS products_promo_remaining_lte_total,
    DROP CONSTRAINT IF EXISTS products_promo_total_nonneg,
    DROP CONSTRAINT IF EXISTS products_promo_remaining_nonneg;

ALTER TABLE products
    DROP COLUMN IF EXISTS promo_quantity_remaining,
    DROP COLUMN IF EXISTS promo_quantity_total,
    DROP COLUMN IF EXISTS promo_margin_percent,
    DROP COLUMN IF EXISTS promo_active;
