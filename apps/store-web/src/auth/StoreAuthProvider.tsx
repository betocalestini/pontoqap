import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import type { AuthMe } from '@store/api-client';
import { api } from '../api';
import {
  clearStoreAccessToken,
  hasValidStoreAccessToken,
  setStoreAccessToken,
} from './token';
import { setStoreUnauthorizedHandler } from './storeUnauthorized';

export type StoreAuthStatus = 'loading' | 'authenticated' | 'unauthenticated';

const publicPaths = new Set(['/', '/login', '/cadastro', '/verificar-email']);

type StoreAuthContextValue = {
  status: StoreAuthStatus;
  user: AuthMe | null;
  completeLogin: (accessToken: string) => void;
  refreshUser: () => Promise<void>;
  expireSession: () => void;
  signOut: () => Promise<void>;
};

const StoreAuthContext = createContext<StoreAuthContextValue | null>(null);

export function StoreAuthProvider({ children }: { children: ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const [status, setStatus] = useState<StoreAuthStatus>('loading');
  const [user, setUser] = useState<AuthMe | null>(null);

  const markUnauthenticated = useCallback(() => {
    clearStoreAccessToken();
    setUser(null);
    setStatus('unauthenticated');
  }, []);

  const expireSession = markUnauthenticated;

  useEffect(() => {
    setStoreUnauthorizedHandler(() => {
      markUnauthenticated();
      if (!publicPaths.has(location.pathname)) {
        navigate('/', { replace: true });
      }
    });
    return () => setStoreUnauthorizedHandler(null);
  }, [location.pathname, markUnauthenticated, navigate]);

  const refreshUser = useCallback(async () => {
    if (!hasValidStoreAccessToken()) {
      markUnauthenticated();
      return;
    }
    try {
      const me = await api.me('store');
      setUser(me);
      setStatus('authenticated');
    } catch {
      markUnauthenticated();
    }
  }, [markUnauthenticated]);

  const verifySession = refreshUser;

  useEffect(() => {
    if (publicPaths.has(location.pathname)) {
      if (hasValidStoreAccessToken()) {
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
      setStoreAccessToken(accessToken);
      void verifySession();
    },
    [verifySession],
  );

  const signOut = useCallback(async () => {
    try {
      await api.logout('store');
    } catch {
      /* ignore */
    }
    markUnauthenticated();
  }, [markUnauthenticated]);

  const value = useMemo(
    () => ({ status, user, completeLogin, refreshUser, expireSession, signOut }),
    [status, user, completeLogin, refreshUser, expireSession, signOut],
  );

  return <StoreAuthContext.Provider value={value}>{children}</StoreAuthContext.Provider>;
}

export function useStoreAuth(): StoreAuthContextValue {
  const ctx = useContext(StoreAuthContext);
  if (!ctx) {
    throw new Error('useStoreAuth must be used within StoreAuthProvider');
  }
  return ctx;
}
