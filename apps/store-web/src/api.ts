import { createApiClient } from '@store/api-client';
import { clearStoreAccessToken, getStoreAccessToken } from './auth/token';

const publicPathPrefixes = ['/', '/login', '/cadastro', '/verificar-email'];

function isPublicStorePath(pathname: string): boolean {
  return publicPathPrefixes.some((p) => pathname === p || (p !== '/' && pathname.startsWith(p)));
}

function onStoreUnauthorized() {
  clearStoreAccessToken();
  if (typeof window !== 'undefined' && !isPublicStorePath(window.location.pathname)) {
    window.location.replace('/login');
  }
}

export const api = createApiClient('/api/v1', {
  getStoreAccessToken,
  onStoreUnauthorized,
});
