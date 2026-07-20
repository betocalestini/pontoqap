package sales

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/customers"
	"github.com/store-platform/store/internal/inventory"
)

type Service struct {
	pool      *pgxpool.Pool
	inventory *inventory.Service
	billing   *billing.Service
	catalog   *catalog.Service
}

func NewService(pool *pgxpool.Pool, inv *inventory.Service, bill *billing.Service, cat *catalog.Service) *Service {
	return &Service{pool: pool, inventory: inv, billing: bill, catalog: cat}
}

type CartItem struct {
	ID             uuid.UUID `json:"id"`
	SKUID          uuid.UUID `json:"sku_id"`
	Quantity       int       `json:"quantity"`
	ProductName    string    `json:"product_name"`
	SKUCode        string    `json:"sku_code,omitempty"`
	UnitPriceCents int64     `json:"unit_price_cents"`
	LineTotalCents int64     `json:"line_total_cents"`
}

type Cart struct {
	ID    uuid.UUID  `json:"id"`
	Items []CartItem `json:"items"`
}

func (s *Service) GetOrCreateCart(ctx context.Context, customerID uuid.UUID) (*Cart, error) {
	var cartID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM carts WHERE customer_id = $1 AND status = 'active' LIMIT 1
	`, customerID).Scan(&cartID)
	if errors.Is(err, pgx.ErrNoRows) {
		cartID = uuid.New()
		_, err = s.pool.Exec(ctx, `INSERT INTO carts (id, customer_id, status) VALUES ($1, $2, 'active')`, cartID, customerID)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return s.loadCart(ctx, cartID)
}

// AddCartItem soma unidades ao item (catálogo / "Comprar" de novo).
func (s *Service) AddCartItem(ctx context.Context, customerID, skuID uuid.UUID, quantity int) (*Cart, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("invalid quantity")
	}
	cart, err := s.GetOrCreateCart(ctx, customerID)
	if err != nil {
		return nil, err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO cart_items (cart_id, sku_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, sku_id) DO UPDATE
		SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = NOW()
	`, cart.ID, skuID, quantity)
	if err != nil {
		return nil, err
	}
	return s.loadCart(ctx, cart.ID)
}

// SetCartItemQuantity define a quantidade absoluta (0 remove o item).
func (s *Service) SetCartItemQuantity(ctx context.Context, customerID, skuID uuid.UUID, quantity int) (*Cart, error) {
	cart, err := s.GetOrCreateCart(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if quantity <= 0 {
		_, err = s.pool.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1 AND sku_id = $2`, cart.ID, skuID)
	} else {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO cart_items (cart_id, sku_id, quantity)
			VALUES ($1, $2, $3)
			ON CONFLICT (cart_id, sku_id) DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = NOW()
		`, cart.ID, skuID, quantity)
	}
	if err != nil {
		return nil, err
	}
	return s.loadCart(ctx, cart.ID)
}

func (s *Service) loadCart(ctx context.Context, cartID uuid.UUID) (*Cart, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ci.id, ci.sku_id, ci.quantity, p.id, p.name, COALESCE(s.code,''), s.sale_price_cents,
		       p.promo_active, p.promo_quantity_remaining
		FROM cart_items ci
		JOIN skus s ON s.id = ci.sku_id
		JOIN products p ON p.id = s.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at
	`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	promoLeft := make(map[uuid.UUID]int)
	var items []CartItem
	for rows.Next() {
		var it CartItem
		var productID uuid.UUID
		var salePrice int64
		var promoActive bool
		var promoRemainDB int
		if err := rows.Scan(&it.ID, &it.SKUID, &it.Quantity, &productID, &it.ProductName, &it.SKUCode, &salePrice,
			&promoActive, &promoRemainDB); err != nil {
			return nil, err
		}
		remain, ok := promoLeft[productID]
		if !ok {
			remain = promoRemainDB
		}
		if s.catalog != nil {
			res, err := s.catalog.PriceLine(ctx, catalog.PriceLineInput{
				ProductID: productID, SKUID: it.SKUID, Quantity: it.Quantity,
				SalePrice: salePrice, PromoActive: promoActive, PromoRemain: remain,
			}, s.inventory.WeightedAverageCostCents)
			if err != nil {
				return nil, err
			}
			it.UnitPriceCents = res.UnitPriceCents
			it.LineTotalCents = res.LineTotalCents
			promoLeft[productID] = remain - res.PromoUnits
		} else {
			it.UnitPriceCents = salePrice
			it.LineTotalCents = salePrice * int64(it.Quantity)
		}
		items = append(items, it)
	}
	return &Cart{ID: cartID, Items: items}, rows.Err()
}

func (s *Service) UpsertCartItem(ctx context.Context, customerID, skuID uuid.UUID, quantity int) (*Cart, error) {
	return s.SetCartItemQuantity(ctx, customerID, skuID, quantity)
}

type Order struct {
	ID          uuid.UUID `json:"id"`
	OrderNumber string    `json:"order_number"`
	TotalCents  int64     `json:"total_cents"`
	Status      string    `json:"status"`
}

func (s *Service) Checkout(ctx context.Context, customerID uuid.UUID, idempotencyKey string, actorUserID uuid.UUID) (*Order, error) {
	if idempotencyKey == "" {
		return nil, fmt.Errorf("idempotency key required")
	}
	var existing Order
	err := s.pool.QueryRow(ctx, `
		SELECT id, order_number, total_cents, status FROM orders WHERE idempotency_key = $1
	`, idempotencyKey).Scan(&existing.ID, &existing.OrderNumber, &existing.TotalCents, &existing.Status)
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	custSvc := customers.NewService(s.pool, nil)
	cust, err := custSvc.GetByID(ctx, customerID)
	if err != nil || cust == nil {
		return nil, customers.ErrNotApproved()
	}
	if cust.Status != "approved" {
		return nil, customers.ErrNotApproved()
	}

	cart, err := s.GetOrCreateCart(ctx, customerID)
	if err != nil || len(cart.Items) == 0 {
		return nil, fmt.Errorf("empty cart")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	type line struct {
		skuID       uuid.UUID
		productID   uuid.UUID
		qty         int
		unitPrice   int64
		lineTotal   int64
		productName string
		skuCode     string
		promoEnded  bool
	}
	var lines []line
	var total int64
	var promoEndedProducts []uuid.UUID
	promoLeft := make(map[uuid.UUID]int)
	for _, it := range cart.Items {
		var l line
		l.skuID = it.SKUID
		l.qty = it.Quantity
		var promoActive bool
		var promoRemainDB int
		err := tx.QueryRow(ctx, `
			SELECT s.sale_price_cents, s.code, p.name, p.id,
			       p.promo_active, p.promo_quantity_remaining
			FROM skus s JOIN products p ON p.id = s.product_id
			WHERE s.id = $1 AND s.active = TRUE AND p.active = TRUE AND p.visible = TRUE
			FOR UPDATE OF s, p
		`, it.SKUID).Scan(&l.unitPrice, &l.skuCode, &l.productName, &l.productID, &promoActive, &promoRemainDB)
		if err != nil {
			return nil, err
		}
		promoRemain, ok := promoLeft[l.productID]
		if !ok {
			promoRemain = promoRemainDB
		}
		salePromo := l.unitPrice
		var priced catalog.PriceLineResult
		if s.catalog != nil {
			priced, err = s.catalog.PriceLine(ctx, catalog.PriceLineInput{
				ProductID: l.productID, SKUID: l.skuID, Quantity: l.qty,
				SalePrice: salePromo, PromoActive: promoActive, PromoRemain: promoRemain,
			}, s.inventory.WeightedAverageCostCents)
			if err != nil {
				return nil, err
			}
		} else {
			priced = catalog.PriceLineResult{
				LineTotalCents: salePromo * int64(l.qty),
				UnitPriceCents: salePromo,
			}
		}
		l.lineTotal = priced.LineTotalCents
		l.unitPrice = priced.UnitPriceCents
		if priced.PromoUnits > 0 && s.catalog != nil {
			_, ended, err := s.catalog.ConsumePromoQuotaTx(ctx, tx, l.productID, priced.PromoUnits)
			if err != nil {
				return nil, err
			}
			promoLeft[l.productID] = promoRemain - priced.PromoUnits
			if ended {
				l.promoEnded = true
				promoEndedProducts = append(promoEndedProducts, l.productID)
			}
		}
		total += l.lineTotal
		lines = append(lines, l)
	}

	if total > custSvc.AvailableLimit(*cust) {
		return nil, customers.ErrInsufficientLimit()
	}

	orderID := uuid.New()
	orderNumber := fmt.Sprintf("ORD-%d", time.Now().UnixNano()%1000000000)
	now := time.Now()
	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, discount_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', $4, 0, $4, $5, $6)
	`, orderID, orderNumber, customerID, total, idempotencyKey, now)
	if err != nil {
		return nil, err
	}

	for _, l := range lines {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, sku_id, product_name_snapshot, sku_code_snapshot, unit_price_cents, quantity, total_cents)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, orderID, l.skuID, l.productName, l.skuCode, l.unitPrice, l.qty, l.lineTotal)
		if err != nil {
			return nil, err
		}
		if err := s.inventory.ReserveAndDecrement(ctx, tx, l.skuID, l.qty, "order", orderID, &actorUserID); err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE customers SET current_exposure_cents = current_exposure_cents + $2, updated_at = NOW()
		WHERE id = $1
	`, customerID, total)
	if err != nil {
		return nil, err
	}

	if err := s.billing.AddOrderEntryTx(ctx, tx, customerID, orderID, total, now); err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `UPDATE carts SET status = 'checked_out', updated_at = NOW() WHERE id = $1`, cart.ID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cart.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	if s.catalog != nil && s.inventory != nil {
		seen := make(map[uuid.UUID]struct{})
		for _, l := range lines {
			if _, ok := seen[l.skuID]; ok {
				continue
			}
			seen[l.skuID] = struct{}{}
			_, _ = s.catalog.RecalculateSKU(ctx, l.skuID, actorUserID, "auto:venda", s.inventory.WeightedAverageCostCents)
		}
		for _, pid := range promoEndedProducts {
			_ = s.catalog.RecalculateProductSKUs(ctx, pid, actorUserID, "auto:promo_fim", s.inventory.WeightedAverageCostCents)
		}
	}
	return &Order{ID: orderID, OrderNumber: orderNumber, TotalCents: total, Status: "confirmed"}, nil
}

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string { return e.Message }

func AsAppError(err error) *AppError {
	var ie *inventory.AppError
	if errors.As(err, &ie) {
		return &AppError{Code: ie.Code, Message: ie.Message, Status: ie.Status}
	}
	var ce *customers.AppError
	if errors.As(err, &ce) {
		return &AppError{Code: ce.Code, Message: ce.Message, Status: ce.Status}
	}
	return nil
}
