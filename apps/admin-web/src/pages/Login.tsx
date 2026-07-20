import { type FormEvent, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api';
import { useAuth } from '../auth/AuthProvider';

export function LoginPage() {
  const navigate = useNavigate();
  const { completeLogin } = useAuth();
  const [email, setEmail] = useState('gerente@loja.local');
  const [password, setPassword] = useState('ChangeMe123!');
  const [mfaCode, setMfaCode] = useState('');
  const [needsMfa, setNeedsMfa] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function finishLogin(res: Record<string, unknown>) {
    const token = res.access_token;
    if (typeof token !== 'string' || !token) {
      setError('Resposta de login inválida');
      return;
    }
    completeLogin(token);
    if (res.mfa_setup_required) {
      navigate('/mfa', { replace: true });
      return;
    }
    navigate('/', { replace: true });
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
      finishLogin(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha no login');
    }
  }

  return (
    <div className="app-shell app-shell--bare">
      <main className="app-main">
        <section className="content-section login-page">
          <h1>Painel administrativo</h1>
          <form onSubmit={onSubmit} className="form">
            {!needsMfa ? (
              <>
                <label>
                  E-mail
                  <input value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="username" />
                </label>
                <label>
                  Senha
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoComplete="current-password"
                  />
                </label>
              </>
            ) : (
              <label>
                Código MFA
                <input
                  value={mfaCode}
                  onChange={(e) => setMfaCode(e.target.value)}
                  inputMode="numeric"
                  autoComplete="one-time-code"
                />
              </label>
            )}
            {error && <p className="error">{error}</p>}
            <button type="submit">{needsMfa ? 'Validar MFA' : 'Entrar'}</button>
          </form>
        </section>
      </main>
    </div>
  );
}
