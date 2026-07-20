package catalog

import (
	"strings"
	"testing"
)

func TestResolveImageURL(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"/product-images/arroz-5kg.svg", "/api/v1/catalog/product-images/arroz-5kg"},
		{"/product-images/arroz-5kg.png", "/api/v1/catalog/product-images/arroz-5kg"},
		{"arroz-5kg", "/api/v1/catalog/product-images/arroz-5kg"},
		{"/product-images/arroz-5kg", "/api/v1/catalog/product-images/arroz-5kg"},
	}
	for _, tc := range tests {
		got := ResolveImageURL(tc.key)
		if got != tc.want {
			t.Fatalf("ResolveImageURL(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestImageURLForSlug(t *testing.T) {
	got := ImageURLForSlug("arroz-5kg")
	want := "/api/v1/catalog/product-images/arroz-5kg"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if ImageURLForSlug("unknown-product-xyz") != "" {
		t.Fatal("unknown slug should be empty")
	}
}

func TestOpenProductImageBySlugAnyFormat(t *testing.T) {
	data, filename, err := OpenProductImageBySlug("arroz-5kg")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("empty image data")
	}
	if !strings.HasPrefix(filename, "arroz-5kg.") {
		t.Fatalf("unexpected file %q", filename)
	}
	_, _, err = OpenProductImageBySlug("arroz-5kg.svg")
	if err != nil {
		t.Fatal("slug with extension in param should still resolve")
	}
}
