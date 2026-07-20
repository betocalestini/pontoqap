import { type FormEvent, useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { api } from '../api';

export function AcceptInvitePage() {
  const [params] = useSearchParams();
  const [token, setToken] = useState('');
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const t = params.get('token');
    if (t) setToken(t);
  }, [params]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    if (password.length < 8) {
      setError('A senha deve ter pelo menos 8 caracteres');
      return;
    }
    if (password !== confirm) {
      setError('As senhas não coincidem');
      return;
    }
    try {
      await api.acceptAdminInvitation({ token, password, name: name || undefined });
      setDone(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao aceitar convite');
    }
  }

  return (
    <div className="app-shell app-shell--bare">
      <main className="app-main">
        <section className="content-section login-page">
          <h1>Ativar conta administrativa</h1>
          {done ? (
            <p>
              Conta ativada.{' '}
              <a href="/login">Ir para o login</a>
            </p>
          ) : (
            <form onSubmit={onSubmit} className="form">
              <label>
                Token (do link)
                <input value={token} onChange={(e) => setToken(e.target.value)} required />
              </label>
              <label>
                Seu nome
                <input value={name} onChange={(e) => setName(e.target.value)} autoComplete="name" />
              </label>
              <label>
                Nova senha
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="new-password"
                  required
                />
              </label>
              <label>
                Confirmar senha
                <input
                  type="password"
                  value={confirm}
                  onChange={(e) => setConfirm(e.target.value)}
                  autoComplete="new-password"
                  required
                />
              </label>
              {error && <p className="error">{error}</p>}
              <button type="submit">Criar senha</button>
            </form>
          )}
        </section>
      </main>
    </div>
  );
}
