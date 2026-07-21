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
	customers := flag.Int("customers", 0, "número de clientes demo (0 = padrão)")
	products := flag.Int("products", 0, "número de produtos seed (0 = padrão)")
	months := flag.Int("months", 0, "meses de histórico (0 = padrão)")
	flag.Parse()

	cfg := devseed.ConfigFromEnv()
	if *customers > 0 {
		cfg.Customers = *customers
	}
	if *products > 0 {
		cfg.Products = *products
	}
	if *months > 0 {
		cfg.Months = *months
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

	log.Printf("Iniciando seed (clientes=%d produtos=%d meses=%d)", cfg.Customers, cfg.Products, cfg.Months)
	res, err := devseed.Run(ctx, pool, cfg)
	if err != nil {
		log.Fatal(err)
	}
	res.Print(cfg)
}
