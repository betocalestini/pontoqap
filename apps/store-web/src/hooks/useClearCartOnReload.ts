import { useEffect } from 'react';
import { api } from '../api';

/** Esvazia o carrinho no servidor após recarregar a página (F5). */
export function useClearCartOnReload() {
  useEffect(() => {
    const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming | undefined;
    if (nav?.type !== 'reload') {
      return;
    }
    api.clearCart().catch(() => {
      // visitante ou sessão expirada — ignorar
    });
  }, []);
}
