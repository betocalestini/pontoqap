import { useCallback, useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { AdminInventoryBalance, AdminInventoryMovement } from '@store/api-client';
import { api } from '../api';

const MOVEMENT_PAGE_SIZE = 50;

const movementLabels: Record<string, string> = {
  entry: 'Entrada',
  initial_stock: 'Estoque inicial',
  sale: 'Venda',
  loss: 'Perda',
  damage: 'Avaria',
  adjustment: 'Ajuste',
  return: 'Devolução',
};

type MovementKind = 'entry' | 'loss' | 'damage' | 'adjustment';

function formatBRL(cents: number) {
  return (cents / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

function parseReaisToCents(value: string): number | null {
  const normalized = value.trim().replace(/\./g, '').replace(',', '.');
  if (!normalized) return null;
  const n = Number.parseFloat(normalized);
  if (!Number.isFinite(n) || n < 0) return null;
  return Math.round(n * 100);
}

export function InventoryPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const filterProductId = searchParams.get('product_id') ?? '';

  const [balances, setBalances] = useState<AdminInventoryBalance[]>([]);
  const [movements, setMovements] = useState<AdminInventoryMovement[]>([]);
  const [movementsTotal, setMovementsTotal] = useState(0);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [productFilterSearch, setProductFilterSearch] = useState('');
  const [skuSearch, setSkuSearch] = useState('');
  const [form, setForm] = useState({
    sku_id: '',
    kind: 'entry' as MovementKind,
    quantity: '',
    physical_count: '',
    unit_cost_reais: '',
    reason: '',
  });
  const [saving, setSaving] = useState(false);

  const products = useMemo(() => {
    const map = new Map<string, string>();
    for (const b of balances) {
      map.set(b.product_id, b.product_name);
    }
    return [...map.entries()]
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name, 'pt-BR'));
  }, [balances]);

  const filteredProducts = useMemo(() => {
    const q = productFilterSearch.trim().toLowerCase();
    if (!q) return products;
    return products.filter((p) => p.name.toLowerCase().includes(q));
  }, [products, productFilterSearch]);

  const displayedBalances = useMemo(() => {
    if (!filterProductId) return balances;
    return balances.filter((b) => b.product_id === filterProductId);
  }, [balances, filterProductId]);

  const loadBalances = useCallback(async () => {
    const balRes = await api.adminInventoryBalances();
    setBalances(balRes.items ?? []);
  }, []);

  const loadMovements = useCallback(
    async (offset: number, append: boolean) => {
      const movRes = await api.adminInventoryMovements({
        product_id: filterProductId || undefined,
        limit: MOVEMENT_PAGE_SIZE,
        offset,
      });
      setMovementsTotal(movRes.total ?? 0);
      setMovements((prev) => (append ? [...prev, ...(movRes.items ?? [])] : movRes.items ?? []));
    },
    [filterProductId],
  );

  useEffect(() => {
    loadBalances().catch((e: Error) => setError(e.message));
  }, [loadBalances]);

  useEffect(() => {
    loadMovements(0, false).catch((e: Error) => setError(e.message));
  }, [loadMovements]);

  const skuOptions = useMemo(() => {
    const q = skuSearch.trim().toLowerCase();
    const base = filterProductId
      ? balances.filter((b) => b.product_id === filterProductId)
      : balances;
    return base.filter(
      (b) =>
        !q ||
        b.product_name.toLowerCase().includes(q) ||
        b.sku_code.toLowerCase().includes(q),
    );
  }, [balances, skuSearch, filterProductId]);

  function setProductFilter(productId: string) {
    const next = new URLSearchParams(searchParams);
    if (productId) next.set('product_id', productId);
    else next.delete('product_id');
    next.delete('sku_id');
    setSearchParams(next, { replace: true });
    setForm((f) => ({
      ...f,
      sku_id: productId && f.sku_id && balances.find((b) => b.sku_id === f.sku_id)?.product_id !== productId ? '' : f.sku_id,
    }));
  }

  async function submitMovement(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      if (!form.sku_id) throw new Error('Selecione um SKU');
      if (!form.reason.trim()) throw new Error('Informe o motivo');
      if (form.kind === 'adjustment') {
        await api.adminCreateInventoryMovement({
          kind: 'adjustment',
          sku_id: form.sku_id,
          physical_count: parseInt(form.physical_count, 10),
          reason: form.reason.trim(),
        });
      } else if (form.kind === 'entry') {
        const qty = parseInt(form.quantity, 10);
        if (!qty || qty <= 0) throw new Error('Quantidade inválida');
        const unitCostCents = parseReaisToCents(form.unit_cost_reais);
        if (unitCostCents == null) throw new Error('Informe o preço unitário pago (R$)');
        await api.adminCreateInventoryMovement({
          kind: 'entry',
          sku_id: form.sku_id,
          quantity: qty,
          reason: form.reason.trim(),
          unit_cost_cents: unitCostCents,
        });
      } else {
        const qty = parseInt(form.quantity, 10);
        if (!qty || qty <= 0) throw new Error('Quantidade inválida');
        await api.adminCreateInventoryMovement({
          kind: form.kind,
          sku_id: form.sku_id,
          quantity: qty,
          reason: form.reason.trim(),
        });
      }
      setForm((f) => ({
        ...f,
        quantity: '',
        physical_count: '',
        unit_cost_reais: '',
        reason: '',
      }));
      await Promise.all([loadBalances(), loadMovements(0, false)]);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro');
    } finally {
      setSaving(false);
    }
  }

  async function onLoadMore() {
    setLoadingMore(true);
    setError(null);
    try {
      await loadMovements(movements.length, true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao carregar');
    } finally {
      setLoadingMore(false);
    }
  }

  const entriesCostOnPage = useMemo(() => {
    return movements.reduce((sum, m) => {
      if ((m.movement_type === 'entry' || m.movement_type === 'initial_stock') && m.unit_cost_cents != null) {
        return sum + m.unit_cost_cents * m.quantity;
      }
      return sum;
    }, 0);
  }, [movements]);

  const hasMoreMovements = movements.length < movementsTotal;

  return (
    <section className="content-section inventory-page">
      <h1>Estoque</h1>
      {error && <p className="error">{error}</p>}

      <div className="inventory-filter-bar">
        <label className="inventory-filter-bar__field">
          Produto (saldos e histórico)
          <select
            value={filterProductId}
            onChange={(e) => setProductFilter(e.target.value)}
          >
            <option value="">Todos os produtos</option>
            {filteredProducts.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </label>
        <label className="inventory-filter-bar__field inventory-filter-bar__search">
          Buscar na lista
          <input
            value={productFilterSearch}
            onChange={(e) => setProductFilterSearch(e.target.value)}
            placeholder="Nome do produto"
          />
        </label>
        {filterProductId && (
          <button type="button" className="inventory-filter-bar__clear" onClick={() => setProductFilter('')}>
            Limpar filtro
          </button>
        )}
      </div>

      <div className="inventory-layout">
        <div className="inventory-panel">
          <h2>Saldos atuais{filterProductId ? ' (filtrado)' : ''}</h2>
          <div className="table-scroll inventory-balances-table">
            <table>
              <thead>
                <tr>
                  <th>Produto</th>
                  <th>SKU</th>
                  <th>Saldo</th>
                  <th>Mínimo</th>
                </tr>
              </thead>
              <tbody>
                {displayedBalances.map((b) => {
                  const low = b.available_quantity < b.minimum_stock;
                  return (
                    <tr key={b.sku_id} className={low ? 'row--warning' : undefined}>
                      <td data-label="Produto">{b.product_name}</td>
                      <td data-label="SKU">{b.sku_code}</td>
                      <td data-label="Saldo">{b.available_quantity}</td>
                      <td data-label="Mínimo">{b.minimum_stock}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
          <ul className="inventory-balances-cards" aria-label="Saldos atuais">
            {displayedBalances.map((b) => {
              const low = b.available_quantity < b.minimum_stock;
              return (
                <li key={b.sku_id} className={low ? 'inventory-balance-card row--warning' : 'inventory-balance-card'}>
                  <p className="inventory-balance-card__title">{b.product_name}</p>
                  <p className="inventory-balance-card__meta">
                    <span>{b.sku_code}</span>
                    <span>
                      Saldo: <strong>{b.available_quantity}</strong>
                      {b.minimum_stock > 0 ? ` · Mín.: ${b.minimum_stock}` : ''}
                    </span>
                  </p>
                </li>
              );
            })}
          </ul>
        </div>

        <div className="inventory-panel inventory-panel--form">
          <h2>Registrar movimentação</h2>
          <form onSubmit={submitMovement} className="form form--wide inventory-form">
            <label className="form__full">
              Buscar produto / SKU
              <input value={skuSearch} onChange={(e) => setSkuSearch(e.target.value)} placeholder="Filtrar lista" />
            </label>
            <label className="form__full">
              SKU
              <select
                value={form.sku_id}
                onChange={(e) => setForm((f) => ({ ...f, sku_id: e.target.value }))}
                required
              >
                <option value="">Selecione…</option>
                {skuOptions.map((b) => (
                  <option key={b.sku_id} value={b.sku_id}>
                    {b.product_name} — {b.sku_code} (saldo {b.available_quantity})
                  </option>
                ))}
              </select>
            </label>
            <label>
              Tipo
              <select
                value={form.kind}
                onChange={(e) => setForm((f) => ({ ...f, kind: e.target.value as MovementKind }))}
              >
                <option value="entry">Entrada</option>
                <option value="loss">Perda</option>
                <option value="damage">Avaria</option>
                <option value="adjustment">Ajuste (contagem física)</option>
              </select>
            </label>
            {form.kind === 'adjustment' ? (
              <label>
                Contagem física
                <input
                  type="number"
                  min={0}
                  value={form.physical_count}
                  onChange={(e) => setForm((f) => ({ ...f, physical_count: e.target.value }))}
                  required
                />
              </label>
            ) : (
              <label>
                Quantidade
                <input
                  type="number"
                  min={1}
                  value={form.quantity}
                  onChange={(e) => setForm((f) => ({ ...f, quantity: e.target.value }))}
                  required
                />
              </label>
            )}
            {form.kind === 'entry' && (
              <label className="form__full">
                Preço unitário pago (R$)
                <input
                  inputMode="decimal"
                  placeholder="Ex.: 12,50"
                  value={form.unit_cost_reais}
                  onChange={(e) => setForm((f) => ({ ...f, unit_cost_reais: e.target.value }))}
                  required
                />
                <small>Usado no custo do SKU e no relatório financeiro de compras.</small>
              </label>
            )}
            <label className="form__full">
              Motivo
              <input value={form.reason} onChange={(e) => setForm((f) => ({ ...f, reason: e.target.value }))} required />
            </label>
            <div className="form__actions form__full">
              <button type="submit" disabled={saving}>
                {saving ? 'Registrando…' : 'Registrar'}
              </button>
            </div>
          </form>
        </div>
      </div>

      <div className="inventory-panel inventory-panel--history">
        <div className="inventory-history-header">
          <h2>Histórico</h2>
          <p className="inventory-history-meta">
            {movementsTotal} movimentação(ões)
            {filterProductId ? ' neste produto' : ' no total'}
            {entriesCostOnPage > 0 && (
              <>
                {' '}
                · Custo de entradas nesta página: <strong>{formatBRL(entriesCostOnPage)}</strong>
              </>
            )}
          </p>
        </div>
        <div className="table-scroll">
          <table className="inventory-history-table">
            <thead>
              <tr>
                <th>Data</th>
                <th>Tipo</th>
                <th>Produto</th>
                <th>Qtd</th>
                <th>Custo un.</th>
                <th>Total compra</th>
                <th>Ant. → Novo</th>
                <th>Referência</th>
                <th>Responsável</th>
                <th>Motivo</th>
              </tr>
            </thead>
            <tbody>
              {movements.map((m) => {
                const isPurchase = m.movement_type === 'entry' || m.movement_type === 'initial_stock';
                const lineTotal =
                  isPurchase && m.unit_cost_cents != null ? m.unit_cost_cents * m.quantity : null;
                return (
                  <tr key={m.id}>
                    <td>{new Date(m.created_at).toLocaleString('pt-BR')}</td>
                    <td>{movementLabels[m.movement_type] ?? m.movement_type}</td>
                    <td>
                      {m.product_name ?? '—'}
                      {m.sku_code ? ` (${m.sku_code})` : ''}
                    </td>
                    <td>{m.quantity > 0 ? `+${m.quantity}` : m.quantity}</td>
                    <td>{m.unit_cost_cents != null ? formatBRL(m.unit_cost_cents) : '—'}</td>
                    <td>{lineTotal != null ? formatBRL(lineTotal) : '—'}</td>
                    <td>
                      {m.previous_balance} → {m.new_balance}
                    </td>
                    <td>
                      {m.movement_type === 'sale' && m.reference_id
                        ? `Pedido ${m.reference_id.slice(0, 8)}…`
                        : m.reference_type ?? '—'}
                    </td>
                    <td>{m.created_by_email ?? '—'}</td>
                    <td>{m.reason ?? '—'}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
        {hasMoreMovements && (
          <div className="inventory-load-more">
            <button type="button" disabled={loadingMore} onClick={() => void onLoadMore()}>
              {loadingMore ? 'Carregando…' : `Carregar mais (${movements.length} de ${movementsTotal})`}
            </button>
          </div>
        )}
      </div>
    </section>
  );
}
