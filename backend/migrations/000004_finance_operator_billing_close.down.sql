SET search_path TO public;

DELETE FROM public.role_permissions
WHERE role_id = 'a0000000-0000-4000-8000-000000000005'
  AND permission_id = 'b0000000-0000-4000-8000-00000000000b';
