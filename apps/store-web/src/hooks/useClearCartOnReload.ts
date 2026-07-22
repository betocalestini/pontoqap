import { useEffect } from 'react';
import { api } from '../api';
import { hasValidStoreAccessToken } from '../auth/token';

/** Esvazia o carrinho no servidor após recarregar a página (F5). */
export function useClearCartOnReload() {
  useEffect(() => {
    const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming | undefined;
    if (nav?.type !== 'reload') {
      return;
    }
    if (!hasValidStoreAccessToken()) {
      return;
    }
    api.clearCart({ skipStoreUnauthorizedHandler: true }).catch(() => {
      // visitante ou sessão expirada — ignorar
    });
  }, []);
}
