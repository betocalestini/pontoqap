package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UpdateCategoryInput struct {
	Name   *string
	Active *bool
}

func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, in UpdateCategoryInput) (Category, error) {
	var c Category
	err := s.pool.QueryRow(ctx, `SELECT id, name, slug, active FROM categories WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, fmt.Errorf("categoria não encontrada")
	}
	if err != nil {
		return Category{}, err
	}
	name := c.Name
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
		if name == "" {
			return Category{}, fmt.Errorf("nome obrigatório")
		}
	}
	active := c.Active
	if in.Active != nil {
		active = *in.Active
	}
	err = s.pool.QueryRow(ctx, `
		UPDATE categories SET name = $2, active = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, slug, active
	`, id, name, active).Scan(&c.ID, &c.Name, &c.Slug, &c.Active)
	return c, err
}

type DeleteCategoryResult struct {
	ProductsUnlinked int `json:"products_unlinked"`
}

func (s *Service) DeleteCategory(ctx context.Context, id uuid.UUID) (DeleteCategoryResult, error) {
	var res DeleteCategoryResult
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return res, err
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return res, err
	}
	if !exists {
		return res, fmt.Errorf("categoria não encontrada")
	}

	tag, err := tx.Exec(ctx, `UPDATE products SET category_id = NULL, updated_at = NOW() WHERE category_id = $1`, id)
	if err != nil {
		return res, err
	}
	res.ProductsUnlinked = int(tag.RowsAffected())

	_, err = tx.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return res, err
	}
	if err := tx.Commit(ctx); err != nil {
		return res, err
	}
	return res, nil
}
