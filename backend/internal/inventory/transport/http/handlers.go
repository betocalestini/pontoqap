package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/inventory"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc *inventory.Service
}

func NewHandler(svc *inventory.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Post("/entries", h.registerEntry)
}

func (h *Handler) RegisterEntry(w http.ResponseWriter, r *http.Request) {
	h.registerEntry(w, r)
}

func (h *Handler) registerEntry(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SKUID    string `json:"sku_id"`
		Quantity int    `json:"quantity"`
		Reason   string `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || body.Quantity <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	skuID, err := uuid.Parse(body.SKUID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "SKU inválido")
		return
	}
	user := identityhttp.UserFromContext(r.Context())
	if err := h.svc.RegisterEntry(r.Context(), skuID, body.Quantity, user.User.ID, body.Reason); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao registrar entrada")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
