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
	dir := ResolveDataDir(cfg)
	rows, err := loadProductsCSV(dir, cfg.DefaultStockQty)
	if err != nil {
		return nil, err
	}

	categoryIDs := map[string]uuid.UUID{}
	out := make([]productSeed, 0, len(rows))

	for _, row := range rows {
		catID, err := ensureCategory(ctx, pool, categoryIDs, row.Category)
		if err != nil {
			return nil, err
		}
		cost := row.UnitCostCents
		p, err := catSvc.CreateProduct(ctx, catalog.CreateProductInput{
			Name:         row.Name,
			Slug:         row.Slug,
			Description:  "Item do seed de desenvolvimento",
			CategoryID:   &catID,
			SKUCode:      row.SKUCode,
			SalePrice:    0,
			CostPrice:    &cost,
			MinimumStock: 2,
			Unit:         row.Unit,
		})
		if err != nil {
			return nil, err
		}
		if len(p.SKUs) == 0 {
			return nil, fmt.Errorf("produto %s sem SKU", row.Name)
		}
		skuID := p.SKUs[0].ID
		if row.MarginPercent != nil {
			if _, err := pool.Exec(ctx, `UPDATE products SET margin_percent = $2 WHERE id = $1`, p.ID, *row.MarginPercent); err != nil {
				return nil, err
			}
		}
		totalPaid := int64(row.StockQty) * row.UnitCostCents
		if err := invSvc.RegisterEntry(ctx, skuID, row.StockQty, actorID, "seed demo", totalPaid, 0); err != nil {
			return nil, err
		}
		if _, err := catSvc.RecalculateSKU(ctx, skuID, actorID, "seed", invSvc.WeightedAverageCostCents); err != nil {
			return nil, err
		}
		if row.ImageSlug != "" {
			key := productImageStorageKey(row.ImageSlug)
			if err := catSvc.UpsertProductImage(ctx, p.ID, key, row.Name); err != nil {
				return nil, err
			}
		}
		var salePrice int64
		if err := pool.QueryRow(ctx, `SELECT sale_price_cents FROM skus WHERE id = $1`, skuID).Scan(&salePrice); err != nil {
			return nil, err
		}
		out = append(out, productSeed{SKUID: skuID, SalePriceCents: salePrice})
	}
	return out, nil
}

func ensureCategory(ctx context.Context, pool *pgxpool.Pool, cache map[string]uuid.UUID, name string) (uuid.UUID, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}
	slug := categorySlug(name)
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO categories (name, slug, active)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, name, slug).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	cache[name] = id
	return id, nil
}
