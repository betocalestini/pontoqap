SET search_path TO public;

DELETE FROM public.role_permissions WHERE role_id IN (
    'a0000000-0000-4000-8000-000000000004',
    'a0000000-0000-4000-8000-000000000005'
);
DELETE FROM public.roles WHERE id IN (
    'a0000000-0000-4000-8000-000000000004',
    'a0000000-0000-4000-8000-000000000005'
);

INSERT INTO public.role_permissions (role_id, permission_id)
VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-00000000000e')
ON CONFLICT DO NOTHING;

DROP TABLE IF EXISTS public.admin_invitations;

UPDATE public.users SET status = 'blocked' WHERE status = 'temporarily_blocked';

ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_status_check;
ALTER TABLE public.users ADD CONSTRAINT users_status_check CHECK (
    (status)::text = ANY (
        ARRAY[
            'active'::character varying,
            'inactive'::character varying,
            'blocked'::character varying,
            'pending_email'::character varying
        ]::text[]
    )
);
