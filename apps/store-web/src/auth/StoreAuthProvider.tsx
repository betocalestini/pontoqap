import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import type { AuthMe } from '@store/api-client';
import { api } from '../api';

export type StoreAuthStatus = 'loading' | 'guest' | 'authenticated';

type StoreAuthContextValue = {
  status: StoreAuthStatus;
  user: AuthMe | null;
  refreshUser: () => Promise<void>;
  signOut: () => Promise<void>;
};

const StoreAuthContext = createContext<StoreAuthContextValue | null>(null);

export function StoreAuthProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<StoreAuthStatus>('loading');
  const [user, setUser] = useState<AuthMe | null>(null);

  const refreshUser = useCallback(async () => {
    try {
      const me = await api.me('store');
      setUser(me);
      setStatus('authenticated');
    } catch {
      setUser(null);
      setStatus('guest');
    }
  }, []);

  useEffect(() => {
    void refreshUser();
  }, [refreshUser]);

  const signOut = useCallback(async () => {
    try {
      await api.logout('store');
    } catch {
      /* ignore */
    }
    setUser(null);
    setStatus('guest');
  }, []);

  const value = useMemo(
    () => ({ status, user, refreshUser, signOut }),
    [status, user, refreshUser, signOut],
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
