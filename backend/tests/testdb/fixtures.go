package testdb

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/identity/security"
)

type Manager struct {
	UserID uuid.UUID
	Email  string
}

type Customer struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Email  string
}

type ProductFixture struct {
	ProductID uuid.UUID
	SKUID     uuid.UUID
	Price     int64
}

// SeedManager cria gerente com papel manager e todas as permissões operacionais.
func SeedManager(ctx context.Context, pool *pgxpool.Pool, email string) (Manager, error) {
	hash, err := security.HashPassword("password123")
	if err != nil {
		return Manager{}, err
	}
	userID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, status)
		VALUES ($1, 'Gerente Teste', $2, $3, 'active')
	`, userID, email, hash)
	if err != nil {
		return Manager{}, err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, 'a0000000-0000-4000-8000-000000000002')
	`, userID)
	if err != nil {
		return Manager{}, err
	}
	return Manager{UserID: userID, Email: email}, nil
}

// SeedCustomer registra cliente pendente (papel customer).
func SeedCustomer(ctx context.Context, pool *pgxpool.Pool, email, name string) (Customer, error) {
	hash, err := security.HashPassword("password123")
	if err != nil {
		return Customer{}, err
	}
	userID := uuid.New()
	customerID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, status)
		VALUES ($1, $2, $3, $4, 'active')
	`, userID, name, email, hash)
	if err != nil {
		return Customer{}, err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, 'a0000000-0000-4000-8000-000000000003')
	`, userID)
	if err != nil {
		return Customer{}, err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO customers (id, user_id, status)
		VALUES ($1, $2, 'pending')
	`, customerID, userID)
	if err != nil {
		return Customer{}, err
	}
	return Customer{ID: customerID, UserID: userID, Email: email}, nil
}

func ApproveCustomer(ctx context.Context, pool *pgxpool.Pool, customerID, managerID uuid.UUID, limitCents int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE customers SET status = 'approved', credit_limit_cents = $3,
		       approved_by = $2, approved_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, customerID, managerID, limitCents)
	return err
}

func SeedProduct(ctx context.Context, pool *pgxpool.Pool, name, skuCode string, priceCents int64) (ProductFixture, error) {
	productID := uuid.New()
	skuID := uuid.New()
	slug := fmt.Sprintf("produto-%s", skuCode)
	_, err := pool.Exec(ctx, `
		INSERT INTO products (id, name, slug, active, visible)
		VALUES ($1, $2, $3, TRUE, TRUE)
	`, productID, name, slug)
	if err != nil {
		return ProductFixture{}, err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO skus (id, product_id, code, unit, sale_price_cents, active)
		VALUES ($1, $2, $3, 'UN', $4, TRUE)
	`, skuID, productID, skuCode, priceCents)
	if err != nil {
		return ProductFixture{}, err
	}
	return ProductFixture{ProductID: productID, SKUID: skuID, Price: priceCents}, nil
}

func UniqueEmail(t *testing.T, prefix string) string {
	t.Helper()
	return fmt.Sprintf("%s-%s@test.local", prefix, uuid.NewString()[:8])
}
