package billing

import "errors"

var (
	ErrPeriodNotOpen    = errors.New("billing period is not open")
	ErrPeriodNotFound   = errors.New("billing period not found")
	ErrNoBusinessDay    = errors.New("not enough business days in month")
	ErrInvoiceExists    = errors.New("invoice already exists for period")
)
