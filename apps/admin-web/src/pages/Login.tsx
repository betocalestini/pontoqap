import { type FormEvent, useState } from 'react';
import { api } from '../api';

export function LoginPage() {
  const [email, setEmail] = useState('gerente@loja.local');
  const [password, setPassword] = useState('ChangeMe123!');
  const [mfaCode, setMfaCode] = useState('');
  const [needsMfa, setNeedsMfa] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function finishLogin(res: Record<string, unknown>) {
    sessionStorage.setItem('admin_authed', '1');
    if (res.mfa_setup_required) {
      window.location.href = '/mfa';
      return;
    }
    window.location.href = '/';
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      const res = await api.login(email, password, 'admin', needsMfa ? mfaCode : undefined);
      if (res.mfa_required) {
        setNeedsMfa(true);
        return;
      }
      await finishLogin(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha no login');
    }
  }

  return (
    <section className="content-section">
      <h1>Painel administrativo</h1>
      <form onSubmit={onSubmit} className="form">
        {!needsMfa ? (
          <>
            <label>
              E-mail
              <input value={email} onChange={(e) => setEmail(e.target.value)} />
            </label>
            <label>
              Senha
              <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
            </label>
          </>
        ) : (
          <label>
            Código MFA
            <input value={mfaCode} onChange={(e) => setMfaCode(e.target.value)} inputMode="numeric" autoComplete="one-time-code" />
          </label>
        )}
        {error && <p className="error">{error}</p>}
        <button type="submit">{needsMfa ? 'Validar MFA' : 'Entrar'}</button>
      </form>
    </section>
  );
}
