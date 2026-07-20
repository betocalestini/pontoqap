import { Link, Navigate, Route, Routes } from 'react-router-dom';
import { LoginPage } from './pages/Login';
import { MfaSetupPage } from './pages/MfaSetup';
import { DashboardPage } from './pages/Dashboard';
import { CustomersPage } from './pages/Customers';
import { BillingPage } from './pages/Billing';
import { ProductsPage } from './pages/Products';
import { ReportsPage } from './pages/Reports';
import './App.css';

export default function App() {
  const authed = sessionStorage.getItem('admin_authed') === '1';

  return (
    <div className="page">
      {authed && (
        <header className="topbar">
          <strong>Painel</strong>
          <nav>
            <Link to="/">Dashboard</Link>
            <Link to="/clientes">Clientes</Link>
            <Link to="/faturamento">Faturamento</Link>
            <Link to="/produtos">Produtos</Link>
            <Link to="/relatorios">Relatórios</Link>
          </nav>
        </header>
      )}
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/mfa" element={authed ? <MfaSetupPage /> : <Navigate to="/login" replace />} />
        <Route path="/" element={authed ? <DashboardPage /> : <Navigate to="/login" replace />} />
        <Route path="/clientes" element={authed ? <CustomersPage /> : <Navigate to="/login" replace />} />
        <Route path="/faturamento" element={authed ? <BillingPage /> : <Navigate to="/login" replace />} />
        <Route path="/produtos" element={authed ? <ProductsPage /> : <Navigate to="/login" replace />} />
        <Route path="/relatorios" element={authed ? <ReportsPage /> : <Navigate to="/login" replace />} />
      </Routes>
    </div>
  );
}
