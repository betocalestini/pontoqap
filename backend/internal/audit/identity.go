package audit

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) LogAdminLogin(ctx context.Context, userID uuid.UUID) error {
	return s.Log(ctx, &userID, "auth.admin_login", "admin_user", &userID, nil, nil)
}
