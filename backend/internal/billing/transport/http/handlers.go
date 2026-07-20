package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc *billing.Service
}

func NewHandler(svc *billing.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) MeRoutes(r chi.Router) {
	r.Get("/invoices", h.ListMyInvoices)
	r.Get("/invoices/{id}", h.GetMyInvoice)
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Get("/invoices", h.ListAllInvoices)
	r.Post("/close", h.ClosePeriods)
	r.Get("/calendar", h.ListCalendar)
	r.Put("/calendar", h.UpsertCalendar)
}

func (h *Handler) ListMyInvoices(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Cliente necessário")
		return
	}
	items, err := h.svc.ListInvoicesByCustomer(r.Context(), *user.CustomerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar faturas")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
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
	inv, err := h.svc.GetInvoice(r.Context(), id)
	if err != nil || inv == nil || inv.CustomerID != *user.CustomerID {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Fatura não encontrada")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, inv)
}

func (h *Handler) ListAllInvoices(w http.ResponseWriter, r *http.Request) {
	// simplificado: lista via query customer_id opcional omitida — retorna vazio se não implementado full scan
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": []billing.Invoice{}})
}

func (h *Handler) ClosePeriods(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year == 0 || month == 0 {
		now := time.Now()
		year, month = billing.PreviousMonth(now.Year(), int(now.Month()))
	}
	n, err := h.svc.CloseOpenPeriodsForReference(r.Context(), year, month)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
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
