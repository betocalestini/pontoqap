import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { useStoreAuth } from './auth/StoreAuthProvider';
import { AppShell } from './layout/AppShell';
import { PublicLayout } from './layout/PublicLayout';
import { LandingPage } from './pages/Landing';
import { LoginPage } from './pages/Login';
import { RegisterPage } from './pages/Register';
import { CatalogPage } from './pages/Catalog';
import { CartPage } from './pages/Cart';
import { InvoicesPage } from './pages/Invoices';
import { OpenBillingPeriodPage } from './pages/OpenBillingPeriod';
import { InvoiceDetailPage } from './pages/InvoiceDetail';
import { VerifyEmailPage } from './pages/VerifyEmail';
import { ProfilePage } from './pages/Profile';
import { useClearCartOnReload } from './hooks/useClearCartOnReload';
import './App.css';

function GuestRoutes() {
  const location = useLocation();

  return (
    <PublicLayout>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/cadastro" element={<RegisterPage />} />
        <Route path="/verificar-email" element={<VerifyEmailPage />} />
        <Route
          path="*"
          element={<Navigate to="/login" replace state={{ from: location.pathname }} />}
        />
      </Routes>
    </PublicLayout>
  );
}

function AuthenticatedRoutes() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<Navigate to="/loja" replace />} />
        <Route path="/login" element={<Navigate to="/loja" replace />} />
        <Route path="/cadastro" element={<Navigate to="/loja" replace />} />
        <Route path="/loja" element={<CatalogPage />} />
        <Route path="/carrinho" element={<CartPage />} />
        <Route path="/faturas" element={<InvoicesPage />} />
        <Route path="/faturas/competencia-atual" element={<OpenBillingPeriodPage />} />
        <Route path="/faturas/:id" element={<InvoiceDetailPage />} />
        <Route path="/perfil" element={<ProfilePage />} />
        <Route path="*" element={<Navigate to="/loja" replace />} />
      </Routes>
    </AppShell>
  );
}

function AppRoutes() {
  const { status } = useStoreAuth();

  if (status === 'loading') {
    return null;
  }

  if (status === 'unauthenticated') {
    return <GuestRoutes />;
  }

  return <AuthenticatedRoutes />;
}

export default function App() {
  useClearCartOnReload();

  return <AppRoutes />;
}
