const STORAGE_KEY = 'store_access_token';

export function getStoreAccessToken(): string | null {
  try {
    return sessionStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

export function setStoreAccessToken(token: string): void {
  sessionStorage.setItem(STORAGE_KEY, token);
}

export function clearStoreAccessToken(): void {
  sessionStorage.removeItem(STORAGE_KEY);
}

export function isStoreAccessTokenExpired(token: string): boolean {
  try {
    const parts = token.split('.');
    if (parts.length < 2) return true;
    const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/'))) as { exp?: number };
    if (!payload.exp) return true;
    return payload.exp * 1000 <= Date.now();
  } catch {
    return true;
  }
}

export function hasValidStoreAccessToken(): boolean {
  const token = getStoreAccessToken();
  if (!token) return false;
  return !isStoreAccessTokenExpired(token);
}
