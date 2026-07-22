package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/audit"
	"github.com/store-platform/store/internal/billing"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc   *billing.Service
	audit *audit.Service
	log   *slog.Logger
}

func NewHandler(svc *billing.Service, auditSvc *audit.Service, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{svc: svc, audit: auditSvc, log: log}
}

func (h *Handler) MeRoutes(r chi.Router) {
	r.Get("/invoices", h.ListMyInvoices)
	r.Get("/invoices/{id}", h.GetMyInvoice)
	r.Get("/invoices/{id}/payment-options", h.GetMyPaymentOptions)
	r.Post("/invoices/{id}/payment-plan", h.SelectMyPaymentPlan)
	r.Get("/invoices/{id}/payment-plan", h.GetMyPaymentPlan)
	r.Get("/invoices/{id}/installments", h.ListMyInstallments)
	r.Get("/billing/open-period", h.GetMyOpenBillingPeriod)
	r.Post("/billing/close-cycle", h.CloseMyBillingCycle)
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Get("/summary", h.AdminSummary)
	r.Get("/invoices", h.ListAllInvoices)
	r.Get("/invoices/{id}", h.GetAdminInvoice)
	r.Post("/invoices/{id}/adjustments", h.AddInvoiceAdjustment)
	r.Post("/invoices/{id}/payment-plan/reset", h.ResetPaymentPlan)
	r.Post("/close", h.ClosePeriods)
	r.Get("/calendar", h.ListCalendar)
	r.Put("/calendar", h.UpsertCalendar)
	r.Get("/installment-policy", h.GetInstallmentPolicy)
	r.Put("/installment-policy", h.PutInstallmentPolicy)
	r.Get("/installment-policies", h.ListInstallmentPolicies)
}

func (h *Handler) ListMyInvoices(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Cliente necessário")
		return
	}
	items, err := h.svc.ListInvoicesByCustomerLimit(r.Context(), *user.CustomerID, 20)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar faturas")
		return
	}
	current, err := h.svc.GetOpenPeriodSummary(r.Context(), *user.CustomerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao carregar competência atual")
		return
	}
	if items == nil {
		items = []billing.Invoice{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"current_period": current, "items": items})
}

func (h *Handler) GetMyOpenBillingPeriod(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Cliente necessário")
		return
	}
	detail, err := h.svc.GetOpenPeriodDetail(r.Context(), *user.CustomerID)
	if err != nil {
		if err == billing.ErrNoOpenPeriod {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Não há período em aberto")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao carregar competência")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, detail)
}

func (h *Handler) GetMyInvoice(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Cliente necessário")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	inv, err := h.svc.GetInvoiceDetail(r.Context(), id)
	if err != nil || inv == nil || inv.CustomerID != *user.CustomerID {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Fatura não encontrada")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, inv)
}

func (h *Handler) CloseMyBillingCycle(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Cliente necessário")
		return
	}
	inv, err := h.svc.CloseCustomerOpenPeriod(r.Context(), *user.CustomerID)
	if err != nil {
		switch err {
		case billing.ErrNoOpenPeriod:
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Não há período em aberto")
		case billing.ErrPeriodEmpty:
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Não há compras para fechar neste ciclo")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, inv)
}

func (h *Handler) AdminSummary(w http.ResponseWriter, r *http.Request) {
	sum, err := h.svc.AdminBillingSummary(r.Context(), time.Now())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao carregar resumo")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sum)
}

func parseAdminInvoiceFilter(r *http.Request) billing.AdminInvoiceFilter {
	var f billing.AdminInvoiceFilter
	f.Status = r.URL.Query().Get("status")
	f.Search = r.URL.Query().Get("search")
	f.Year, _ = strconv.Atoi(r.URL.Query().Get("year"))
	f.Month, _ = strconv.Atoi(r.URL.Query().Get("month"))
	f.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	f.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	if cid := r.URL.Query().Get("customer_id"); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			f.CustomerID = &id
		}
	}
	return f
}

func (h *Handler) ListAllInvoices(w http.ResponseWriter, r *http.Request) {
	f := parseAdminInvoiceFilter(r)
	items, total, err := h.svc.ListInvoicesAdmin(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar faturas")
		return
	}
	if items == nil {
		items = []billing.AdminInvoiceListRow{}
	}
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": f.Offset,
	})
}

func (h *Handler) GetAdminInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	inv, err := h.svc.GetInvoiceDetail(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao carregar fatura")
		return
	}
	if inv == nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Fatura não encontrada")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, inv)
}

func (h *Handler) AddInvoiceAdjustment(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	if !identityhttp.HasPermission(r.Context(), "billing.close") {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Permissão insuficiente")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		AdjustmentType string `json:"adjustment_type"`
		AmountCents    int64  `json:"amount_cents"`
		Reason         string `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	inv, err := h.svc.AddInvoiceAdjustment(r.Context(), id, user.User.ID, body.AdjustmentType, body.AmountCents, body.Reason)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if h.audit != nil {
		_ = h.audit.Log(r.Context(), &user.User.ID, "billing.invoice_adjustment", "invoice", &id, nil, map[string]any{
			"adjustment_type": body.AdjustmentType,
			"amount_cents":    body.AmountCents,
			"reason":          body.Reason,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, inv)
}

func (h *Handler) ClosePeriods(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	reason := strings.TrimSpace(r.URL.Query().Get("reason"))
	var body struct {
		Year   int    `json:"year"`
		Month  int    `json:"month"`
		Reason string `json:"reason"`
	}
	_ = httpx.DecodeJSON(r, &body)
	if body.Year > 0 {
		year = body.Year
	}
	if body.Month > 0 {
		month = body.Month
	}
	if body.Reason != "" {
		reason = strings.TrimSpace(body.Reason)
	}
	if year == 0 || month == 0 {
		now := time.Now()
		year, month = billing.PreviousMonth(now.Year(), int(now.Month()))
	}
	if len(reason) < 10 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Informe o motivo do fechamento (mínimo 10 caracteres)")
		return
	}
	n, err := h.svc.CloseOpenPeriodsForReference(r.Context(), year, month, billing.CloseTypeAdminManual)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if h.audit != nil && user != nil {
		_ = h.audit.Log(r.Context(), &user.User.ID, "billing.close_manual", "billing_period", nil, nil, map[string]any{
			"year":           year,
			"month":          month,
			"closed_periods": n,
			"reason":         reason,
		})
	}
	h.log.Info("billing manual close",
		slog.String("request_id", httpx.RequestIDFromContext(r.Context())),
		slog.Int("year", year),
		slog.Int("month", month),
		slog.Int("closed_periods", n),
	)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"closed_periods": n, "year": year, "month": month})
}

func (h *Handler) ListCalendar(w http.ResponseWriter, r *http.Request) {
	from, _ := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	to, _ := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if from.IsZero() {
		from = time.Now().AddDate(0, -1, 0)
	}
	if to.IsZero() {
		to = time.Now().AddDate(0, 2, 0)
	}
	items, err := h.svc.ListCalendar(r.Context(), from, to)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar calendário")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) UpsertCalendar(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	var body struct {
		Date          string `json:"date"`
		Name          string `json:"name"`
		Scope         string `json:"scope"`
		IsBusinessDay bool   `json:"is_business_day"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	d, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Data inválida")
		return
	}
	if err := h.svc.UpsertCalendarDay(r.Context(), billing.CalendarEntry{
		Date: d, Name: body.Name, Scope: body.Scope, IsBusinessDay: body.IsBusinessDay,
	}, user.User.ID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao salvar")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
