package catalog

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/inventory"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Category struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Slug   string    `json:"slug"`
	Active bool      `json:"active"`
}

type SKU struct {
	ID                uuid.UUID `json:"id"`
	Code              string    `json:"code"`
	Barcode           string    `json:"barcode,omitempty"`
	Unit              string    `json:"unit"`
	SalePriceCents    int64     `json:"sale_price_cents"`
	CostPriceCents    *int64    `json:"cost_price_cents,omitempty"`
	AverageCostCents  *int64    `json:"average_cost_cents,omitempty"`
	MinimumStock      int       `json:"minimum_stock"`
	Active            bool      `json:"active"`
	AvailableQty      *int      `json:"available_quantity,omitempty"`
}

type Product struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description,omitempty"`
	CategoryID    *uuid.UUID `json:"category_id,omitempty"`
	MarginPercent          float64   `json:"margin_percent"`
	PromoActive            bool      `json:"promo_active"`
	PromoMarginPercent     *float64  `json:"promo_margin_percent,omitempty"`
	PromoQuantityTotal     int       `json:"promo_quantity_total"`
	PromoQuantityRemaining int       `json:"promo_quantity_remaining"`
	OnPromotion            bool      `json:"on_promotion"`
	Active        bool      `json:"active"`
	Visible       bool      `json:"visible"`
	UpdatedAt     time.Time `json:"updated_at"`
	ImageURL    string    `json:"image_url,omitempty"`
	ImageAlt    string    `json:"image_alt,omitempty"`
	Images      []ProductImage `json:"images,omitempty"`
	SKUs        []SKU     `json:"skus,omitempty"`
}

type ProductImage struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`
	Alt string    `json:"alt,omitempty"`
}

type ListProductsFilter struct {
	Search     string
	Category   string
	Sort       string
	CustomerID *uuid.UUID
	Page       int
	PageSize   int
	Admin      bool
}

func (s *Service) ListCategories(ctx context.Context, admin bool) ([]Category, error) {
	q := `SELECT id, name, slug, active FROM categories`
	if !admin {
		q += ` WHERE active = TRUE`
	}
	q += ` ORDER BY name`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Active); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Service) ListProducts(ctx context.Context, f ListProductsFilter) ([]Product, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	offset := (f.Page - 1) * f.PageSize

	where := []string{"1=1"}
	args := []any{}
	arg := 1
	if !f.Admin {
		where = append(where, "p.active = TRUE AND p.visible = TRUE")
		locID, _ := uuid.Parse(inventory.DefaultLocationID)
		where = append(where, `EXISTS (
			SELECT 1 FROM skus s
			JOIN inventory_balances ib ON ib.sku_id = s.id AND ib.location_id = $`+itoa(arg)+`
			WHERE s.product_id = p.id AND s.active = TRUE AND ib.available_quantity > 0
		)`)
		args = append(args, locID)
		arg++
	}
	if f.Search != "" {
		where = append(where, "(p.name ILIKE $"+itoa(arg)+" OR p.slug ILIKE $"+itoa(arg)+")")
		args = append(args, "%"+f.Search+"%")
		arg++
	}
	if f.Category == "promocoes" {
		where = append(where, "p.promo_active = TRUE AND p.promo_quantity_remaining > 0")
	} else if f.Category != "" {
		where = append(where, "c.slug = $"+itoa(arg))
		args = append(args, f.Category)
		arg++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	countQ := `SELECT COUNT(*) FROM products p LEFT JOIN categories c ON c.id = p.category_id WHERE ` + whereSQL
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sort := strings.TrimSpace(f.Sort)
	needPrice := !f.Admin && (sort == "price_asc" || sort == "price_desc")
	needPurchases := !f.Admin && sort == "purchases" && f.CustomerID != nil

	joins := ""
	if needPrice {
		joins += `
		LEFT JOIN LATERAL (
			SELECT MIN(s.sale_price_cents) AS min_price
			FROM skus s WHERE s.product_id = p.id AND s.active = TRUE
		) sku_price ON TRUE`
	}
	if needPurchases {
		joins += `
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(oi.quantity), 0)::bigint AS purchase_qty
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id AND o.status = 'confirmed'
			JOIN skus s ON s.id = oi.sku_id AND s.product_id = p.id
			WHERE o.customer_id = $` + itoa(arg) + `
		) cust_purch ON TRUE`
		args = append(args, *f.CustomerID)
		arg++
	}

	orderSQL := catalogOrderClause(f)
	limitArg := arg
	offsetArg := arg + 1
	args = append(args, f.PageSize, offset)
	q := `
		SELECT p.id, p.name, p.slug, COALESCE(p.description,''), p.category_id, p.margin_percent,
		       p.promo_active, p.promo_margin_percent, p.promo_quantity_total, p.promo_quantity_remaining,
		       p.active, p.visible, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id` + joins + `
		WHERE ` + whereSQL + `
		` + orderSQL + `
		LIMIT $` + itoa(limitArg) + ` OFFSET $` + itoa(offsetArg)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []Product
	var ids []uuid.UUID
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CategoryID, &p.MarginPercent,
			&p.PromoActive, &p.PromoMarginPercent, &p.PromoQuantityTotal, &p.PromoQuantityRemaining,
			&p.Active, &p.Visible, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		p.OnPromotion = p.IsOnPromotion()
		products = append(products, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return products, total, nil
	}
	skuMap, err := s.loadSKUs(ctx, ids, !f.Admin)
	if err != nil {
		return nil, 0, err
	}
	imgMap, err := s.loadPrimaryImages(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range products {
		products[i].SKUs = skuMap[products[i].ID]
		if img, ok := imgMap[products[i].ID]; ok {
			products[i].ImageURL = img.URL
			products[i].ImageAlt = img.Alt
		}
		if products[i].ImageURL == "" {
			if u := ImageURLForSlug(products[i].Slug); u != "" {
				products[i].ImageURL = u
				if products[i].ImageAlt == "" {
					products[i].ImageAlt = products[i].Name
				}
			}
		}
	}
	return products, total, nil
}

func (s *Service) GetProduct(ctx context.Context, id uuid.UUID, publicOnly bool) (*Product, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(description,''), category_id, margin_percent,
		       promo_active, promo_margin_percent, promo_quantity_total, promo_quantity_remaining,
		       active, visible, updated_at
		FROM products WHERE id = $1
	`, id)
	var p Product
	if err := row.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CategoryID, &p.MarginPercent,
		&p.PromoActive, &p.PromoMarginPercent, &p.PromoQuantityTotal, &p.PromoQuantityRemaining,
		&p.Active, &p.Visible, &p.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p.OnPromotion = p.IsOnPromotion()
	if publicOnly && (!p.Active || !p.Visible) {
		return nil, nil
	}
	m, err := s.loadSKUs(ctx, []uuid.UUID{p.ID}, publicOnly)
	if err != nil {
		return nil, err
	}
	p.SKUs = m[p.ID]
	if publicOnly && !productHasAvailableStock(p.SKUs) {
		return nil, nil
	}
	if img, err := s.loadPrimaryImages(ctx, []uuid.UUID{p.ID}); err != nil {
		return nil, err
	} else if im, ok := img[p.ID]; ok {
		p.ImageURL = im.URL
		p.ImageAlt = im.Alt
	}
	if p.ImageURL == "" {
		if u := ImageURLForSlug(p.Slug); u != "" {
			p.ImageURL = u
			if p.ImageAlt == "" {
				p.ImageAlt = p.Name
			}
		}
	}
	if !publicOnly {
		imgs, err := s.loadProductImages(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.Images = imgs
	}
	return &p, nil
}

func (s *Service) loadSKUs(ctx context.Context, productIDs []uuid.UUID, activeOnly bool) (map[uuid.UUID][]SKU, error) {
	locID, _ := uuid.Parse(inventory.DefaultLocationID)
	q := `
		SELECT s.id, s.product_id, s.code, COALESCE(s.barcode,''), s.unit, s.sale_price_cents,
		       s.cost_price_cents, s.minimum_stock, s.active,
		       COALESCE(ib.available_quantity, 0),
		       CASE WHEN COALESCE(SUM(il.quantity_remaining), 0) > 0 THEN
		         (
		           (COALESCE(SUM(il.quantity_remaining::bigint * il.unit_cost_cents), 0)::bigint
		             + COALESCE(SUM(il.quantity_remaining), 0)::bigint / 2)
		           / COALESCE(SUM(il.quantity_remaining), 1)::bigint
		         )
		       ELSE NULL END
		FROM skus s
		LEFT JOIN inventory_balances ib ON ib.sku_id = s.id AND ib.location_id = $2
		LEFT JOIN inventory_lots il ON il.sku_id = s.id AND il.location_id = $2 AND il.quantity_remaining > 0
		WHERE s.product_id = ANY($1)
	`
	if activeOnly {
		q += ` AND s.active = TRUE`
	}
	q += ` GROUP BY s.id, s.product_id, s.code, s.barcode, s.unit, s.sale_price_cents, s.cost_price_cents, s.minimum_stock, s.active, ib.available_quantity`
	rows, err := s.pool.Query(ctx, q, productIDs, locID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID][]SKU)
	for rows.Next() {
		var sku SKU
		var productID uuid.UUID
		var qty int
		var avgCost *int64
		if err := rows.Scan(&sku.ID, &productID, &sku.Code, &sku.Barcode, &sku.Unit, &sku.SalePriceCents,
			&sku.CostPriceCents, &sku.MinimumStock, &sku.Active, &qty, &avgCost); err != nil {
			return nil, err
		}
		sku.AverageCostCents = avgCost
		sku.AvailableQty = &qty
		out[productID] = append(out[productID], sku)
	}
	return out, rows.Err()
}

type productImage struct {
	URL string
	Alt string
}

func (s *Service) loadPrimaryImages(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]productImage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (product_id) product_id, storage_key, COALESCE(alt_text, '')
		FROM product_images
		WHERE product_id = ANY($1)
		ORDER BY product_id, position ASC, created_at ASC
	`, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID]productImage)
	for rows.Next() {
		var productID uuid.UUID
		var key, alt string
		if err := rows.Scan(&productID, &key, &alt); err != nil {
			return nil, err
		}
		out[productID] = productImage{URL: ResolveImageURL(key), Alt: alt}
	}
	return out, rows.Err()
}

func (s *Service) loadProductImages(ctx context.Context, productID uuid.UUID) ([]ProductImage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, storage_key, COALESCE(alt_text, '')
		FROM product_images
		WHERE product_id = $1
		ORDER BY position ASC, created_at ASC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProductImage
	for rows.Next() {
		var im ProductImage
		var key string
		if err := rows.Scan(&im.ID, &key, &im.Alt); err != nil {
			return nil, err
		}
		im.URL = ResolveImageURL(key)
		out = append(out, im)
	}
	return out, rows.Err()
}

type CreateCategoryInput struct {
	Name string
	Slug string
}

func (s *Service) CreateCategory(ctx context.Context, in CreateCategoryInput) (Category, error) {
	if in.Slug == "" {
		in.Slug = slugify(in.Name)
	}
	var c Category
	err := s.pool.QueryRow(ctx, `
		INSERT INTO categories (name, slug) VALUES ($1, $2)
		RETURNING id, name, slug, active
	`, in.Name, in.Slug).Scan(&c.ID, &c.Name, &c.Slug, &c.Active)
	return c, err
}

type CreateProductInput struct {
	Name         string
	Slug         string
	Description  string
	CategoryID   *uuid.UUID
	SKUCode      string
	SalePrice    int64
	CostPrice    *int64
	MinimumStock int
	Barcode      string
	Unit         string
}

func (s *Service) CreateProduct(ctx context.Context, in CreateProductInput) (*Product, error) {
	if in.Slug == "" {
		in.Slug = slugify(in.Name)
	}
	if in.Unit == "" {
		in.Unit = "UN"
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	margin, err := s.GetDefaultMarginPercent(ctx)
	if err != nil {
		margin = 30
	}
	var p Product
	err = tx.QueryRow(ctx, `
		INSERT INTO products (name, slug, description, category_id, margin_percent, active, visible)
		VALUES ($1, $2, $3, $4, $5, TRUE, TRUE)
		RETURNING id, name, slug, COALESCE(description,''), category_id, margin_percent, active, visible
	`, in.Name, in.Slug, in.Description, in.CategoryID, margin).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.CategoryID, &p.MarginPercent, &p.Active, &p.Visible,
	)
	if err != nil {
		return nil, err
	}
	var sku SKU
	err = tx.QueryRow(ctx, `
		INSERT INTO skus (product_id, code, barcode, unit, sale_price_cents, cost_price_cents, minimum_stock, active)
		VALUES ($1, $2, NULLIF($3,''), $4, $5, $6, $7, TRUE)
		RETURNING id, code, COALESCE(barcode,''), unit, sale_price_cents, cost_price_cents, minimum_stock, active
	`, p.ID, in.SKUCode, in.Barcode, in.Unit, RoundSalePriceCents(in.SalePrice), in.CostPrice, in.MinimumStock).Scan(
		&sku.ID, &sku.Code, &sku.Barcode, &sku.Unit, &sku.SalePriceCents, &sku.CostPriceCents, &sku.MinimumStock, &sku.Active,
	)
	if err != nil {
		return nil, err
	}
	p.SKUs = []SKU{sku}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) ChangeSKUPrice(ctx context.Context, skuID uuid.UUID, newPrice int64, changedBy uuid.UUID, reason string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var prev int64
	if err := tx.QueryRow(ctx, `SELECT sale_price_cents FROM skus WHERE id = $1 FOR UPDATE`, skuID).Scan(&prev); err != nil {
		return err
	}
	newPrice = RoundSalePriceCents(newPrice)
	if _, err := tx.Exec(ctx, `UPDATE skus SET sale_price_cents = $2, updated_at = NOW() WHERE id = $1`, skuID, newPrice); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO price_history (sku_id, previous_price_cents, new_price_cents, changed_by, reason)
		VALUES ($1, $2, $3, $4, $5)
	`, skuID, prev, newPrice, changedBy, reason); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func productHasAvailableStock(skus []SKU) bool {
	for _, sku := range skus {
		if sku.AvailableQty != nil && *sku.AvailableQty > 0 {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
