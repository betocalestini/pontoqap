type Card = { label: string; value: string };

export function ReportSummaryCards({ cards }: { cards: Card[] }) {
  if (!cards.length) return null;
  return (
    <div className="report-summary-cards">
      {cards.map((c) => (
        <div key={c.label} className="report-summary-card">
          <span className="report-summary-card__label">{c.label}</span>
          <strong className="report-summary-card__value">{c.value}</strong>
        </div>
      ))}
    </div>
  );
}
