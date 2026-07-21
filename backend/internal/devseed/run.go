package devseed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/inventory"
)

// Run popula o banco com dados de demonstração (somente dev).
func Run(ctx context.Context, pool *pgxpool.Pool, cfg Config) (Result, error) {
	var res Result
	if err := cfg.Guard(); err != nil {
		return res, err
	}

	people, err := seedPeople(ctx, pool, cfg)
	if err != nil {
		return res, fmt.Errorf("pessoas: %w", err)
	}
	res.Staff = len(staffSpecs(cfg.Domain))
	res.Customers = len(people.CustomerIDs)

	actorID, err := resolveActorID(ctx, pool, people)
	if err != nil {
		return res, err
	}

	catSvc := catalog.NewService(pool)
	invSvc := inventory.NewService(pool)
	products, err := seedProducts(ctx, pool, catSvc, invSvc, actorID, cfg)
	if err != nil {
		return res, fmt.Errorf("produtos: %w", err)
	}
	res.Products = len(products)

	stats, err := seedCommerce(ctx, pool, cfg, people, products, actorID)
	if err != nil {
		return res, fmt.Errorf("pedidos/faturas: %w", err)
	}
	res.Orders = stats.orders
	res.ClosedPeriods = stats.closedPeriods
	res.OpenPeriods = stats.openPeriods

	return res, nil
}
