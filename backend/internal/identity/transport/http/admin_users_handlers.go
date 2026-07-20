package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/identity"
	"github.com/store-platform/store/internal/platform/httpx"
)

type AdminUsersHandler struct {
	svc *identity.AdminUsersService
}

func NewAdminUsersHandler(svc *identity.AdminUsersService) *AdminUsersHandler {
	return &AdminUsersHandler{svc: svc}
}

func (h *AdminUsersHandler) Routes(r chi.Router) {
	r.With(RequirePermission("users.manage")).Get("/users", h.listUsers)
	r.With(RequirePermission("users.manage")).Get("/users/{id}", h.getUser)
	r.With(RequirePermission("users.manage")).Post("/users/invitations", h.createInvitation)
	r.With(RequirePermission("users.manage")).Post("/users/invitations/{id}/revoke", h.revokeInvitation)
	r.With(RequirePermission("users.manage")).Patch("/users/{id}/role", h.setRole)
	r.With(RequirePermission("users.manage")).Patch("/users/{id}/status", h.setStatus)
	r.With(RequirePermission("users.manage")).Post("/users/{id}/sessions/revoke", h.revokeSessions)
	r.With(RequirePermission("users.manage")).Get("/roles", h.listRoles)
}

func (h *AdminUsersHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListStaff(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar usuários")
		return
	}
	if items == nil {
		items = []identity.StaffUserSummary{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *AdminUsersHandler) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	user, perms, err := h.svc.GetStaff(r.Context(), id)
	if err != nil {
		writeAppError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"user":        user,
		"permissions": perms,
	})
}

func (h *AdminUsersHandler) createInvitation(w http.ResponseWriter, r *http.Request) {
	actor := UserFromContext(r.Context())
	if actor == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	var body struct {
		Email  string `json:"email"`
		Name   string `json:"name"`
		RoleID string `json:"role_id"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	roleID, err := uuid.Parse(body.RoleID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Papel inválido")
		return
	}
	if err := h.svc.CreateInvitation(r.Context(), *actor, identity.CreateInvitationInput{
		Email: body.Email, Name: body.Name, RoleID: roleID,
	}); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminUsersHandler) revokeInvitation(w http.ResponseWriter, r *http.Request) {
	actor := UserFromContext(r.Context())
	if actor == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	if err := h.svc.RevokeInvitation(r.Context(), *actor, id); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminUsersHandler) setRole(w http.ResponseWriter, r *http.Request) {
	actor := UserFromContext(r.Context())
	if actor == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		RoleID   string `json:"role_id"`
		Password string `json:"password"`
		MFACode  string `json:"mfa_code"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	roleID, err := uuid.Parse(body.RoleID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Papel inválido")
		return
	}
	if err := h.svc.SetUserRole(r.Context(), *actor, id, roleID, body.Password, body.MFACode); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminUsersHandler) setStatus(w http.ResponseWriter, r *http.Request) {
	actor := UserFromContext(r.Context())
	if actor == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		Status   string `json:"status"`
		Password string `json:"password"`
		MFACode  string `json:"mfa_code"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	if err := h.svc.SetUserStatus(r.Context(), *actor, id, body.Status, body.Password, body.MFACode); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminUsersHandler) revokeSessions(w http.ResponseWriter, r *http.Request) {
	actor := UserFromContext(r.Context())
	if actor == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		Password string `json:"password"`
		MFACode  string `json:"mfa_code"`
	}
	_ = httpx.DecodeJSON(r, &body)
	if err := h.svc.RevokeUserSessions(r.Context(), *actor, id, body.Password, body.MFACode); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminUsersHandler) listRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.svc.ListInternalRoles(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar papéis")
		return
	}
	if roles == nil {
		roles = []identity.RoleInfo{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": roles})
}

func (h *Handler) acceptInvitation(w http.ResponseWriter, r *http.Request) {
	if h.adminUsers == nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Indisponível")
		return
	}
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	if err := h.adminUsers.AcceptInvitation(r.Context(), identity.AcceptInvitationInput{
		Token: body.Token, Password: body.Password, Name: body.Name,
	}); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
