package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/identity/security"
	"github.com/store-platform/store/internal/jobs"
	"github.com/store-platform/store/internal/notification"
	"github.com/store-platform/store/internal/platform/config"
	platformerrors "github.com/store-platform/store/internal/platform/errors"
)

const verificationTTL = 24 * time.Hour

type VerificationService struct {
	pool          *pgxpool.Pool
	jobs          *jobs.Repository
	storeWebURL   string
	defaultLimit  int64
}

func NewVerificationService(pool *pgxpool.Pool, jobRepo *jobs.Repository, app config.AppConfig, cust config.CustomerConfig) *VerificationService {
	return &VerificationService{
		pool:         pool,
		jobs:         jobRepo,
		storeWebURL:  app.StoreWebURL,
		defaultLimit: cust.DefaultCreditLimitCents,
	}
}

func (v *VerificationService) EnqueueVerification(ctx context.Context, tx pgx.Tx, userID uuid.UUID, email, name string) error {
	raw, hash, err := newVerificationToken()
	if err != nil {
		return err
	}
	expires := time.Now().Add(verificationTTL)
	_, err = tx.Exec(ctx, `
		UPDATE email_verification_tokens SET used_at = NOW()
		WHERE user_id = $1 AND used_at IS NULL
	`, userID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO email_verification_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, uuid.New(), userID, hash, expires)
	if err != nil {
		return err
	}
	verifyURL := notification.BuildVerifyURL(v.storeWebURL, raw)
	payload := map[string]string{
		"to":         email,
		"name":       name,
		"verify_url": verifyURL,
	}
	return v.jobs.PublishOutbox(ctx, tx, notification.EventUserVerification, "user", userID, payload)
}

func (v *VerificationService) VerifyEmail(ctx context.Context, rawToken string) error {
	hash := security.HashToken(rawToken)
	tx, err := v.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID uuid.UUID
	var expires time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id, expires_at FROM email_verification_tokens
		WHERE token_hash = $1 AND used_at IS NULL
		FOR UPDATE
	`, hash).Scan(&userID, &expires)
	if err == pgx.ErrNoRows {
		return errNotFound()
	}
	if err != nil {
		return err
	}
	if time.Now().After(expires) {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Link expirado", Status: 400}
	}

	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE email_verification_tokens SET used_at = $2 WHERE token_hash = $1
	`, hash, now)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE users SET status = 'active', email_verified_at = $2, updated_at = NOW()
		WHERE id = $1
	`, userID, now)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE customers SET status = 'approved', credit_limit_cents = $2,
		       approved_at = $3, updated_at = NOW()
		WHERE user_id = $1 AND status = 'pending'
	`, userID, v.defaultLimit, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		_, _ = tx.Exec(ctx, `
			UPDATE customers SET status = 'approved', updated_at = NOW()
			WHERE user_id = $1 AND status != 'blocked'
		`, userID)
	}
	return tx.Commit(ctx)
}

func (v *VerificationService) ResendVerification(ctx context.Context, email string) error {
	var userID uuid.UUID
	var name string
	var status string
	err := v.pool.QueryRow(ctx, `
		SELECT id, name, status FROM users WHERE LOWER(email) = LOWER($1)
	`, email).Scan(&userID, &name, &status)
	if err == pgx.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if status != "pending_email" {
		return nil
	}
	tx, err := v.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := v.EnqueueVerification(ctx, tx, userID, email, name); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func newVerificationToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	hash = security.HashToken(raw)
	return raw, hash, nil
}

func errEmailNotVerified() error {
	return &AppError{
		Code:    platformerrors.CodeEmailNotVerified,
		Message: "Confirme seu e-mail antes de entrar. Verifique sua caixa de entrada ou solicite um novo envio.",
		Status:  401,
	}
}

// ConfirmEmailForTest ativa usuário e cliente (somente testes).
func ConfirmEmailForTest(ctx context.Context, pool *pgxpool.Pool, email string, limit int64) error {
	now := time.Now()
	_, err := pool.Exec(ctx, `
		UPDATE users u SET status = 'active', email_verified_at = $2, updated_at = NOW()
		FROM customers c
		WHERE LOWER(u.email) = LOWER($1) AND c.user_id = u.id
	`, email, now)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		UPDATE customers c SET status = 'approved', credit_limit_cents = $2,
		       approved_at = $3, updated_at = NOW()
		FROM users u
		WHERE LOWER(u.email) = LOWER($1) AND c.user_id = u.id
	`, email, limit, now)
	return err
}

func (v *VerificationService) DefaultLimit() int64 {
	return v.defaultLimit
}
