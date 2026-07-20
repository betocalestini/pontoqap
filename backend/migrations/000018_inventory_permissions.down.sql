DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN ('inventory.entry', 'inventory.loss', 'inventory.adjust')
);
DELETE FROM permissions WHERE code IN ('inventory.entry', 'inventory.loss', 'inventory.adjust');
