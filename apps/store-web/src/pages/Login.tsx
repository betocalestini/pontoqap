import { type FormEvent, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { ApiError } from '@store/api-client';
import { api } from '../api';
import { useStoreAuth } from '../auth/StoreAuthProvider';

export function LoginPage() {
  const nav = useNavigate();
  const location = useLocation();
  const { refreshUser } = useStoreAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [needsVerify, setNeedsVerify] = useState(false);
  const [resent, setResent] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setNeedsVerify(false);
    setResent(false);
    try {
      await api.login(email, password, 'store');
      await refreshUser();
      const from = (location.state as { from?: string } | null)?.from;
      nav(from && from !== '/login' ? from : '/');
    } catch (err) {
      if (err instanceof ApiError && err.code === 'EMAIL_NOT_VERIFIED') {
        setNeedsVerify(true);
      }
      setError(err instanceof Error ? err.message : 'Falha no login');
    }
  }

  async function resend() {
    setResent(false);
    try {
      await api.resendVerification(email);
      setResent(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao reenviar');
    }
  }

  return (
    <section className="content-section">
      <h1>Entrar</h1>
      <form onSubmit={onSubmit} className="form">
        <label>
          E-mail
          <input value={email} onChange={(e) => setEmail(e.target.value)} />
        </label>
        <label>
          Senha
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
        </label>
        {error && <p className="error">{error}</p>}
        {needsVerify && (
          <p>
            <button type="button" onClick={resend}>Reenviar e-mail de confirmação</button>
          </p>
        )}
        {resent && <p className="ok">Se o e-mail existir, enviamos um novo link.</p>}
        <button type="submit" className="btn-block btn-block--sm-auto">Entrar</button>
      </form>
      <p>
        Ainda não tem conta? <Link to="/cadastro">Criar conta</Link>
      </p>
    </section>
  );
}
