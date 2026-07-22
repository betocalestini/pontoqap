package billing

// InstallmentPolicyParams holds values used to calculate installment options.
type InstallmentPolicyParams struct {
	MinimumInvoiceAmountCents      int64
	MinimumInstallmentAmountCents  int64
	MaximumInstallments            int
	InstallmentEnabled             bool
}

// MaxInstallments returns allowed installment counts (1..N). liveEnabled is the active policy kill-switch.
func MaxInstallments(totalCents int64, p InstallmentPolicyParams, liveInstallmentEnabled bool) int {
	if totalCents <= 0 {
		return 1
	}
	if !liveInstallmentEnabled {
		return 1
	}
	if totalCents < p.MinimumInvoiceAmountCents {
		return 1
	}
	byMin := int(totalCents / p.MinimumInstallmentAmountCents)
	if byMin < 1 {
		byMin = 1
	}
	if byMin > p.MaximumInstallments {
		return p.MaximumInstallments
	}
	return byMin
}

// DistributeInstallmentAmounts splits totalCents into count parts; extra cents go to the last parcels.
func DistributeInstallmentAmounts(totalCents int64, count int) []int64 {
	if count < 1 {
		return nil
	}
	base := totalCents / int64(count)
	rest := int(totalCents % int64(count))
	out := make([]int64, count)
	for i := range out {
		out[i] = base
	}
	for i := count - rest; i < count; i++ {
		out[i]++
	}
	return out
}

// InstallmentEligible is true when multi-installment is offered (max > 1).
func InstallmentEligible(max int, liveInstallmentEnabled bool) bool {
	return liveInstallmentEnabled && max > 1
}
