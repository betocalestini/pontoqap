SET search_path TO public;

DROP INDEX IF EXISTS public.payment_charges_one_pending_per_installment;
DROP INDEX IF EXISTS public.payments_provider_external_payment_id_unique;
ALTER TABLE public.payment_charges DROP COLUMN IF EXISTS last_synced_at;
