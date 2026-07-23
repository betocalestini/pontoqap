package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TypeMonthlyClose      = "billing.monthly_close"
	TypeMarkOverdue       = "billing.mark_overdue"
	TypeMercadoPagoOrder  = "payments.mercadopago_order"
)

type Job struct {
	ID      uuid.UUID
	Type    string
	Payload json.RawMessage
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Enqueue(ctx context.Context, jobType string, payload any) error {
	return r.EnqueueAvailableAt(ctx, jobType, payload, time.Now())
}

func (r *Repository) EnqueueAvailableAt(ctx context.Context, jobType string, payload any, availableAt time.Time) error {
	b, _ := json.Marshal(payload)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO jobs (id, type, payload, status, available_at)
		VALUES ($1, $2, $3, 'pending', $4)
	`, uuid.New(), jobType, b, availableAt)
	return err
}

func (r *Repository) Acquire(ctx context.Context, workerID string, limit int) ([]Job, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, type, payload FROM jobs
		WHERE status = 'pending' AND available_at <= NOW()
		ORDER BY available_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	var ids []uuid.UUID
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.Type, &j.Payload); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
		ids = append(ids, j.ID)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	for _, id := range ids {
		_, err = tx.Exec(ctx, `
			UPDATE jobs SET status = 'running', locked_at = NOW(), locked_by = $2, attempts = attempts + 1
			WHERE id = $1
		`, id, workerID)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *Repository) Complete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE jobs SET status = 'completed', completed_at = NOW(), locked_at = NULL, locked_by = NULL WHERE id = $1
	`, id)
	return err
}

func (r *Repository) Fail(ctx context.Context, id uuid.UUID, jobErr error, retryAfter time.Duration) error {
	status := "pending"
	available := time.Now().Add(retryAfter)
	var lastErr *string
	if jobErr != nil {
		s := jobErr.Error()
		lastErr = &s
	}
	if retryAfter <= 0 {
		status = "failed"
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE jobs SET status = $2, available_at = $3, last_error = $4, locked_at = NULL, locked_by = NULL WHERE id = $1
	`, id, status, available, lastErr)
	return err
}

func (r *Repository) PublishOutbox(ctx context.Context, tx pgx.Tx, eventType, aggregateType string, aggregateID uuid.UUID, payload any) error {
	b, _ := json.Marshal(payload)
	_, err := tx.Exec(ctx, `
		INSERT INTO outbox_events (id, event_type, aggregate_type, aggregate_id, payload, status, available_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', NOW())
	`, uuid.New(), eventType, aggregateType, aggregateID, b)
	return err
}
