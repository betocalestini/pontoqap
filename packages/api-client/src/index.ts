export type ApiErrorBody = { code: string; message: string };

const defaultBase = 'http://localhost:8080/api/v1';

export function createApiClient(baseUrl = defaultBase) {
  async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const res = await fetch(`${baseUrl}${path}`, {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        ...(init.headers ?? {}),
      },
      ...init,
    });
    if (!res.ok) {
      const body = (await res.json().catch(() => ({}))) as ApiErrorBody;
      throw new Error(body.message ?? res.statusText);
    }
    if (res.status === 204) {
      return undefined as T;
    }
    return res.json() as Promise<T>;
  }

  return {
    login: (email: string, password: string, audience: 'store' | 'admin') =>
      request('/auth/login', {
        method: 'POST',
        headers: { 'X-App-Audience': audience },
        body: JSON.stringify({ email, password, audience }),
      }),
    me: () => request('/auth/me'),
    listProducts: () => request<{ items: unknown[]; total: number }>('/catalog/products'),
    registerCustomer: (body: Record<string, string>) =>
      request('/customers/register', { method: 'POST', body: JSON.stringify(body) }),
  };
}
