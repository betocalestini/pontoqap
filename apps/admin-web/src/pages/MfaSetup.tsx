import { type FormEvent, useEffect, useState } from 'react';
import { api } from '../api';

export function MfaSetupPage() {
  const [secret, setSecret] = useState('');
  const [uri, setUri] = useState('');
  const [code, setCode] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  useEffect(() => {
    api.mfaSetup()
      .then((r) => {
        setSecret(r.secret);
        setUri(r.otpauth_url);
      })
      .catch((e: Error) => setError(e.message));
  }, []);

  async function onVerify(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await api.mfaVerify(code);
      setDone(true);
      window.location.href = '/';
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Código inválido');
    }
  }

  return (
    <section>
      <h1>Configurar MFA</h1>
      <p>Escaneie no aplicativo autenticador (Google Authenticator, etc.).</p>
      {secret && (
        <p>
          Segredo manual: <code>{secret}</code>
        </p>
      )}
      {uri && <code style={{ wordBreak: 'break-all', fontSize: '0.75rem' }}>{uri}</code>}
      <form onSubmit={onVerify} className="form">
        <label>
          Código de 6 dígitos
          <input value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" required />
        </label>
        {error && <p className="error">{error}</p>}
        {done && <p className="ok">MFA ativado.</p>}
        <button type="submit">Confirmar</button>
      </form>
    </section>
  );
}
