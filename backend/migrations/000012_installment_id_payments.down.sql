SET search_path TO public;

DROP INDEX IF EXISTS public.idx_payments_installment;
ALTER TABLE public.payments DROP COLUMN IF EXISTS installment_id;
