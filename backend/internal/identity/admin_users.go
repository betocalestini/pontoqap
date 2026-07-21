package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/audit"
	"github.com/store-platform/store/internal/identity/security"
	"github.com/store-platform/store/internal/jobs"
	"github.com/store-platform/store/internal/notification"
	platformerrors "github.com/store-platform/store/internal/platform/errors"
)

const invitationTTL = 48 * time.Hour

type AdminUsersService struct {
	pool        *pgxpool.Pool
	repo        Repository
	adminRepo   AdminUsersRepository
	jobs        *jobs.Repository
	audit       *audit.Service
	adminWebURL string
}

func NewAdminUsersService(pool *pgxpool.Pool, repo Repository, adminRepo AdminUsersRepository, jobRepo *jobs.Repository, auditSvc *audit.Service, adminWebURL string) *AdminUsersService {
	return &AdminUsersService{
		pool:        pool,
		repo:        repo,
		adminRepo:   adminRepo,
		jobs:        jobRepo,
		audit:       auditSvc,
		adminWebURL: adminWebURL,
	}
}

type StaffUserSummary struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	Roles       []string   `json:"roles"`
	MFAEnabled  bool       `json:"mfa_enabled"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type RoleInfo struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

type CreateInvitationInput struct {
	Email  string
	Name   string
	RoleID uuid.UUID
}

func (s *AdminUsersService) ListStaff(ctx context.Context) ([]StaffUserSummary, error) {
	return s.adminRepo.ListStaff(ctx)
}

func (s *AdminUsersService) GetStaff(ctx context.Context, userID uuid.UUID) (*StaffUserSummary, []string, error) {
	summary, err := s.adminRepo.GetStaffByID(ctx, userID)
	if err != nil || summary == nil {
		return nil, nil, errNotFound()
	}
	perms, err := s.repo.ListUserPermissions(ctx, userID)
	return summary, perms, err
}

func (s *AdminUsersService) ListInternalRoles(ctx context.Context) ([]RoleInfo, error) {
	return s.adminRepo.ListInternalRoles(ctx)
}

func (s *AdminUsersService) CreateInvitation(ctx context.Context, actor AuthUser, in CreateInvitationInput) error {
	if !HasSystemAdminRole(actor.Roles) {
		roleCode, err := s.adminRepo.RoleCodeByID(ctx, in.RoleID)
		if err != nil {
			return err
		}
		if roleCode == "system_admin" {
			return errForbidden("Somente administradores podem convidar outro administrador")
		}
	}
	roleCode, err := s.adminRepo.RoleCodeByID(ctx, in.RoleID)
	if err != nil {
		return err
	}
	if !IsStaffRole(roleCode) {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Papel inválido", Status: 400}
	}
	email := strings.TrimSpace(strings.ToLower(in.Email))
	name := strings.TrimSpace(in.Name)
	if email == "" || name == "" {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Nome e e-mail obrigatórios", Status: 400}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	userID, existingStatus, err := s.adminRepo.FindUserIDByEmailTx(ctx, tx, email)
	if err != nil {
		return err
	}
	if userID == nil {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Cadastro na loja obrigatório antes de função interna", Status: 409}
	}
	hasCustomerRecord, err := s.adminRepo.UserHasCustomerRecordTx(ctx, tx, *userID)
	if err != nil {
		return err
	}
	if !hasCustomerRecord {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Cadastro na loja obrigatório antes de função interna", Status: 409}
	}
	roles, _ := s.repo.ListUserRoles(ctx, *userID)
	if existingStatus == "active" || existingStatus == "suspended" {
		if HasStaffRole(roles) {
			return &AppError{Code: platformerrors.CodeValidation, Message: "Usuário já possui função interna; altere o papel em Usuários ou no cadastro do cliente", Status: 409}
		}
		_, err = tx.Exec(ctx, `UPDATE users SET name = $2, updated_at = NOW() WHERE id = $1`, *userID, name)
		if err != nil {
			return err
		}
	} else if existingStatus == "invited" {
		_, err = tx.Exec(ctx, `UPDATE users SET name = $2, status = 'invited', updated_at = NOW() WHERE id = $1`, *userID, name)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec(ctx, `UPDATE users SET name = $2, updated_at = NOW() WHERE id = $1`, *userID, name)
		if err != nil {
			return err
		}
	}

	raw, hash, err := newInviteToken()
	if err != nil {
		return err
	}
	invID := uuid.New()
	expires := time.Now().Add(invitationTTL)
	_, err = tx.Exec(ctx, `
		UPDATE admin_invitations SET revoked_at = NOW()
		WHERE LOWER(email) = LOWER($1) AND accepted_at IS NULL AND revoked_at IS NULL
	`, email)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO admin_invitations (id, email, role_id, token_hash, expires_at, invited_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, invID, email, in.RoleID, hash, expires, actor.User.ID)
	if err != nil {
		return err
	}

	inviteURL := notification.BuildAdminInviteURL(s.adminWebURL, raw)
	payload := map[string]string{
		"to":         email,
		"name":       name,
		"invite_url": inviteURL,
	}
	if err := s.jobs.PublishOutbox(ctx, tx, notification.EventAdminInvitation, "admin_invitation", invID, payload); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &actor.User.ID, "invitation.created", "admin_invitation", &invID, nil, map[string]any{
		"email": email, "role_id": in.RoleID.String(),
	})
	return nil
}

func (s *AdminUsersService) RevokeInvitation(ctx context.Context, actor AuthUser, invitationID uuid.UUID) error {
	inv, err := s.adminRepo.FindInvitation(ctx, invitationID)
	if err != nil || inv == nil {
		return errNotFound()
	}
	if inv.AcceptedAt != nil || inv.RevokedAt != nil {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Convite já finalizado", Status: 400}
	}
	if err := s.adminRepo.RevokeInvitation(ctx, invitationID); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &actor.User.ID, "invitation.revoked", "admin_invitation", &invitationID, nil, nil)
	return nil
}

type AcceptInvitationInput struct {
	Token    string
	Password string
	Name     string
}

func (s *AdminUsersService) AcceptInvitation(ctx context.Context, in AcceptInvitationInput) error {
	hash := security.HashToken(strings.TrimSpace(in.Token))
	inv, err := s.adminRepo.FindInvitationByTokenHash(ctx, hash)
	if err != nil || inv == nil {
		return errNotFound()
	}
	if inv.RevokedAt != nil {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Convite revogado", Status: 400}
	}
	if inv.AcceptedAt != nil {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Convite já utilizado", Status: 400}
	}
	if time.Now().After(inv.ExpiresAt) {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Convite expirado", Status: 400}
	}
	pwHash, err := security.HashPassword(in.Password)
	if err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID uuid.UUID
	var userStatus string
	var passwordHash string
	err = tx.QueryRow(ctx, `
		SELECT id, status, password_hash FROM users WHERE LOWER(email) = LOWER($1) FOR UPDATE
	`, inv.Email).Scan(&userID, &userStatus, &passwordHash)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		_ = tx.QueryRow(ctx, `SELECT name FROM users WHERE id = $1`, userID).Scan(&name)
	}
	if userStatus == "active" {
		ok, verr := security.VerifyPassword(passwordHash, in.Password)
		if verr != nil || !ok {
			return errInvalidCredentials()
		}
		_, err = tx.Exec(ctx, `UPDATE users SET name = $2, updated_at = NOW() WHERE id = $1`, userID, name)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE users SET name = $2, password_hash = $3, status = 'active', updated_at = NOW()
			WHERE id = $1
		`, userID, name, pwHash)
	}
	if err != nil {
		return err
	}
	if err := s.adminRepo.ReplaceUserRoleTx(ctx, tx, userID, inv.RoleID); err != nil {
		return err
	}
	now := time.Now()
	_, err = tx.Exec(ctx, `UPDATE admin_invitations SET accepted_at = $2 WHERE id = $1`, inv.ID, now)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &userID, "invitation.accepted", "admin_invitation", &inv.ID, nil, map[string]any{"role_id": inv.RoleID.String()})
	return nil
}

func (s *AdminUsersService) AssignStaffRoleFromCustomer(ctx context.Context, actor AuthUser, customerID uuid.UUID, roleID uuid.UUID, password, mfaCode string) error {
	if err := VerifyStepUp(ctx, s.repo, actor.User, password, mfaCode); err != nil {
		return err
	}
	var userID uuid.UUID
	var userStatus string
	err := s.pool.QueryRow(ctx, `
		SELECT c.user_id, u.status FROM customers c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = $1
	`, customerID).Scan(&userID, &userStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return errNotFound()
	}
	if err != nil {
		return err
	}
	if userStatus == "disabled" {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Conta desativada", Status: 400}
	}
	newRole, err := s.adminRepo.RoleCodeByID(ctx, roleID)
	if err != nil || !IsStaffRole(newRole) {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Papel inválido", Status: 400}
	}
	if newRole == "system_admin" && !HasSystemAdminRole(actor.Roles) {
		return errForbidden("Somente administradores podem atribuir papel de administrador")
	}
	targetRoles, err := s.repo.ListUserRoles(ctx, userID)
	if err != nil {
		return err
	}
	if HasSystemAdminRole(targetRoles) && !HasSystemAdminRole(actor.Roles) {
		return errForbidden("Sem permissão para alterar administrador")
	}
	oldRole := firstStaffRole(targetRoles)
	if newRole == "system_admin" && oldRole != "system_admin" && !HasSystemAdminRole(actor.Roles) {
		return errForbidden("Somente administradores podem promover a administrador")
	}
	if oldRole == "system_admin" && newRole != "system_admin" {
		if err := s.ensureNotLastActiveAdmin(ctx, userID); err != nil {
			return err
		}
	}
	if oldRole == newRole {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Cliente já possui este papel interno", Status: 409}
	}
	if err := s.adminRepo.ReplaceUserRole(ctx, userID, roleID); err != nil {
		return err
	}
	_ = s.repo.RevokeUserSessions(ctx, userID, "admin")
	_ = s.audit.Log(ctx, &actor.User.ID, "user.role_assigned_from_customer", "customer", &customerID,
		map[string]string{"role": oldRole}, map[string]string{"role": newRole, "user_id": userID.String()})
	return nil
}

func firstStaffRole(roles []string) string {
	for _, r := range roles {
		if IsStaffRole(r) {
			return r
		}
	}
	return ""
}

func (s *AdminUsersService) SetUserRole(ctx context.Context, actor AuthUser, targetID uuid.UUID, roleID uuid.UUID, password, mfaCode string) error {
	if err := VerifyStepUp(ctx, s.repo, actor.User, password, mfaCode); err != nil {
		return err
	}
	target, err := s.adminRepo.GetStaffByID(ctx, targetID)
	if err != nil || target == nil {
		return errNotFound()
	}
	newRole, err := s.adminRepo.RoleCodeByID(ctx, roleID)
	if err != nil || !IsStaffRole(newRole) {
		return &AppError{Code: platformerrors.CodeValidation, Message: "Papel inválido", Status: 400}
	}
	if newRole == "system_admin" && !HasSystemAdminRole(actor.Roles) {
		return errForbidden("Somente administradores podem atribuir papel de administrador")
	}
	if HasSystemAdminRole(target.Roles) && !HasSystemAdminRole(actor.Roles) {
		return errForbidden("Sem permissão para alterar administrador")
	}
	oldRole := ""
	if len(target.Roles) > 0 {
		oldRole = target.Roles[0]
	}
	if targetID == actor.User.ID {
		return errForbidden("Não é possível alterar o próprio papel")
	}
	if newRole == "system_admin" && oldRole != "system_admin" && !HasSystemAdminRole(actor.Roles) {
		return errForbidden("Somente administradores podem promover a administrador")
	}
	if oldRole == "system_admin" && newRole != "system_admin" {
		if err := s.ensureNotLastActiveAdmin(ctx, targetID); err != nil {
			return err
		}
	}
	if err := s.adminRepo.ReplaceUserRole(ctx, targetID, roleID); err != nil {
		return err
	}
	_ = s.repo.RevokeUserSessions(ctx, targetID, "admin")
	_ = s.audit.Log(ctx, &actor.User.ID, "user.role_changed", "admin_user", &targetID,
		map[string]string{"role": oldRole}, map[string]string{"role": newRole})
	return nil
}

func (s *AdminUsersService) SetUserStatus(ctx context.Context, actor AuthUser, targetID uuid.UUID, status, password, mfaCode string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "active", "suspended", "disabled":
	default:
		return &AppError{Code: platformerrors.CodeValidation, Message: "Status inválido", Status: 400}
	}
	if status == "disabled" || status == "suspended" {
		if err := VerifyStepUp(ctx, s.repo, actor.User, password, mfaCode); err != nil {
			return err
		}
	}
	target, err := s.adminRepo.GetStaffByID(ctx, targetID)
	if err != nil || target == nil {
		return errNotFound()
	}
	if targetID == actor.User.ID && status != "active" {
		return errForbidden("Não é possível alterar o próprio status")
	}
	if HasSystemAdminRole(target.Roles) && (status == "disabled" || status == "suspended") {
		if err := s.ensureNotLastActiveAdmin(ctx, targetID); err != nil {
			return err
		}
	}
	oldStatus := target.Status
	if err := s.adminRepo.UpdateUserStatus(ctx, targetID, status); err != nil {
		return err
	}
	if status == "disabled" || status == "suspended" {
		_ = s.repo.RevokeUserSessions(ctx, targetID, "admin")
	}
	_ = s.audit.Log(ctx, &actor.User.ID, "user.status_changed", "admin_user", &targetID,
		map[string]string{"status": oldStatus}, map[string]string{"status": status})
	return nil
}

func (s *AdminUsersService) RevokeUserSessions(ctx context.Context, actor AuthUser, targetID uuid.UUID, password, mfaCode string) error {
	if err := VerifyStepUp(ctx, s.repo, actor.User, password, mfaCode); err != nil {
		return err
	}
	target, err := s.adminRepo.GetStaffByID(ctx, targetID)
	if err != nil || target == nil {
		return errNotFound()
	}
	if err := s.repo.RevokeUserSessions(ctx, targetID, "admin"); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &actor.User.ID, "user.sessions_revoked", "admin_user", &targetID, nil, nil)
	return nil
}

func (s *AdminUsersService) ensureNotLastActiveAdmin(ctx context.Context, excludeUserID uuid.UUID) error {
	n, err := s.adminRepo.CountActiveSystemAdmins(ctx, excludeUserID)
	if err != nil {
		return err
	}
	if n < 1 {
		return errForbidden("Não é possível desativar o último administrador ativo")
	}
	return nil
}

func newInviteToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	hash = security.HashToken(raw)
	return raw, hash, nil
}

func randomUnusablePassword() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
