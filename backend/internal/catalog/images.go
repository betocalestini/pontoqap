package catalog

import (
	"embed"
	"errors"
	"io/fs"
	"path"
	"strings"
)

//go:embed static/product-images/*
var productImageFS embed.FS

var productImagesSub fs.FS

// Ordem de preferência ao resolver imagem pelo nome (slug), sem extensão na URL.
var productImageExtensions = []string{".webp", ".png", ".jpg", ".jpeg", ".svg"}

var errProductImageNotFound = errors.New("product image not found")

func init() {
	sub, err := fs.Sub(productImageFS, "static/product-images")
	if err != nil {
		panic("catalog product images: " + err.Error())
	}
	productImagesSub = sub
}

// ProductImageAPIPathForSlug retorna a URL pública sem extensão (formato resolvido no servidor).
func ProductImageAPIPathForSlug(slug string) string {
	slug = normalizeImageSlug(slug)
	if slug == "" {
		return ""
	}
	return "/api/v1/catalog/product-images/" + slug
}

// ResolveImageURL turns a storage_key into a URL the store can load (via API proxy).
func ResolveImageURL(storageKey string) string {
	storageKey = strings.TrimSpace(storageKey)
	if storageKey == "" {
		return ""
	}
	if strings.HasPrefix(storageKey, "http://") || strings.HasPrefix(storageKey, "https://") {
		return storageKey
	}
	if strings.HasPrefix(storageKey, "/api/v1/catalog/product-images/") {
		slug := slugFromAPIPath(storageKey)
		if slug != "" {
			return ProductImageAPIPathForSlug(slug)
		}
		return storageKey
	}

	slug := SlugFromStorageKey(storageKey)
	if slug == "" {
		if strings.HasPrefix(storageKey, "/") {
			return storageKey
		}
		return ""
	}
	if HasProductImage(slug) {
		return ProductImageAPIPathForSlug(slug)
	}
	// Chave aponta para nome conhecido no catálogo embutido / convenção product-images
	if strings.HasPrefix(storageKey, "/product-images/") || !strings.Contains(storageKey, "/") {
		return ProductImageAPIPathForSlug(slug)
	}
	if strings.HasPrefix(storageKey, "/") {
		return storageKey
	}
	return ProductImageAPIPathForSlug(slug)
}

// SlugFromStorageKey extrai o identificador do produto (sem extensão) de uma storage_key.
func SlugFromStorageKey(storageKey string) string {
	return normalizeImageSlug(storageKey)
}

func slugFromAPIPath(apiPath string) string {
	const prefix = "/api/v1/catalog/product-images/"
	if !strings.HasPrefix(apiPath, prefix) {
		return ""
	}
	return normalizeImageSlug(strings.TrimPrefix(apiPath, prefix))
}

func normalizeImageSlug(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "/product-images/")
	raw = strings.TrimPrefix(raw, "/")
	if raw == "" {
		return ""
	}
	base := path.Base(raw)
	if ext := path.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return base
}

// ImageURLForSlug returns an API image URL when a bundled asset exists for the slug (any format).
func ImageURLForSlug(slug string) string {
	slug = normalizeImageSlug(slug)
	if slug == "" || !HasProductImage(slug) {
		return ""
	}
	return ProductImageAPIPathForSlug(slug)
}

// HasProductImage reports whether any file exists for the slug in the embedded catalog.
func HasProductImage(slug string) bool {
	_, _, err := OpenProductImageBySlug(slug)
	return err == nil
}

// OpenProductImageBySlug loads the first matching file for slug across supported extensions.
func OpenProductImageBySlug(slug string) ([]byte, string, error) {
	slug = normalizeImageSlug(slug)
	if slug == "" {
		return nil, "", errProductImageNotFound
	}
	for _, ext := range productImageExtensions {
		filename := slug + ext
		data, err := fs.ReadFile(productImagesSub, filename)
		if err == nil {
			return data, filename, nil
		}
	}
	return nil, "", errProductImageNotFound
}

// ProductImageContentType returns a MIME type for known product image extensions.
func ProductImageContentType(filename string) string {
	switch strings.ToLower(path.Ext(filename)) {
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
