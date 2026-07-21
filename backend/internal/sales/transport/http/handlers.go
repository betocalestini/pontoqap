package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/audit"
	"github.com/store-platform/store/internal/identity"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
	"github.com/store-platform/store/internal/sales"
)

type Handler struct {
	svc   *sales.Service
	audit *audit.Service
}

func NewHandler(svc *sales.Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

func (h *Handler) MeRoutes(r chi.Router) {
	r.Get("/cart", h.getCart)
	r.Delete("/cart", h.clearCart)
	r.Post("/cart/items", h.addItem)
	r.Patch("/cart/items/{skuId}", h.setItemQuantity)
	r.Post("/cart/checkout", h.checkout)
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Get("/", h.AdminListOrders)
	r.Get("/{id}", h.AdminGetOrder)
	r.Post("/{id}/cancel", h.AdminCancelOrder)
}

func (h *Handler) AdminListOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.svc.AdminListOrders(r.Context(), sales.AdminOrderFilter{
		Status: r.URL.Query().Get("status"),
		Search: r.URL.Query().Get("search"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar pedidos")
		return
	}
	if items == nil {
		items = []sales.AdminOrderListItem{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *Handler) AdminGetOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	order, err := h.svc.AdminGetOrder(r.Context(), id)
	if err != nil {
		writeSalesError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, order)
}

func (h *Handler) AdminCancelOrder(w http.ResponseWriter, r *http.Request) {
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
		Password string `json:"password"`
		MFACode  string `json:"mfa_code"`
	}
	_ = httpx.DecodeJSON(r, &body)
	if err := identity.VerifyStepUp(r.Context(), nil, user.User, body.Password, body.MFACode); err != nil {
		if ae := identity.AsAppError(err); ae != nil {
			httpx.WriteError(w, ae.Status, ae.Code, ae.Message)
			return
		}
		httpx.WriteError(w, http.StatusForbidden, "STEP_UP_REQUIRED", "Confirme com sua senha ou código MFA")
		return
	}
	order, err := h.svc.AdminCancelOrder(r.Context(), id, user.User.ID)
	if err != nil {
		writeSalesError(w, err)
		return
	}
	if h.audit != nil {
		_ = h.audit.Log(r.Context(), &user.User.ID, "order.cancelled", "order", &id,
			map[string]string{"status": "confirmed"}, map[string]string{"status": "cancelled"})
	}
	httpx.WriteJSON(w, http.StatusOK, order)
}
