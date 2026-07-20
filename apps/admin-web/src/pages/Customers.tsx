import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { AdminCustomer, AdminStaffRole, CollaboratorCategory } from '@store/api-client';
import { api } from '../api';

function formatBRL(cents: number) {
  return (cents / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

const emptyCategoryForm = () => ({ name: '', margin_percent: '15', active: true });

const statusLabels: Record<string, string> = {
  pending: 'Pendente',
  approved: 'Aprovado',
  rejected: 'Rejeitado',
  blocked: 'Bloqueado',
};

function CustomerStatus({ customer }: { customer: AdminCustomer }) {
  const status = customer.status ?? '';
  const label = statusLabels[status] ?? status;
  return (
    <span className={`customer-status customer-status--${status}`}>
      {label}
      {status === 'blocked' && customer.blocked_reason ? (
        <span className="customer-status__reason"> — {customer.blocked_reason}</span>
      ) : null}
    </span>
  );
}

function CustomerActions({
  customer,
  onApprove,
  onEdit,
  onUnblock,
}: {
  customer: AdminCustomer;
  onApprove: (id: string) => void;
  onEdit: (c: AdminCustomer) => void;
  onUnblock: (id: string) => void;
}) {
  return (
    <div className="table-actions">
      {customer.status === 'pending' && (
        <button type="button" onClick={() => onApprove(customer.id)}>
          Aprovar
        </button>
      )}
      <button type="button" onClick={() => onEdit(customer)}>
        Editar
      </button>
      {customer.status === 'blocked' && (
        <button type="button" onClick={() => onUnblock(customer.id)}>
          Desbloquear
        </button>
      )}
    </div>
  );
}

function CustomerTableRow({
  customer,
  selected,
  onApprove,
  onEdit,
  onUnblock,
}: {
  customer: AdminCustomer;
  selected: boolean;
  onApprove: (id: string) => void;
  onEdit: (c: AdminCustomer) => void;
  onUnblock: (id: string) => void;
}) {
  return (
    <tr className={selected ? 'row--selected' : undefined}>
      <td data-label="Nome">{customer.name}</td>
      <td data-label="E-mail">
        {customer.email}
        {customer.email_verified === false && (
          <span className="customer-email-hint">E-mail pendente</span>
        )}
      </td>
      <td data-label="Status">
        <CustomerStatus customer={customer} />
      </td>
      <td data-label="Colaborador">{customer.collaborator_category_name || '—'}</td>
      <td data-label="Limite">{formatBRL(customer.credit_limit_cents ?? 0)}</td>
      <td data-label="Exposição">{formatBRL(customer.current_exposure_cents ?? 0)}</td>
      <td>
        <CustomerActions customer={customer} onApprove={onApprove} onEdit={onEdit} onUnblock={onUnblock} />
      </td>
    </tr>
  );
}

function CustomerCard({
  customer,
  selected,
  onApprove,
  onEdit,
  onUnblock,
}: {
  customer: AdminCustomer;
  selected: boolean;
  onApprove: (id: string) => void;
  onEdit: (c: AdminCustomer) => void;
  onUnblock: (id: string) => void;
}) {
  return (
    <li className={`customer-card${selected ? ' customer-card--selected' : ''}`}>
      <p className="customer-card__title">{customer.name}</p>
      <p className="customer-card__meta">
        <span>{customer.email}</span>
        {customer.email_verified === false && (
          <span className="customer-email-hint">E-mail pendente</span>
        )}
      </p>
      <p className="customer-card__meta">
        <CustomerStatus customer={customer} />
      </p>
      <dl className="customer-card__dl">
        <div>
          <dt>Colaborador</dt>
          <dd>{customer.collaborator_category_name || '—'}</dd>
        </div>
        <div>
          <dt>Limite</dt>
          <dd>{formatBRL(customer.credit_limit_cents ?? 0)}</dd>
        </div>
        <div>
          <dt>Exposição</dt>
          <dd>{formatBRL(customer.current_exposure_cents ?? 0)}</dd>
        </div>
      </dl>
      <CustomerActions customer={customer} onApprove={onApprove} onEdit={onEdit} onUnblock={onUnblock} />
    </li>
  );
}

export function CustomersPage() {
  const [items, setItems] = useState<AdminCustomer[]>([]);
  const [categories, setCategories] = useState<CollaboratorCategory[]>([]);
  const [staffRoles, setStaffRoles] = useState<AdminStaffRole[]>([]);
  const [editingCustomer, setEditingCustomer] = useState<AdminCustomer | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({
    name: '',
    phone: '',
    document: '',
    collaborator_category_id: '',
    staff_role_id: '',
    limit_cents: '',
    limit_reason: '',
    block_reason: '',
  });
  const [catForm, setCatForm] = useState(emptyCategoryForm);
  const [editingCatId, setEditingCatId] = useState<string | null>(null);
  const [initialLimitCents, setInitialLimitCents] = useState(0);
  const [approveLimitReais, setApproveLimitReais] = useState('1000');
  const [clientFilter, setClientFilter] = useState('');
  const editPanelRef = useRef<HTMLDivElement>(null);

  const filteredItems = useMemo(() => {
    const q = clientFilter.trim().toLowerCase();
    if (!q) return items;
    return items.filter((c) => {
      const name = (c.name ?? '').toLowerCase();
      const email = (c.email ?? '').toLowerCase();
      return name.includes(q) || email.includes(q);
    });
  }, [items, clientFilter]);

  const load = useCallback(async () => {
    const [custRes, catRes, rolesRes] = await Promise.all([
      api.adminListCustomers(),
      api.adminListCollaboratorCategories(),
      api.adminListStaffRoles(),
    ]);
    setItems((custRes.items ?? []) as AdminCustomer[]);
    setCategories((catRes.items ?? []) as CollaboratorCategory[]);
    setStaffRoles((rolesRes.items ?? []) as AdminStaffRole[]);
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
    setEditingCustomer(c);
    setInitialLimitCents(c.credit_limit_cents ?? 0);
    const currentStaff = (c.staff_roles ?? [])[0] ?? '';
    const roleMatch = staffRoles.find((r) => r.code === currentStaff);
    setForm({
      name: c.name ?? '',
      phone: c.phone ?? '',
      document: c.document ?? '',
      collaborator_category_id: c.collaborator_category_id ?? '',
      staff_role_id: roleMatch?.id ?? '',
      limit_cents: String((c.credit_limit_cents ?? 0) / 100),
      limit_reason: '',
      block_reason: '',
    });
    void api.adminGetCustomer(c.id).then((fresh) => {
      setEditingCustomer(fresh);
      const code = (fresh.staff_roles ?? [])[0] ?? '';
      const match = staffRoles.find((r) => r.code === code);
      setForm((f) => ({ ...f, staff_role_id: match?.id ?? '' }));
    });
    requestAnimationFrame(() => {
      editPanelRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    });
  }

  async function assignStaffRole() {
    if (!editingId || !form.staff_role_id) return;
    const password = window.prompt('Confirme sua senha para atribuir acesso administrativo:');
    if (!password) return;
    setError(null);
    try {
      await api.adminAssignCustomerStaffRole(editingId, { role_id: form.staff_role_id, password });
      await load();
      const fresh = await api.adminGetCustomer(editingId);
      setEditingCustomer(fresh);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao atribuir papel');
    }
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

      <details className="collaborator-categories-bar">
        <summary className="collaborator-categories-bar__summary">Categorias de colaboradores</summary>
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
            <label className="collaborator-cat-form__checkbox form__checkbox">
              <input
                type="checkbox"
                checked={catForm.active}
                onChange={(e) => setCatForm((f) => ({ ...f, active: e.target.checked }))}
              />
              Categoria ativa
            </label>
          )}
          <div className="collaborator-cat-form__actions form__full">
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
          </div>
        </form>
        <ul className="collaborator-cat-list">
          {categories.map((cat) => (
            <li key={cat.id}>
              <span className="collaborator-cat-list__label">
                <strong>{cat.name}</strong> — {cat.margin_percent}%
                {!cat.active && <span className="collaborator-cat-list__inactive"> (inativa)</span>}
              </span>
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
      </details>

      <div className="customers-layout">
        <div className="customers-panel">
          <div className="customers-toolbar">
            <h2 className="customers-panel__title">Cadastros</h2>
            <label className="customers-toolbar__limit">
              Limite na aprovação (R$)
              <input
                value={approveLimitReais}
                onChange={(e) => setApproveLimitReais(e.target.value)}
                inputMode="decimal"
                aria-label="Limite na aprovação em reais"
              />
            </label>
            <label className="customers-toolbar__search">
              Buscar cliente
              <input
                type="search"
                value={clientFilter}
                onChange={(e) => setClientFilter(e.target.value)}
                placeholder="Nome ou e-mail"
                aria-label="Filtrar por nome ou e-mail"
              />
            </label>
          </div>

          {filteredItems.length === 0 ? (
            <p className="customers-panel__placeholder">
              {items.length === 0
                ? 'Nenhum cliente cadastrado.'
                : 'Nenhum cliente corresponde à busca.'}
            </p>
          ) : (
            <>
          <div className="table-scroll customers-table-desktop">
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
                {filteredItems.map((c) => (
                  <CustomerTableRow
                    key={c.id}
                    customer={c}
                    selected={editingId === c.id}
                    onApprove={(id) => void approve(id)}
                    onEdit={startEdit}
                    onUnblock={(id) => void unblockCustomer(id)}
                  />
                ))}
              </tbody>
            </table>
          </div>

          <ul className="customers-cards" aria-label="Lista de clientes">
            {filteredItems.map((c) => (
              <CustomerCard
                key={c.id}
                customer={c}
                selected={editingId === c.id}
                onApprove={(id) => void approve(id)}
                onEdit={startEdit}
                onUnblock={(id) => void unblockCustomer(id)}
              />
            ))}
          </ul>
            </>
          )}
        </div>

        <div ref={editPanelRef} className="customers-panel customers-panel--form">
          <h2 className="customers-panel__title">Editar cliente</h2>
          {editingId ? (
            <form className="form form--wide customer-edit-form" onSubmit={saveCustomer}>
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
              <label className="form__full">
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
              <fieldset className="customer-staff-fieldset form__full">
                <legend>Acesso administrativo</legend>
                <p className="form-hint">
                  Funcionários devem ter cadastro na loja. O papel interno não remove o perfil de cliente.
                </p>
                {(editingCustomer?.staff_roles?.length ?? 0) > 0 && (
                  <p>
                    Papel atual: <strong>{(editingCustomer?.staff_roles ?? []).join(', ')}</strong>
                  </p>
                )}
                <label>
                  Papel interno
                  <select
                    value={form.staff_role_id}
                    onChange={(e) => setForm((f) => ({ ...f, staff_role_id: e.target.value }))}
                  >
                    <option value="">— Sem acesso ao painel —</option>
                    {staffRoles.map((r) => (
                      <option key={r.id} value={r.id}>
                        {r.name} ({r.code})
                      </option>
                    ))}
                  </select>
                </label>
                <button type="button" disabled={!form.staff_role_id} onClick={() => void assignStaffRole()}>
                  Atribuir / alterar papel
                </button>
              </fieldset>
              <div className="form__actions form__full">
                <button type="submit">Salvar</button>
                <button type="button" onClick={() => setEditingId(null)}>
                  Cancelar
                </button>
              </div>
              <fieldset className="customer-block-fieldset form__full">
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
          ) : (
            <p className="customers-panel__placeholder">
              Selecione <strong>Editar</strong> em um cliente da lista para alterar dados, limite ou bloqueio.
            </p>
          )}
        </div>
      </div>
    </section>
  );
}
