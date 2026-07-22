package devseed

import (
	"os"
	"strconv"
)

const (
	DefaultPassword     = "DemoStore123!"
	DefaultDomain       = "demo.loja.local"
	DefaultStockQty     = 50
)

type Config struct {
	Customers int
	Products  int // legado: ignorado; produtos vêm de products.csv
	Months    int
	Password  string
	Domain    string
	DataDir         string
	DefaultStockQty int
	AppEnv          string
	SeedAllow bool
}

func DefaultConfig() Config {
	return Config{
		Customers: 60,
		Products:  40,
		Months:    4,
		Password:  DefaultPassword,
		Domain:          DefaultDomain,
		DefaultStockQty: DefaultStockQty,
		AppEnv:          "development",
		SeedAllow: false,
	}
}

func ConfigFromEnv() Config {
	cfg := DefaultConfig()
	cfg.AppEnv = os.Getenv("APP_ENV")
	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}
	cfg.SeedAllow = os.Getenv("SEED_ALLOW") == "true"
	if v := os.Getenv("SEED_CUSTOMERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Customers = n
		}
	}
	if v := os.Getenv("SEED_PRODUCTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Products = n
		}
	}
	if v := os.Getenv("SEED_MONTHS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Months = n
		}
	}
	if v := os.Getenv("SEED_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("SEED_DOMAIN"); v != "" {
		cfg.Domain = v
	}
	if v := os.Getenv("SEED_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}
	if v := os.Getenv("SEED_DEFAULT_STOCK_QTY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.DefaultStockQty = n
		}
	}
	return cfg
}
