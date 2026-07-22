package http

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/identity"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

func (h *Handler) GetMyPaymentOptions(w http.ResponseWriter, r *http.Request) {
	user, id, ok := h.meInvoiceID(w, r)
	if !ok {
		return
	}
	opts, err := h.svc.GetPaymentOptions(r.Context(), id, *user.CustomerID)
	if err != nil {
		h.writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"data": opts})
}

func (h *Handler) SelectMyPaymentPlan(w http.ResponseWriter, r *http.Request) {
	user, id, ok := h.meInvoiceID(w, r)
	if !ok {
		return
	}
	var body struct {
		InstallmentCount int `json:"installment_count"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	plan, err := h.svc.SelectPaymentPlan(r.Context(), id, *user.CustomerID, user.User.ID, body.InstallmentCount)
	if err != nil {
		h.writePlanError(w, err)
		return
	}
	if h.audit != nil {
		_ = h.audit.Log(r.Context(), &user.User.ID, "PAYMENT_PLAN_SELECTED", "invoice", &id, nil, map[string]any{
			"installment_count": body.InstallmentCount,
		})
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"data": plan})
}

func (h *Handler) GetMyPaymentPlan(w http.ResponseWriter, r *http.Request) {
	user, id, ok := h.meInvoiceID(w, r)
	if !ok {
		return
	}
	plan, err := h.svc.GetPaymentPlan(r.Context(), id, *user.CustomerID)
	if err != nil {
		h.writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"data": plan})
}

func (h *Handler) ListMyInstallments(w http.ResponseWriter, r *http.Request) {
	user, id, ok := h.meInvoiceID(w, r)
	if !ok {
		return
	}
	items, err := h.svc.ListInvoiceInstallments(r.Context(), id, *user.CustomerID)
	if err != nil {
		h.writePlanError(w, err)
		return
	}
	if items == nil {
		items = []billing.InvoiceInstallment{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) GetInstallmentPolicy(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetActiveInstallmentPolicy(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) ListInstallmentPolicies(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListInstallmentPolicies(r.Context(), 20)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) PutInstallmentPolicy(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	var body billing.UpdateInstallmentPolicyInput
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	old, _ := h.svc.GetActiveInstallmentPolicy(r.Context())
	p, err := h.svc.CreateInstallmentPolicyVersion(r.Context(), body, user.User.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if h.audit != nil {
		_ = h.audit.Log(r.Context(), &user.User.ID, "INSTALLMENT_POLICY_CHANGED", "installment_policy", &p.ID, old, p)
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) ResetPaymentPlan(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	reason := strings.TrimSpace(body.Reason)
	if len(reason) < 10 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Informe justificativa (mínimo 10 caracteres)")
		return
	}
	if err := h.svc.ResetPaymentPlan(r.Context(), id, user.User.ID, reason); err != nil {
		h.writePlanError(w, err)
		return
	}
	if h.audit != nil {
		_ = h.audit.Log(r.Context(), &user.User.ID, "PAYMENT_PLAN_OVERRIDDEN", "invoice", &id, nil, map[string]any{"reason": reason})
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) meInvoiceID(w http.ResponseWriter, r *http.Request) (*identity.AuthUser, uuid.UUID, bool) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Cliente necessário")
		return nil, uuid.Nil, false
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return nil, uuid.Nil, false
	}
	return user, id, true
}

func (h *Handler) writePlanError(w http.ResponseWriter, err error) {
	switch err {
	case billing.ErrInstallmentsDisabled:
		httpx.WriteError(w, http.StatusBadRequest, "INSTALLMENTS_DISABLED", err.Error())
	case billing.ErrInvalidInstallmentCount:
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case billing.ErrPaymentPlanNotPending, billing.ErrPaymentPlanImmutable:
		httpx.WriteError(w, http.StatusConflict, "CONFLICT", err.Error())
	case billing.ErrInstallmentAfterDue:
		httpx.WriteError(w, http.StatusBadRequest, "INSTALLMENT_AFTER_DUE", err.Error())
	case billing.ErrPeriodNotFound:
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Fatura não encontrada")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
