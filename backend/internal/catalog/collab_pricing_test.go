package catalog_test

import (
	"testing"

	"github.com/store-platform/store/internal/catalog"
)

func TestApplyCollaboratorMin(t *testing.T) {
	promo, reg := catalog.ApplyCollaboratorMinForTest(1000, 1500, 1200, true)
	if promo != 1000 || reg != 1200 {
		t.Fatalf("promo cheaper: got %d %d", promo, reg)
	}
	promo, reg = catalog.ApplyCollaboratorMinForTest(1000, 1500, 800, true)
	if promo != 800 || reg != 800 {
		t.Fatalf("collab cheapest: got %d %d", promo, reg)
	}
}
