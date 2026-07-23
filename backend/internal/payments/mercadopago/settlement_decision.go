package mercadopago

import (
	"fmt"
	"strings"
)

// SettlementOutcome classifies whether an Order can settle a charge.
type SettlementOutcome int

const (
	SettlementPending SettlementOutcome = iota
	SettlementSettle
	SettlementRequiresReview
)

type SettlementEvaluation struct {
	Outcome       SettlementOutcome
	PaymentID     string
	AmountCents   int64
	Reason        string
}

// EvaluateSettlement checks Order state against an expected charge reference and amount.
func EvaluateSettlement(order OrderDetail, expectedExternalRef string, expectedAmountCents int64) SettlementEvaluation {
	ref := strings.TrimSpace(order.ExternalReference)
	if ref != strings.TrimSpace(expectedExternalRef) {
		return SettlementEvaluation{
			Outcome: SettlementRequiresReview,
			Reason:  fmt.Sprintf("external_reference mismatch: got %q want %q", ref, expectedExternalRef),
		}
	}
	if order.TotalAmountCents != expectedAmountCents {
		return SettlementEvaluation{
			Outcome: SettlementRequiresReview,
			Reason:  fmt.Sprintf("order total mismatch: got %d want %d", order.TotalAmountCents, expectedAmountCents),
		}
	}

	for _, p := range order.Payments {
		if !isPixPayment(p.PaymentMethod) {
			continue
		}
		if p.Status == "processed" && p.StatusDetail == "accredited" {
			if p.AmountCents != expectedAmountCents {
				return SettlementEvaluation{
					Outcome: SettlementRequiresReview,
					Reason:  fmt.Sprintf("payment amount mismatch: got %d want %d", p.AmountCents, expectedAmountCents),
				}
			}
			return SettlementEvaluation{
				Outcome:     SettlementSettle,
				PaymentID:   p.ID,
				AmountCents: p.AmountCents,
			}
		}
	}
	return SettlementEvaluation{Outcome: SettlementPending}
}

func isPixPayment(method string) bool {
	m := strings.ToLower(strings.TrimSpace(method))
	return m == "pix" || m == "bank_transfer"
}

// ExpectedExternalReferenceForInstallment builds the MP external_reference for a parcela.
func ExpectedExternalReferenceForInstallment(installmentID string) string {
	return fmt.Sprintf("INSTALLMENT-%s", installmentID)
}
