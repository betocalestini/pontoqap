SET search_path TO public;

ALTER TABLE public.customers DROP COLUMN IF EXISTS status_before_block;
