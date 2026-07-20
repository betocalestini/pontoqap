package http

import (
	"net/http"
	"strconv"
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
	r.Get("/top-products", h.TopProducts)
	r.Get("/inventory", h.Inventory)
}

func (h *ReportsHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	y, m := yearMonth(r)
	d, err := h.svc.Dashboard(r.Context(), y, m)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no dashboard")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, d)
}

func (h *ReportsHandler) TopProducts(w http.ResponseWriter, r *http.Request) {
	y, m := yearMonth(r)
	items, err := h.svc.TopProducts(r.Context(), y, m, 10)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ReportsHandler) Inventory(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.InventoryReport(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha no relatório")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func yearMonth(r *http.Request) (int, int) {
	y, _ := strconv.Atoi(r.URL.Query().Get("year"))
	m, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if y == 0 {
		now := time.Now()
		y, m = now.Year(), int(now.Month())
	}
	if m == 0 {
		m = int(time.Now().Month())
	}
	return y, m
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
	items, err := h.svc.ListLatest(r.Context(), 50)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha na previsão")
		return
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
