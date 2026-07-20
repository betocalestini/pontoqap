import { Link, Navigate, Route, Routes } from 'react-router-dom';
import { LoginPage } from './pages/Login';
import { DashboardPage } from './pages/Dashboard';
import { CustomersPage } from './pages/Customers';
import { BillingPage } from './pages/Billing';
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
            <Link to="/relatorios">Relatórios</Link>
          </nav>
        </header>
      )}
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={authed ? <DashboardPage /> : <Navigate to="/login" replace />} />
        <Route path="/clientes" element={authed ? <CustomersPage /> : <Navigate to="/login" replace />} />
        <Route path="/faturamento" element={authed ? <BillingPage /> : <Navigate to="/login" replace />} />
        <Route path="/relatorios" element={authed ? <ReportsPage /> : <Navigate to="/login" replace />} />
      </Routes>
    </div>
  );
}
