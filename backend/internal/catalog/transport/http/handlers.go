package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/catalog"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc *catalog.Service
}

func NewHandler(svc *catalog.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) PublicRoutes(r chi.Router) {
	r.Get("/categories", h.listCategories)
	r.Get("/products", h.listProducts)
	r.Get("/products/{id}", h.getProduct)
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Get("/categories", h.ListCategoriesAdmin)
	r.Post("/categories", h.CreateCategory)
	r.Get("/products", h.ListProductsAdmin)
	r.Post("/products", h.CreateProduct)
	r.Patch("/skus/{skuId}/price", h.ChangePrice)
}

func (h *Handler) ListCategoriesAdmin(w http.ResponseWriter, r *http.Request) {
	h.listCategoriesAdmin(w, r)
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	h.createCategory(w, r)
}

func (h *Handler) ListProductsAdmin(w http.ResponseWriter, r *http.Request) {
	h.listProductsAdmin(w, r)
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	h.createProduct(w, r)
}

func (h *Handler) ChangePrice(w http.ResponseWriter, r *http.Request) {
	h.changePrice(w, r)
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListCategories(r.Context(), false)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar categorias")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) listCategoriesAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListCategories(r.Context(), true)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar categorias")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	h.writeProductList(w, r, false)
}

func (h *Handler) listProductsAdmin(w http.ResponseWriter, r *http.Request) {
	h.writeProductList(w, r, true)
}

func (h *Handler) writeProductList(w http.ResponseWriter, r *http.Request, admin bool) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	items, total, err := h.svc.ListProducts(r.Context(), catalog.ListProductsFilter{
		Search:   r.URL.Query().Get("search"),
		Category: r.URL.Query().Get("category"),
		Page:     page,
		PageSize: pageSize,
		Admin:    admin,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao listar produtos")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	p, err := h.svc.GetProduct(r.Context(), id, true)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao buscar produto")
		return
	}
	if p == nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Produto não encontrado")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || body.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Nome obrigatório")
		return
	}
	c, err := h.svc.CreateCategory(r.Context(), catalog.CreateCategoryInput{Name: body.Name, Slug: body.Slug})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao criar categoria")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, c)
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string  `json:"name"`
		Slug        string  `json:"slug"`
		Description string  `json:"description"`
		CategoryID  *string `json:"category_id"`
		SKUCode     string  `json:"sku_code"`
		SalePrice   int64   `json:"sale_price_cents"`
		Unit        string  `json:"unit"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || body.Name == "" || body.SKUCode == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	var catID *uuid.UUID
	if body.CategoryID != nil {
		id, err := uuid.Parse(*body.CategoryID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Categoria inválida")
			return
		}
		catID = &id
	}
	p, err := h.svc.CreateProduct(r.Context(), catalog.CreateProductInput{
		Name:        body.Name,
		Slug:        body.Slug,
		Description: body.Description,
		CategoryID:  catID,
		SKUCode:     body.SKUCode,
		SalePrice:   body.SalePrice,
		Unit:        body.Unit,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao criar produto")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, p)
}

func (h *Handler) changePrice(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	skuID, err := uuid.Parse(chi.URLParam(r, "skuId"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "SKU inválido")
		return
	}
	var body struct {
		SalePriceCents int64  `json:"sale_price_cents"`
		Reason         string `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	if err := h.svc.ChangeSKUPrice(r.Context(), skuID, body.SalePriceCents, user.User.ID, body.Reason); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao alterar preço")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
