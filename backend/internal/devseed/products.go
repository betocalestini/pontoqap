package devseed

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/inventory"
)

type productSeed struct {
	SKUID          uuid.UUID
	SalePriceCents int64
}

func seedProducts(ctx context.Context, pool *pgxpool.Pool, catSvc *catalog.Service, invSvc *inventory.Service, actorID uuid.UUID, cfg Config) ([]productSeed, error) {
	categoryNames := []string{"Seed Mercearia", "Seed Bebidas", "Seed Limpeza"}
	categoryIDs := make([]uuid.UUID, 0, len(categoryNames))
	for i, name := range categoryNames {
		slug := fmt.Sprintf("seed-cat-%d", i+1)
		var id uuid.UUID
		err := pool.QueryRow(ctx, `
			INSERT INTO categories (name, slug, active)
			VALUES ($1, $2, TRUE)
			ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, name, slug).Scan(&id)
		if err != nil {
			return nil, err
		}
		categoryIDs = append(categoryIDs, id)
	}

	out := make([]productSeed, 0, cfg.Products)
	for i := 0; i < cfg.Products; i++ {
		catID := categoryIDs[i%len(categoryIDs)]
		name := fmt.Sprintf("Produto seed %03d", i+1)
		slug := fmt.Sprintf("seed-produto-%03d", i+1)
		skuCode := fmt.Sprintf("SEED-%03d", i+1)
		cost := int64(500 + (i%20)*75)
		sale := cost * 130 / 100
		if sale < 100 {
			sale = 100
		}
		p, err := catSvc.CreateProduct(ctx, catalog.CreateProductInput{
			Name:         name,
			Slug:         slug,
			Description:  "Item gerado pelo seed de desenvolvimento",
			CategoryID:   &catID,
			SKUCode:      skuCode,
			SalePrice:    sale,
			CostPrice:    &cost,
			MinimumStock: 2,
			Unit:         "UN",
		})
		if err != nil {
			return nil, err
		}
		if len(p.SKUs) == 0 {
			return nil, fmt.Errorf("produto sem SKU")
		}
		skuID := p.SKUs[0].ID
		qty := 400 + (i % 50)
		if err := invSvc.RegisterEntry(ctx, skuID, qty, actorID, "seed demo", int64(qty)*cost, 0); err != nil {
			return nil, err
		}
		out = append(out, productSeed{SKUID: skuID, SalePriceCents: sale})
	}
	return out, nil
}
