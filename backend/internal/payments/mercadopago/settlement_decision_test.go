package mercadopago_test

import (
	"testing"

	"github.com/store-platform/store/internal/payments/mercadopago"
)

func TestEvaluateSettlementAccredited(t *testing.T) {
	order := mercadopago.OrderDetail{
		ExternalReference: "INSTALLMENT-abc",
		TotalAmountCents:  35000,
		Payments: []mercadopago.OrderPaymentDetail{{
			ID:            "PAY1",
			Status:        "processed",
			StatusDetail:  "accredited",
			AmountCents:   35000,
			PaymentMethod: "bank_transfer",
		}},
	}
	eval := mercadopago.EvaluateSettlement(order, "INSTALLMENT-abc", 35000)
	if eval.Outcome != mercadopago.SettlementSettle || eval.PaymentID != "PAY1" {
		t.Fatalf("got %+v", eval)
	}
}

func TestEvaluateSettlementPending(t *testing.T) {
	order := mercadopago.OrderDetail{
		ExternalReference: "INV-1",
		TotalAmountCents:  1000,
		Payments: []mercadopago.OrderPaymentDetail{{
			Status:        "action_required",
			StatusDetail:  "waiting_transfer",
			AmountCents:   1000,
			PaymentMethod: "pix",
		}},
	}
	eval := mercadopago.EvaluateSettlement(order, "INV-1", 1000)
	if eval.Outcome != mercadopago.SettlementPending {
		t.Fatalf("got %+v", eval)
	}
}

func TestEvaluateSettlementAmountMismatch(t *testing.T) {
	order := mercadopago.OrderDetail{
		ExternalReference: "INV-1",
		TotalAmountCents:  1000,
		Payments: []mercadopago.OrderPaymentDetail{{
			Status:        "processed",
			StatusDetail:  "accredited",
			AmountCents:   999,
			PaymentMethod: "pix",
		}},
	}
	eval := mercadopago.EvaluateSettlement(order, "INV-1", 1000)
	if eval.Outcome != mercadopago.SettlementRequiresReview {
		t.Fatalf("got %+v", eval)
	}
}
