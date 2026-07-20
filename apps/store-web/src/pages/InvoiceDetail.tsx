import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';

type Charge = { id: string; qr_code_text: string; amount_cents: number };

export function InvoiceDetailPage() {
  const { id } = useParams();
  const [inv, setInv] = useState<Record<string, unknown> | null>(null);
  const [charge, setCharge] = useState<Charge | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    api.getMyInvoice(id).then((v) => setInv(v as Record<string, unknown>)).catch((e: Error) => setError(e.message));
  }, [id]);

  async function payPix() {
    if (!id) return;
    setError(null);
    try {
      const c = (await api.createPixCharge(id)) as Charge;
      setCharge(c);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  async function simulate() {
    if (!charge) return;
    try {
      await api.simulatePixPayment(charge.id);
      if (id) setInv((await api.getMyInvoice(id)) as Record<string, unknown>);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  return (
    <section className="content-section">
      <h1>Fatura</h1>
      {error && <p className="error">{error}</p>}
      {inv && <pre className="invoice-pre">{JSON.stringify(inv, null, 2)}</pre>}
      <div className="stack-sm stack-sm--row">
        <button type="button" onClick={payPix}>Gerar Pix</button>
      </div>
      {charge && (
        <div className="pix">
          <p>Valor: {formatMoney(charge.amount_cents)}</p>
          <code>{charge.qr_code_text}</code>
          <button type="button" onClick={simulate}>Simular pagamento (dev)</button>
        </div>
      )}
    </section>
  );
}
