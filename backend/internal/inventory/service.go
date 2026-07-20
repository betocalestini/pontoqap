package inventory

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	platformerrors "github.com/store-platform/store/internal/platform/errors"
)

const DefaultLocationID = "c0000000-0000-4000-8000-000000000001"

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Balance struct {
	SKUID             uuid.UUID `json:"sku_id"`
	AvailableQuantity int       `json:"available_quantity"`
}

func (s *Service) ReserveAndDecrement(ctx context.Context, tx pgx.Tx, skuID uuid.UUID, quantity int, refType string, refID uuid.UUID, createdBy *uuid.UUID) error {
	locID, _ := uuid.Parse(DefaultLocationID)
	createdByID := uuid.Nil
	if createdBy != nil {
		createdByID = *createdBy
	}
	return s.applyMovementTx(ctx, tx, locID, skuID, movementApply{
		movementType:  MovementSale,
		quantityDelta: -quantity,
		refType:       refType,
		refID:         &refID,
		createdBy:     createdByID,
	})
}

func ErrInsufficientStock() error {
	return &AppError{Code: platformerrors.CodeInsufficientStock, Message: "Estoque insuficiente", Status: 422}
}

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string { return e.Message }
