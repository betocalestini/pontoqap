import { type FormEvent, useCallback, useEffect, useState } from 'react';
import type { AdminStaffRole, AdminStaffUser } from '@store/api-client';
import { useDialog } from '@store/ui';
import { api } from '../api';
import { STAFF_ROLE_SUMMARIES } from '../content/staffRoles';

function staffRoleCode(user: AdminStaffUser): string {
  return (user.roles ?? []).find((r) => r !== 'customer') ?? '';
}

export function UsersPage() {
  const { prompt } = useDialog();

  async function promptStepUp(): Promise<string | null> {
    return prompt({
      title: 'Confirmar identidade',
      message: 'Confirme sua senha para esta ação.',
      label: 'Senha',
      inputType: 'password',
      confirmLabel: 'Confirmar',
    });
  }

  const [users, setUsers] = useState<AdminStaffUser[]>([]);
  const [roles, setRoles] = useState<AdminStaffRole[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteName, setInviteName] = useState('');
  const [inviteRoleId, setInviteRoleId] = useState('');
  const [inviteMsg, setInviteMsg] = useState<string | null>(null);
  const [expandedPermsUserId, setExpandedPermsUserId] = useState<string | null>(null);
  const [expandedPerms, setExpandedPerms] = useState<string[]>([]);

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
    const password = await promptStepUp();
    if (!password) return;
    try {
      await api.adminSetStaffUserStatus(user.id, { status, password });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao alterar status');
    }
  }

  async function changeRole(user: AdminStaffUser, roleId: string) {
    if (!roleId) return;
    const password = await promptStepUp();
    if (!password) return;
    try {
      await api.adminSetStaffUserRole(user.id, { role_id: roleId, password });
      if (expandedPermsUserId === user.id) {
        setExpandedPermsUserId(null);
        setExpandedPerms([]);
      }
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao alterar papel');
    }
  }

  async function revokeSessions(user: AdminStaffUser) {
    const password = await promptStepUp();
    if (!password) return;
    try {
      await api.adminRevokeStaffUserSessions(user.id, { password });
      setInviteMsg(`Sessões do painel revogadas para ${user.email}.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao revogar sessões');
    }
  }

  async function togglePermissions(user: AdminStaffUser) {
    if (expandedPermsUserId === user.id) {
      setExpandedPermsUserId(null);
      setExpandedPerms([]);
      return;
    }
    try {
      const res = await api.adminGetStaffUser(user.id);
      const perms = Array.isArray(res.permissions) ? (res.permissions as string[]) : [];
      setExpandedPermsUserId(user.id);
      setExpandedPerms(perms.sort());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao carregar permissões');
    }
  }

  return (
    <section className="content-section">
      <h1>Usuários internos</h1>
      <p className="form-hint">
        Funcionários precisam de cadastro na loja como clientes. Papéis são fixos (sem permissões customizadas). Ver
        também <strong>Clientes → Editar → Acesso administrativo</strong>. Suspender bloqueia só o painel; a conta de
        cliente na loja continua ativa.
      </p>
      {error && <p className="error">{error}</p>}
      {inviteMsg && <p className="success">{inviteMsg}</p>}

      <h2>Papéis do sistema</h2>
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Papel</th>
              <th>Responsabilidades</th>
            </tr>
          </thead>
          <tbody>
            {STAFF_ROLE_SUMMARIES.map((s) => (
              <tr key={s.code}>
                <td>
                  <strong>{s.title}</strong>
                  <br />
                  <code>{s.code}</code>
                </td>
                <td>
                  <ul>
                    {s.bullets.map((b) => (
                      <li key={b}>{b}</li>
                    ))}
                  </ul>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

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
                <th>Papel interno</th>
                <th>MFA</th>
                <th>Ações</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => {
                const code = staffRoleCode(u);
                const currentRoleId = roles.find((r) => r.code === code)?.id ?? '';
                return (
                  <tr key={u.id}>
                    <td>{u.name}</td>
                    <td>{u.email}</td>
                    <td>{u.status}</td>
                    <td>
                      <select
                        value={currentRoleId}
                        onChange={(e) => void changeRole(u, e.target.value)}
                        aria-label={`Papel de ${u.name}`}
                      >
                        {roles.map((r) => (
                          <option key={r.id} value={r.id}>
                            {r.name}
                          </option>
                        ))}
                      </select>
                      {(u.roles ?? []).includes('customer') ? (
                        <span className="customer-email-hint"> · também cliente</span>
                      ) : null}
                      {expandedPermsUserId === u.id && (
                        <p className="form-hint">
                          Permissões: {expandedPerms.length ? expandedPerms.join(', ') : '—'}
                        </p>
                      )}
                    </td>
                    <td>{u.mfa_enabled ? 'Sim' : 'Não'}</td>
                    <td>
                      <button type="button" className="btn-link" onClick={() => void togglePermissions(u)}>
                        {expandedPermsUserId === u.id ? 'Ocultar permissões' : 'Ver permissões'}
                      </button>
                      <button type="button" className="btn-link" onClick={() => void revokeSessions(u)}>
                        Revogar sessões
                      </button>
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
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
