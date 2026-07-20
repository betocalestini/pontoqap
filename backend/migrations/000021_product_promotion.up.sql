ALTER TABLE products
    ADD COLUMN promo_active BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN promo_margin_percent NUMERIC(5, 2),
    ADD COLUMN promo_quantity_total INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN promo_quantity_remaining INTEGER NOT NULL DEFAULT 0;

ALTER TABLE products
    ADD CONSTRAINT products_promo_remaining_nonneg CHECK (promo_quantity_remaining >= 0),
    ADD CONSTRAINT products_promo_total_nonneg CHECK (promo_quantity_total >= 0),
    ADD CONSTRAINT products_promo_remaining_lte_total CHECK (promo_quantity_remaining <= promo_quantity_total);

ALTER TABLE products
    ADD CONSTRAINT products_promo_margin_when_active CHECK (
        NOT promo_active OR (promo_margin_percent IS NOT NULL AND promo_quantity_total > 0)
    );
