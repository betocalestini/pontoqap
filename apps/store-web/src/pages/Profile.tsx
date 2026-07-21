import { type FormEvent, useEffect, useState } from 'react';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { useStoreAuth } from '../auth/StoreAuthProvider';
import { api } from '../api';

export function ProfilePage() {
  const { status, user, refreshUser, signOut } = useStoreAuth();
  const nav = useNavigate();
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [document, setDocument] = useState('');
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!user) return;
    setName(user.name ?? '');
    setPhone(user.phone ?? '');
    setDocument(user.document ?? '');
  }, [user]);

  if (status === 'loading') {
    return null;
  }

  if (status === 'guest') {
    return <Navigate to="/login" replace state={{ from: '/perfil' }} />;
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const body: { name?: string; phone?: string; document?: string } = { name: name.trim() };
      if (phone.trim()) body.phone = phone.trim();
      if (user?.customer_id) body.document = document.trim();
      await api.updateMyProfile(body, 'store');
      await refreshUser();
      setMessage('Dados salvos.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Não foi possível salvar');
    } finally {
      setSaving(false);
    }
  }

  async function onSignOut() {
    await signOut();
    nav('/');
  }

  return (
    <section className="content-section">
      <h1>Meu perfil</h1>
      <form onSubmit={onSubmit} className="form">
        <label>
          Nome
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </label>
        <label>
          E-mail
          <input value={user?.email ?? ''} disabled />
        </label>
        <label>
          Telefone
          <input value={phone} onChange={(e) => setPhone(e.target.value)} />
        </label>
        {user?.customer_id && (
          <label>
            Documento (CPF/CNPJ)
            <input value={document} onChange={(e) => setDocument(e.target.value)} />
          </label>
        )}
        {error && <p className="error">{error}</p>}
        {message && <p className="ok">{message}</p>}
        <button type="submit" className="btn-block btn-block--sm-auto" disabled={saving}>
          {saving ? 'Salvando…' : 'Salvar'}
        </button>
      </form>
      <p>
        <button type="button" className="btn-link" onClick={() => void onSignOut()}>
          Sair da conta
        </button>
        {' · '}
        <Link to="/">Voltar ao catálogo</Link>
      </p>
    </section>
  );
}
