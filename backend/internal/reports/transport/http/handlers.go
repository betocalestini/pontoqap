package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/store-platform/store/internal/forecasting"
	"github.com/store-platform/store/internal/platform/httpx"
	"github.com/store-platform/store/internal/reports"
)

type ReportsHandler struct {
	svc *reports.Service
}

func NewReportsHandler(svc *reports.Service) *ReportsHandler {
	return &ReportsHandler{svc: svc}
}

func (h *ReportsHandler) Routes(r chi.Router) {
	r.Get("/dashboard", h.Dashboard)
	r.Get("/dashboard/series", h.DashboardSeries)
	r.Get("/top-products", h.TopProducts)
	r.Get("/top-customers", h.TopCustomers)
	r.Get("/inventory", h.InventoryLegacy)
	r.Get("/sales/orders", h.SalesOrders)
	r.Get("/sales/orders/export.csv", h.SalesOrdersCSV)
	r.Get("/inventory/position", h.InventoryPosition)
	r.Get("/inventory/position/export.csv", h.InventoryPositionCSV)
	r.Get("/inventory/movements", h.InventoryMovements)
	r.Get("/inventory/movements/export.csv", h.InventoryMovementsCSV)
	r.Get("/receivables/invoices", h.Receivables)
	r.Get("/receivables/invoices/export.csv", h.ReceivablesCSV)
	r.Get("/payments/pix-reconciliation", h.PixReconciliation)
	r.Get("/customers/exposure", h.CustomerExposure)
	r.Get("/exceptions", h.Exceptions)
}

func (h *ReportsHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	y, m := reports.ParseYearMonthHTTP(r)
	d, err := h.svc.Dashboard(r.Context(), y, m)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no dashboard")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, d)
}

func (h *ReportsHandler) DashboardSeries(w http.ResponseWriter, r *http.Request) {
	months, _ := strconv.Atoi(r.URL.Query().Get("months"))
	items, err := h.svc.DashboardSeries(r.Context(), months)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha na série do dashboard")
		return
	}
	if items == nil {
		items = []reports.DashboardSeriesPoint{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReportsHandler) TopProducts(w http.ResponseWriter, r *http.Request) {
	y, m := reports.ParseYearMonthHTTP(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.svc.TopProducts(r.Context(), y, m, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório")
		return
	}
	if items == nil {
		items = []reports.RankRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReportsHandler) TopCustomers(w http.ResponseWriter, r *http.Request) {
	y, m := reports.ParseYearMonthHTTP(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.svc.TopCustomers(r.Context(), y, m, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório")
		return
	}
	if items == nil {
		items = []reports.RankRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReportsHandler) InventoryLegacy(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.InventoryReport(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReportsHandler) SalesOrders(w http.ResponseWriter, r *http.Request) {
	f := salesFilterFromRequest(r)
	items, summary, total, err := h.svc.SalesOrders(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório de vendas")
		return
	}
	if items == nil {
		items = []reports.SalesOrderRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"summary": summary, "items": items, "total": total, "limit": f.Limit, "offset": f.Offset,
	})
}

func (h *ReportsHandler) SalesOrdersCSV(w http.ResponseWriter, r *http.Request) {
	f := salesFilterFromRequest(r)
	f.Limit = 10000
	f.Offset = 0
	items, _, _, err := h.svc.SalesOrders(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha na exportação")
		return
	}
	var rows [][]string
	for _, it := range items {
		rows = append(rows, []string{
			it.OrderNumber, reports.FormatTimeHTTP(it.ConfirmedAt), it.CustomerName,
			strconv.Itoa(it.ItemCount), strconv.FormatInt(it.TotalCents, 10), it.Status,
		})
	}
	reports.WriteCSV(w, "vendas-pedidos.csv",
		[]string{"pedido", "confirmado_em", "cliente", "itens", "total_centavos", "status"}, rows)
}

func salesFilterFromRequest(r *http.Request) reports.SalesOrdersFilter {
	f := reports.SalesOrdersFilter{
		DateRange:  reports.ParseDateRangeHTTP(r),
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		CustomerID: reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("customer_id")),
		ProductID:  reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("product_id")),
		CategoryID: reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("category_id")),
		MinTotal:   reports.Int64QueryHTTP(r, "min_total_cents"),
		MaxTotal:   reports.Int64QueryHTTP(r, "max_total_cents"),
		Cancelled:  reports.BoolQueryHTTP(r, "cancelled"),
		PageFilter: reports.ParsePageHTTP(r),
	}
	return f
}

func (h *ReportsHandler) InventoryPosition(w http.ResponseWriter, r *http.Request) {
	f := inventoryPositionFilter(r)
	items, total, err := h.svc.InventoryPosition(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório de estoque")
		return
	}
	if items == nil {
		items = []reports.InventoryPositionRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"items": items, "total": total, "limit": f.Limit, "offset": f.Offset,
	})
}

func (h *ReportsHandler) InventoryPositionCSV(w http.ResponseWriter, r *http.Request) {
	f := inventoryPositionFilter(r)
	f.Limit = 10000
	items, _, err := h.svc.InventoryPosition(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha na exportação")
		return
	}
	var rows [][]string
	for _, it := range items {
		rows = append(rows, []string{
			it.ProductName, it.SKUCode, it.CategoryName,
			strconv.Itoa(it.AvailableQuantity), strconv.Itoa(it.MinimumStock), it.Situation,
		})
	}
	reports.WriteCSV(w, "estoque-posicao.csv",
		[]string{"produto", "sku", "categoria", "disponivel", "minimo", "situacao"}, rows)
}

func inventoryPositionFilter(r *http.Request) reports.InventoryPositionFilter {
	below := r.URL.Query().Get("below_minimum") == "true"
	zero := r.URL.Query().Get("zero_stock") == "true"
	var activeOnly *bool
	if v := r.URL.Query().Get("active"); v != "" {
		b := v == "true"
		activeOnly = &b
	}
	return reports.InventoryPositionFilter{
		CategoryID:   reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("category_id")),
		Situation:    strings.TrimSpace(r.URL.Query().Get("situation")),
		ProductID:    reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("product_id")),
		BelowMinimum: below,
		ZeroStock:    zero,
		ActiveOnly:   activeOnly,
		PageFilter:   reports.ParsePageHTTP(r),
	}
}

func (h *ReportsHandler) InventoryMovements(w http.ResponseWriter, r *http.Request) {
	f := inventoryMovementsFilter(r)
	items, total, err := h.svc.InventoryMovements(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório de movimentações")
		return
	}
	if items == nil {
		items = []reports.InventoryMovementRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"items": items, "total": total, "limit": f.Limit, "offset": f.Offset,
	})
}

func (h *ReportsHandler) InventoryMovementsCSV(w http.ResponseWriter, r *http.Request) {
	f := inventoryMovementsFilter(r)
	f.Limit = 10000
	items, _, err := h.svc.InventoryMovements(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha na exportação")
		return
	}
	var rows [][]string
	for _, it := range items {
		rows = append(rows, []string{
			it.CreatedAt.Format(time.RFC3339), it.ProductName, it.SKUCode, it.MovementType,
			strconv.Itoa(it.Quantity), strconv.Itoa(it.PreviousBalance), strconv.Itoa(it.NewBalance),
		})
	}
	reports.WriteCSV(w, "estoque-movimentacoes.csv",
		[]string{"quando", "produto", "sku", "tipo", "qtd", "saldo_anterior", "saldo_posterior"}, rows)
}

func inventoryMovementsFilter(r *http.Request) reports.InventoryMovementsFilter {
	return reports.InventoryMovementsFilter{
		DateRange:    reports.ParseDateRangeHTTP(r),
		SKUID:        reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("sku_id")),
		ProductID:    reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("product_id")),
		MovementType: r.URL.Query().Get("movement_type"),
		UserID:       reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("user_id")),
		OrderID:      reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("order_id")),
		ManualOnly:   r.URL.Query().Get("manual_only") == "true",
		PageFilter:   reports.ParsePageHTTP(r),
	}
}

func (h *ReportsHandler) Receivables(w http.ResponseWriter, r *http.Request) {
	f := receivablesFilter(r)
	items, summary, total, err := h.svc.ReceivablesInvoices(r.Context(), f, time.Now())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório de recebíveis")
		return
	}
	if items == nil {
		items = []reports.ReceivableRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"summary": summary, "items": items, "total": total, "limit": f.Limit, "offset": f.Offset,
	})
}

func (h *ReportsHandler) ReceivablesCSV(w http.ResponseWriter, r *http.Request) {
	f := receivablesFilter(r)
	f.Limit = 10000
	items, _, _, err := h.svc.ReceivablesInvoices(r.Context(), f, time.Now())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha na exportação")
		return
	}
	var rows [][]string
	for _, it := range items {
		rows = append(rows, []string{
			it.InvoiceNumber, it.CustomerName, it.Status,
			strconv.FormatInt(it.TotalCents, 10), strconv.FormatInt(it.RemainingCents, 10), it.AgingBucket,
		})
	}
	reports.WriteCSV(w, "contas-a-receber.csv",
		[]string{"fatura", "cliente", "status", "total_centavos", "saldo_centavos", "faixa_atraso"}, rows)
}

func receivablesFilter(r *http.Request) reports.ReceivablesFilter {
	y, _ := strconv.Atoi(r.URL.Query().Get("year"))
	m, _ := strconv.Atoi(r.URL.Query().Get("month"))
	return reports.ReceivablesFilter{
		Year:         y,
		Month:        m,
		Status:       r.URL.Query().Get("status"),
		CustomerID:   reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("customer_id")),
		OverdueOnly:  r.URL.Query().Get("overdue_only") == "true",
		PartialOnly:  r.URL.Query().Get("partial_only") == "true",
		MinRemaining: reports.Int64QueryHTTP(r, "min_remaining_cents"),
		PageFilter:   reports.ParsePageHTTP(r),
	}
}

func (h *ReportsHandler) PixReconciliation(w http.ResponseWriter, r *http.Request) {
	f := reports.PixReconciliationFilter{
		DateRange:      reports.ParseDateRangeHTTP(r),
		CustomerID:     reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("customer_id")),
		InvoiceID:      reports.ParseOptionalUUIDHTTP(r.URL.Query().Get("invoice_id")),
		Status:         r.URL.Query().Get("status"),
		DivergenceOnly: r.URL.Query().Get("divergence_only") == "true",
		PageFilter:     reports.ParsePageHTTP(r),
	}
	items, total, err := h.svc.PixReconciliation(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha na conciliação Pix")
		return
	}
	if items == nil {
		items = []reports.PixReconciliationRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"items": items, "total": total, "limit": f.Limit, "offset": f.Offset,
	})
}

func (h *ReportsHandler) CustomerExposure(w http.ResponseWriter, r *http.Request) {
	minU, _ := strconv.Atoi(r.URL.Query().Get("min_utilization_percent"))
	f := reports.CustomerExposureFilter{
		Status:         r.URL.Query().Get("status"),
		MinUtilization: minU,
		OverdueOnly:    r.URL.Query().Get("overdue_only") == "true",
		LimitExhausted: r.URL.Query().Get("limit_exhausted") == "true",
		BlockedOnly:    r.URL.Query().Get("blocked_only") == "true",
		PageFilter:     reports.ParsePageHTTP(r),
	}
	items, total, err := h.svc.CustomerExposure(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório de exposição")
		return
	}
	if items == nil {
		items = []reports.CustomerExposureRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"items": items, "total": total, "limit": f.Limit, "offset": f.Offset,
	})
}

func (h *ReportsHandler) Exceptions(w http.ResponseWriter, r *http.Request) {
	dr := reports.ParseDateRangeHTTP(r)
	f := reports.ExceptionsFilter{
		DateRange:  dr,
		EventType:  r.URL.Query().Get("event_type"),
		ManualOnly: r.URL.Query().Get("manual_only") == "true",
		PageFilter: reports.ParsePageHTTP(r),
	}
	items, total, err := h.svc.Exceptions(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório de exceções")
		return
	}
	summary, _ := h.svc.ExceptionsSummary(r.Context(), dr)
	if items == nil {
		items = []reports.ExceptionRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"summary": summary, "items": items, "total": total, "limit": f.Limit, "offset": f.Offset,
	})
}

type ForecastHandler struct {
	svc *forecasting.Service
}

func NewForecastHandler(svc *forecasting.Service) *ForecastHandler {
	return &ForecastHandler{svc: svc}
}

func (h *ForecastHandler) Routes(r chi.Router) {
	r.Get("/forecast", h.List)
	r.Post("/forecast/generate", h.Generate)
}

func (h *ForecastHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.svc.ListLatest(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha na previsão")
		return
	}
	if items == nil {
		items = []forecasting.Snapshot{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ForecastHandler) Generate(w http.ResponseWriter, r *http.Request) {
	n, err := h.svc.GenerateMonthlySnapshots(r.Context(), time.Now())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"snapshots_created": n})
}
