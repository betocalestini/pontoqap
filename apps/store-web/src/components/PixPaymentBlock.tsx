import { useEffect, useState } from 'react';
import QRCode from 'qrcode';
import { formatMoney } from '@store/shared-core';

export type PixChargeView = {
  id: string;
  qr_code_text: string;
  amount_cents: number;
  expires_at?: string;
  qr_code_base64?: string;
  ticket_url?: string;
  provider?: string;
  simulatable?: boolean;
};

type Props = {
  charge: PixChargeView;
  onClose: () => void;
  onSimulate?: () => void;
};

export function PixPaymentBlock({ charge, onClose, onSimulate }: Props) {
  const [qrDataUrl, setQrDataUrl] = useState<string | null>(null);
  const [copyLabel, setCopyLabel] = useState('Copiar código Pix');

  useEffect(() => {
    let cancelled = false;
    if (charge.qr_code_base64) {
      const mime = charge.qr_code_base64.startsWith('data:')
        ? charge.qr_code_base64
        : `data:image/png;base64,${charge.qr_code_base64}`;
      setQrDataUrl(mime);
      return;
    }
    if (!charge.qr_code_text) {
      setQrDataUrl(null);
      return;
    }
    QRCode.toDataURL(charge.qr_code_text, { width: 220, margin: 1 })
      .then((url) => {
        if (!cancelled) setQrDataUrl(url);
      })
      .catch(() => {
        if (!cancelled) setQrDataUrl(null);
      });
    return () => {
      cancelled = true;
    };
  }, [charge.qr_code_base64, charge.qr_code_text]);

  async function copyPix() {
    try {
      await navigator.clipboard.writeText(charge.qr_code_text);
      setCopyLabel('Copiado!');
      window.setTimeout(() => setCopyLabel('Copiar código Pix'), 2000);
    } catch {
      setCopyLabel('Não foi possível copiar');
    }
  }

  const expiresLabel =
    charge.expires_at &&
    new Date(charge.expires_at).toLocaleString('pt-BR', {
      dateStyle: 'short',
      timeStyle: 'short',
    });

  return (
    <div className="invoice-card pix">
      <h2>Pagamento Pix</h2>
      <p className="pix-status-hint">Aguardando pagamento — atualizamos o status automaticamente.</p>
      <p>Valor: {formatMoney(charge.amount_cents)}</p>
      {expiresLabel && (
        <p className="invoice-card-meta">Válido até: {expiresLabel}</p>
      )}
      {qrDataUrl && (
        <p className="pix-qr-wrap">
          <img src={qrDataUrl} alt="QR Code Pix" className="pix-qr-image" width={220} height={220} />
        </p>
      )}
      <code className="pix-copy-paste">{charge.qr_code_text}</code>
      <div className="invoice-actions invoice-actions--stack-mobile">
        <button
          type="button"
          className="invoice-action-btn invoice-action-btn--primary invoice-action-btn--block"
          onClick={() => void copyPix()}
        >
          {copyLabel}
        </button>
        {charge.ticket_url && (
          <a
            className="invoice-action-btn invoice-action-btn--secondary invoice-action-btn--block"
            href={charge.ticket_url}
            target="_blank"
            rel="noreferrer"
          >
            Abrir instruções no Mercado Pago
          </a>
        )}
        {charge.provider === 'mercadopago' && (
          <p className="invoice-card-meta">
            Mercado Pago: com <code>MERCADO_PAGO_TEST_AUTO_APPROVE=true</code>, o pagamento pode ser aprovado
            automaticamente (payer APRO). Mantenha o worker ativo e o webhook configurado. Se a parcela não baixar
            em ~1 min, peça ao admin para sincronizar a cobrança.
          </p>
        )}
        <button
          type="button"
          className="invoice-action-btn invoice-action-btn--secondary invoice-action-btn--block"
          onClick={onClose}
        >
          Fechar
        </button>
        {charge.simulatable && onSimulate && (
          <button
            type="button"
            className="invoice-action-btn invoice-action-btn--primary invoice-action-btn--block"
            onClick={onSimulate}
          >
            Simular pagamento (dev)
          </button>
        )}
      </div>
    </div>
  );
}
