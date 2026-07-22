package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

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
	var actorID *uuid.UUID
	if s := r.URL.Query().Get("actor_user_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			actorID = &id
		}
	}
	var dateFrom, dateTo *time.Time
	if s := r.URL.Query().Get("date_from"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			dateFrom = &t
		}
	}
	if s := r.URL.Query().Get("date_to"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			end := t.AddDate(0, 0, 1)
			dateTo = &end
		}
	}
	items, total, err := h.svc.List(r.Context(), audit.ListFilter{
		Action:      r.URL.Query().Get("action"),
		EntityType:  r.URL.Query().Get("entity_type"),
		ActorUserID: actorID,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		Limit:       limit,
		Offset:      offset,
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
