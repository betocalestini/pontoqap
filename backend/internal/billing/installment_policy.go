package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InstallmentPolicy is a versioned installment configuration.
type InstallmentPolicy struct {
	ID                            uuid.UUID  `json:"id"`
	Version                       int        `json:"version"`
	Active                        bool       `json:"active"`
	InstallmentEnabled            bool       `json:"installment_enabled"`
	MinimumInvoiceAmountCents     int64      `json:"minimum_invoice_amount_cents"`
	MinimumInstallmentAmountCents int64      `json:"minimum_installment_amount_cents"`
	MaximumInstallments           int        `json:"maximum_installments"`
	InstallmentIntervalMonths     int        `json:"installment_interval_months"`
	AllowInstallmentAfterDueDate  bool       `json:"allow_installment_after_due_date"`
	AllowEarlyInstallmentPayment  bool       `json:"allow_early_installment_payment"`
	RequireSequentialPayment      bool       `json:"require_sequential_payment"`
	AdjustDueDateToBusinessDay    bool       `json:"adjust_due_date_to_business_day"`
	ValidFrom                     time.Time  `json:"valid_from"`
	ValidUntil                    *time.Time `json:"valid_until,omitempty"`
	CreatedAt                     time.Time  `json:"created_at"`
}

type UpdateInstallmentPolicyInput struct {
	InstallmentEnabled            bool  `json:"installment_enabled"`
	MinimumInvoiceAmountCents     int64 `json:"minimum_invoice_amount_cents"`
	MinimumInstallmentAmountCents int64 `json:"minimum_installment_amount_cents"`
	MaximumInstallments           int   `json:"maximum_installments"`
	InstallmentIntervalMonths     int   `json:"installment_interval_months"`
	AllowInstallmentAfterDueDate  bool  `json:"allow_installment_after_due_date"`
	AllowEarlyInstallmentPayment  bool  `json:"allow_early_installment_payment"`
	RequireSequentialPayment      bool  `json:"require_sequential_payment"`
	AdjustDueDateToBusinessDay    bool  `json:"adjust_due_date_to_business_day"`
}

func scanInstallmentPolicy(row pgx.Row) (*InstallmentPolicy, error) {
	var p InstallmentPolicy
	err := row.Scan(
		&p.ID, &p.Version, &p.Active, &p.InstallmentEnabled,
		&p.MinimumInvoiceAmountCents, &p.MinimumInstallmentAmountCents, &p.MaximumInstallments,
		&p.InstallmentIntervalMonths, &p.AllowInstallmentAfterDueDate, &p.AllowEarlyInstallmentPayment,
		&p.RequireSequentialPayment, &p.AdjustDueDateToBusinessDay,
		&p.ValidFrom, &p.ValidUntil, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

const installmentPolicySelect = `
	SELECT id, version, active, installment_enabled,
	       minimum_invoice_amount_cents, minimum_installment_amount_cents, maximum_installments,
	       installment_interval_months, allow_installment_after_due_date, allow_early_installment_payment,
	       require_sequential_payment, adjust_due_date_to_business_day,
	       valid_from, valid_until, created_at
	FROM installment_policies
`

func (s *Service) GetActiveInstallmentPolicy(ctx context.Context) (*InstallmentPolicy, error) {
	p, err := scanInstallmentPolicy(s.pool.QueryRow(ctx, installmentPolicySelect+` WHERE active = true LIMIT 1`))
	if err == pgx.ErrNoRows {
		return nil, ErrNoActiveInstallmentPolicy
	}
	return p, err
}

func (s *Service) ListInstallmentPolicies(ctx context.Context, limit int) ([]InstallmentPolicy, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, installmentPolicySelect+` ORDER BY version DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InstallmentPolicy
	for rows.Next() {
		p, err := scanInstallmentPolicy(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func (s *Service) CreateInstallmentPolicyVersion(ctx context.Context, in UpdateInstallmentPolicyInput, actorID uuid.UUID) (*InstallmentPolicy, error) {
	if in.MaximumInstallments < 1 || in.InstallmentIntervalMonths < 1 {
		return nil, ErrInvalidInstallmentCount
	}
	if in.MinimumInstallmentAmountCents < 1 || in.MinimumInvoiceAmountCents < 1 {
		return nil, ErrInvalidInstallmentCount
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var nextVersion int
	err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) + 1 FROM installment_policies`).Scan(&nextVersion)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `
		UPDATE installment_policies SET active = false, valid_until = NOW()
		WHERE active = true
	`)
	if err != nil {
		return nil, err
	}
	id := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO installment_policies (
			id, version, active, installment_enabled,
			minimum_invoice_amount_cents, minimum_installment_amount_cents, maximum_installments,
			installment_interval_months, allow_installment_after_due_date, allow_early_installment_payment,
			require_sequential_payment, adjust_due_date_to_business_day,
			valid_from, created_by
		) VALUES ($1,$2,true,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),$12)
	`, id, nextVersion, in.InstallmentEnabled,
		in.MinimumInvoiceAmountCents, in.MinimumInstallmentAmountCents, in.MaximumInstallments,
		in.InstallmentIntervalMonths, in.AllowInstallmentAfterDueDate, in.AllowEarlyInstallmentPayment,
		in.RequireSequentialPayment, in.AdjustDueDateToBusinessDay, actorID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetActiveInstallmentPolicy(ctx)
}

func (p *InstallmentPolicy) Params() InstallmentPolicyParams {
	return InstallmentPolicyParams{
		MinimumInvoiceAmountCents:     p.MinimumInvoiceAmountCents,
		MinimumInstallmentAmountCents: p.MinimumInstallmentAmountCents,
		MaximumInstallments:           p.MaximumInstallments,
		InstallmentEnabled:            p.InstallmentEnabled,
	}
}
