import { Link } from 'react-router-dom';

export type RankingPanelItem = {
  key: string;
  primary: string;
  secondary?: string;
  href?: string;
};

type RankingPanelProps = {
  title: string;
  emptyMessage: string;
  items: RankingPanelItem[];
};

export function RankingPanel({ title, emptyMessage, items }: RankingPanelProps) {
  return (
    <article className="report-ranking">
      <h2 className="report-ranking__title">{title}</h2>
      {items.length === 0 ? (
        <p className="report-ranking__empty">{emptyMessage}</p>
      ) : (
        <ol className="report-ranking__list">
          {items.map((item, index) => (
            <li key={item.key} className="report-ranking__item">
              <div className="report-ranking__row">
                <span className="report-ranking__rank" aria-hidden>
                  {index + 1}
                </span>
                <div className="report-ranking__main">
                  {item.href ? (
                    <Link to={item.href} className="report-ranking__label report-ranking__label--link">
                      {item.primary}
                    </Link>
                  ) : (
                    <span className="report-ranking__label">{item.primary}</span>
                  )}
                </div>
                {item.secondary ? (
                  <span className="report-ranking__value">{item.secondary}</span>
                ) : null}
              </div>
            </li>
          ))}
        </ol>
      )}
    </article>
  );
}
