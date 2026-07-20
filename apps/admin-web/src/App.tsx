import { Navigate, Route, Routes } from 'react-router-dom';
import { AppShell } from './layout/AppShell';
import { LoginPage } from './pages/Login';
import { MfaSetupPage } from './pages/MfaSetup';
import { DashboardPage } from './pages/Dashboard';
import { CustomersPage } from './pages/Customers';
import { BillingPage } from './pages/Billing';
import { ProductsPage } from './pages/Products';
import { InventoryPage } from './pages/Inventory';
import { ReportsPage } from './pages/Reports';
import './App.css';

export default function App() {
  const authed = sessionStorage.getItem('admin_authed') === '1';

  const routes = (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/mfa" element={authed ? <MfaSetupPage /> : <Navigate to="/login" replace />} />
      <Route path="/" element={authed ? <DashboardPage /> : <Navigate to="/login" replace />} />
      <Route path="/clientes" element={authed ? <CustomersPage /> : <Navigate to="/login" replace />} />
      <Route path="/faturamento" element={authed ? <BillingPage /> : <Navigate to="/login" replace />} />
      <Route path="/produtos" element={authed ? <ProductsPage /> : <Navigate to="/login" replace />} />
      <Route path="/estoque" element={authed ? <InventoryPage /> : <Navigate to="/login" replace />} />
      <Route path="/relatorios" element={authed ? <ReportsPage /> : <Navigate to="/login" replace />} />
    </Routes>
  );

  if (authed) {
    return <AppShell>{routes}</AppShell>;
  }

  return (
    <div className="app-shell app-shell--bare">
      <main className="app-main">{routes}</main>
    </div>
  );
}
