import type { MouseEvent, ReactElement } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { storeNavLinks } from '../layout/navLinks';

const TAB_LABELS: Record<(typeof storeNavLinks)[number]['to'], string> = {
  '/loja': 'Catálogo',
  '/carrinho': 'Carrinho',
  '/faturas': 'Faturas',
};

function isTabActive(pathname: string, to: string) {
  return pathname === to || pathname.startsWith(`${to}/`);
}

function CatalogIcon() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="14" y="14" width="7" height="7" rx="1" />
    </svg>
  );
}

function CartIcon() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
      <circle cx="9" cy="21" r="1" />
      <circle cx="20" cy="21" r="1" />
      <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function InvoiceIcon() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

const TAB_ICONS: Record<(typeof storeNavLinks)[number]['to'], () => ReactElement> = {
  '/loja': CatalogIcon,
  '/carrinho': CartIcon,
  '/faturas': InvoiceIcon,
};

export function StoreTabBar() {
  const { pathname } = useLocation();

  function onCatalogClick(e: MouseEvent<HTMLAnchorElement>) {
    if (pathname === '/loja') {
      e.preventDefault();
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  }

  return (
    <nav className="store-tab-bar" aria-label="Atalhos da loja">
      <ul className="store-tab-bar__list">
        {storeNavLinks.map(({ to }) => {
          const Icon = TAB_ICONS[to];
          const label = TAB_LABELS[to];
          const active = isTabActive(pathname, to);
          return (
            <li key={to} className="store-tab-bar__item">
              <Link
                to={to}
                className={active ? 'store-tab-bar__link is-active' : 'store-tab-bar__link'}
                aria-current={active ? 'page' : undefined}
                onClick={to === '/loja' ? onCatalogClick : undefined}
              >
                <Icon />
                <span className="store-tab-bar__label">{label}</span>
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
