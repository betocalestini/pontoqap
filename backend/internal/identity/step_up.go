package identity

import (
	"context"

	"github.com/pquerna/otp/totp"

	"github.com/store-platform/store/internal/identity/security"
)

// VerifyStepUp confirms password or MFA for sensitive admin operations.
func VerifyStepUp(ctx context.Context, repo Repository, actor User, password, mfaCode string) error {
	if mfaCode != "" && actor.MFAEnabled {
		if totp.Validate(mfaCode, actor.MFASecret) {
			return nil
		}
		return errInvalidCredentials()
	}
	if password != "" {
		ok, err := security.VerifyPassword(actor.PasswordHash, password)
		if err == nil && ok {
			return nil
		}
	}
	return &AppError{Code: "STEP_UP_REQUIRED", Message: "Confirme com sua senha ou código MFA", Status: 403}
}
