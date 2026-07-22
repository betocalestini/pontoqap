package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/devseed"
	"github.com/store-platform/store/internal/platform/config"
)

func main() {
	customers := flag.Int("customers", 0, "total de clientes demo incl. CSV (0 = só CSV)")
	months := flag.Int("months", 0, "meses de histórico (0 = padrão)")
	dataDir := flag.String("data-dir", "", "pasta com products.csv e customers.csv (default: devdata ou SEED_DATA_DIR)")
	flag.Parse()

	cfg := devseed.ConfigFromEnv()
	if *customers > 0 {
		cfg.Customers = *customers
	}
	if *months > 0 {
		cfg.Months = *months
	}
	if *dataDir != "" {
		cfg.DataDir = *dataDir
	}

	appCfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.AppEnv == "" || cfg.AppEnv == "development" {
		cfg.AppEnv = appCfg.AppEnv
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, appCfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	log.Printf("Iniciando seed (clientes alvo=%d meses=%d data=%s)", cfg.Customers, cfg.Months, devseed.ResolveDataDir(cfg))
	res, err := devseed.Run(ctx, pool, cfg)
	if err != nil {
		log.Fatal(err)
	}
	res.Print(cfg)
}
