package mercadopago

import (
	"errors"
	"testing"
)

func TestParseMPErrorBody(t *testing.T) {
	code, msg := parseMPErrorBody([]byte(`{"message":"bad","error":"forbidden","cause":[{"code":"c1","description":"d1"}]}`))
	if code != "forbidden" || msg != "bad" {
		t.Fatalf("got %q %q", code, msg)
	}
}

func TestHTTPStatusFromCall(t *testing.T) {
	err := httpStatusFromCall(callError{
		status:      403,
		mpRequestID: "abc",
		body:        []byte(`{"error":"x","message":"y"}`),
	}, "user msg")
	var mp HTTPStatusError
	if !errors.As(err, &mp) {
		t.Fatal("expected HTTPStatusError")
	}
	if mp.Status != 403 || mp.Message != "user msg" || mp.MPRequestID != "abc" {
		t.Fatalf("%+v", mp)
	}
}
