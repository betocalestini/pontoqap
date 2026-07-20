import { useCallback, useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { AdminInventoryBalance, AdminInventoryMovement } from '@store/api-client';
import { api } from '../api';

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

export function InventoryPage() {
  const [searchParams] = useSearchParams();
  const filterSku = searchParams.get('sku_id') ?? '';

  const [balances, setBalances] = useState<AdminInventoryBalance[]>([]);
  const [movements, setMovements] = useState<AdminInventoryMovement[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [skuSearch, setSkuSearch] = useState('');
  const [form, setForm] = useState({
    sku_id: filterSku,
    kind: 'entry' as MovementKind,
    quantity: '',
    physical_count: '',
    reason: '',
  });
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    const [balRes, movRes] = await Promise.all([
      api.adminInventoryBalances(),
      api.adminInventoryMovements({
        sku_id: filterSku || undefined,
        limit: 100,
      }),
    ]);
    setBalances(balRes.items ?? []);
    setMovements(movRes.items ?? []);
  }, [filterSku]);

  useEffect(() => {
    setForm((f) => ({ ...f, sku_id: filterSku || f.sku_id }));
    load().catch((e: Error) => setError(e.message));
  }, [load, filterSku]);

  const skuOptions = useMemo(() => {
    const q = skuSearch.trim().toLowerCase();
    return balances.filter(
      (b) =>
        !q ||
        b.product_name.toLowerCase().includes(q) ||
        b.sku_code.toLowerCase().includes(q),
    );
  }, [balances, skuSearch]);

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
      setForm((f) => ({ ...f, quantity: '', physical_count: '', reason: '' }));
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro');
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="content-section">
      <h1>Estoque</h1>
      {error && <p className="error">{error}</p>}

      <h2>Saldos atuais</h2>
      <div className="table-scroll">
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
            {balances.map((b) => {
              const low = b.available_quantity < b.minimum_stock;
              return (
                <tr key={b.sku_id} className={low ? 'error' : undefined}>
                  <td>{b.product_name}</td>
                  <td>{b.sku_code}</td>
                  <td>{b.available_quantity}</td>
                  <td>{b.minimum_stock}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <h2>Registrar movimentação</h2>
      <form onSubmit={submitMovement} className="form">
        <label>
          Buscar produto / SKU
          <input value={skuSearch} onChange={(e) => setSkuSearch(e.target.value)} placeholder="Filtrar lista" />
        </label>
        <label>
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
        <label>
          Motivo
          <input value={form.reason} onChange={(e) => setForm((f) => ({ ...f, reason: e.target.value }))} required />
        </label>
        <button type="submit" disabled={saving}>
          {saving ? 'Registrando…' : 'Registrar'}
        </button>
      </form>

      <h2>Histórico{filterSku ? ' (SKU filtrado)' : ''}</h2>
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Data</th>
              <th>Tipo</th>
              <th>Produto</th>
              <th>Qtd</th>
              <th>Ant. → Novo</th>
              <th>Referência</th>
              <th>Responsável</th>
              <th>Motivo</th>
            </tr>
          </thead>
          <tbody>
            {movements.map((m) => (
              <tr key={m.id}>
                <td>{new Date(m.created_at).toLocaleString('pt-BR')}</td>
                <td>{movementLabels[m.movement_type] ?? m.movement_type}</td>
                <td>
                  {m.product_name ?? '—'}
                  {m.sku_code ? ` (${m.sku_code})` : ''}
                </td>
                <td>{m.quantity > 0 ? `+${m.quantity}` : m.quantity}</td>
                <td>
                  {m.previous_balance} → {m.new_balance}
                </td>
                <td>
                  {m.movement_type === 'sale' && m.reference_id ? `Pedido ${m.reference_id.slice(0, 8)}…` : m.reference_type ?? '—'}
                </td>
                <td>{m.created_by_email ?? '—'}</td>
                <td>{m.reason ?? '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
