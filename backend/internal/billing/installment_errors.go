package billing

import "errors"

var (
	ErrInstallmentsDisabled     = errors.New("installment payments are disabled")
	ErrInvalidInstallmentCount  = errors.New("invalid installment count")
	ErrPaymentPlanNotPending    = errors.New("payment plan is not pending selection")
	ErrPaymentPlanImmutable     = errors.New("payment plan cannot be changed")
	ErrInstallmentNotPayable    = errors.New("installment is not available for payment")
	ErrInstallmentNotSequential = errors.New("previous installment must be paid first")
	ErrInstallmentAfterDue      = errors.New("installment selection after due date is not allowed")
	ErrNoActiveInstallmentPolicy = errors.New("no active installment policy")
)
