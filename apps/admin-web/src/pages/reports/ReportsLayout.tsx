import { NavLink, Outlet } from 'react-router-dom';
import { usePermissions } from '../../auth/usePermissions';
import { reportNavLinks } from './reportNav';

function navVisible(permissions: string[], permission?: string | string[]) {
  if (!permission) return true;
  const codes = Array.isArray(permission) ? permission : [permission];
  return codes.some((c) => permissions.includes(c));
}

export function ReportsLayout() {
  const permissions = usePermissions();
  const links = reportNavLinks.filter((l) => navVisible(permissions, l.permission));

  return (
    <div className="reports-layout">
      <div className="reports-subnav-wrap">
        <nav className="reports-subnav" aria-label="Relatórios">
          {links.map((l) => (
            <NavLink
              key={l.path}
              to={l.to}
              end={l.to.startsWith('/relatorios/')}
              className={({ isActive }) =>
                isActive ? 'reports-subnav__link reports-subnav__link--active' : 'reports-subnav__link'
              }
            >
              {l.label}
            </NavLink>
          ))}
        </nav>
      </div>
      <div className="reports-layout__content">
        <Outlet />
      </div>
    </div>
  );
}
