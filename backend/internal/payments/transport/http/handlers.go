package http

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/payments"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc *payments.Service
}

func NewHandler(svc *payments.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) MeRoutes(r chi.Router) {
	r.Post("/invoices/{id}/pix-charge", h.CreatePixCharge)
	r.Post("/installments/{installmentId}/pix-charge", h.CreateInstallmentPixCharge)
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Post("/invoices/{id}/pix-charge", h.CreatePixCharge)
}

func (h *Handler) WebhookRoutes(r chi.Router) {
	r.Post("/pix", h.WebhookPix)
}

func (h *Handler) CreatePixCharge(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	charge, err := h.svc.CreateOrReusePixCharge(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, charge)
}

func (h *Handler) CreateInstallmentPixCharge(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Cliente necessário")
		return
	}
	installmentID, err := uuid.Parse(chi.URLParam(r, "installmentId"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	charge, err := h.svc.CreateOrReusePixChargeForInstallment(r.Context(), installmentID, *user.CustomerID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, charge)
}

func (h *Handler) WebhookPix(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "payload inválido")
		return
	}
	sig := r.Header.Get("X-Webhook-Signature")
	if err := h.svc.ProcessWebhook(r.Context(), body, sig); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DevSimulate expõe confirmação sandbox (somente desenvolvimento).
func (h *Handler) DevSimulate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "chargeId"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	if err := h.svc.SimulateSandboxPayment(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
