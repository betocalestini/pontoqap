import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { AuthProvider, useAuth } from './auth/AuthProvider';
import { AppShell } from './layout/AppShell';
import { LoginPage } from './pages/Login';
import { MfaSetupPage } from './pages/MfaSetup';
import { DashboardPage } from './pages/Dashboard';
import { CustomersPage } from './pages/Customers';
import { BillingPage } from './pages/Billing';
import { BillingInvoiceDetailPage } from './pages/BillingInvoiceDetail';
import { ProductsPage } from './pages/Products';
import { InventoryPage } from './pages/Inventory';
import { ReportsLayout } from './pages/reports/ReportsLayout';
import { ReportsIndexRedirect } from './pages/reports/ReportsIndexRedirect';
import { SalesReportPage } from './pages/reports/SalesReport';
import { InventoryPositionReportPage } from './pages/reports/InventoryPositionReport';
import { InventoryMovementsReportPage } from './pages/reports/InventoryMovementsReport';
import { ReceivablesReportPage } from './pages/reports/ReceivablesReport';
import { PixReconciliationReportPage } from './pages/reports/PixReconciliationReport';
import { CustomerExposureReportPage } from './pages/reports/CustomerExposureReport';
import { ExceptionsReportPage } from './pages/reports/ExceptionsReport';
import { ForecastReportPage } from './pages/reports/ForecastReport';
import { OrdersPage } from './pages/Orders';
import { AuditPage } from './pages/Audit';
import { UsersPage } from './pages/Users';
import { AcceptInvitePage } from './pages/AcceptInvite';
import { ProfilePage } from './pages/Profile';
import './App.css';

function AppRoutes() {
  const { status } = useAuth();
  const location = useLocation();
  const publicInvite = location.pathname === '/convite/aceitar';

  if (publicInvite) {
    return (
      <Routes>
        <Route path="/convite/aceitar" element={<AcceptInvitePage />} />
      </Routes>
    );
  }

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
        <Route path="/perfil" element={<ProfilePage />} />
        <Route path="/" element={<DashboardPage />} />
        <Route path="/clientes" element={<CustomersPage />} />
        <Route path="/faturamento" element={<BillingPage />} />
        <Route path="/faturamento/:id" element={<BillingInvoiceDetailPage />} />
        <Route path="/produtos" element={<ProductsPage />} />
        <Route path="/estoque" element={<InventoryPage />} />
        <Route path="/relatorios" element={<ReportsLayout />}>
          <Route index element={<ReportsIndexRedirect />} />
          <Route path="visao-geral" element={<Navigate to="/" replace />} />
          <Route path="vendas" element={<SalesReportPage />} />
          <Route path="estoque" element={<InventoryPositionReportPage />} />
          <Route path="movimentacoes" element={<InventoryMovementsReportPage />} />
          <Route path="recebiveis" element={<ReceivablesReportPage />} />
          <Route path="pix" element={<PixReconciliationReportPage />} />
          <Route path="limites" element={<CustomerExposureReportPage />} />
          <Route path="excecoes" element={<ExceptionsReportPage />} />
          <Route path="previsao" element={<ForecastReportPage />} />
        </Route>
        <Route path="/pedidos" element={<OrdersPage />} />
        <Route path="/auditoria" element={<AuditPage />} />
        <Route path="/usuarios" element={<UsersPage />} />
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
