package testdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultURL = "postgres://store:store@localhost:5432/store?sslmode=disable"

var resetMu sync.Mutex

// Pool conecta ao PostgreSQL de testes ou pula o teste se DATABASE_URL não estiver disponível.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = defaultURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("PostgreSQL indisponível: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("PostgreSQL indisponível: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// MigrateUp aplica migrations pendentes (idempotente). Pula se o banco não estiver acessível.
func MigrateUp(t *testing.T) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = defaultURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	probe, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("PostgreSQL indisponível: %v", err)
	}
	if err := probe.Ping(ctx); err != nil {
		probe.Close()
		t.Skipf("PostgreSQL indisponível: %v", err)
	}
	probe.Close()
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("runtime.Caller failed")
		}
		migrationsPath = filepath.Join(filepath.Dir(file), "..", "..", "migrations")
	}
	abs, err := filepath.Abs(migrationsPath)
	if err != nil {
		t.Fatal(err)
	}
	sourceURL := fmt.Sprintf("file://%s", abs)
	dbURL := url
	if strings.HasPrefix(dbURL, "postgres://") {
		dbURL = "pgx5://" + strings.TrimPrefix(dbURL, "postgres://")
	}
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		t.Skipf("migrate indisponível: %v", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}
}

// Reset remove dados transacionais preservando papéis, permissões e local de estoque padrão.
func Reset(ctx context.Context, pool *pgxpool.Pool) error {
	resetMu.Lock()
	defer resetMu.Unlock()
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			audit_logs, outbox_events, jobs, forecast_snapshots,
			payment_events, payments, payment_charges,
			invoice_installments, invoice_payment_plans,
			billing_adjustments, invoice_items, invoices, billing_entries, billing_periods, business_calendar,
			order_return_items, order_returns, order_items, orders, cart_items, carts,
			stock_movements, inventory_balances,
			price_history, product_images, skus, products, categories,
			customer_limit_history, customers,
			email_verification_tokens, password_reset_tokens, sessions, user_roles, users
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		return err
	}
	return EnsureDefaultInstallmentPolicy(ctx, pool)
}

// EnsureDefaultInstallmentPolicy garante política ativa após reset de dados de teste.
func EnsureDefaultInstallmentPolicy(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `UPDATE installment_policies SET active = false`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO installment_policies (
			id, version, active, installment_enabled,
			minimum_invoice_amount_cents, minimum_installment_amount_cents, maximum_installments,
			installment_interval_months, allow_installment_after_due_date, allow_early_installment_payment,
			require_sequential_payment, adjust_due_date_to_business_day, valid_from
		) VALUES (
			'c0000000-0000-4000-8000-000000000001', 1, true, true,
			30000, 10000, 10,
			1, false, false,
			true, true, now()
		)
		ON CONFLICT (id) DO UPDATE SET active = true, installment_enabled = EXCLUDED.installment_enabled
	`)
	return err
}
