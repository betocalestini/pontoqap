INSERT INTO permissions (id, code, name) VALUES
    ('b0000000-0000-4000-8000-000000000011', 'inventory.entry', 'Registrar entrada de estoque'),
    ('b0000000-0000-4000-8000-000000000012', 'inventory.loss', 'Registrar perda e avaria'),
    ('b0000000-0000-4000-8000-000000000013', 'inventory.adjust', 'Ajustar inventário')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'a0000000-0000-4000-8000-000000000001', id FROM permissions
WHERE code IN ('inventory.entry', 'inventory.loss', 'inventory.adjust')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'a0000000-0000-4000-8000-000000000002', id FROM permissions
WHERE code IN ('inventory.entry', 'inventory.loss', 'inventory.adjust')
ON CONFLICT DO NOTHING;

-- Quem já tinha inventory.adjust legado (b000...004) ganha as novas permissões
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p.id
FROM role_permissions rp
JOIN permissions legacy ON legacy.id = rp.permission_id AND legacy.code = 'inventory.adjust'
CROSS JOIN permissions p
WHERE p.code IN ('inventory.entry', 'inventory.loss')
ON CONFLICT DO NOTHING;
