SET search_path TO public;

ALTER TABLE public.customers
    ADD COLUMN IF NOT EXISTS status_before_block VARCHAR(32);
