SET search_path TO public;

DROP INDEX IF EXISTS public.idx_payment_charges_installment;
ALTER TABLE public.payment_charges DROP COLUMN IF EXISTS installment_id;
