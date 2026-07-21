package devseed_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/devseed"
)

func TestRun_smoke(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://store:store@localhost:5432/store?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("postgres indisponível: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("postgres indisponível: %v", err)
	}

	cfg := devseed.DefaultConfig()
	cfg.Customers = 2
	cfg.Products = 2
	cfg.Months = 2
	cfg.SeedAllow = true
	cfg.Domain = "smoke-" + time.Now().Format("150405") + ".demo.loja.local"

	_, err = devseed.Run(ctx, pool, cfg)
	if err != nil {
		t.Fatal(err)
	}
}
