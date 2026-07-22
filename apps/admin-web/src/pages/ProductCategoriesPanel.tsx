import { type FormEvent, useState } from 'react';
import type { AdminCategory } from '@store/api-client';
import { useDialog } from '@store/ui';
import { api } from '../api';

function slugify(name: string) {
  return name
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '');
}

const emptyCatForm = () => ({ name: '', slug: '', active: true });

export type ProductCategoriesPanelProps = {
  categories: AdminCategory[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onReload: () => Promise<void>;
  onError: (message: string | null) => void;
  onCategoryDeleted: (categoryId: string) => void;
};

export function ProductCategoriesPanel({
  categories,
  open,
  onOpenChange,
  onReload,
  onError,
  onCategoryDeleted,
}: ProductCategoriesPanelProps) {
  const { confirm } = useDialog();
  const [editingCatId, setEditingCatId] = useState<string | null>(null);
  const [catForm, setCatForm] = useState(emptyCatForm);

  async function saveCategory(e: FormEvent) {
    e.preventDefault();
    onError(null);
    const name = catForm.name.trim();
    if (!name) {
      onError('Nome da categoria é obrigatório');
      return;
    }
    try {
      if (editingCatId) {
        await api.adminUpdateCategory(editingCatId, { name, active: catForm.active });
      } else {
        const slug = catForm.slug.trim() || slugify(name);
        await api.adminCreateCategory({ name, slug });
      }
      setCatForm(emptyCatForm());
      setEditingCatId(null);
      await onReload();
    } catch (err) {
      onError(err instanceof Error ? err.message : 'Erro na categoria');
    }
  }

  async function deleteCategory(cat: AdminCategory) {
    const message = `Excluir a categoria "${cat.name}"? Produtos vinculados ficarão sem categoria.`;
    const ok = await confirm({
      title: 'Excluir categoria',
      message,
      confirmLabel: 'Excluir',
      variant: 'danger',
    });
    if (!ok) return;
    onError(null);
    try {
      await api.adminDeleteCategory(cat.id);
      if (editingCatId === cat.id) {
        setEditingCatId(null);
        setCatForm(emptyCatForm());
      }
      onCategoryDeleted(cat.id);
      await onReload();
    } catch (err) {
      onError(err instanceof Error ? err.message : 'Falha ao excluir categoria');
    }
  }

  return (
    <details
      className="product-categories-bar"
      open={open}
      onToggle={(e) => onOpenChange((e.target as HTMLDetailsElement).open)}
    >
      <summary className="product-categories-bar__summary">Categorias de produtos</summary>
      <p className="form-hint">Organize o catálogo da loja. O slug é definido na criação e não pode ser alterado.</p>
      <form className="product-cat-form" onSubmit={saveCategory}>
        <label>
          Nome
          <input
            value={catForm.name}
            onChange={(e) =>
              setCatForm((f) => ({
                ...f,
                name: e.target.value,
                slug: editingCatId ? f.slug : f.slug || slugify(e.target.value),
              }))
            }
            required
          />
        </label>
        {!editingCatId && (
          <label>
            Slug (opcional)
            <input value={catForm.slug} onChange={(e) => setCatForm((f) => ({ ...f, slug: e.target.value }))} />
          </label>
        )}
        {editingCatId && (
          <label className="product-cat-form__checkbox form__checkbox">
            <input
              type="checkbox"
              checked={catForm.active}
              onChange={(e) => setCatForm((f) => ({ ...f, active: e.target.checked }))}
            />
            Categoria ativa na loja
          </label>
        )}
        <div className="product-cat-form__actions form__full">
          <button type="submit">{editingCatId ? 'Atualizar categoria' : 'Adicionar categoria'}</button>
          {editingCatId && (
            <button
              type="button"
              onClick={() => {
                setEditingCatId(null);
                setCatForm(emptyCatForm());
              }}
            >
              Cancelar edição
            </button>
          )}
        </div>
      </form>
      <ul className="product-cat-list">
        {categories.map((cat) => (
          <li key={cat.id}>
            <span className="product-cat-list__label">
              <strong>{cat.name}</strong>
              <span className="product-cat-list__slug"> ({cat.slug})</span>
              {!cat.active && <span className="product-cat-list__inactive"> — inativa</span>}
            </span>
            <span className="product-cat-list__actions">
              <button
                type="button"
                onClick={() => {
                  setEditingCatId(cat.id);
                  setCatForm({ name: cat.name, slug: cat.slug, active: cat.active });
                }}
              >
                Editar
              </button>
              <button type="button" className="button--danger" onClick={() => void deleteCategory(cat)}>
                Excluir
              </button>
            </span>
          </li>
        ))}
      </ul>
    </details>
  );
}
