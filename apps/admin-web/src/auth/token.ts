const STORAGE_KEY = 'admin_access_token';

export function getAdminAccessToken(): string | null {
  try {
    return sessionStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

export function setAdminAccessToken(token: string): void {
  sessionStorage.setItem(STORAGE_KEY, token);
}

export function clearAdminAccessToken(): void {
  sessionStorage.removeItem(STORAGE_KEY);
  sessionStorage.removeItem('admin_authed');
}

export function isAdminAccessTokenExpired(token: string): boolean {
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

export function hasValidAdminAccessToken(): boolean {
  const token = getAdminAccessToken();
  if (!token) return false;
  return !isAdminAccessTokenExpired(token);
}
