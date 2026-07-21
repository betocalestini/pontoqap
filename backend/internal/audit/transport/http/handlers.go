package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/store-platform/store/internal/audit"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc *audit.Service
}

func NewHandler(svc *audit.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/logs", h.ListLogs)
}

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.svc.List(r.Context(), audit.ListFilter{
		Action:     r.URL.Query().Get("action"),
		EntityType: r.URL.Query().Get("entity_type"),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar auditoria")
		return
	}
	if items == nil {
		items = []audit.LogEntry{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}
