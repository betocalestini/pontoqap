import { useCallback, useEffect, useState } from 'react';
import type { AuditLogEntry } from '@store/api-client';
import { api } from '../api';

export function AuditPage() {
  const [items, setItems] = useState<AuditLogEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [action, setAction] = useState('');
  const [entityType, setEntityType] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.adminListAuditLogs({
        action: action || undefined,
        entity_type: entityType || undefined,
        limit: 100,
      });
      setItems(res.items ?? []);
      setTotal(res.total ?? 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao carregar auditoria');
    } finally {
      setLoading(false);
    }
  }, [action, entityType]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <section className="content-section">
      <h1>Auditoria</h1>
      <p className="form-hint">Registros administrativos e operações sensíveis (somente administrador).</p>
      {error && <p className="error">{error}</p>}
      <div className="form form--wide customers-list-filters">
        <label>
          Ação
          <input value={action} onChange={(e) => setAction(e.target.value)} placeholder="ex.: order.cancelled" />
        </label>
        <label>
          Entidade
          <input value={entityType} onChange={(e) => setEntityType(e.target.value)} placeholder="ex.: order" />
        </label>
        <button type="button" onClick={() => void load()}>
          Filtrar
        </button>
      </div>
      {loading ? (
        <p>Carregando…</p>
      ) : (
        <>
          <p className="form-hint">{total} registro(s)</p>
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Quando</th>
                  <th>Quem</th>
                  <th>Ação</th>
                  <th>Entidade</th>
                </tr>
              </thead>
              <tbody>
                {items.map((row) => (
                  <tr key={row.id}>
                    <td>{new Date(row.created_at).toLocaleString('pt-BR')}</td>
                    <td>{row.actor_email ?? row.actor_user_id ?? '—'}</td>
                    <td>
                      <code>{row.action}</code>
                    </td>
                    <td>
                      {row.entity_type}
                      {row.entity_id ? ` / ${row.entity_id.slice(0, 8)}…` : ''}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </section>
  );
}
