import { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { api } from '../api';

export function VerifyEmailPage() {
  const [params] = useSearchParams();
  const [status, setStatus] = useState<'loading' | 'ok' | 'error'>('loading');
  const [message, setMessage] = useState('');

  useEffect(() => {
    const token = params.get('token');
    if (!token) {
      setStatus('error');
      setMessage('Link inválido.');
      return;
    }
    api
      .verifyEmail(token)
      .then(() => {
        setStatus('ok');
        setMessage('E-mail confirmado! Você já pode entrar e comprar.');
      })
      .catch((e: Error) => {
        setStatus('error');
        setMessage(e.message);
      });
  }, [params]);

  return (
    <section>
      <h1>Confirmação de e-mail</h1>
      {status === 'loading' && <p>Validando…</p>}
      {status !== 'loading' && <p className={status === 'ok' ? 'ok' : 'error'}>{message}</p>}
      <p>
        <Link to="/login">Ir para entrar</Link>
      </p>
    </section>
  );
}
