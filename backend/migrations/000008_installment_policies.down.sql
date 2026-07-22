SET search_path TO public;

DELETE FROM public.role_permissions
WHERE permission_id IN (
  'b0000000-0000-4000-8000-00000000001d',
  'b0000000-0000-4000-8000-00000000001e'
);

DELETE FROM public.permissions
WHERE id IN (
  'b0000000-0000-4000-8000-00000000001d',
  'b0000000-0000-4000-8000-00000000001e'
);

DROP TABLE IF EXISTS public.installment_policies;
