package catalog

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	ID             uuid.UUID `json:"id"`
	Code           string    `json:"code"`
	Barcode        string    `json:"barcode,omitempty"`
	Unit           string    `json:"unit"`
	SalePriceCents int64     `json:"sale_price_cents"`
	Active         bool      `json:"active"`
	AvailableQty   *int      `json:"available_quantity,omitempty"`
}

type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	Active      bool      `json:"active"`
	Visible     bool      `json:"visible"`
	SKUs        []SKU     `json:"skus,omitempty"`
}

type ListProductsFilter struct {
	Search   string
	Category string
	Page     int
	PageSize int
	Admin    bool
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
	}
	if f.Search != "" {
		where = append(where, "(p.name ILIKE $"+itoa(arg)+" OR p.slug ILIKE $"+itoa(arg)+")")
		args = append(args, "%"+f.Search+"%")
		arg++
	}
	if f.Category != "" {
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

	args = append(args, f.PageSize, offset)
	q := `
		SELECT p.id, p.name, p.slug, COALESCE(p.description,''), p.category_id, p.active, p.visible
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE ` + whereSQL + `
		ORDER BY p.name
		LIMIT $` + itoa(arg) + ` OFFSET $` + itoa(arg+1)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []Product
	var ids []uuid.UUID
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CategoryID, &p.Active, &p.Visible); err != nil {
			return nil, 0, err
		}
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
	for i := range products {
		products[i].SKUs = skuMap[products[i].ID]
	}
	return products, total, nil
}

func (s *Service) GetProduct(ctx context.Context, id uuid.UUID, publicOnly bool) (*Product, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(description,''), category_id, active, visible
		FROM products WHERE id = $1
	`, id)
	var p Product
	if err := row.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CategoryID, &p.Active, &p.Visible); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if publicOnly && (!p.Active || !p.Visible) {
		return nil, nil
	}
	m, err := s.loadSKUs(ctx, []uuid.UUID{p.ID}, publicOnly)
	if err != nil {
		return nil, err
	}
	p.SKUs = m[p.ID]
	return &p, nil
}

func (s *Service) loadSKUs(ctx context.Context, productIDs []uuid.UUID, activeOnly bool) (map[uuid.UUID][]SKU, error) {
	q := `
		SELECT s.id, s.product_id, s.code, COALESCE(s.barcode,''), s.unit, s.sale_price_cents, s.active,
		       COALESCE(ib.available_quantity, 0)
		FROM skus s
		LEFT JOIN inventory_balances ib ON ib.sku_id = s.id
		WHERE s.product_id = ANY($1)
	`
	if activeOnly {
		q += ` AND s.active = TRUE`
	}
	rows, err := s.pool.Query(ctx, q, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID][]SKU)
	for rows.Next() {
		var sku SKU
		var productID uuid.UUID
		var qty int
		if err := rows.Scan(&sku.ID, &productID, &sku.Code, &sku.Barcode, &sku.Unit, &sku.SalePriceCents, &sku.Active, &qty); err != nil {
			return nil, err
		}
		sku.AvailableQty = &qty
		out[productID] = append(out[productID], sku)
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
	Name        string
	Slug        string
	Description string
	CategoryID  *uuid.UUID
	SKUCode     string
	SalePrice   int64
	Unit        string
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

	var p Product
	err = tx.QueryRow(ctx, `
		INSERT INTO products (name, slug, description, category_id, active, visible)
		VALUES ($1, $2, $3, $4, TRUE, TRUE)
		RETURNING id, name, slug, COALESCE(description,''), category_id, active, visible
	`, in.Name, in.Slug, in.Description, in.CategoryID).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.CategoryID, &p.Active, &p.Visible,
	)
	if err != nil {
		return nil, err
	}
	var sku SKU
	err = tx.QueryRow(ctx, `
		INSERT INTO skus (product_id, code, unit, sale_price_cents, active)
		VALUES ($1, $2, $3, $4, TRUE)
		RETURNING id, code, COALESCE(barcode,''), unit, sale_price_cents, active
	`, p.ID, in.SKUCode, in.Unit, in.SalePrice).Scan(
		&sku.ID, &sku.Code, &sku.Barcode, &sku.Unit, &sku.SalePriceCents, &sku.Active,
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

func itoa(n int) string {
	return strconv.Itoa(n)
}
