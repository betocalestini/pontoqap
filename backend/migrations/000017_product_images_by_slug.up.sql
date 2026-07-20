-- storage_key pelo nome do produto (slug), sem extensão — o servidor resolve .svg, .png, etc.
INSERT INTO product_images (id, product_id, storage_key, position, alt_text) VALUES
    ('d0000000-0000-4000-8000-000000000041', 'd0000000-0000-4000-8000-000000000011',
     '/product-images/arroz-5kg', 0, 'Arroz 5kg'),
    ('d0000000-0000-4000-8000-000000000042', 'd0000000-0000-4000-8000-000000000012',
     '/product-images/feijao-carioca-1kg', 0, 'Feijão Carioca 1kg'),
    ('d0000000-0000-4000-8000-000000000043', 'd0000000-0000-4000-8000-000000000013',
     '/product-images/oleo-soja-900ml', 0, 'Óleo de Soja 900ml'),
    ('d0000000-0000-4000-8000-000000000044', 'd0000000-0000-4000-8000-000000000014',
     '/product-images/macarrao-espaguete-500g', 0, 'Macarrão Espaguete 500g')
ON CONFLICT (id) DO UPDATE SET
    storage_key = EXCLUDED.storage_key,
    alt_text = EXCLUDED.alt_text,
    position = EXCLUDED.position;
