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
      <p className="reports-readonly-banner form-hint">
        Relatórios são somente consulta e exportação. Para alterar dados, use Clientes, Pedidos, Faturamento,
        Estoque ou Produtos.
      </p>
      <nav className="reports-subnav" aria-label="Relatórios">
        {links.map((l) => (
          <NavLink key={l.path} to={l.to} className={({ isActive }) => (isActive ? 'active' : undefined)}>
            {l.label}
          </NavLink>
        ))}
      </nav>
      <div className="reports-layout__content">
        <Outlet />
      </div>
    </div>
  );
}
