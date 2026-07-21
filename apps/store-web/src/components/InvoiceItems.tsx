import type { BillingEntryView, InvoiceDetailItem, InvoiceProductLine } from '@store/api-client';
import { formatMoney } from '@store/shared-core';

function ProductLines({ lines }: { lines: InvoiceProductLine[] }) {
  if (lines.length === 0) {
    return null;
  }
  return (
    <ul className="invoice-product-lines">
      {lines.map((p, i) => (
        <li key={`${p.sku_code}-${i}`}>
          {p.product_name} ({p.sku_code}) — {p.quantity} × {formatMoney(p.unit_price_cents)} ={' '}
          {formatMoney(p.total_cents)}
        </li>
      ))}
    </ul>
  );
}

export function BillingEntriesList({ entries }: { entries: BillingEntryView[] }) {
  if (entries.length === 0) {
    return <p className="invoice-card-meta">Nenhum lançamento nesta competência.</p>;
  }
  return (
    <ul className="invoice-entry-list">
      {entries.map((e) => (
        <li key={e.id} className="invoice-entry-block">
          <p className="invoice-entry-heading">
            <strong>{e.order_number || e.description}</strong>
            {' — '}
            {formatMoney(e.amount_cents)}
          </p>
          {e.order_number && e.description !== e.order_number && (
            <p className="invoice-card-meta">{e.description}</p>
          )}
          <ProductLines lines={e.products} />
        </li>
      ))}
    </ul>
  );
}

export function InvoiceItemsList({ items }: { items: InvoiceDetailItem[] }) {
  if (items.length === 0) {
    return <p className="invoice-card-meta">Nenhum item nesta fatura.</p>;
  }
  return (
    <ul className="invoice-item-list">
      {items.map((it) => (
        <li key={it.id} className="invoice-entry-block">
          {it.products && it.products.length > 0 ? (
            <>
              <p className="invoice-entry-heading">
                <strong>{it.description}</strong> — {formatMoney(it.total_cents)}
              </p>
              <ProductLines lines={it.products} />
            </>
          ) : (
            <>
              {it.description} — {it.quantity} × {formatMoney(it.unit_price_cents)} ={' '}
              {formatMoney(it.total_cents)}
            </>
          )}
        </li>
      ))}
    </ul>
  );
}
