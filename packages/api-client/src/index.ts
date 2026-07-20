export type ApiErrorBody = { code: string; message: string };

export class ApiError extends Error {
  code: string;
  constructor(body: ApiErrorBody) {
    super(body.message);
    this.code = body.code;
  }
}

const defaultBase = 'http://localhost:8080/api/v1';

type Audience = 'store' | 'admin';

export function createApiClient(baseUrl = defaultBase) {
  async function request<T>(
    path: string,
    init: RequestInit = {},
    audience?: Audience,
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(init.headers as Record<string, string> | undefined),
    };
    if (audience) {
      headers['X-App-Audience'] = audience;
    }
    const res = await fetch(`${baseUrl}${path}`, {
      credentials: 'include',
      headers,
      ...init,
    });
    if (!res.ok) {
      const body = (await res.json().catch(() => ({}))) as ApiErrorBody;
      throw new ApiError({ code: body.code ?? 'ERROR', message: body.message ?? res.statusText });
    }
    if (res.status === 204) {
      return undefined as T;
    }
    return res.json() as Promise<T>;
  }

  return {
    login: (email: string, password: string, audience: Audience, mfaCode?: string) =>
      request<Record<string, unknown>>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          email,
          password,
          audience,
          ...(mfaCode ? { mfa_code: mfaCode } : {}),
        }),
      }, audience),
    mfaSetup: () => request<{ secret: string; otpauth_url: string }>('/auth/mfa/setup', { method: 'POST' }, 'admin'),
    mfaVerify: (code: string) =>
      request('/auth/mfa/verify', { method: 'POST', body: JSON.stringify({ code }) }, 'admin'),
    logout: (audience: Audience) =>
      request('/auth/logout', { method: 'POST' }, audience),
    me: (audience: Audience) => request('/auth/me', {}, audience),

    verifyEmail: (token: string) =>
      request<{ status: string }>(`/auth/verify-email?token=${encodeURIComponent(token)}`),
    resendVerification: (email: string) =>
      request('/auth/resend-verification', { method: 'POST', body: JSON.stringify({ email }) }),

    listProducts: () => request<{ items: unknown[]; total: number }>('/catalog/products'),
    registerCustomer: (body: Record<string, string>) =>
      request('/customers/register', { method: 'POST', body: JSON.stringify(body) }),

    getCart: () => request('/me/cart', {}, 'store'),
    addToCart: (skuId: string, quantity: number) =>
      request('/me/cart/items', {
        method: 'POST',
        body: JSON.stringify({ sku_id: skuId, quantity }),
      }, 'store'),
    setCartItemQuantity: (skuId: string, quantity: number) =>
      request('/me/cart/items/' + encodeURIComponent(skuId), {
        method: 'PATCH',
        body: JSON.stringify({ quantity }),
      }, 'store'),
    checkout: () =>
      request('/me/cart/checkout', {
        method: 'POST',
        headers: { 'Idempotency-Key': crypto.randomUUID() },
      }, 'store'),

    listMyInvoices: () => request<{ items: unknown[] }>('/me/invoices', {}, 'store'),
    getMyInvoice: (id: string) => request(`/me/invoices/${id}`, {}, 'store'),
    createPixCharge: (invoiceId: string) =>
      request(`/me/invoices/${invoiceId}/pix-charge`, { method: 'POST' }, 'store'),
    simulatePixPayment: (chargeId: string) =>
      request(`/dev/pix/simulate/${chargeId}`, { method: 'POST' }),

    adminListCustomers: () => request<{ items: unknown[] }>('/admin/customers', {}, 'admin'),
    adminApproveCustomer: (id: string) =>
      request(`/admin/customers/${id}/approve`, { method: 'PATCH' }, 'admin'),
    adminInventoryEntry: (body: { sku_id: string; quantity: number; note?: string }) =>
      request('/admin/inventory/entries', { method: 'POST', body: JSON.stringify(body) }, 'admin'),
    adminCloseBilling: (body?: { year?: number; month?: number }) =>
      request('/admin/billing/close', { method: 'POST', body: JSON.stringify(body ?? {}) }, 'admin'),
    adminListInvoices: () => request<{ items: unknown[] }>('/admin/billing/invoices', {}, 'admin'),
    adminDashboard: (year?: number, month?: number) => {
      const q = new URLSearchParams();
      if (year) q.set('year', String(year));
      if (month) q.set('month', String(month));
      const suffix = q.toString() ? `?${q}` : '';
      return request(`/admin/reports/dashboard${suffix}`, {}, 'admin');
    },
    adminTopProducts: (year?: number, month?: number) => {
      const q = new URLSearchParams();
      if (year) q.set('year', String(year));
      if (month) q.set('month', String(month));
      const suffix = q.toString() ? `?${q}` : '';
      return request(`/admin/reports/top-products${suffix}`, {}, 'admin');
    },
    adminInventoryReport: () => request('/admin/reports/inventory', {}, 'admin'),
    adminForecast: () => request('/admin/reports/forecast', {}, 'admin'),
    adminGenerateForecast: () =>
      request('/admin/reports/forecast/generate', { method: 'POST' }, 'admin'),
  };
}
