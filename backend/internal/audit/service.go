package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/platform/httpx"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Log(ctx context.Context, actorID *uuid.UUID, action, entityType string, entityID *uuid.UUID, oldVal, newVal any) error {
	reqID := httpx.RequestIDFromContext(ctx)
	var oldJSON, newJSON []byte
	if oldVal != nil {
		oldJSON, _ = json.Marshal(oldVal)
	}
	if newVal != nil {
		newJSON, _ = json.Marshal(newVal)
	}
	var rid *uuid.UUID
	if reqID != "" {
		if u, err := uuid.Parse(reqID); err == nil {
			rid = &u
		}
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, request_id, old_values, new_values)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, actorID, action, entityType, entityID, rid, oldJSON, newJSON)
	return err
}
