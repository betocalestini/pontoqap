package http

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/platform/httpx"
)

func (h *Handler) getPricingSettings(w http.ResponseWriter, r *http.Request) {
	m, err := h.svc.GetDefaultMarginPercent(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao ler configuração")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]float64{"default_margin_percent": m})
}

func (h *Handler) patchPricingSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DefaultMarginPercent float64 `json:"default_margin_percent"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	if err := h.svc.SetDefaultMarginPercent(r.Context(), body.DefaultMarginPercent); err != nil {
		if _, ok := err.(catalog.ValidationError); ok {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao salvar")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) repriceAllProducts(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	var body struct {
		MarginPercent float64 `json:"margin_percent"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	n, err := h.svc.RepriceAllProducts(r.Context(), body.MarginPercent, user.User.ID, h.inv.WeightedAverageCostCents)
	if err != nil {
		if _, ok := err.(catalog.ValidationError); ok {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao recalcular")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]int{"products_updated": n})
}

func (h *Handler) recalculateProductSKUs(ctx context.Context, productID, changedBy uuid.UUID, reason string) {
	if h.inv == nil {
		return
	}
	_ = h.svc.RecalculateProductSKUs(ctx, productID, changedBy, reason, h.inv.WeightedAverageCostCents)
}
