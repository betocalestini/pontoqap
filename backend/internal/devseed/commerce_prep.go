package devseed

import (
	"context"
	"math/rand"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const demoMinCreditLimitCents = 5_000_000

func prepareCustomersForDemoCommerce(ctx context.Context, pool *pgxpool.Pool, products []productSeed, months int) error {
	var maxSale int64
	for _, p := range products {
		if p.SalePriceCents > maxSale {
			maxSale = p.SalePriceCents
		}
	}
	if months < 1 {
		months = 1
	}
	minLimit := maxSale * 4 * int64(months)
	if minLimit < demoMinCreditLimitCents {
		minLimit = demoMinCreditLimitCents
	}
	if _, err := pool.Exec(ctx, `UPDATE customers SET current_exposure_cents = 0, updated_at = NOW()`); err != nil {
		return err
	}
	_, err := pool.Exec(ctx, `
		UPDATE customers SET credit_limit_cents = GREATEST(credit_limit_cents, $1), updated_at = NOW()
	`, minLimit)
	return err
}

func customerAvailableCents(ctx context.Context, pool *pgxpool.Pool, customerID uuid.UUID) (int64, error) {
	var limit, exposure int64
	err := pool.QueryRow(ctx, `
		SELECT credit_limit_cents, current_exposure_cents FROM customers WHERE id = $1
	`, customerID).Scan(&limit, &exposure)
	if err != nil {
		return 0, err
	}
	return limit - exposure, nil
}

// pickAffordableLine chooses a product line that fits in availableLimit (unit * qty).
func pickAffordableLine(products []productSeed, available int64, rng *rand.Rand, maxQty int) (productSeed, int, bool) {
	if available <= 0 || len(products) == 0 || maxQty < 1 {
		return productSeed{}, 0, false
	}
	const tries = 24
	for i := 0; i < tries; i++ {
		p := products[rng.Intn(len(products))]
		if p.SalePriceCents <= 0 {
			continue
		}
		qty := 1 + rng.Intn(maxQty)
		if p.SalePriceCents*int64(qty) <= available {
			return p, qty, true
		}
	}
	cheapest := products[0]
	for _, p := range products[1:] {
		if p.SalePriceCents > 0 && p.SalePriceCents < cheapest.SalePriceCents {
			cheapest = p
		}
	}
	if cheapest.SalePriceCents > 0 && cheapest.SalePriceCents <= available {
		return cheapest, 1, true
	}
	return productSeed{}, 0, false
}
