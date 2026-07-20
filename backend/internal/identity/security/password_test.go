package security_test

import (
	"testing"

	"github.com/store-platform/store/internal/identity/security"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := security.HashPassword("ChangeMe123!")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := security.VerifyPassword(hash, "ChangeMe123!")
	if err != nil || !ok {
		t.Fatalf("expected valid password, ok=%v err=%v", ok, err)
	}
	ok, err = security.VerifyPassword(hash, "wrong")
	if err != nil || ok {
		t.Fatalf("expected invalid password")
	}
}
