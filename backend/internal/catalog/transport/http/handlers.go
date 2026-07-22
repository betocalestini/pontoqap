package http

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/inventory"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc       *catalog.Service
	inv       *inventory.Service
	uploadDir string
}

func NewHandler(svc *catalog.Service, inv *inventory.Service, uploadDir string) *Handler {
	return &Handler{svc: svc, inv: inv, uploadDir: uploadDir}
}

func (h *Handler) PublicRoutes(r chi.Router) {
	r.Get("/categories", h.listCategories)
	r.Get("/products", h.listProducts)
	r.Get("/products/{id}", h.getProduct)
	r.Get("/product-images/{name}", h.serveProductImage)
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Get("/categories", h.ListCategoriesAdmin)
	r.Post("/categories", h.CreateCategory)
	r.Patch("/categories/{id}", h.UpdateCategory)
	r.Delete("/categories/{id}", h.DeleteCategory)
	r.Get("/products", h.ListProductsAdmin)
	r.Get("/products/{id}", h.GetProductAdmin)
	r.Post("/products", h.CreateProduct)
	r.Patch("/products/{id}", h.UpdateProduct)
	r.Post("/products/{id}/images", h.UploadProductImage)
	r.Delete("/products/{id}/images/{imageId}", h.DeleteProductImage)
	r.Patch("/skus/{skuId}", h.UpdateSKU)
	r.Patch("/skus/{skuId}/price", h.ChangePrice)
	r.Get("/settings/pricing", h.getPricingSettings)
	r.Patch("/settings/pricing", h.patchPricingSettings)
	r.Post("/products/reprice-all", h.repriceAllProducts)
}

func (h *Handler) ListCategoriesAdmin(w http.ResponseWriter, r *http.Request) {
	h.listCategoriesAdmin(w, r)
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	h.createCategory(w, r)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	h.updateCategory(w, r)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	h.deleteCategory(w, r)
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

func (h *Handler) GetProductAdmin(w http.ResponseWriter, r *http.Request) {
	h.getProductAdmin(w, r)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	h.updateProduct(w, r)
}

func (h *Handler) UpdateSKU(w http.ResponseWriter, r *http.Request) {
	h.updateSKU(w, r)
}

func (h *Handler) UploadProductImage(w http.ResponseWriter, r *http.Request) {
	h.uploadProductImage(w, r)
}

func (h *Handler) DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	h.deleteProductImage(w, r)
}

func (h *Handler) GetPricingSettings(w http.ResponseWriter, r *http.Request) {
	h.getPricingSettings(w, r)
}

func (h *Handler) PatchPricingSettings(w http.ResponseWriter, r *http.Request) {
	h.patchPricingSettings(w, r)
}

func (h *Handler) RepriceAllProducts(w http.ResponseWriter, r *http.Request) {
	h.repriceAllProducts(w, r)
}

func (h *Handler) serveProductImage(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Arquivo inválido")
		return
	}
	data, filename, err := catalog.OpenProductImageBySlug(name)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Imagem não encontrada")
		return
	}
	sum := sha256.Sum256(data)
	etag := `"` + hex.EncodeToString(sum[:12]) + `"`
	if inm := r.Header.Get("If-None-Match"); inm == etag {
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "public, max-age=86400, must-revalidate")
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", catalog.ProductImageContentType(filename))
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=86400, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
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
	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	if admin {
		sort = ""
	}
	var customerID *uuid.UUID
	if !admin && sort == "purchases" {
		user := identityhttp.UserFromContext(r.Context())
		if user == nil || user.CustomerID == nil {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ordenação por compras exige login na loja")
			return
		}
		customerID = user.CustomerID
	}
	var activeFilter, visibleFilter *bool
	if admin {
		var ok bool
		activeFilter, ok = parseOptionalBoolQuery(r.URL.Query().Get("active"))
		if !ok {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Parâmetro active inválido")
			return
		}
		visibleFilter, ok = parseOptionalBoolQuery(r.URL.Query().Get("visible"))
		if !ok {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Parâmetro visible inválido")
			return
		}
	}
	items, total, err := h.svc.ListProducts(r.Context(), catalog.ListProductsFilter{
		Search:     r.URL.Query().Get("search"),
		Category:   r.URL.Query().Get("category"),
		Sort:       sort,
		CustomerID: customerID,
		Page:       page,
		PageSize:   pageSize,
		Admin:      admin,
		Active:     activeFilter,
		Visible:    visibleFilter,
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

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		Name   *string `json:"name"`
		Active *bool   `json:"active"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	if body.Name == nil && body.Active == nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Nada para atualizar")
		return
	}
	c, err := h.svc.UpdateCategory(r.Context(), id, catalog.UpdateCategoryInput{Name: body.Name, Active: body.Active})
	if err != nil {
		if strings.Contains(err.Error(), "não encontrada") {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		if strings.Contains(err.Error(), "obrigatório") {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao atualizar categoria")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	res, err := h.svc.DeleteCategory(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "não encontrada") {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao excluir categoria")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
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

func parseOptionalBoolQuery(v string) (*bool, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, true
	}
	switch v {
	case "true", "1":
		b := true
		return &b, true
	case "false", "0":
		b := false
		return &b, true
	default:
		return nil, false
	}
}
