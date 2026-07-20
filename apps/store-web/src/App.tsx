import { Route, Routes } from 'react-router-dom';
import { AppShell } from './layout/AppShell';
import { CatalogPage } from './pages/Catalog';
import { LoginPage } from './pages/Login';
import { RegisterPage } from './pages/Register';
import { CartPage } from './pages/Cart';
import { InvoicesPage } from './pages/Invoices';
import { InvoiceDetailPage } from './pages/InvoiceDetail';
import { VerifyEmailPage } from './pages/VerifyEmail';
import './App.css';

export default function App() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<CatalogPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/cadastro" element={<RegisterPage />} />
        <Route path="/carrinho" element={<CartPage />} />
        <Route path="/faturas" element={<InvoicesPage />} />
        <Route path="/verificar-email" element={<VerifyEmailPage />} />
        <Route path="/faturas/:id" element={<InvoiceDetailPage />} />
      </Routes>
    </AppShell>
  );
}
