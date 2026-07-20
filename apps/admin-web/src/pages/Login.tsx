import { type FormEvent, useState } from 'react';
import { api } from '../api';

export function LoginPage() {
  const [email, setEmail] = useState('gerente@loja.local');
  const [password, setPassword] = useState('ChangeMe123!');
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await api.login(email, password, 'admin');
      sessionStorage.setItem('admin_authed', '1');
      window.location.href = '/';
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha no login');
    }
  }

  return (
    <section>
      <h1>Painel administrativo</h1>
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
        <button type="submit">Entrar</button>
      </form>
    </section>
  );
}
