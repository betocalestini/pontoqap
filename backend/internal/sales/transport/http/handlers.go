package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
	"github.com/store-platform/store/internal/sales"
)

type Handler struct {
	svc *sales.Service
}

func NewHandler(svc *sales.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) MeRoutes(r chi.Router) {
	r.Get("/cart", h.getCart)
	r.Post("/cart/items", h.upsertItem)
	r.Post("/cart/checkout", h.checkout)
}

func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Perfil de cliente necessário")
		return
	}
	cart, err := h.svc.GetOrCreateCart(r.Context(), *user.CustomerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao carregar carrinho")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, cart)
}

func (h *Handler) upsertItem(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Perfil de cliente necessário")
		return
	}
	var body struct {
		SKUID    string `json:"sku_id"`
		Quantity int    `json:"quantity"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	skuID, err := uuid.Parse(body.SKUID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "SKU inválido")
		return
	}
	cart, err := h.svc.UpsertCartItem(r.Context(), *user.CustomerID, skuID, body.Quantity)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, cart)
}

func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Perfil de cliente necessário")
		return
	}
	idem := r.Header.Get("Idempotency-Key")
	if idem == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Cabeçalho Idempotency-Key obrigatório")
		return
	}
	order, err := h.svc.Checkout(r.Context(), *user.CustomerID, idem, user.User.ID)
	if err != nil {
		if ae := sales.AsAppError(err); ae != nil {
			httpx.WriteError(w, ae.Status, ae.Code, ae.Message)
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, order)
}
