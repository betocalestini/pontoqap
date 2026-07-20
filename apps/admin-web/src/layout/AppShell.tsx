import { useEffect, useState, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Link, useLocation } from 'react-router-dom';
import { adminNavLinks } from './navLinks';
import { usePermissions } from '../auth/usePermissions';

function navVisible(permissions: string[], permission?: string | string[]) {
  if (!permission) return true;
  const codes = Array.isArray(permission) ? permission : [permission];
  return codes.some((c) => permissions.includes(c));
}

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
  permissions,
}: {
  navClassName: string;
  onNavigate?: () => void;
  pathname: string;
  id?: string;
  permissions: string[];
}) {
  return (
    <nav id={id} className={navClassName} aria-label="Principal">
      <ul className="app-nav__list">
        {adminNavLinks
          .filter((link) => navVisible(permissions, link.permission))
          .map(({ to, label }) => (
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
  const permissions = usePermissions();

  useEffect(() => {
    setMenuOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    if (!menuOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMenuOpen(false);
    };
    const onResize = () => {
      if (window.innerWidth >= 768) setMenuOpen(false);
    };
    window.addEventListener('keydown', onKey);
    window.addEventListener('resize', onResize);
    return () => {
      window.removeEventListener('keydown', onKey);
      window.removeEventListener('resize', onResize);
    };
  }, [menuOpen]);

  const closeMenu = () => setMenuOpen(false);

  const mobileMenu =
    menuOpen &&
    createPortal(
      <>
        <button type="button" className="nav-backdrop" aria-label="Fechar menu" onClick={closeMenu} />
        <NavLinks
          navClassName="app-nav app-nav--drawer is-open"
          id="admin-primary-nav"
          onNavigate={closeMenu}
          pathname={location.pathname}
          permissions={permissions}
        />
      </>,
      document.body,
    );

  return (
    <div className={`app-shell${menuOpen ? ' app-shell--menu-open' : ''}`}>
      <div className="app-top">
        <div className="app-top__inner">
          <Link to="/" className="app-brand">
            Painel
          </Link>
          <NavLinks navClassName="app-nav app-nav--bar" pathname={location.pathname} permissions={permissions} />
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

      {mobileMenu}

      <main className="app-main">{children}</main>
    </div>
  );
}
