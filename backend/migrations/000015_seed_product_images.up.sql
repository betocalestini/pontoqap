-- Imagens de demonstração (servidas pelo store-web em /product-images/*)
INSERT INTO product_images (id, product_id, storage_key, position, alt_text) VALUES
    ('d0000000-0000-4000-8000-000000000041', 'd0000000-0000-4000-8000-000000000011',
     '/product-images/arroz-5kg.svg', 0, 'Arroz 5kg'),
    ('d0000000-0000-4000-8000-000000000042', 'd0000000-0000-4000-8000-000000000012',
     '/product-images/feijao-carioca-1kg.svg', 0, 'Feijão Carioca 1kg'),
    ('d0000000-0000-4000-8000-000000000043', 'd0000000-0000-4000-8000-000000000013',
     '/product-images/oleo-soja-900ml.svg', 0, 'Óleo de Soja 900ml'),
    ('d0000000-0000-4000-8000-000000000044', 'd0000000-0000-4000-8000-000000000014',
     '/product-images/macarrao-espaguete-500g.svg', 0, 'Macarrão Espaguete 500g')
ON CONFLICT (id) DO NOTHING;
