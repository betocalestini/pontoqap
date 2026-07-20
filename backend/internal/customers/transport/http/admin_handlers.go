package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/customers"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

func (h *Handler) getCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil || c == nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Cliente não encontrado")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

func (h *Handler) updateCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		Name                   *string `json:"name"`
		Phone                  *string `json:"phone"`
		Document               *string `json:"document"`
		CollaboratorCategoryID *string `json:"collaborator_category_id"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	in := customers.UpdateCustomerInput{
		Name: body.Name, Phone: body.Phone, Document: body.Document,
	}
	if body.CollaboratorCategoryID != nil {
		if *body.CollaboratorCategoryID == "" {
			in.ClearCollaborator = true
		} else {
			cid, err := uuid.Parse(*body.CollaboratorCategoryID)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Categoria inválida")
				return
			}
			in.CollaboratorCategoryID = &cid
		}
	}
	c, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		if ae := customers.AsAppError(err); ae != nil {
			httpx.WriteError(w, ae.Status, ae.Code, ae.Message)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao atualizar cliente")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

func (h *Handler) blockCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = httpx.DecodeJSON(r, &body)
	if err := h.svc.Block(r.Context(), id, body.Reason); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Não foi possível bloquear")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) unblockCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	if err := h.svc.Unblock(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Não foi possível desbloquear")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listCollaboratorCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListCollaboratorCategories(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar categorias")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) createCollaboratorCategory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name          string  `json:"name"`
		Slug          string  `json:"slug"`
		MarginPercent float64 `json:"margin_percent"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || body.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	c, err := h.svc.CreateCollaboratorCategory(r.Context(), customers.CreateCollaboratorCategoryInput{
		Name: body.Name, Slug: body.Slug, MarginPercent: body.MarginPercent,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, c)
}

func (h *Handler) updateCollaboratorCategory(w http.ResponseWriter, r *http.Request) {
	_ = identityhttp.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		Name          *string  `json:"name"`
		MarginPercent *float64 `json:"margin_percent"`
		Active        *bool    `json:"active"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	c, err := h.svc.UpdateCollaboratorCategory(r.Context(), id, customers.UpdateCollaboratorCategoryInput{
		Name: body.Name, MarginPercent: body.MarginPercent, Active: body.Active,
	})
	if err != nil {
		if ae := customers.AsAppError(err); ae != nil {
			httpx.WriteError(w, ae.Status, ae.Code, ae.Message)
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if c == nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Categoria não encontrada")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}
