package customers

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CollaboratorCategory struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	MarginPercent float64   `json:"margin_percent"`
	Active        bool      `json:"active"`
}

type CreateCollaboratorCategoryInput struct {
	Name          string
	Slug          string
	MarginPercent float64
}

type UpdateCollaboratorCategoryInput struct {
	Name          *string
	MarginPercent *float64
	Active        *bool
}

func (s *Service) ListCollaboratorCategories(ctx context.Context) ([]CollaboratorCategory, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, slug, margin_percent, active
		FROM collaborator_categories
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CollaboratorCategory
	for rows.Next() {
		var c CollaboratorCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.MarginPercent, &c.Active); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Service) CreateCollaboratorCategory(ctx context.Context, in CreateCollaboratorCategoryInput) (*CollaboratorCategory, error) {
	if err := validateCollabMargin(in.MarginPercent); err != nil {
		return nil, err
	}
	slug := in.Slug
	if slug == "" {
		slug = slugifyCollab(in.Name)
	}
	var c CollaboratorCategory
	err := s.pool.QueryRow(ctx, `
		INSERT INTO collaborator_categories (name, slug, margin_percent)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, margin_percent, active
	`, in.Name, slug, in.MarginPercent).Scan(&c.ID, &c.Name, &c.Slug, &c.MarginPercent, &c.Active)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) UpdateCollaboratorCategory(ctx context.Context, id uuid.UUID, in UpdateCollaboratorCategoryInput) (*CollaboratorCategory, error) {
	cur, err := s.getCollaboratorCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, ErrNotFound()
	}
	name := cur.Name
	margin := cur.MarginPercent
	active := cur.Active
	if in.Name != nil {
		name = *in.Name
	}
	if in.MarginPercent != nil {
		if err := validateCollabMargin(*in.MarginPercent); err != nil {
			return nil, err
		}
		margin = *in.MarginPercent
	}
	if in.Active != nil {
		active = *in.Active
	}
	var c CollaboratorCategory
	err = s.pool.QueryRow(ctx, `
		UPDATE collaborator_categories SET name = $2, margin_percent = $3, active = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, slug, margin_percent, active
	`, id, name, margin, active).Scan(&c.ID, &c.Name, &c.Slug, &c.MarginPercent, &c.Active)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// AssertCollaboratorCategoryAssignable ensures the category exists and is active.
func (s *Service) AssertCollaboratorCategoryAssignable(ctx context.Context, id uuid.UUID) error {
	cur, err := s.getCollaboratorCategory(ctx, id)
	if err != nil {
		return err
	}
	if cur == nil || !cur.Active {
		return ErrInvalidCollaboratorCategory()
	}
	return nil
}

func (s *Service) getCollaboratorCategory(ctx context.Context, id uuid.UUID) (*CollaboratorCategory, error) {
	var c CollaboratorCategory
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, slug, margin_percent, active FROM collaborator_categories WHERE id = $1
	`, id).Scan(&c.ID, &c.Name, &c.Slug, &c.MarginPercent, &c.Active)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func validateCollabMargin(m float64) error {
	if m < 0 || m > 1000 {
		return fmt.Errorf("margem inválida")
	}
	return nil
}

func slugifyCollab(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

// CollaboratorMarginForCustomer returns active category margin for pricing, if any.
func (s *Service) CollaboratorMarginForCustomer(ctx context.Context, customerID uuid.UUID) (*float64, error) {
	var margin float64
	err := s.pool.QueryRow(ctx, `
		SELECT cc.margin_percent
		FROM customers c
		JOIN collaborator_categories cc ON cc.id = c.collaborator_category_id
		WHERE c.id = $1 AND c.status = 'approved' AND cc.active = TRUE
	`, customerID).Scan(&margin)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &margin, nil
}
