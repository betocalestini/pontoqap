import { Route, Routes } from 'react-router-dom';
import { AppShell } from './layout/AppShell';
import { CatalogPage } from './pages/Catalog';
import { LoginPage } from './pages/Login';
import { RegisterPage } from './pages/Register';
import { CartPage } from './pages/Cart';
import { InvoicesPage } from './pages/Invoices';
import { OpenBillingPeriodPage } from './pages/OpenBillingPeriod';
import { InvoiceDetailPage } from './pages/InvoiceDetail';
import { VerifyEmailPage } from './pages/VerifyEmail';
import { ProfilePage } from './pages/Profile';
import { useClearCartOnReload } from './hooks/useClearCartOnReload';
import './App.css';

export default function App() {
  useClearCartOnReload();

  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<CatalogPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/cadastro" element={<RegisterPage />} />
        <Route path="/carrinho" element={<CartPage />} />
        <Route path="/faturas" element={<InvoicesPage />} />
        <Route path="/faturas/competencia-atual" element={<OpenBillingPeriodPage />} />
        <Route path="/verificar-email" element={<VerifyEmailPage />} />
        <Route path="/perfil" element={<ProfilePage />} />
        <Route path="/faturas/:id" element={<InvoiceDetailPage />} />
      </Routes>
    </AppShell>
  );
}
