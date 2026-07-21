package identity

import (
	"context"
	"strings"

	"github.com/google/uuid"

	platformerrors "github.com/store-platform/store/internal/platform/errors"
)

// UpdateProfileInput campos editáveis pelo próprio usuário.
type UpdateProfileInput struct {
	Name     *string
	Phone    *string
	Document *string
}

func (s *Service) GetCustomerDocument(ctx context.Context, customerID uuid.UUID) (string, error) {
	return s.repo.GetCustomerDocument(ctx, customerID)
}

func (s *Service) UpdateProfile(ctx context.Context, auth AuthUser, in UpdateProfileInput) (AuthUser, error) {
	name := auth.User.Name
	phone := auth.User.Phone
	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return AuthUser{}, errValidation("Nome é obrigatório")
		}
		name = n
	}
	if in.Phone != nil {
		phone = strings.TrimSpace(*in.Phone)
	}
	if err := s.repo.UpdateUserProfile(ctx, auth.User.ID, name, phone); err != nil {
		return AuthUser{}, err
	}
	if auth.CustomerID != nil && in.Document != nil {
		doc := strings.TrimSpace(*in.Document)
		if err := s.repo.UpdateCustomerDocument(ctx, *auth.CustomerID, doc); err != nil {
			return AuthUser{}, err
		}
	}
	user, err := s.repo.FindUserByID(ctx, auth.User.ID)
	if err != nil || user == nil {
		return AuthUser{}, errNotFound()
	}
	return s.buildAuthUser(ctx, *user)
}

func errValidation(msg string) error {
	return &AppError{Code: platformerrors.CodeValidation, Message: msg, Status: 400}
}
