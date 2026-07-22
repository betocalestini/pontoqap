package integration_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestOpenAPIDocumentsCoreMVPPaths fails when documented paths drift from the router's public contract.
func TestOpenAPIDocumentsCoreMVPPaths(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	backendDir := filepath.Join(filepath.Dir(thisFile), "..", "..")
	openAPIPath := filepath.Join(backendDir, "openapi", "openapi.yaml")
	data, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	doc := string(data)

	required := []string{
		"/auth/login",
		"/auth/me",
		"/catalog/products",
		"/me/cart",
		"/me/cart/checkout",
		"/admin/reports/dashboard",
		"/admin/billing/invoices",
		"/webhooks/pix",
		"/webhooks/mercado-pago/orders",
	}
	pathRe := func(p string) *regexp.Regexp {
		return regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(p) + `\s*:`)
	}
	for _, p := range required {
		if !pathRe(p).MatchString(doc) && !strings.Contains(doc, p) {
			t.Errorf("openapi.yaml missing path %s", p)
		}
	}
}
