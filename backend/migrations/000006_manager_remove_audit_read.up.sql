SET search_path TO public;

-- Auditoria administrativa completa: somente system_admin (ver docs/access-control.md)
DELETE FROM public.role_permissions
WHERE role_id = 'a0000000-0000-4000-8000-000000000002'
  AND permission_id = 'b0000000-0000-4000-8000-00000000000f';
