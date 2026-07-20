package catalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UpdateProductInput struct {
	Name          *string
	Description   *string
	CategoryID    *uuid.UUID
	ClearCategory bool
	Active        *bool
	Visible       *bool
}

type UpdateSKUInput struct {
	Code            *string
	Barcode         *string
	Unit            *string
	SalePriceCents  *int64
	CostPriceCents  *int64
	MinimumStock    *int
	Active          *bool
}

func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, in UpdateProductInput) (*Product, error) {
	p, err := s.GetProduct(ctx, id, false)
	if err != nil || p == nil {
		return nil, err
	}
	name := p.Name
	desc := p.Description
	catID := p.CategoryID
	active := p.Active
	visible := p.Visible
	if in.Name != nil {
		name = *in.Name
	}
	if in.Description != nil {
		desc = *in.Description
	}
	if in.ClearCategory {
		catID = nil
	} else if in.CategoryID != nil {
		catID = in.CategoryID
	}
	if in.Active != nil {
		active = *in.Active
	}
	if in.Visible != nil {
		visible = *in.Visible
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE products SET name = $2, description = $3, category_id = $4, active = $5, visible = $6, updated_at = NOW()
		WHERE id = $1
	`, id, name, desc, catID, active, visible)
	if err != nil {
		return nil, err
	}
	return s.GetProduct(ctx, id, false)
}

func (s *Service) UpdateSKU(ctx context.Context, skuID uuid.UUID, in UpdateSKUInput, priceChangedBy uuid.UUID, priceReason string) error {
	row := s.pool.QueryRow(ctx, `
		SELECT code, COALESCE(barcode,''), unit, sale_price_cents, cost_price_cents, minimum_stock, active
		FROM skus WHERE id = $1
	`, skuID)
	var code, barcode, unit string
	var salePrice int64
	var costPrice *int64
	var minStock int
	var active bool
	if err := row.Scan(&code, &barcode, &unit, &salePrice, &costPrice, &minStock, &active); err != nil {
		return err
	}
	if in.Code != nil {
		code = *in.Code
	}
	if in.Barcode != nil {
		barcode = *in.Barcode
	}
	if in.Unit != nil {
		unit = *in.Unit
	}
	if in.MinimumStock != nil {
		minStock = *in.MinimumStock
	}
	if in.Active != nil {
		active = *in.Active
	}
	if in.CostPriceCents != nil {
		costPrice = in.CostPriceCents
	}
	newSale := salePrice
	if in.SalePriceCents != nil && *in.SalePriceCents != salePrice {
		newSale = *in.SalePriceCents
		if err := s.ChangeSKUPrice(ctx, skuID, newSale, priceChangedBy, priceReason); err != nil {
			return err
		}
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE skus SET code = $2, barcode = NULLIF($3,''), unit = $4, cost_price_cents = $5, minimum_stock = $6, active = $7, updated_at = NOW()
		WHERE id = $1
	`, skuID, code, barcode, unit, costPrice, minStock, active)
	return err
}

func (s *Service) UpsertProductImage(ctx context.Context, productID uuid.UUID, storageKey, alt string) error {
	var existing uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM product_images WHERE product_id = $1 ORDER BY position ASC LIMIT 1
	`, productID).Scan(&existing)
	if err == pgx.ErrNoRows {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO product_images (product_id, storage_key, position, alt_text)
			VALUES ($1, $2, 0, $3)
		`, productID, storageKey, alt)
		if err != nil {
			return err
		}
		_, err = s.pool.Exec(ctx, `UPDATE products SET updated_at = NOW() WHERE id = $1`, productID)
		return err
	}
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE product_images SET storage_key = $2, alt_text = $3 WHERE id = $1
	`, existing, storageKey, alt)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `UPDATE products SET updated_at = NOW() WHERE id = $1`, productID)
	return err
}

func (s *Service) DeleteProductImage(ctx context.Context, productID, imageID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM product_images WHERE id = $1 AND product_id = $2`, imageID, productID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Service) PrimarySKUID(ctx context.Context, productID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM skus WHERE product_id = $1 ORDER BY created_at ASC LIMIT 1
	`, productID).Scan(&id)
	return id, err
}
