import type { PaymentOption } from '@store/api-client';
import { formatMoney } from '@store/shared-core';

type Props = {
  remainingCents: number;
  option: PaymentOption | undefined;
};

export function PaymentPlanConfirmMessage({ remainingCents, option }: Props) {
  return (
    <>
      <p>
        Confirmar pagamento da fatura de <strong>{formatMoney(remainingCents)}</strong> na forma escolhida?
      </p>
      {option && option.installments.length > 0 && (
        <ul className="payment-plan-preview payment-plan-preview--dialog">
          {option.installments.map((inst) => (
            <li key={inst.number}>
              <span>
                Parcela {inst.number}/{option.installments.length}
              </span>
              <span>{formatMoney(inst.amount_cents)}</span>
              <span>{new Date(inst.due_date).toLocaleDateString('pt-BR')}</span>
            </li>
          ))}
        </ul>
      )}
      <p className="invoice-card-meta">
        Parcelas em atraso podem bloquear novas compras até a regularização.
      </p>
    </>
  );
}
