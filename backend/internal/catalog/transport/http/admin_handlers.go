package http

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/catalog"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

func (h *Handler) getProductAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	p, err := h.svc.GetProduct(r.Context(), id, false)
	if err != nil || p == nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Produto não encontrado")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		CategoryID  *string `json:"category_id"`
		Active      *bool   `json:"active"`
		Visible     *bool   `json:"visible"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	in := catalog.UpdateProductInput{
		Name: body.Name, Description: body.Description, Active: body.Active, Visible: body.Visible,
	}
	if body.CategoryID != nil {
		if strings.TrimSpace(*body.CategoryID) == "" {
			in.ClearCategory = true
		} else {
			cid, err := uuid.Parse(*body.CategoryID)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Categoria inválida")
				return
			}
			in.CategoryID = &cid
		}
	}
	p, err := h.svc.UpdateProduct(r.Context(), id, in)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao atualizar produto")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) updateSKU(w http.ResponseWriter, r *http.Request) {
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
		Code           *string `json:"code"`
		Barcode        *string `json:"barcode"`
		Unit           *string `json:"unit"`
		SalePriceCents *int64  `json:"sale_price_cents"`
		CostPriceCents *int64  `json:"cost_price_cents"`
		MinimumStock   *int    `json:"minimum_stock"`
		Active         *bool   `json:"active"`
		PriceReason    string  `json:"price_reason"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dados inválidos")
		return
	}
	if err := h.svc.UpdateSKU(r.Context(), skuID, catalog.UpdateSKUInput{
		Code: body.Code, Barcode: body.Barcode, Unit: body.Unit,
		SalePriceCents: body.SalePriceCents, CostPriceCents: body.CostPriceCents,
		MinimumStock: body.MinimumStock, Active: body.Active,
	}, user.User.ID, body.PriceReason); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao atualizar SKU")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	var body struct {
		Name         string  `json:"name"`
		Slug         string  `json:"slug"`
		Description  string  `json:"description"`
		CategoryID   *string `json:"category_id"`
		SKUCode      string  `json:"sku_code"`
		Barcode      string  `json:"barcode"`
		SalePrice    int64   `json:"sale_price_cents"`
		CostPrice    *int64  `json:"cost_price_cents"`
		MinimumStock int     `json:"minimum_stock"`
		Unit         string  `json:"unit"`
		InitialStock int     `json:"initial_stock"`
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
		Name: body.Name, Slug: body.Slug, Description: body.Description, CategoryID: catID,
		SKUCode: body.SKUCode, Barcode: body.Barcode, SalePrice: body.SalePrice,
		CostPrice: body.CostPrice, MinimumStock: body.MinimumStock, Unit: body.Unit,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao criar produto")
		return
	}
	if body.InitialStock > 0 && h.inv != nil && user != nil && len(p.SKUs) > 0 {
		if err := h.inv.RegisterInitialStock(r.Context(), p.SKUs[0].ID, body.InitialStock, user.User.ID); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		p, _ = h.svc.GetProduct(r.Context(), p.ID, false)
	}
	httpx.WriteJSON(w, http.StatusCreated, p)
}

func (h *Handler) uploadProductImage(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	p, err := h.svc.GetProduct(r.Context(), productID, false)
	if err != nil || p == nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Produto não encontrado")
		return
	}
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Upload inválido")
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Arquivo image obrigatório")
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".svg": true}
	if !allowed[ext] {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Formato não suportado")
		return
	}
	dir := filepath.Join(h.uploadDir, "product-images")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao salvar imagem")
		return
	}
	filename := p.Slug + ext
	dest := filepath.Join(dir, filename)
	for _, other := range catalog.ProductImageExtensions() {
		if other == ext {
			continue
		}
		_ = os.Remove(filepath.Join(dir, p.Slug+other))
	}
	out, err := os.Create(dest)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao salvar imagem")
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao salvar imagem")
		return
	}
	out.Close()
	info, _ := os.Stat(dest)
	size := int64(0)
	if info != nil {
		size = info.Size()
	}
	slog.Info("product image uploaded",
		"product_id", productID.String(),
		"slug", p.Slug,
		"file", filename,
		"bytes", size,
	)
	storageKey := "/product-images/" + p.Slug
	if err := h.svc.UpsertProductImage(r.Context(), productID, storageKey, p.Name); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao registrar imagem")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"image_url": catalog.ResolveImageURL(storageKey),
	})
}

func (h *Handler) deleteProductImage(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	imageID, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Imagem inválida")
		return
	}
	if err := h.svc.DeleteProductImage(r.Context(), productID, imageID); err != nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Imagem não encontrada")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
