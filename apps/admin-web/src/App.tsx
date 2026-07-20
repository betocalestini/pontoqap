import { Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider, useAuth } from './auth/AuthProvider';
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

function AppRoutes() {
  const { status } = useAuth();

  if (status === 'loading') {
    return null;
  }

  if (status === 'unauthenticated') {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    );
  }

  return (
    <AppShell>
      <Routes>
        <Route path="/login" element={<Navigate to="/" replace />} />
        <Route path="/mfa" element={<MfaSetupPage />} />
        <Route path="/" element={<DashboardPage />} />
        <Route path="/clientes" element={<CustomersPage />} />
        <Route path="/faturamento" element={<BillingPage />} />
        <Route path="/produtos" element={<ProductsPage />} />
        <Route path="/estoque" element={<InventoryPage />} />
        <Route path="/relatorios" element={<ReportsPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AppShell>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <AppRoutes />
    </AuthProvider>
  );
}
