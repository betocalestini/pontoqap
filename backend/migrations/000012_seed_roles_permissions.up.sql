-- Roles and permissions seed (BK-012)

INSERT INTO roles (id, code, name) VALUES
    ('a0000000-0000-4000-8000-000000000001', 'system_admin', 'Administrador do sistema'),
    ('a0000000-0000-4000-8000-000000000002', 'manager', 'Gerente'),
    ('a0000000-0000-4000-8000-000000000003', 'customer', 'Cliente')
ON CONFLICT (code) DO NOTHING;

INSERT INTO permissions (id, code, name) VALUES
    ('b0000000-0000-4000-8000-000000000001', 'products.read', 'Consultar produtos'),
    ('b0000000-0000-4000-8000-000000000002', 'products.write', 'Gerenciar produtos'),
    ('b0000000-0000-4000-8000-000000000003', 'inventory.read', 'Consultar estoque'),
    ('b0000000-0000-4000-8000-000000000004', 'inventory.adjust', 'Ajustar estoque'),
    ('b0000000-0000-4000-8000-000000000005', 'customers.read', 'Consultar clientes'),
    ('b0000000-0000-4000-8000-000000000006', 'customers.approve', 'Aprovar clientes'),
    ('b0000000-0000-4000-8000-000000000007', 'customers.change_limit', 'Alterar limite'),
    ('b0000000-0000-4000-8000-000000000008', 'orders.read', 'Consultar pedidos'),
    ('b0000000-0000-4000-8000-000000000009', 'orders.cancel', 'Cancelar pedidos'),
    ('b0000000-0000-4000-8000-00000000000a', 'billing.read', 'Consultar faturamento'),
    ('b0000000-0000-4000-8000-00000000000b', 'billing.close', 'Fechar período'),
    ('b0000000-0000-4000-8000-00000000000c', 'payments.read', 'Consultar pagamentos'),
    ('b0000000-0000-4000-8000-00000000000d', 'reports.read', 'Consultar relatórios'),
    ('b0000000-0000-4000-8000-00000000000e', 'settings.write', 'Alterar configurações'),
    ('b0000000-0000-4000-8000-00000000000f', 'audit.read', 'Consultar auditoria'),
    ('b0000000-0000-4000-8000-000000000010', 'users.manage', 'Gerenciar usuários')
ON CONFLICT (code) DO NOTHING;

-- system_admin: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'a0000000-0000-4000-8000-000000000001', id FROM permissions
ON CONFLICT DO NOTHING;

-- manager: operational permissions (no users.manage)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'a0000000-0000-4000-8000-000000000002', id FROM permissions
WHERE code != 'users.manage'
ON CONFLICT DO NOTHING;

-- customer: no admin permissions (store uses customer profile, not permission checks for catalog)

INSERT INTO inventory_locations (id, name, active)
SELECT 'c0000000-0000-4000-8000-000000000001', 'Principal', TRUE
WHERE NOT EXISTS (SELECT 1 FROM inventory_locations WHERE id = 'c0000000-0000-4000-8000-000000000001');
