package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	PlanStatusPendingSelection = "pending_selection"
	PlanStatusActive           = "active"
	PlanStatusCompleted        = "completed"
	PlanStatusCanceled         = "canceled"

	InstStatusScheduled = "scheduled"
	InstStatusOpen      = "open"
	InstStatusPixActive = "pix_active"
	InstStatusPaid      = "paid"
	InstStatusOverdue   = "overdue"
	InstStatusCanceled  = "canceled"
)

type PaymentPlan struct {
	ID                         uuid.UUID  `json:"id"`
	InvoiceID                  uuid.UUID  `json:"invoice_id"`
	PolicyID                   uuid.UUID  `json:"policy_id"`
	Status                     string     `json:"status"`
	SelectedInstallmentCount   *int       `json:"selected_installment_count,omitempty"`
	InvoiceTotalCents          int64      `json:"invoice_total_cents"`
	PaidCents                  int64      `json:"paid_cents"`
	RemainingCents             int64      `json:"remaining_cents"`
	InstallmentEnabledSnapshot bool       `json:"installment_enabled_snapshot"`
	SelectedAt                 *time.Time `json:"selected_at,omitempty"`
}

type InvoiceInstallment struct {
	ID                uuid.UUID `json:"id"`
	PaymentPlanID     uuid.UUID `json:"payment_plan_id"`
	InvoiceID         uuid.UUID `json:"invoice_id"`
	InstallmentNumber int       `json:"installment_number"`
	AmountCents       int64     `json:"amount_cents"`
	PaidCents         int64     `json:"paid_cents"`
	RemainingCents    int64     `json:"remaining_cents"`
	DueDate           time.Time `json:"due_date"`
	Status            string    `json:"status"`
}

type PaymentOptionInstallment struct {
	Number      int       `json:"number"`
	AmountCents int64     `json:"amount_cents"`
	DueDate     time.Time `json:"due_date"`
}

type PaymentOption struct {
	InstallmentCount int                        `json:"installment_count"`
	Installments     []PaymentOptionInstallment `json:"installments"`
}

type PaymentOptionsResult struct {
	InvoiceID           uuid.UUID       `json:"invoice_id"`
	TotalCents          int64           `json:"total_cents"`
	InstallmentEligible bool            `json:"installment_eligible"`
	MaximumInstallments int             `json:"maximum_installments"`
	Options             []PaymentOption `json:"options"`
	PlanStatus          string          `json:"plan_status,omitempty"`
}

func planParamsFromSnapshots(
	minInv, minInst int64,
	maxInst int,
) InstallmentPolicyParams {
	return InstallmentPolicyParams{
		MinimumInvoiceAmountCents:     minInv,
		MinimumInstallmentAmountCents: minInst,
		MaximumInstallments:           maxInst,
	}
}

func (s *Service) createPendingPaymentPlanTx(ctx context.Context, tx pgx.Tx, invoiceID uuid.UUID, totalCents int64, policy *InstallmentPolicy) error {
	remaining := totalCents
	_, err := tx.Exec(ctx, `
		INSERT INTO invoice_payment_plans (
			invoice_id, policy_id, status,
			invoice_total_cents, paid_cents, remaining_cents,
			minimum_invoice_amount_cents_snapshot, minimum_installment_amount_cents_snapshot,
			maximum_installments_snapshot, installment_interval_months_snapshot,
			installment_enabled_snapshot, allow_early_payment_snapshot,
			require_sequential_payment_snapshot, adjust_business_day_snapshot
		) VALUES ($1,$2,$3,$4,0,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (invoice_id) DO NOTHING
	`, invoiceID, policy.ID, PlanStatusPendingSelection, totalCents, remaining,
		policy.MinimumInvoiceAmountCents, policy.MinimumInstallmentAmountCents,
		policy.MaximumInstallments, policy.InstallmentIntervalMonths,
		policy.InstallmentEnabled, policy.AllowEarlyInstallmentPayment,
		policy.RequireSequentialPayment, policy.AdjustDueDateToBusinessDay)
	return err
}

func (s *Service) EnsurePaymentPlanForInvoice(ctx context.Context, invoiceID uuid.UUID) error {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM invoice_payment_plans WHERE invoice_id = $1)`, invoiceID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	inv, err := s.GetInvoice(ctx, invoiceID)
	if err != nil || inv == nil {
		return ErrPeriodNotFound
	}
	policy, err := s.GetActiveInstallmentPolicy(ctx)
	if err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := s.createPendingPaymentPlanTx(ctx, tx, invoiceID, inv.TotalCents, policy); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type planRow struct {
	plan PaymentPlan
	minInv, minInst int64
	maxInst, interval int
	allowEarly, requireSeq, adjustBiz bool
}

func (s *Service) loadPlanForInvoice(ctx context.Context, invoiceID uuid.UUID) (*planRow, error) {
	var pr planRow
	var selected *int
	err := s.pool.QueryRow(ctx, `
		SELECT id, invoice_id, policy_id, status, selected_installment_count,
		       invoice_total_cents, paid_cents, remaining_cents, installment_enabled_snapshot,
		       minimum_invoice_amount_cents_snapshot, minimum_installment_amount_cents_snapshot,
		       maximum_installments_snapshot, installment_interval_months_snapshot,
		       allow_early_payment_snapshot, require_sequential_payment_snapshot, adjust_business_day_snapshot,
		       selected_at
		FROM invoice_payment_plans WHERE invoice_id = $1
	`, invoiceID).Scan(
		&pr.plan.ID, &pr.plan.InvoiceID, &pr.plan.PolicyID, &pr.plan.Status, &selected,
		&pr.plan.InvoiceTotalCents, &pr.plan.PaidCents, &pr.plan.RemainingCents, &pr.plan.InstallmentEnabledSnapshot,
		&pr.minInv, &pr.minInst, &pr.maxInst, &pr.interval,
		&pr.allowEarly, &pr.requireSeq, &pr.adjustBiz,
		&pr.plan.SelectedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	pr.plan.SelectedInstallmentCount = selected
	return &pr, nil
}

func (s *Service) GetPaymentOptions(ctx context.Context, invoiceID, customerID uuid.UUID) (*PaymentOptionsResult, error) {
	inv, err := s.GetInvoice(ctx, invoiceID)
	if err != nil || inv == nil || inv.CustomerID != customerID {
		return nil, ErrPeriodNotFound
	}
	if err := s.EnsurePaymentPlanForInvoice(ctx, invoiceID); err != nil {
		return nil, err
	}
	pr, err := s.loadPlanForInvoice(ctx, invoiceID)
	if err != nil || pr == nil {
		return nil, err
	}
	if pr.plan.Status != PlanStatusPendingSelection {
		max := 1
		if pr.plan.SelectedInstallmentCount != nil {
			max = *pr.plan.SelectedInstallmentCount
		}
		return &PaymentOptionsResult{
			InvoiceID: invoiceID, TotalCents: inv.TotalCents,
			InstallmentEligible: false, MaximumInstallments: max,
			PlanStatus: pr.plan.Status,
		}, nil
	}
	live, err := s.GetActiveInstallmentPolicy(ctx)
	if err != nil {
		return nil, err
	}
	params := planParamsFromSnapshots(pr.minInv, pr.minInst, pr.maxInst)
	max := MaxInstallments(inv.RemainingCents(), params, live.InstallmentEnabled)
	eligible := InstallmentEligible(max, live.InstallmentEnabled)
	opts := make([]PaymentOption, 0, max)
	for n := 1; n <= max; n++ {
		amounts := DistributeInstallmentAmounts(inv.RemainingCents(), n)
		dates, err := BuildInstallmentDueDates(ctx, s.pool, inv.DueAt, n, pr.interval, pr.adjustBiz)
		if err != nil {
			return nil, err
		}
		po := PaymentOption{InstallmentCount: n}
		for i := 0; i < n; i++ {
			po.Installments = append(po.Installments, PaymentOptionInstallment{
				Number: i + 1, AmountCents: amounts[i], DueDate: dates[i],
			})
		}
		opts = append(opts, po)
	}
	return &PaymentOptionsResult{
		InvoiceID: invoiceID, TotalCents: inv.RemainingCents(),
		InstallmentEligible: eligible, MaximumInstallments: max,
		Options: opts, PlanStatus: pr.plan.Status,
	}, nil
}

func (s *Service) SelectPaymentPlan(ctx context.Context, invoiceID, customerID, userID uuid.UUID, installmentCount int) (*PaymentPlan, error) {
	if installmentCount < 1 {
		return nil, ErrInvalidInstallmentCount
	}
	inv, err := s.GetInvoice(ctx, invoiceID)
	if err != nil || inv == nil || inv.CustomerID != customerID {
		return nil, ErrPeriodNotFound
	}
	if inv.RemainingCents() <= 0 {
		return nil, ErrInvalidInstallmentCount
	}
	if err := s.EnsurePaymentPlanForInvoice(ctx, invoiceID); err != nil {
		return nil, err
	}
	live, err := s.GetActiveInstallmentPolicy(ctx)
	if err != nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	pr, err := s.loadPlanForInvoiceTx(ctx, tx, invoiceID)
	if err != nil || pr == nil {
		return nil, ErrPaymentPlanNotPending
	}
	if pr.plan.Status != PlanStatusPendingSelection {
		return nil, ErrPaymentPlanImmutable
	}
	params := planParamsFromSnapshots(pr.minInv, pr.minInst, pr.maxInst)
	max := MaxInstallments(inv.RemainingCents(), params, live.InstallmentEnabled)
	if installmentCount > max {
		if !live.InstallmentEnabled && installmentCount > 1 {
			return nil, ErrInstallmentsDisabled
		}
		return nil, ErrInvalidInstallmentCount
	}
	if !live.AllowInstallmentAfterDueDate && time.Now().After(inv.DueAt) && installmentCount > 1 {
		return nil, ErrInstallmentAfterDue
	}

	amounts := DistributeInstallmentAmounts(inv.RemainingCents(), installmentCount)
	if installmentCount > 1 {
		base := amounts[0]
		if base < pr.minInst {
			return nil, ErrInvalidInstallmentCount
		}
	}
	dates, err := BuildInstallmentDueDates(ctx, s.pool, inv.DueAt, installmentCount, pr.interval, pr.adjustBiz)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE invoice_payment_plans SET
			status = $2, selected_installment_count = $3,
			selected_by_user_id = $4, selected_at = $5, updated_at = NOW()
		WHERE id = $1 AND status = $6
	`, pr.plan.ID, PlanStatusActive, installmentCount, userID, now, PlanStatusPendingSelection)
	if err != nil {
		return nil, err
	}
	for i := 0; i < installmentCount; i++ {
		st := InstStatusScheduled
		var openedAt *time.Time
		if i == 0 {
			st = InstStatusOpen
			openedAt = &now
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO invoice_installments (
				payment_plan_id, invoice_id, installment_number,
				amount_cents, paid_cents, remaining_cents, due_date, status, opened_at
			) VALUES ($1,$2,$3,$4,0,$4,$5,$6,$7)
		`, pr.plan.ID, invoiceID, i+1, amounts[i], dates[i], st, openedAt)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	plan, _ := s.GetPaymentPlan(ctx, invoiceID, customerID)
	return plan, nil
}

func (s *Service) loadPlanForInvoiceTx(ctx context.Context, tx pgx.Tx, invoiceID uuid.UUID) (*planRow, error) {
	var pr planRow
	var selected *int
	err := tx.QueryRow(ctx, `
		SELECT id, invoice_id, policy_id, status, selected_installment_count,
		       invoice_total_cents, paid_cents, remaining_cents, installment_enabled_snapshot,
		       minimum_invoice_amount_cents_snapshot, minimum_installment_amount_cents_snapshot,
		       maximum_installments_snapshot, installment_interval_months_snapshot,
		       allow_early_payment_snapshot, require_sequential_payment_snapshot, adjust_business_day_snapshot,
		       selected_at
		FROM invoice_payment_plans WHERE invoice_id = $1 FOR UPDATE
	`, invoiceID).Scan(
		&pr.plan.ID, &pr.plan.InvoiceID, &pr.plan.PolicyID, &pr.plan.Status, &selected,
		&pr.plan.InvoiceTotalCents, &pr.plan.PaidCents, &pr.plan.RemainingCents, &pr.plan.InstallmentEnabledSnapshot,
		&pr.minInv, &pr.minInst, &pr.maxInst, &pr.interval,
		&pr.allowEarly, &pr.requireSeq, &pr.adjustBiz,
		&pr.plan.SelectedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	pr.plan.SelectedInstallmentCount = selected
	return &pr, nil
}

func (s *Service) GetPaymentPlan(ctx context.Context, invoiceID, customerID uuid.UUID) (*PaymentPlan, error) {
	inv, err := s.GetInvoice(ctx, invoiceID)
	if err != nil || inv == nil || inv.CustomerID != customerID {
		return nil, ErrPeriodNotFound
	}
	pr, err := s.loadPlanForInvoice(ctx, invoiceID)
	if err != nil || pr == nil {
		return nil, ErrPeriodNotFound
	}
	return &pr.plan, nil
}

func (s *Service) ListInvoiceInstallments(ctx context.Context, invoiceID, customerID uuid.UUID) ([]InvoiceInstallment, error) {
	inv, err := s.GetInvoice(ctx, invoiceID)
	if err != nil || inv == nil || inv.CustomerID != customerID {
		return nil, ErrPeriodNotFound
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, payment_plan_id, invoice_id, installment_number,
		       amount_cents, paid_cents, remaining_cents, due_date, status
		FROM invoice_installments WHERE invoice_id = $1 ORDER BY installment_number
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InvoiceInstallment
	for rows.Next() {
		var inst InvoiceInstallment
		if err := rows.Scan(&inst.ID, &inst.PaymentPlanID, &inst.InvoiceID, &inst.InstallmentNumber,
			&inst.AmountCents, &inst.PaidCents, &inst.RemainingCents, &inst.DueDate, &inst.Status); err != nil {
			return nil, err
		}
		out = append(out, inst)
	}
	return out, rows.Err()
}

// InvoiceRequiresInstallmentPix is true when the customer must pay via installment flow.
func (s *Service) InvoiceRequiresInstallmentPix(ctx context.Context, invoiceID uuid.UUID) (bool, error) {
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT status FROM invoice_payment_plans WHERE invoice_id = $1
	`, invoiceID).Scan(&status)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == PlanStatusPendingSelection || status == PlanStatusActive, nil
}

func (s *Service) GetInstallmentForCustomer(ctx context.Context, installmentID, customerID uuid.UUID) (*InvoiceInstallment, error) {
	var inst InvoiceInstallment
	err := s.pool.QueryRow(ctx, `
		SELECT ii.id, ii.payment_plan_id, ii.invoice_id, ii.installment_number,
		       ii.amount_cents, ii.paid_cents, ii.remaining_cents, ii.due_date, ii.status
		FROM invoice_installments ii
		JOIN invoices i ON i.id = ii.invoice_id
		WHERE ii.id = $1 AND i.customer_id = $2
	`, installmentID, customerID).Scan(
		&inst.ID, &inst.PaymentPlanID, &inst.InvoiceID, &inst.InstallmentNumber,
		&inst.AmountCents, &inst.PaidCents, &inst.RemainingCents, &inst.DueDate, &inst.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (s *Service) GetInstallmentByID(ctx context.Context, installmentID uuid.UUID) (*InvoiceInstallment, error) {
	var inst InvoiceInstallment
	err := s.pool.QueryRow(ctx, `
		SELECT id, payment_plan_id, invoice_id, installment_number,
		       amount_cents, paid_cents, remaining_cents, due_date, status
		FROM invoice_installments WHERE id = $1
	`, installmentID).Scan(
		&inst.ID, &inst.PaymentPlanID, &inst.InvoiceID, &inst.InstallmentNumber,
		&inst.AmountCents, &inst.PaidCents, &inst.RemainingCents, &inst.DueDate, &inst.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// ApplyInstallmentPaymentTx settles one installment and updates plan + invoice.
func (s *Service) ApplyInstallmentPaymentTx(ctx context.Context, tx pgx.Tx, installmentID uuid.UUID, amountCents int64) error {
	var inst InvoiceInstallment
	var planID uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT id, payment_plan_id, invoice_id, installment_number,
		       amount_cents, paid_cents, remaining_cents, status
		FROM invoice_installments WHERE id = $1 FOR UPDATE
	`, installmentID).Scan(
		&inst.ID, &planID, &inst.InvoiceID, &inst.InstallmentNumber,
		&inst.AmountCents, &inst.PaidCents, &inst.RemainingCents, &inst.Status,
	)
	if err != nil {
		return err
	}
	if inst.Status != InstStatusOpen && inst.Status != InstStatusPixActive {
		return ErrInstallmentNotPayable
	}
	if amountCents != inst.RemainingCents {
		return ErrInstallmentNotPayable
	}
	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE invoice_installments SET
			paid_cents = amount_cents, remaining_cents = 0, status = $2, paid_at = $3, updated_at = NOW()
		WHERE id = $1
	`, installmentID, InstStatusPaid, now)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE invoice_payment_plans SET
			paid_cents = paid_cents + $2,
			remaining_cents = GREATEST(0, remaining_cents - $2),
			updated_at = NOW()
		WHERE id = $1
	`, planID, amountCents)
	if err != nil {
		return err
	}
	var invTotal, invPaid int64
	var invStatus string
	err = tx.QueryRow(ctx, `
		SELECT total_cents, paid_cents, status FROM invoices WHERE id = $1 FOR UPDATE
	`, inst.InvoiceID).Scan(&invTotal, &invPaid, &invStatus)
	if err != nil {
		return err
	}
	newPaid := invPaid + amountCents
	newStatus := invStatus
	markPaidAt := false
	if newPaid >= invTotal {
		newStatus = "paid"
		markPaidAt = true
	} else if newPaid > 0 {
		newStatus = "partially_paid"
	}
	if markPaidAt {
		_, err = tx.Exec(ctx, `
			UPDATE invoices SET paid_cents = $2, status = $3, paid_at = COALESCE(paid_at, NOW()), updated_at = NOW()
			WHERE id = $1
		`, inst.InvoiceID, newPaid, newStatus)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE invoices SET paid_cents = $2, status = $3, updated_at = NOW()
			WHERE id = $1
		`, inst.InvoiceID, newPaid, newStatus)
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE customers SET current_exposure_cents = GREATEST(0, current_exposure_cents - $2), updated_at = NOW()
		WHERE id = (SELECT customer_id FROM invoices WHERE id = $1)
	`, inst.InvoiceID, amountCents)
	if err != nil {
		return err
	}
	var remaining int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM invoice_installments
		WHERE payment_plan_id = $1 AND status NOT IN ($2, $3)
	`, planID, InstStatusPaid, InstStatusCanceled).Scan(&remaining)
	if err != nil {
		return err
	}
	if remaining == 0 {
		_, err = tx.Exec(ctx, `UPDATE invoice_payment_plans SET status = $2, updated_at = NOW() WHERE id = $1`, planID, PlanStatusCompleted)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE invoice_installments SET status = $2, opened_at = COALESCE(opened_at, NOW()), updated_at = NOW()
			WHERE payment_plan_id = $1 AND installment_number = $3 AND status = $4
		`, planID, InstStatusOpen, inst.InstallmentNumber+1, InstStatusScheduled)
	}
	if err != nil {
		return err
	}
	if !markPaidAt {
		return s.syncInvoiceDueFromInstallmentsTx(ctx, tx, inst.InvoiceID)
	}
	return nil
}

func (s *Service) syncInvoiceDueFromInstallmentsTx(ctx context.Context, tx pgx.Tx, invoiceID uuid.UUID) error {
	next, err := s.nextUnpaidInstallmentDueDateTx(ctx, tx, invoiceID)
	if err != nil {
		return err
	}
	if next == nil {
		return nil
	}
	_, err = tx.Exec(ctx, `UPDATE invoices SET due_at = $2, updated_at = NOW() WHERE id = $1`, invoiceID, *next)
	return err
}

func (s *Service) MarkInstallmentPixActive(ctx context.Context, installmentID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE invoice_installments SET status = $2, updated_at = NOW()
		WHERE id = $1 AND status = $3
	`, installmentID, InstStatusPixActive, InstStatusOpen)
	return err
}

func (s *Service) MarkOverdueInstallments(ctx context.Context, now time.Time) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `
		UPDATE invoice_installments SET status = $1, overdue_at = NOW(), updated_at = NOW()
		WHERE status IN ('open', 'pix_active') AND due_date < $2::date
	`, InstStatusOverdue, now)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(ctx, `
		UPDATE invoices i SET status = 'overdue', updated_at = NOW()
		WHERE i.status IN ('open', 'partially_paid')
		  AND EXISTS (
		    SELECT 1 FROM invoice_installments ii
		    WHERE ii.invoice_id = i.id AND ii.status = $1
		  )
	`, InstStatusOverdue)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (s *Service) ResetPaymentPlan(ctx context.Context, invoiceID uuid.UUID, actorID uuid.UUID, reason string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	pr, err := s.loadPlanForInvoiceTx(ctx, tx, invoiceID)
	if err != nil || pr == nil {
		return ErrPeriodNotFound
	}
	var paidCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM invoice_installments
		WHERE payment_plan_id = $1 AND status = $2
	`, pr.plan.ID, InstStatusPaid).Scan(&paidCount)
	if err != nil {
		return err
	}
	if paidCount > 0 {
		return ErrPaymentPlanImmutable
	}
	_, err = tx.Exec(ctx, `DELETE FROM invoice_installments WHERE payment_plan_id = $1`, pr.plan.ID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE invoice_payment_plans SET
			status = $2, selected_installment_count = NULL,
			selected_by_user_id = NULL, selected_at = NULL,
			canceled_by_user_id = $3, canceled_at = NOW(), cancellation_reason = $4,
			updated_at = NOW()
		WHERE id = $1
	`, pr.plan.ID, PlanStatusPendingSelection, actorID, reason)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) ValidateInstallmentForPix(ctx context.Context, installmentID uuid.UUID) error {
	var inst InvoiceInstallment
	var requireSeq bool
	err := s.pool.QueryRow(ctx, `
		SELECT ii.id, ii.installment_number, ii.status, ii.remaining_cents,
		       pp.require_sequential_payment_snapshot
		FROM invoice_installments ii
		JOIN invoice_payment_plans pp ON pp.id = ii.payment_plan_id
		WHERE ii.id = $1
	`, installmentID).Scan(&inst.ID, &inst.InstallmentNumber, &inst.Status, &inst.RemainingCents, &requireSeq)
	if err != nil {
		return err
	}
	if inst.Status != InstStatusOpen {
		return ErrInstallmentNotPayable
	}
	if requireSeq && inst.InstallmentNumber > 1 {
		var prevStatus string
		err = s.pool.QueryRow(ctx, `
			SELECT status FROM invoice_installments
			WHERE payment_plan_id = (SELECT payment_plan_id FROM invoice_installments WHERE id = $1)
			  AND installment_number = $2
		`, installmentID, inst.InstallmentNumber-1).Scan(&prevStatus)
		if err != nil || prevStatus != InstStatusPaid {
			return ErrInstallmentNotSequential
		}
	}
	return nil
}
