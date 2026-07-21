import { type FormEvent, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthProvider';
import { api } from '../api';

export function ProfilePage() {
  const { user, refreshUser, signOut } = useAuth();
  const navigate = useNavigate();
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

  if (!user) {
    return null;
  }

  const customerId = user.customer_id;

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const body: { name?: string; phone?: string; document?: string } = { name: name.trim() };
      if (phone.trim()) body.phone = phone.trim();
      if (customerId) body.document = document.trim();
      await api.updateMyProfile(body, 'admin');
      await refreshUser();
      setMessage('Dados salvos.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Não foi possível salvar');
    } finally {
      setSaving(false);
    }
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
          <input value={user.email} disabled />
        </label>
        <label>
          Telefone
          <input value={phone} onChange={(e) => setPhone(e.target.value)} />
        </label>
        {customerId && (
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
      {!user.mfa_enabled && (
        <p>
          <Link to="/mfa">Configurar autenticação em duas etapas (MFA)</Link>
        </p>
      )}
      <p>
        <button
          type="button"
          className="btn-link"
          onClick={async () => {
            await signOut();
            navigate('/login');
          }}
        >
          Sair da conta
        </button>
        {' · '}
        <Link to="/">Voltar ao painel</Link>
      </p>
    </section>
  );
}
