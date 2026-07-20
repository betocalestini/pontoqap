import { useEffect, useState, type ReactNode } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { adminNavLinks } from './navLinks';

type AppShellProps = {
  children: ReactNode;
};

function isActive(pathname: string, to: string) {
  if (to === '/') return pathname === '/';
  return pathname === to || pathname.startsWith(`${to}/`);
}

function NavLinks({
  navClassName,
  onNavigate,
  pathname,
  id,
}: {
  navClassName: string;
  onNavigate?: () => void;
  pathname: string;
  id?: string;
}) {
  return (
    <nav id={id} className={navClassName} aria-label="Principal">
      <ul className="app-nav__list">
        {adminNavLinks.map(({ to, label }) => (
          <li key={to}>
            <Link
              to={to}
              className={isActive(pathname, to) ? 'app-nav__link is-active' : 'app-nav__link'}
              onClick={onNavigate}
            >
              {label}
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  );
}

export function AppShell({ children }: AppShellProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const location = useLocation();

  useEffect(() => {
    setMenuOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    document.body.style.overflow = menuOpen ? 'hidden' : '';
    return () => {
      document.body.style.overflow = '';
    };
  }, [menuOpen]);

  const closeMenu = () => setMenuOpen(false);

  return (
    <div className={`app-shell${menuOpen ? ' app-shell--menu-open' : ''}`}>
      <div className="app-top">
        <div className="app-top__inner">
          <Link to="/" className="app-brand">
            Painel
          </Link>
          <NavLinks navClassName="app-nav app-nav--bar" pathname={location.pathname} />
          <button
            type="button"
            className="nav-toggle"
            aria-expanded={menuOpen}
            aria-controls="admin-primary-nav"
            onClick={() => setMenuOpen((open) => !open)}
          >
            <span className="nav-toggle__bars" aria-hidden />
            <span className="visually-hidden">{menuOpen ? 'Fechar menu' : 'Abrir menu'}</span>
          </button>
        </div>
      </div>

      {menuOpen && (
        <>
          <button type="button" className="nav-backdrop" aria-label="Fechar menu" onClick={closeMenu} />
          <NavLinks
            navClassName="app-nav app-nav--drawer is-open"
            id="admin-primary-nav"
            onNavigate={closeMenu}
            pathname={location.pathname}
          />
        </>
      )}

      <main className="app-main">{children}</main>
    </div>
  );
}
