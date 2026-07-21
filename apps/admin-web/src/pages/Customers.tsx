import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { AdminCustomer, AdminStaffRole, CollaboratorCategory } from '@store/api-client';
import { api } from '../api';
import { useHasPermission } from '../auth/usePermissions';

function formatBRL(cents: number) {
  return (cents / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

function availableCreditCents(c: AdminCustomer) {
  return Math.max(0, (c.credit_limit_cents ?? 0) - (c.current_exposure_cents ?? 0));
}

function limitUsagePercent(c: AdminCustomer) {
  const limit = c.credit_limit_cents ?? 0;
  const exposure = c.current_exposure_cents ?? 0;
  if (limit <= 0) {
    return exposure > 0 ? 100 : 0;
  }
  return Math.min(999, Math.round((exposure / limit) * 100));
}

type CustomerSortKey = 'exposure' | 'open_invoices';

type ColumnFilters = {
  search: string;
  status: string;
  collaborator: string;
  billing: string;
};

const defaultColumnFilters = (): ColumnFilters => ({
  search: '',
  status: '',
  collaborator: '',
  billing: '',
});

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
  const usage = limitUsagePercent(customer);
  const warn =
    usage >= 80 || (customer.current_exposure_cents ?? 0) > (customer.credit_limit_cents ?? 0);
  const overdue = customer.overdue_invoices_count ?? 0;
  const open = customer.open_invoices_count ?? 0;

  return (
    <tr className={selected ? 'row--selected' : undefined}>
      <td data-label="Cliente">
        <div className="customer-cell-identity">
          <span className="customer-cell-identity__name">{customer.name}</span>
          <span className="customer-cell-identity__email">{customer.email}</span>
          {customer.document?.trim() ? (
            <span className="customer-cell-identity__meta">{customer.document}</span>
          ) : null}
          {customer.email_verified === false ? (
            <span className="customer-email-hint">E-mail não confirmado</span>
          ) : null}
        </div>
      </td>
      <td data-label="Situação">
        <div className="customer-cell-stack">
          <CustomerStatus customer={customer} />
          {customer.collaborator_category_name ? (
            <span className="customer-collab-tag">{customer.collaborator_category_name}</span>
          ) : null}
        </div>
      </td>
      <td data-label="Crédito">
        <dl className="customer-credit-grid">
          <div>
            <dt>Limite</dt>
            <dd>{formatBRL(customer.credit_limit_cents ?? 0)}</dd>
          </div>
          <div>
            <dt>Disponível</dt>
            <dd>{formatBRL(availableCreditCents(customer))}</dd>
          </div>
          <div>
            <dt>Exposição</dt>
            <dd>{formatBRL(customer.current_exposure_cents ?? 0)}</dd>
          </div>
          <div>
            <dt>Uso</dt>
            <dd className={warn ? 'customer-credit--warn' : undefined}>{usage}%</dd>
          </div>
        </dl>
      </td>
      <td data-label="Faturas">
        <dl className="customer-billing-grid">
          <div>
            <dt>Aberto</dt>
            <dd>{open}</dd>
          </div>
          <div>
            <dt>Atraso</dt>
            <dd className={overdue > 0 ? 'customer-overdue--highlight' : undefined}>{overdue}</dd>
          </div>
        </dl>
      </td>
      <td className="customer-cell-actions">
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
          <dt>Documento</dt>
          <dd>{customer.document?.trim() || '—'}</dd>
        </div>
        <div>
          <dt>Disponível</dt>
          <dd>{formatBRL(availableCreditCents(customer))}</dd>
        </div>
        <div>
          <dt>Uso do limite</dt>
          <dd>{limitUsagePercent(customer)}%</dd>
        </div>
        <div>
          <dt>Meses em aberto</dt>
          <dd>{customer.open_invoices_count ?? 0}</dd>
        </div>
        <div>
          <dt>Em atraso</dt>
          <dd>{customer.overdue_invoices_count ?? 0}</dd>
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
  const canManageUsers = useHasPermission('users.manage');
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
  const [columnFilters, setColumnFilters] = useState<ColumnFilters>(defaultColumnFilters);
  const [sortKey, setSortKey] = useState<CustomerSortKey | null>(null);
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');
  const editPanelRef = useRef<HTMLDivElement>(null);

  const filteredItems = useMemo(() => {
    const q = columnFilters.search.trim().toLowerCase();
    return items.filter((c) => {
      if (columnFilters.status && c.status !== columnFilters.status) return false;
      if (columnFilters.collaborator === '__none__') {
        if (c.collaborator_category_id) return false;
      } else if (columnFilters.collaborator && c.collaborator_category_id !== columnFilters.collaborator) {
        return false;
      }
      const open = c.open_invoices_count ?? 0;
      const overdue = c.overdue_invoices_count ?? 0;
      if (columnFilters.billing === 'overdue' && overdue < 1) return false;
      if (columnFilters.billing === 'open' && open < 1) return false;
      if (columnFilters.billing === 'ok' && (open > 0 || overdue > 0)) return false;
      if (!q) return true;
      const hay = [c.name, c.email, c.document, c.collaborator_category_name]
        .filter(Boolean)
        .join(' ')
        .toLowerCase();
      return hay.includes(q);
    });
  }, [items, columnFilters]);

  const sortedItems = useMemo(() => {
    if (!sortKey) return filteredItems;
    const dir = sortDir === 'asc' ? 1 : -1;
    return [...filteredItems].sort((a, b) => {
      const av =
        sortKey === 'exposure' ? (a.current_exposure_cents ?? 0) : (a.open_invoices_count ?? 0);
      const bv =
        sortKey === 'exposure' ? (b.current_exposure_cents ?? 0) : (b.open_invoices_count ?? 0);
      return (av - bv) * dir;
    });
  }, [filteredItems, sortKey, sortDir]);

  function toggleSort(key: CustomerSortKey) {
    if (sortKey === key) {
      setSortDir((d) => (d === 'desc' ? 'asc' : 'desc'));
    } else {
      setSortKey(key);
      setSortDir('desc');
    }
  }

  const load = useCallback(async () => {
    if (canManageUsers) {
      const [custRes, catRes, rolesRes] = await Promise.all([
        api.adminListCustomers(),
        api.adminListCollaboratorCategories(),
        api.adminListStaffRoles(),
      ]);
      setItems((custRes.items ?? []) as AdminCustomer[]);
      setCategories((catRes.items ?? []) as CollaboratorCategory[]);
      setStaffRoles((rolesRes.items ?? []) as AdminStaffRole[]);
      return;
    }
    const [custRes, catRes] = await Promise.all([api.adminListCustomers(), api.adminListCollaboratorCategories()]);
    setItems((custRes.items ?? []) as AdminCustomer[]);
    setCategories((catRes.items ?? []) as CollaboratorCategory[]);
    setStaffRoles([]);
  }, [canManageUsers]);

  useEffect(() => {
    load().catch((e: Error) => setError(e.message));
  }, [load]);

  async function approve(id: string) {
    setError(null);
    const existing = items.find((c) => c.id === id);
    const defaultReais =
      existing && (existing.credit_limit_cents ?? 0) > 0
        ? String((existing.credit_limit_cents ?? 0) / 100)
        : '1000';
    const raw = window.prompt(
      'Limite de crédito (R$) para aprovar este cadastro pendente.\nCadastros que confirmam o e-mail na loja costumam ser aprovados automaticamente.',
      defaultReais,
    );
    if (raw === null) return;
    const limitCents = Math.round(parseFloat(raw.replace(',', '.')) * 100);
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
      editPanelRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    });
  }

  function closeEdit() {
    setEditingId(null);
    setEditingCustomer(null);
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
      closeEdit();
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

      <div className="customers-main">
        {editingId ? (
          <div ref={editPanelRef} className="customers-edit-band">
            <div className="customers-edit-band__header">
              <h2>Editando: {form.name || editingCustomer?.name || 'Cliente'}</h2>
              <button type="button" onClick={closeEdit}>
                Fechar
              </button>
            </div>
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
                {canManageUsers ? (
                  <>
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
                  </>
                ) : (
                  <p className="form-hint">
                    Atribuição de papel interno é feita por um administrador (menu Usuários ou permissão{' '}
                    <code>users.manage</code>).
                  </p>
                )}
              </fieldset>
              <div className="form__actions form__full">
                <button type="submit">Salvar</button>
                <button type="button" onClick={closeEdit}>
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
          </div>
        ) : null}

        <div className="customers-panel customers-panel--cadastros">
          <h2 className="customers-panel__title">Cadastros</h2>

          <div className="customers-list-filters" role="search">
            <label className="customers-list-filters__field customers-list-filters__field--search">
              <span className="customers-list-filters__label">Buscar</span>
              <input
                type="search"
                className="customers-filter-input"
                placeholder="Nome, e-mail, documento…"
                value={columnFilters.search}
                onChange={(e) => setColumnFilters((f) => ({ ...f, search: e.target.value }))}
              />
            </label>
            <label className="customers-list-filters__field">
              <span className="customers-list-filters__label">Situação</span>
              <select
                className="customers-filter-input"
                value={columnFilters.status}
                onChange={(e) => setColumnFilters((f) => ({ ...f, status: e.target.value }))}
              >
                <option value="">Todas</option>
                <option value="pending">Pendente</option>
                <option value="approved">Aprovado</option>
                <option value="blocked">Bloqueado</option>
                <option value="rejected">Rejeitado</option>
              </select>
            </label>
            <label className="customers-list-filters__field">
              <span className="customers-list-filters__label">Colaborador</span>
              <select
                className="customers-filter-input"
                value={columnFilters.collaborator}
                onChange={(e) => setColumnFilters((f) => ({ ...f, collaborator: e.target.value }))}
              >
                <option value="">Todos</option>
                <option value="__none__">Sem categoria</option>
                {categories.map((cat) => (
                  <option key={cat.id} value={cat.id}>
                    {cat.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="customers-list-filters__field">
              <span className="customers-list-filters__label">Faturas</span>
              <select
                className="customers-filter-input"
                value={columnFilters.billing}
                onChange={(e) => setColumnFilters((f) => ({ ...f, billing: e.target.value }))}
              >
                <option value="">Todas</option>
                <option value="overdue">Com atraso</option>
                <option value="open">Com meses em aberto</option>
                <option value="ok">Em dia</option>
              </select>
            </label>
          </div>

          {sortedItems.length === 0 ? (
            <p className="customers-panel__placeholder">
              {items.length === 0
                ? 'Nenhum cliente cadastrado.'
                : 'Nenhum cliente corresponde à busca.'}
            </p>
          ) : (
            <>
              <div className="customers-table-wrap">
                <table className="customers-table customers-table--compact">
                  <thead>
                    <tr>
                      <th>Cliente</th>
                      <th>Situação</th>
                      <th
                        className="sortable"
                        onClick={() => toggleSort('exposure')}
                        title="Ordenar por exposição de crédito"
                      >
                        Crédito{sortKey === 'exposure' ? (sortDir === 'desc' ? ' ↓' : ' ↑') : ''}
                      </th>
                      <th
                        className="sortable"
                        onClick={() => toggleSort('open_invoices')}
                        title="Ordenar por meses em aberto"
                      >
                        Faturas{sortKey === 'open_invoices' ? (sortDir === 'desc' ? ' ↓' : ' ↑') : ''}
                      </th>
                      <th className="customer-cell-actions">Ações</th>
                    </tr>
                  </thead>
                  <tbody>
                    {sortedItems.map((c) => (
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

              <ul className="customers-cards customers-cards--filters" aria-label="Filtros mobile">
                <li className="customer-card customer-card--filters">
                  <label>
                    Buscar
                    <input
                      type="search"
                      value={columnFilters.search}
                      onChange={(e) => setColumnFilters((f) => ({ ...f, search: e.target.value }))}
                    />
                  </label>
                  <label>
                    Status
                    <select
                      value={columnFilters.status}
                      onChange={(e) => setColumnFilters((f) => ({ ...f, status: e.target.value }))}
                    >
                      <option value="">Todos</option>
                      <option value="pending">Pendente</option>
                      <option value="approved">Aprovado</option>
                      <option value="blocked">Bloqueado</option>
                    </select>
                  </label>
                  <label>
                    Faturas
                    <select
                      value={columnFilters.billing}
                      onChange={(e) => setColumnFilters((f) => ({ ...f, billing: e.target.value }))}
                    >
                      <option value="">Todas</option>
                      <option value="overdue">Com atraso</option>
                      <option value="open">Em aberto</option>
                      <option value="ok">Em dia</option>
                    </select>
                  </label>
                </li>
              </ul>

              <ul className="customers-cards" aria-label="Lista de clientes">
                {sortedItems.map((c) => (
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
      </div>
    </section>
  );
}
