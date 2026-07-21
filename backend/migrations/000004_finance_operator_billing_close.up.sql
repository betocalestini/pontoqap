SET search_path TO public;

-- Financeiro pode fechar competência (homologação / operação diária)
INSERT INTO public.role_permissions (role_id, permission_id)
VALUES ('a0000000-0000-4000-8000-000000000005', 'b0000000-0000-4000-8000-00000000000b')
ON CONFLICT DO NOTHING;
