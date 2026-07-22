SET search_path TO public;

ALTER TABLE public.payment_charges
    ADD COLUMN installment_id uuid REFERENCES public.invoice_installments(id);

CREATE INDEX idx_payment_charges_installment ON public.payment_charges (installment_id, status)
    WHERE installment_id IS NOT NULL;
