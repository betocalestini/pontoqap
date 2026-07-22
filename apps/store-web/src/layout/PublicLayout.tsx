import type { ReactNode } from 'react';
import { AppBrand } from '../components/AppBrand';
import { ThemeToggle } from '../components/ThemeToggle';

type PublicLayoutProps = {
  children: ReactNode;
};

export function PublicLayout({ children }: PublicLayoutProps) {
  return (
    <div className="public-shell">
      <header className="public-shell__header">
        <div className="public-shell__inner">
          <AppBrand to="/" />
          <ThemeToggle />
        </div>
      </header>
      <main className="public-shell__main">{children}</main>
    </div>
  );
}
