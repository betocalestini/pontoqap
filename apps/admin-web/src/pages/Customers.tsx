import { useCallback, useEffect, useState } from 'react';
import type { AdminCustomer, CollaboratorCategory } from '@store/api-client';
import { api } from '../api';

function formatBRL(cents: number) {
  return (cents / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

const emptyCategoryForm = () => ({ name: '', margin_percent: '15', active: true });

export function CustomersPage() {
  const [items, setItems] = useState<AdminCustomer[]>([]);
  const [categories, setCategories] = useState<CollaboratorCategory[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({
    name: '',
    phone: '',
    document: '',
    collaborator_category_id: '',
    limit_cents: '',
    limit_reason: '',
    block_reason: '',
  });
  const [catForm, setCatForm] = useState(emptyCategoryForm);
  const [editingCatId, setEditingCatId] = useState<string | null>(null);
  const [initialLimitCents, setInitialLimitCents] = useState(0);
  const [approveLimitReais, setApproveLimitReais] = useState('1000');

  const load = useCallback(async () => {
    const [custRes, catRes] = await Promise.all([
      api.adminListCustomers(),
      api.adminListCollaboratorCategories(),
    ]);
    setItems((custRes.items ?? []) as AdminCustomer[]);
    setCategories((catRes.items ?? []) as CollaboratorCategory[]);
  }, []);

  useEffect(() => {
    load().catch((e: Error) => setError(e.message));
  }, [load]);

  async function approve(id: string) {
    setError(null);
    const limitCents = Math.round(parseFloat(approveLimitReais.replace(',', '.')) * 100);
    if (!Number.isFinite(limitCents) || limitCents < 0) {
      setError('Informe um limite válido para aprovação');
      return;
    }
    try {
      await api.adminApproveCustomer(id, limitCents);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  function startEdit(c: AdminCustomer) {
    setEditingId(c.id);
    setInitialLimitCents(c.credit_limit_cents ?? 0);
    setForm({
      name: c.name ?? '',
      phone: c.phone ?? '',
      document: c.document ?? '',
      collaborator_category_id: c.collaborator_category_id ?? '',
      limit_cents: String((c.credit_limit_cents ?? 0) / 100),
      limit_reason: '',
      block_reason: '',
    });
  }

  async function saveCustomer(e: React.FormEvent) {
    e.preventDefault();
    if (!editingId) return;
    setError(null);
    try {
      await api.adminUpdateCustomer(editingId, {
        name: form.name,
        phone: form.phone,
        document: form.document,
        collaborator_category_id: form.collaborator_category_id || null,
      });
      const limitCents = Math.round(parseFloat(form.limit_cents.replace(',', '.')) * 100);
      if (Number.isFinite(limitCents) && limitCents >= 0 && limitCents !== initialLimitCents) {
        await api.adminChangeCustomerLimit(editingId, limitCents, form.limit_reason || 'Ajuste admin');
      }
      setEditingId(null);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar');
    }
  }

  async function blockCustomer() {
    if (!editingId) return;
    setError(null);
    try {
      await api.adminBlockCustomer(editingId, form.block_reason);
      setEditingId(null);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro');
    }
  }

  async function unblockCustomer(id: string) {
    setError(null);
    try {
      await api.adminUnblockCustomer(id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro');
    }
  }

  async function saveCategory(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const margin = parseFloat(catForm.margin_percent.replace(',', '.'));
    if (!catForm.name.trim() || !Number.isFinite(margin)) {
      setError('Nome e margem da categoria são obrigatórios');
      return;
    }
    try {
      if (editingCatId) {
        await api.adminUpdateCollaboratorCategory(editingCatId, {
          name: catForm.name,
          margin_percent: margin,
          active: catForm.active,
        });
      } else {
        await api.adminCreateCollaboratorCategory({ name: catForm.name, margin_percent: margin });
      }
      setCatForm(emptyCategoryForm());
      setEditingCatId(null);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro na categoria');
    }
  }

  return (
    <section className="content-section customers-page">
      <h1>Clientes</h1>
      {error && <p className="error">{error}</p>}

      <div className="collaborator-categories-bar">
        <h2>Categorias de colaboradores</h2>
        <p className="form-hint">
          Margem sobre o custo médio dos lotes; na loja o colaborador paga o menor entre esse preço e o varejo/promoção.
        </p>
        <form className="collaborator-cat-form" onSubmit={saveCategory}>
          <label>
            Nome
            <input value={catForm.name} onChange={(e) => setCatForm((f) => ({ ...f, name: e.target.value }))} required />
          </label>
          <label>
            Margem (%)
            <input
              value={catForm.margin_percent}
              onChange={(e) => setCatForm((f) => ({ ...f, margin_percent: e.target.value }))}
              inputMode="decimal"
              required
            />
          </label>
          {editingCatId && (
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={catForm.active}
                onChange={(e) => setCatForm((f) => ({ ...f, active: e.target.checked }))}
              />
              Categoria ativa
            </label>
          )}
          <button type="submit">{editingCatId ? 'Atualizar categoria' : 'Adicionar categoria'}</button>
          {editingCatId && (
            <button
              type="button"
              onClick={() => {
                setEditingCatId(null);
                setCatForm(emptyCategoryForm());
              }}
            >
              Cancelar edição
            </button>
          )}
        </form>
        <ul className="collaborator-cat-list">
          {categories.map((cat) => (
            <li key={cat.id}>
              <strong>{cat.name}</strong> — {cat.margin_percent}% {!cat.active && '(inativa)'}
              <button
                type="button"
                onClick={() => {
                  setEditingCatId(cat.id);
                  setCatForm({
                    name: cat.name,
                    margin_percent: String(cat.margin_percent),
                    active: cat.active,
                  });
                }}
              >
                Editar
              </button>
            </li>
          ))}
        </ul>
      </div>

      {editingId && (
        <form className="form form--wide customer-edit-form" onSubmit={saveCustomer}>
          <h2>Editar cliente</h2>
          <label>
            Nome
            <input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} required />
          </label>
          <label>
            Telefone
            <input value={form.phone} onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))} />
          </label>
          <label>
            Documento
            <input value={form.document} onChange={(e) => setForm((f) => ({ ...f, document: e.target.value }))} />
          </label>
          <label>
            Categoria colaborador
            <select
              value={form.collaborator_category_id}
              onChange={(e) => setForm((f) => ({ ...f, collaborator_category_id: e.target.value }))}
            >
              <option value="">— Não é colaborador —</option>
              {categories.filter((c) => c.active).map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name} ({c.margin_percent}%)
                </option>
              ))}
            </select>
          </label>
          <label>
            Limite de crédito (R$)
            <input value={form.limit_cents} onChange={(e) => setForm((f) => ({ ...f, limit_cents: e.target.value }))} />
          </label>
          <label>
            Motivo do limite (opcional)
            <input value={form.limit_reason} onChange={(e) => setForm((f) => ({ ...f, limit_reason: e.target.value }))} />
          </label>
          <div className="form__actions">
            <button type="submit">Salvar</button>
            <button type="button" onClick={() => setEditingId(null)}>
              Cancelar
            </button>
          </div>
          <fieldset className="customer-block-fieldset">
            <legend>Bloqueio</legend>
            <label>
              Motivo do bloqueio
              <input value={form.block_reason} onChange={(e) => setForm((f) => ({ ...f, block_reason: e.target.value }))} />
            </label>
            <button type="button" className="button--danger" onClick={() => void blockCustomer()}>
              Bloquear cliente
            </button>
          </fieldset>
        </form>
      )}

      <div className="table-wrap">
        <p className="form-hint approve-limit-hint">
          Limite padrão na aprovação (R$):{' '}
          <input
            className="approve-limit-input"
            value={approveLimitReais}
            onChange={(e) => setApproveLimitReais(e.target.value)}
            inputMode="decimal"
            aria-label="Limite na aprovação em reais"
          />
        </p>
        <table className="customers-table">
          <thead>
            <tr>
              <th>Nome</th>
              <th>E-mail</th>
              <th>Status</th>
              <th>Colaborador</th>
              <th>Limite</th>
              <th>Exposição</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {items.map((c) => (
              <tr key={c.id}>
                <td>{c.name}</td>
                <td>
                  {c.email}
                  {c.email_verified === false && ' (e-mail pendente)'}
                </td>
                <td>
                  {c.status}
                  {c.status === 'blocked' && c.blocked_reason ? ` — ${c.blocked_reason}` : ''}
                </td>
                <td>{c.collaborator_category_name || '—'}</td>
                <td>{formatBRL(c.credit_limit_cents ?? 0)}</td>
                <td>{formatBRL(c.current_exposure_cents ?? 0)}</td>
                <td className="table-actions">
                  {c.status === 'pending' && (
                    <button type="button" onClick={() => void approve(c.id)}>
                      Aprovar
                    </button>
                  )}
                  <button type="button" onClick={() => startEdit(c)}>
                    Editar
                  </button>
                  {c.status === 'blocked' && (
                    <button type="button" onClick={() => void unblockCustomer(c.id)}>
                      Desbloquear
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
