import { type FormEvent, useState } from 'react';
import { createApiClient } from '@store/api-client';
import './App.css';

const api = createApiClient('/api/v1');

export default function App() {
  const [email, setEmail] = useState('gerente@loja.local');
  const [password, setPassword] = useState('ChangeMe123!');
  const [user, setUser] = useState<Record<string, unknown> | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      const res = await api.login(email, password, 'admin');
      setUser(res as Record<string, unknown>);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha no login');
    }
  }

  return (
    <div className="page">
      <h1>Painel administrativo</h1>
      {!user ? (
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
      ) : (
        <pre>{JSON.stringify(user, null, 2)}</pre>
      )}
    </div>
  );
}
