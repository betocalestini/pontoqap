export type ApiErrorBody = { code: string; message: string };

export type AdminCategory = { id: string; name: string; slug: string; active: boolean };

export type AdminSku = {
  id: string;
  code: string;
  barcode?: string;
  unit: string;
  sale_price_cents: number;
  cost_price_cents?: number;
  minimum_stock: number;
  active: boolean;
  available_quantity?: number;
};

export type AdminProduct = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  category_id?: string;
  active: boolean;
  visible: boolean;
  image_url?: string;
  image_alt?: string;
  images?: { id: string; url: string; alt?: string }[];
  skus?: AdminSku[];
};

export type AdminCreateProductBody = {
  name: string;
  slug?: string;
  description?: string;
  category_id?: string;
  sku_code: string;
  barcode?: string;
  sale_price_cents: number;
  cost_price_cents?: number;
  minimum_stock?: number;
  unit?: string;
  initial_stock?: number;
};

export type AdminUpdateProductBody = {
  name?: string;
  description?: string;
  category_id?: string | null;
  active?: boolean;
  visible?: boolean;
};

export type AdminUpdateSkuBody = {
  code?: string;
  barcode?: string;
  unit?: string;
  sale_price_cents?: number;
  cost_price_cents?: number | null;
  minimum_stock?: number;
  active?: boolean;
  price_reason?: string;
};

export type AdminInventoryBalance = {
  sku_id: string;
  sku_code: string;
  product_id: string;
  product_name: string;
  minimum_stock: number;
  available_quantity: number;
};

export type AdminInventoryMovement = {
  id: string;
  sku_id: string;
  product_name?: string;
  sku_code?: string;
  movement_type: string;
  quantity: number;
  previous_balance: number;
  new_balance: number;
  reference_type?: string;
  reference_id?: string;
  reason?: string;
  created_by_email?: string;
  created_at: string;
};

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

    listProducts: (params?: { search?: string; page?: number; page_size?: number }) => {
      const q = new URLSearchParams();
      if (params?.search?.trim()) q.set('search', params.search.trim());
      if (params?.page != null) q.set('page', String(params.page));
      if (params?.page_size != null) q.set('page_size', String(params.page_size));
      const qs = q.toString();
      return request<{ items: unknown[]; total: number }>(
        `/catalog/products${qs ? `?${qs}` : ''}`,
      );
    },
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

    listMyInvoices: () =>
      request<{ current_period: unknown; items: unknown[] }>('/me/invoices', {}, 'store'),
    getMyInvoice: (id: string) => request(`/me/invoices/${id}`, {}, 'store'),
    createPixCharge: (invoiceId: string) =>
      request(`/me/invoices/${invoiceId}/pix-charge`, { method: 'POST' }, 'store'),
    simulatePixPayment: (chargeId: string) =>
      request(`/dev/pix/simulate/${chargeId}`, { method: 'POST' }),

    adminListCustomers: () => request<{ items: unknown[] }>('/admin/customers', {}, 'admin'),
    adminApproveCustomer: (id: string) =>
      request(`/admin/customers/${id}/approve`, { method: 'PATCH' }, 'admin'),
    adminInventoryEntry: (body: { sku_id: string; quantity: number; reason?: string; note?: string }) =>
      request('/admin/inventory/entries', {
        method: 'POST',
        body: JSON.stringify({
          sku_id: body.sku_id,
          quantity: body.quantity,
          reason: body.reason ?? body.note ?? '',
        }),
      }, 'admin'),
    adminInventoryBalances: () =>
      request<{ items: AdminInventoryBalance[] }>('/admin/inventory/balances', {}, 'admin'),
    adminInventoryMovements: (params?: { sku_id?: string; limit?: number; offset?: number }) => {
      const q = new URLSearchParams();
      if (params?.sku_id) q.set('sku_id', params.sku_id);
      if (params?.limit != null) q.set('limit', String(params.limit));
      if (params?.offset != null) q.set('offset', String(params.offset));
      const qs = q.toString();
      return request<{ items: AdminInventoryMovement[] }>(
        `/admin/inventory/movements${qs ? `?${qs}` : ''}`,
        {},
        'admin',
      );
    },
    adminCreateInventoryMovement: (body: {
      kind: 'entry' | 'loss' | 'damage' | 'adjustment';
      sku_id: string;
      quantity?: number;
      physical_count?: number;
      reason: string;
    }) =>
      request('/admin/inventory/movements', { method: 'POST', body: JSON.stringify(body) }, 'admin'),

    adminListCategories: () =>
      request<{ items: AdminCategory[] }>('/admin/categories', {}, 'admin'),
    adminListProducts: (params?: { search?: string; page?: number; page_size?: number }) => {
      const q = new URLSearchParams();
      if (params?.search?.trim()) q.set('search', params.search.trim());
      if (params?.page != null) q.set('page', String(params.page));
      if (params?.page_size != null) q.set('page_size', String(params.page_size));
      const qs = q.toString();
      return request<{ items: AdminProduct[]; total: number }>(
        `/admin/products${qs ? `?${qs}` : ''}`,
        {},
        'admin',
      );
    },
    adminGetProduct: (id: string) => request<AdminProduct>(`/admin/products/${id}`, {}, 'admin'),
    adminCreateProduct: (body: AdminCreateProductBody) =>
      request<AdminProduct>('/admin/products', { method: 'POST', body: JSON.stringify(body) }, 'admin'),
    adminUpdateProduct: (id: string, body: AdminUpdateProductBody) =>
      request<AdminProduct>(`/admin/products/${id}`, { method: 'PATCH', body: JSON.stringify(body) }, 'admin'),
    adminUpdateSku: (skuId: string, body: AdminUpdateSkuBody) =>
      request(`/admin/skus/${skuId}`, { method: 'PATCH', body: JSON.stringify(body) }, 'admin'),
    adminUploadProductImage: async (productId: string, file: File) => {
      const form = new FormData();
      form.append('image', file);
      const res = await fetch(`${baseUrl}/admin/products/${productId}/images`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'X-App-Audience': 'admin' },
        body: form,
      });
      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as ApiErrorBody;
        throw new ApiError({ code: body.code ?? 'ERROR', message: body.message ?? res.statusText });
      }
      return res.json() as Promise<{ image_url: string }>;
    },
    adminDeleteProductImage: (productId: string, imageId: string) =>
      request(`/admin/products/${productId}/images/${imageId}`, { method: 'DELETE' }, 'admin'),

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
