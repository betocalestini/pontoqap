import { useEffect, useState, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { UserMenu } from '@store/ui';
import '@store/ui/user-menu.css';
import { storeNavLinks } from './navLinks';
import { AppBrand } from '../components/AppBrand';
import { ThemeToggle } from '../components/ThemeToggle';
import { useStoreAuth } from '../auth/StoreAuthProvider';

type AppShellProps = {
  children: ReactNode;
};

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
        {storeNavLinks.map(({ to, label }) => (
          <li key={to}>
            <Link
              to={to}
              className={pathname === to || pathname.startsWith(`${to}/`) ? 'app-nav__link is-active' : 'app-nav__link'}
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
  const navigate = useNavigate();
  const { user, signOut } = useStoreAuth();

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
          id="store-primary-nav"
          onNavigate={closeMenu}
          pathname={location.pathname}
        />
      </>,
      document.body,
    );

  return (
    <div className={`app-shell${menuOpen ? ' app-shell--menu-open' : ''}`}>
      <div className="app-top">
        <div className="app-top__inner">
          <AppBrand to="/loja" />
          <NavLinks navClassName="app-nav app-nav--bar" pathname={location.pathname} />
          <div className="app-top__actions">
            {user && (
              <UserMenu
                name={user.name}
                email={user.email}
                onSignOut={async () => {
                  await signOut();
                  navigate('/');
                }}
              />
            )}
            <ThemeToggle />
            <button
            type="button"
            className="nav-toggle"
            aria-expanded={menuOpen}
            aria-controls="store-primary-nav"
            onClick={() => setMenuOpen((open) => !open)}
          >
            <span className="nav-toggle__bars" aria-hidden />
            <span className="visually-hidden">{menuOpen ? 'Fechar menu' : 'Abrir menu'}</span>
          </button>
          </div>
        </div>
      </div>

      {mobileMenu}

      <main className="app-main">{children}</main>
    </div>
  );
}
