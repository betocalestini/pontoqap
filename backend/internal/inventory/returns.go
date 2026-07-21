package inventory

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RestoreFromSaleTx devolve quantidade ao estoque (cancelamento de pedido).
func (s *Service) RestoreFromSaleTx(ctx context.Context, tx pgx.Tx, skuID uuid.UUID, quantity int, orderID uuid.UUID, createdBy uuid.UUID) error {
	locID, _ := uuid.Parse(DefaultLocationID)
	ref := orderID
	return s.applyMovementTx(ctx, tx, locID, skuID, movementApply{
		movementType:  MovementReturn,
		quantityDelta: quantity,
		reason:        "Cancelamento de pedido",
		refType:       "order_cancel",
		refID:         &ref,
		createdBy:     createdBy,
	})
}
