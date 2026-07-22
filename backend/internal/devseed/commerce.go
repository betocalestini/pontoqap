package devseed

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/customers"
	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/internal/sales"
)

type commerceStats struct {
	orders        int
	closedPeriods int
	openPeriods   int
}

func seedCommerce(
	ctx context.Context,
	pool *pgxpool.Pool,
	cfg Config,
	people *People,
	products []productSeed,
	actorID uuid.UUID,
) (commerceStats, error) {
	if err := prepareCustomersForDemoCommerce(ctx, pool, products, cfg.Months); err != nil {
		return commerceStats{}, err
	}

	invSvc := inventory.NewService(pool)
	billSvc := billing.NewService(pool, nil, "http://localhost:5173", nil)
	catSvc := catalog.NewService(pool)
	custSvc := customers.NewService(pool, nil)
	salesSvc := sales.NewService(pool, invSvc, billSvc, catSvc, custSvc, nil)

	rng := rand.New(rand.NewSource(42))
	var stats commerceStats
	now := time.Now()

	for ci, custID := range people.CustomerIDs {
		userID := people.CustomerUsers[ci]
		for monthBack := cfg.Months - 1; monthBack >= 1; monthBack-- {
			ref := now.AddDate(0, -monthBack, 0)
			at := time.Date(ref.Year(), ref.Month(), 10+rng.Intn(10), 12, 0, 0, 0, time.UTC)
			nOrders := 1 + rng.Intn(3)
			var periodID uuid.UUID
			for o := 0; o < nOrders; o++ {
				avail, err := customerAvailableCents(ctx, pool, custID)
				if err != nil {
					return stats, err
				}
				prod, qty, ok := pickAffordableLine(products, avail, rng, 3)
				if !ok {
					continue
				}
				qty, ok = clampOrderQty(ctx, pool, prod, qty, avail)
				if !ok {
					continue
				}
				pid, err := placeBackdatedOrder(ctx, pool, billSvc, invSvc, salesSvc, custID, userID, actorID, prod.SKUID, qty, prod.SalePriceCents, at, fmt.Sprintf("seed-%s-%d-%d", custID.String()[:8], monthBack, o))
				if err != nil {
					return stats, err
				}
				if periodID == uuid.Nil {
					periodID = pid
				}
				stats.orders++
			}
			if periodID != uuid.Nil {
				if _, err := billSvc.ClosePeriod(ctx, periodID); err != nil {
					return stats, err
				}
				stats.closedPeriods++
			}
		}

		nCurrent := 1 + rng.Intn(3)
		for o := 0; o < nCurrent; o++ {
			avail, err := customerAvailableCents(ctx, pool, custID)
			if err != nil {
				return stats, err
			}
			prod, qty, ok := pickAffordableLine(products, avail, rng, 2)
			if !ok {
				continue
			}
			qty, ok = clampOrderQty(ctx, pool, prod, qty, avail)
			if !ok {
				continue
			}
			if _, err := salesSvc.UpsertCartItem(ctx, custID, prod.SKUID, qty); err != nil {
				return stats, err
			}
			key := fmt.Sprintf("seed-live-%s-%d", custID.String(), o)
			if _, err := salesSvc.Checkout(ctx, custID, key, userID); err != nil {
				return stats, err
			}
			stats.orders++
		}
		stats.openPeriods++
	}
	return stats, nil
}

func placeBackdatedOrder(
	ctx context.Context,
	pool *pgxpool.Pool,
	billSvc *billing.Service,
	invSvc *inventory.Service,
	salesSvc *sales.Service,
	customerID, userID, actorID, skuID uuid.UUID,
	qty int,
	unitPrice int64,
	at time.Time,
	idemKey string,
) (periodID uuid.UUID, err error) {
	total := unitPrice * int64(qty)
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	pid, err := billSvc.EnsureOpenPeriodTx(ctx, tx, customerID, at)
	if err != nil {
		return uuid.Nil, err
	}
	orderID := uuid.New()
	orderNumber := fmt.Sprintf("SEED-%s", idemKey)
	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE idempotency_key = $1)`, idemKey).Scan(&exists)
	if err != nil {
		return uuid.Nil, err
	}
	if exists {
		return pid, tx.Commit(ctx)
	}

	var productName, skuCode string
	err = tx.QueryRow(ctx, `
		SELECT p.name, s.code FROM skus s JOIN products p ON p.id = s.product_id WHERE s.id = $1
	`, skuID).Scan(&productName, &skuCode)
	if err != nil {
		return uuid.Nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, discount_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', $4, 0, $4, $5, $6)
	`, orderID, orderNumber, customerID, total, idemKey, at)
	if err != nil {
		return uuid.Nil, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO order_items (order_id, sku_id, product_name_snapshot, sku_code_snapshot, unit_price_cents, quantity, total_cents)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, orderID, skuID, productName, skuCode, unitPrice, qty, total)
	if err != nil {
		return uuid.Nil, err
	}
	if err := invSvc.ReserveAndDecrement(ctx, tx, skuID, qty, "order", orderID, &actorID); err != nil {
		return uuid.Nil, err
	}
	if err := billSvc.AddOrderEntryTx(ctx, tx, customerID, orderID, total, at); err != nil {
		return uuid.Nil, err
	}
	_, err = tx.Exec(ctx, `
		UPDATE customers SET current_exposure_cents = current_exposure_cents + $2, updated_at = NOW()
		WHERE id = $1
	`, customerID, total)
	if err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	_ = salesSvc
	return pid, nil
}
