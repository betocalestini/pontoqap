SET search_path TO public;

-- Expand user status values for staff lifecycle
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_status_check;
ALTER TABLE public.users ADD CONSTRAINT users_status_check CHECK (
    (status)::text = ANY (
        ARRAY[
            'active'::character varying,
            'inactive'::character varying,
            'blocked'::character varying,
            'pending_email'::character varying,
            'invited'::character varying,
            'temporarily_blocked'::character varying,
            'suspended'::character varying,
            'disabled'::character varying
        ]::text[]
    )
);

UPDATE public.users u
SET status = 'temporarily_blocked'
WHERE u.status = 'blocked'
  AND NOT EXISTS (
      SELECT 1 FROM public.user_roles ur
      JOIN public.roles r ON r.id = ur.role_id
      WHERE ur.user_id = u.id AND r.code = 'customer'
  );

CREATE TABLE public.admin_invitations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    role_id uuid NOT NULL,
    token_hash character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    invited_by uuid NOT NULL,
    accepted_at timestamp with time zone,
    revoked_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT admin_invitations_pkey PRIMARY KEY (id),
    CONSTRAINT admin_invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES public.users(id),
    CONSTRAINT admin_invitations_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id)
);

CREATE INDEX idx_admin_invitations_email ON public.admin_invitations USING btree (lower((email)::text));
CREATE INDEX idx_admin_invitations_token_hash ON public.admin_invitations USING btree (token_hash) WHERE (revoked_at IS NULL AND accepted_at IS NULL);

INSERT INTO public.roles VALUES ('a0000000-0000-4000-8000-000000000004', 'inventory_operator', 'Operador de estoque');
INSERT INTO public.roles VALUES ('a0000000-0000-4000-8000-000000000005', 'finance_operator', 'Financeiro');

-- Manager: remove settings.write (financial calendar / critical settings stay with system_admin)
DELETE FROM public.role_permissions
WHERE role_id = 'a0000000-0000-4000-8000-000000000002'
  AND permission_id = 'b0000000-0000-4000-8000-00000000000e';

-- inventory_operator
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000004', 'b0000000-0000-4000-8000-000000000001');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000004', 'b0000000-0000-4000-8000-000000000003');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000004', 'b0000000-0000-4000-8000-000000000011');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000004', 'b0000000-0000-4000-8000-000000000012');

-- finance_operator
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000005', 'b0000000-0000-4000-8000-000000000008');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000005', 'b0000000-0000-4000-8000-00000000000a');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000005', 'b0000000-0000-4000-8000-00000000000c');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000005', 'b0000000-0000-4000-8000-00000000000d');
