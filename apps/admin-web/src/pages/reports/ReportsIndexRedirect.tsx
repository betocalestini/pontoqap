import { Navigate } from 'react-router-dom';
import { usePermissions } from '../../auth/usePermissions';
import { reportNavLinks } from './reportNav';

function navVisible(permissions: string[], permission?: string | string[]) {
  if (!permission) return true;
  const codes = Array.isArray(permission) ? permission : [permission];
  return codes.some((c) => permissions.includes(c));
}

/** Redireciona /relatorios para o primeiro sub-relatório permitido (não duplica o dashboard em /). */
export function ReportsIndexRedirect() {
  const permissions = usePermissions();
  const first = reportNavLinks.find((l) => navVisible(permissions, l.permission));
  if (!first) {
    return <Navigate to="/" replace />;
  }
  return <Navigate to={first.to} replace />;
}
