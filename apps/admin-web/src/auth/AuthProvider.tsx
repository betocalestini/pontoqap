import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import { useLocation } from 'react-router-dom';
import { api } from '../api';
import {
  clearAdminAccessToken,
  hasValidAdminAccessToken,
  setAdminAccessToken,
} from './token';

export type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated';

type AuthContextValue = {
  status: AuthStatus;
  permissions: string[];
  completeLogin: (accessToken: string) => void;
  signOut: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const location = useLocation();
  const [status, setStatus] = useState<AuthStatus>('loading');
  const [permissions, setPermissions] = useState<string[]>([]);

  const markUnauthenticated = useCallback(() => {
    clearAdminAccessToken();
    setPermissions([]);
    setStatus('unauthenticated');
  }, []);

  const verifySession = useCallback(async () => {
    if (!hasValidAdminAccessToken()) {
      markUnauthenticated();
      return;
    }
    try {
      const me = (await api.me('admin')) as { permissions?: unknown };
      const perms = Array.isArray(me.permissions) ? (me.permissions as string[]) : [];
      setPermissions(perms);
      setStatus('authenticated');
    } catch {
      markUnauthenticated();
    }
  }, [markUnauthenticated]);

  useEffect(() => {
    if (location.pathname === '/convite/aceitar') {
      setStatus('unauthenticated');
      return;
    }
    if (location.pathname === '/login') {
      if (hasValidAdminAccessToken()) {
        void verifySession();
        return;
      }
      setStatus('unauthenticated');
      return;
    }
    void verifySession();
  }, [location.pathname, verifySession]);

  const completeLogin = useCallback(
    (accessToken: string) => {
      setAdminAccessToken(accessToken);
      void verifySession();
    },
    [verifySession],
  );

  const signOut = useCallback(async () => {
    try {
      await api.logout('admin');
    } catch {
      /* ignore */
    }
    clearAdminAccessToken();
    setStatus('unauthenticated');
  }, []);

  const value = useMemo(
    () => ({ status, permissions, completeLogin, signOut }),
    [status, permissions, completeLogin, signOut],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
}
