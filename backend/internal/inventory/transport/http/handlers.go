package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/internal/catalog"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc *inventory.Service
	cat *catalog.Service
}

func NewHandler(svc *inventory.Service, cat *catalog.Service) *Handler {
	return &Handler{svc: svc, cat: cat}
}

func (h *Handler) ListBalances(w http.ResponseWriter, r *http.Request) {
	h.listBalances(w, r)
}

func (h *Handler) ListMovements(w http.ResponseWriter, r *http.Request) {
	h.listMovements(w, r)
}

func (h *Handler) CreateMovement(w http.ResponseWriter, r *http.Request) {
	h.createMovement(w, r)
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Get("/balances", h.listBalances)
	r.Get("/movements", h.listMovements)
	r.Post("/movements", h.createMovement)
	r.Post("/entries", h.registerEntry)
}

func (h *Handler) RegisterEntry(w http.ResponseWriter, r *http.Request) {
	h.registerEntry(w, r)
}

func (h *Handler) listBalances(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListBalances(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar saldos")
		return
	}
	if items == nil {
		items = []inventory.BalanceRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func parseMovementFilter(r *http.Request) (inventory.MovementFilter, error) {
	var f inventory.MovementFilter
	if s := r.URL.Query().Get("sku_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return f, err
		}
		f.SKUID = &id
	}
	if s := r.URL.Query().Get("product_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return f, err
		}
		f.ProductID = &id
	}
	f.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	f.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	return f, nil
}

func (h *Handler) listMovements(w http.ResponseWriter, r *http.Request) {
	filter, err := parseMovementFilter(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Filtro inválido")
		return
	}
	items, total, err := h.svc.ListMovements(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar movimentações")
		return
	}
	if items == nil {
		items = []inventory.Movement{}
	}
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": filter.Offset,
	})
}

func (h *Handler) createMovement(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	var body struct {
		Kind               string `json:"kind"`
		SKUID              string `json:"sku_id"`
		Quantity           int    `json:"quantity"`
		PhysicalCount      *int   `json:"physical_count"`
		Reason             string `json:"reason"`
		TotalPaidCents     *int64 `json:"total_paid_cents"`
		OtherExpensesCents *int64 `json:"other_expenses_cents"`
		UnitCostCents      *int64 `json:"unit_cost_cents"`
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
	kind := strings.TrimSpace(strings.ToLower(body.Kind))
	switch kind {
	case "entry":
		if !identityhttp.HasPermission(r.Context(), "inventory.entry") &&
			!identityhttp.HasPermission(r.Context(), "inventory.adjust") {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Permissão insuficiente")
			return
		}
		if body.Quantity <= 0 || strings.TrimSpace(body.Reason) == "" {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Quantidade e motivo obrigatórios")
			return
		}
		totalPaid, otherExpenses, parseErr := parseEntryPurchaseCents(body.Quantity, body.TotalPaidCents, body.OtherExpensesCents, body.UnitCostCents)
		if parseErr != nil {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", parseErr.Error())
			return
		}
		err = h.svc.RegisterEntry(r.Context(), skuID, body.Quantity, user.User.ID, body.Reason, totalPaid, otherExpenses)
	case "loss", "damage":
		if !identityhttp.HasPermission(r.Context(), "inventory.loss") &&
			!identityhttp.HasPermission(r.Context(), "inventory.adjust") {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Permissão insuficiente")
			return
		}
		if body.Quantity <= 0 || strings.TrimSpace(body.Reason) == "" {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Quantidade e motivo obrigatórios")
			return
		}
		err = h.svc.RegisterOutbound(r.Context(), skuID, body.Quantity, kind, body.Reason, user.User.ID)
	case "adjustment":
		if !identityhttp.HasPermission(r.Context(), "inventory.adjust") {
			httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Permissão insuficiente")
			return
		}
		if body.PhysicalCount == nil {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Contagem física obrigatória")
			return
		}
		err = h.svc.RegisterAdjustment(r.Context(), skuID, *body.PhysicalCount, body.Reason, user.User.ID)
	default:
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Tipo de movimentação inválido")
		return
	}
	if err != nil {
		if invErr, ok := err.(*inventory.AppError); ok {
			httpx.WriteError(w, invErr.Status, invErr.Code, invErr.Message)
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
	h.maybeRecalcSKU(r, skuID)
}

func parseEntryPurchaseCents(quantity int, totalPaid, otherExpenses, unitCostLegacy *int64) (int64, int64, error) {
	if totalPaid != nil {
		if *totalPaid < 0 {
			return 0, 0, fmt.Errorf("valor total pago inválido")
		}
		var other int64
		if otherExpenses != nil {
			if *otherExpenses < 0 {
				return 0, 0, fmt.Errorf("outros gastos inválidos")
			}
			other = *otherExpenses
		}
		return *totalPaid, other, nil
	}
	if unitCostLegacy != nil {
		if *unitCostLegacy < 0 {
			return 0, 0, fmt.Errorf("custo unitário inválido")
		}
		return *unitCostLegacy * int64(quantity), 0, nil
	}
	return 0, 0, fmt.Errorf("informe o valor total pago da entrada")
}

func (h *Handler) maybeRecalcSKU(r *http.Request, skuID uuid.UUID) {
	if h.cat == nil {
		return
	}
	user := identityhttp.UserFromContext(r.Context())
	if user == nil {
		return
	}
	_, _ = h.cat.RecalculateSKU(r.Context(), skuID, user.User.ID, "auto:entrada", h.svc.WeightedAverageCostCents)
}

func (h *Handler) registerEntry(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	var body struct {
		SKUID              string `json:"sku_id"`
		Quantity           int    `json:"quantity"`
		Reason             string `json:"reason"`
		Note               string `json:"note"`
		TotalPaidCents     *int64 `json:"total_paid_cents"`
		OtherExpensesCents *int64 `json:"other_expenses_cents"`
		UnitCostCents      *int64 `json:"unit_cost_cents"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || body.Quantity <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	reason := body.Reason
	if reason == "" {
		reason = body.Note
	}
	skuID, err := uuid.Parse(body.SKUID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "SKU inválido")
		return
	}
	totalPaid, otherExpenses, parseErr := parseEntryPurchaseCents(body.Quantity, body.TotalPaidCents, body.OtherExpensesCents, body.UnitCostCents)
	if parseErr != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", parseErr.Error())
		return
	}
	if err := h.svc.RegisterEntry(r.Context(), skuID, body.Quantity, user.User.ID, reason, totalPaid, otherExpenses); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao registrar entrada")
		return
	}
	w.WriteHeader(http.StatusNoContent)
	h.maybeRecalcSKU(r, skuID)
}
