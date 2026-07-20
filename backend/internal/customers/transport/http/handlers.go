package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/customers"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc *customers.Service
}

func NewHandler(svc *customers.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) StoreRoutes(r chi.Router) {
	r.Post("/register", h.register)
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Get("/", h.list)
	r.Patch("/{id}/approve", h.approve)
	r.Patch("/{id}/credit-limit", h.changeLimit)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
		Document string `json:"document"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || body.Email == "" || body.Password == "" || body.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	c, err := h.svc.Register(r.Context(), customers.RegisterInput{
		Name: body.Name, Email: body.Email, Password: body.Password, Phone: body.Phone, Document: body.Document,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusConflict, "CONFLICT", "Não foi possível registrar o cliente")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, c)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	h.list(w, r)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	h.approve(w, r)
}

func (h *Handler) ChangeLimit(w http.ResponseWriter, r *http.Request) {
	h.changeLimit(w, r)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar clientes")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) approve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		CreditLimitCents int64 `json:"credit_limit_cents"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	user := identityhttp.UserFromContext(r.Context())
	if err := h.svc.Approve(r.Context(), id, user.User.ID, body.CreditLimitCents); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Não foi possível aprovar")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) changeLimit(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		CreditLimitCents int64  `json:"credit_limit_cents"`
		Reason           string `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	user := identityhttp.UserFromContext(r.Context())
	if err := h.svc.ChangeLimit(r.Context(), id, user.User.ID, body.CreditLimitCents, body.Reason); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao alterar limite")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
