import { type FormEvent, useCallback, useEffect, useState } from 'react';
import type { AdminStaffRole, AdminStaffUser } from '@store/api-client';
import { api } from '../api';

export function UsersPage() {
  const [users, setUsers] = useState<AdminStaffUser[]>([]);
  const [roles, setRoles] = useState<AdminStaffRole[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteName, setInviteName] = useState('');
  const [inviteRoleId, setInviteRoleId] = useState('');
  const [inviteMsg, setInviteMsg] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [u, r] = await Promise.all([api.adminListStaffUsers(), api.adminListStaffRoles()]);
      setUsers(u.items ?? []);
      setRoles(r.items ?? []);
      if (!inviteRoleId && r.items?.length) {
        const mgr = r.items.find((x) => x.code === 'manager');
        setInviteRoleId(mgr?.id ?? r.items[0].id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao carregar');
    } finally {
      setLoading(false);
    }
  }, [inviteRoleId]);

  useEffect(() => {
    void load();
  }, [load]);

  async function onInvite(e: FormEvent) {
    e.preventDefault();
    setInviteMsg(null);
    setError(null);
    try {
      await api.adminCreateStaffInvitation({
        email: inviteEmail,
        name: inviteName,
        role_id: inviteRoleId,
      });
      setInviteMsg('Convite enviado por e-mail.');
      setInviteEmail('');
      setInviteName('');
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao convidar');
    }
  }

  async function setStatus(user: AdminStaffUser, status: string) {
    const password = window.prompt('Confirme sua senha para esta ação:');
    if (!password) return;
    try {
      await api.adminSetStaffUserStatus(user.id, { status, password });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao alterar status');
    }
  }

  return (
    <section className="content-section">
      <h1>Usuários internos</h1>
      <p className="form-hint">
        Novos funcionários devem se cadastrar na loja como clientes. Atribua papéis administrativos em{' '}
        <strong>Clientes → Editar → Acesso administrativo</strong>. Suspender aqui bloqueia só o painel; a conta de
        cliente na loja continua ativa.
      </p>
      {error && <p className="error">{error}</p>}
      {inviteMsg && <p className="success">{inviteMsg}</p>}

      <h2>Convidar por e-mail (cliente já cadastrado)</h2>
      <form onSubmit={onInvite} className="form form--wide">
        <label>
          Nome
          <input value={inviteName} onChange={(e) => setInviteName(e.target.value)} required />
        </label>
        <label>
          E-mail
          <input type="email" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} required />
        </label>
        <label>
          Papel
          <select value={inviteRoleId} onChange={(e) => setInviteRoleId(e.target.value)} required>
            {roles.map((r) => (
              <option key={r.id} value={r.id}>
                {r.name} ({r.code})
              </option>
            ))}
          </select>
        </label>
        <button type="submit">Enviar convite</button>
      </form>

      <h2>Equipe</h2>
      {loading ? (
        <p>Carregando…</p>
      ) : (
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Nome</th>
                <th>E-mail</th>
                <th>Status</th>
                <th>Papéis</th>
                <th>MFA</th>
                <th>Ações</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id}>
                  <td>{u.name}</td>
                  <td>{u.email}</td>
                  <td>{u.status}</td>
                  <td>
                    {(u.roles ?? []).filter((r) => r !== 'customer').join(', ') || '—'}
                    {(u.roles ?? []).includes('customer') ? (
                      <span className="customer-email-hint"> · também cliente</span>
                    ) : null}
                  </td>
                  <td>{u.mfa_enabled ? 'Sim' : 'Não'}</td>
                  <td>
                    {u.status === 'active' && (
                      <button type="button" className="btn-link" onClick={() => void setStatus(u, 'suspended')}>
                        Suspender
                      </button>
                    )}
                    {u.status === 'suspended' && (
                      <button type="button" className="btn-link" onClick={() => void setStatus(u, 'active')}>
                        Reativar
                      </button>
                    )}
                    {u.status !== 'disabled' && (
                      <button type="button" className="btn-link" onClick={() => void setStatus(u, 'disabled')}>
                        Desativar
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
