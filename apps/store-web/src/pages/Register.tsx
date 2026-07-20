import { type FormEvent, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api';

export function RegisterPage() {
  const [form, setForm] = useState({ name: '', email: '', password: '', document: '', phone: '' });
  const [msg, setMsg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await api.registerCustomer(form);
      setMsg('Cadastro recebido. Enviamos um e-mail de confirmação — clique no link para ativar sua conta e começar a comprar.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro');
    }
  }

  return (
    <section>
      <h1>Cadastro</h1>
      <form
        className="form"
        onSubmit={onSubmit}
      >
        <label>
          Nome
          <input name="name" required value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} />
        </label>
        <label>
          E-mail
          <input name="email" type="email" required value={form.email} onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))} />
        </label>
        <label>
          Senha
          <input name="password" type="password" required value={form.password} onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))} />
        </label>
        <label>
          Documento
          <input name="document" value={form.document} onChange={(e) => setForm((f) => ({ ...f, document: e.target.value }))} />
        </label>
        <label>
          Telefone
          <input name="phone" value={form.phone} onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))} />
        </label>
        {error && <p className="error">{error}</p>}
        {msg && <p className="ok">{msg}</p>}
        <button type="submit">Enviar</button>
      </form>
      <p>
        Já tem conta? <Link to="/login">Entrar</Link>
      </p>
    </section>
  );
}
