package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID        uuid.UUID
	EventType string
	Payload   json.RawMessage
	Attempts  int
}

func (r *Repository) AcquireOutbox(ctx context.Context, limit int) ([]OutboxEvent, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, event_type, payload, attempts FROM outbox_events
		WHERE status = 'pending' AND available_at <= NOW()
		ORDER BY available_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent
	var ids []uuid.UUID
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.Attempts); err != nil {
			return nil, err
		}
		events = append(events, e)
		ids = append(ids, e.ID)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	for _, id := range ids {
		_, err = tx.Exec(ctx, `
			UPDATE outbox_events SET status = 'processing', attempts = attempts + 1 WHERE id = $1
		`, id)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *Repository) CompleteOutbox(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox_events SET status = 'sent', processed_at = NOW(), last_error = NULL WHERE id = $1
	`, id)
	return err
}

func (r *Repository) FailOutbox(ctx context.Context, id uuid.UUID, errMsg string, retryAfter time.Duration) error {
	status := "pending"
	available := time.Now().Add(retryAfter)
	if retryAfter <= 0 {
		status = "failed"
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox_events SET status = $2, available_at = $3, last_error = $4 WHERE id = $1
	`, id, status, available, errMsg)
	return err
}

// ProcessOutbox runs notification handlers for pending outbox rows.
func (r *Repository) ProcessOutbox(ctx context.Context, limit int, handle func(eventType string, payload json.RawMessage) error) error {
	events, err := r.AcquireOutbox(ctx, limit)
	if err != nil {
		return err
	}
	for _, ev := range events {
		hErr := handle(ev.EventType, ev.Payload)
		if hErr != nil {
			retry := 30 * time.Second
			if ev.Attempts >= 5 {
				retry = 0
			}
			_ = r.FailOutbox(ctx, ev.ID, hErr.Error(), retry)
			continue
		}
		_ = r.CompleteOutbox(ctx, ev.ID)
	}
	return nil
}
