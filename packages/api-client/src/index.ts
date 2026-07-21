export type ApiErrorBody = { code: string; message: string };

export type AuthMe = {
  id: string;
  name: string;
  email: string;
  phone?: string;
  document?: string;
  customer_id?: string;
  roles?: string[];
  permissions?: string[];
  mfa_enabled?: boolean;
};

export type UpdateMyProfileBody = {
  name?: string;
  phone?: string;
  document?: string;
};

export type AdminCategory = { id: string; name: string; slug: string; active: boolean };

export type CollaboratorCategory = {
  id: string;
  name: string;
  slug: string;
  margin_percent: number;
  active: boolean;
};

export type AdminCustomer = {
  id: string;
  user_id: string;
  name: string;
  email: string;
  phone?: string;
  document?: string;
  status: string;
  credit_limit_cents: number;
  current_exposure_cents: number;
  email_verified?: boolean;
  collaborator_category_id?: string;
  collaborator_category_name?: string;
  blocked_reason?: string;
  staff_roles?: string[];
  open_invoices_count?: number;
  overdue_invoices_count?: number;
};

export type AdminUpdateCustomerBody = {
  name?: string;
  phone?: string;
  document?: string;
  collaborator_category_id?: string | null;
};

export type AdminSku = {
  id: string;
  code: string;
  barcode?: string;
  unit: string;
  sale_price_cents: number;
  cost_price_cents?: number;
  average_cost_cents?: number;
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
  margin_percent?: number;
  promo_active?: boolean;
  promo_margin_percent?: number;
  promo_quantity_total?: number;
  promo_quantity_remaining?: number;
  on_promotion?: boolean;
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
  margin_percent?: number;
  promo_active?: boolean;
  promo_margin_percent?: number;
  promo_quantity?: number;
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
  unit_cost_cents?: number;
  created_by_email?: string;
  created_at: string;
};

export type AdminBillingSummary = {
  open_receivables_cents: number;
  overdue_invoices_count: number;
  open_periods_count: number;
  open_periods_total_cents: number;
  scheduled_closing_today: boolean;
  scheduled_monthly_close_today?: boolean;
};

export type MyInvoiceListItem = {
  id: string;
  invoice_number: string;
  status: string;
  total_cents: number;
  paid_cents: number;
  due_at?: string;
  close_type?: string;
  reference_year?: number;
  reference_month?: number;
  cycle_number?: number;
};

export type OpenBillingPeriod = {
  billing_period_id: string;
  reference_year: number;
  reference_month: number;
  cycle_number?: number;
  status: string;
  total_cents: number;
  entry_count: number;
};

export type AdminInvoiceListItem = {
  id: string;
  invoice_number: string;
  customer_id: string;
  customer_name: string;
  customer_email: string;
  reference_year: number;
  reference_month: number;
  status: string;
  total_cents: number;
  paid_cents: number;
  remaining_cents: number;
  due_at?: string;
  closed_at?: string;
};

export type InvoiceDetailItem = {
  id: string;
  description: string;
  quantity: number;
  unit_price_cents: number;
  total_cents: number;
  products?: InvoiceProductLine[];
};

export type InvoiceProductLine = {
  product_name: string;
  sku_code: string;
  quantity: number;
  unit_price_cents: number;
  total_cents: number;
};

export type BillingEntryView = {
  id: string;
  description: string;
  amount_cents: number;
  occurred_at: string;
  order_number?: string;
  products: InvoiceProductLine[];
};

export type OpenBillingPeriodDetail = {
  period: OpenBillingPeriod;
  entries: BillingEntryView[];
};

/** Detalhe de fatura na loja (mesmo formato do admin, sem campos sensíveis obrigatórios). */
export type MyInvoiceDetail = AdminInvoiceDetail;

export type InvoiceDetailAdjustment = {
  id: string;
  adjustment_type: string;
  amount_cents: number;
  reason: string;
  created_at: string;
};

export type AdminInvoiceDetail = {
  id: string;
  invoice_number: string;
  customer_id: string;
  customer_name?: string;
  customer_email?: string;
  billing_period_id: string;
  reference_year: number;
  reference_month: number;
  cycle_number?: number;
  close_type?: string;
  status: string;
  subtotal_cents: number;
  credit_cents: number;
  adjustment_cents: number;
  total_cents: number;
  paid_cents: number;
  remaining_cents: number;
  due_at?: string;
  closed_at?: string;
  paid_at?: string;
  items: InvoiceDetailItem[];
  adjustments: InvoiceDetailAdjustment[];
};

export type BillingCalendarEntry = {
  date: string;
  name: string;
  scope: string;
  is_business_day: boolean;
};

export type AdminOrderListItem = {
  id: string;
  order_number: string;
  status: string;
  total_cents: number;
  customer_id: string;
  customer_name: string;
  customer_email: string;
  confirmed_at?: string;
  created_at: string;
};

export type AdminOrderDetail = AdminOrderListItem & {
  items: {
    id: string;
    sku_id: string;
    product_name: string;
    sku_code: string;
    unit_price_cents: number;
    quantity: number;
    total_cents: number;
  }[];
  cancelled_at?: string;
};

export type AuditLogEntry = {
  id: string;
  actor_user_id?: string;
  actor_email?: string;
  action: string;
  entity_type: string;
  entity_id?: string;
  created_at: string;
  old_values?: unknown;
  new_values?: unknown;
};

export type AdminStaffUser = {
  id: string;
  name: string;
  email: string;
  status: string;
  roles: string[];
  mfa_enabled: boolean;
  last_login_at?: string;
  created_at: string;
};

export type AdminStaffRole = {
  id: string;
  code: string;
  name: string;
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

export type ApiClientOptions = {
  getAdminAccessToken?: () => string | null;
  onAdminUnauthorized?: () => void;
};

export function createApiClient(baseUrl = defaultBase, options: ApiClientOptions = {}) {
  async function request<T>(
    path: string,
    init: RequestInit = {},
    audience?: Audience,
  ): Promise<T> {
    const headers: Record<string, string> = {
      ...(init.headers as Record<string, string> | undefined),
    };
    const hasBody = init.body != null && init.method !== 'GET' && init.method !== 'HEAD';
    const isFormData = typeof FormData !== 'undefined' && init.body instanceof FormData;
    if (hasBody && !isFormData && !headers['Content-Type']) {
      headers['Content-Type'] = 'application/json';
    }
    if (audience) {
      headers['X-App-Audience'] = audience;
    }
    if (audience === 'admin') {
      const token = options.getAdminAccessToken?.();
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }
    }
    const res = await fetch(`${baseUrl}${path}`, {
      credentials: audience === 'admin' ? 'omit' : 'include',
      headers,
      ...init,
    });
    if (!res.ok) {
      if (res.status === 401 && audience === 'admin') {
        options.onAdminUnauthorized?.();
      }
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
    me: (audience: Audience) => request<AuthMe>('/auth/me', {}, audience),
    updateMyProfile: (body: UpdateMyProfileBody, audience: Audience) =>
      request<AuthMe>('/auth/me', { method: 'PATCH', body: JSON.stringify(body) }, audience),

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
    clearCart: () => request('/me/cart', { method: 'DELETE' }, 'store'),
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
      request<{ current_period: OpenBillingPeriod | null; items: MyInvoiceListItem[] }>(
        '/me/invoices',
        {},
        'store',
      ),
    closeMyBillingCycle: () =>
      request<MyInvoiceListItem>('/me/billing/close-cycle', { method: 'POST' }, 'store'),
    getMyOpenBillingPeriod: () =>
      request<OpenBillingPeriodDetail>('/me/billing/open-period', {}, 'store'),
    getMyInvoice: (id: string) => request<MyInvoiceDetail>(`/me/invoices/${id}`, {}, 'store'),
    createPixCharge: (invoiceId: string) =>
      request(`/me/invoices/${invoiceId}/pix-charge`, { method: 'POST' }, 'store'),
    simulatePixPayment: (chargeId: string) =>
      request(`/dev/pix/simulate/${chargeId}`, { method: 'POST' }),

    adminListCustomers: () => request<{ items: AdminCustomer[] }>('/admin/customers', {}, 'admin'),
    adminGetCustomer: (id: string) => request<AdminCustomer>(`/admin/customers/${id}`, {}, 'admin'),
    adminUpdateCustomer: (id: string, body: AdminUpdateCustomerBody) =>
      request<AdminCustomer>(`/admin/customers/${id}`, { method: 'PATCH', body: JSON.stringify(body) }, 'admin'),
    adminApproveCustomer: (id: string, credit_limit_cents = 100_000) =>
      request(`/admin/customers/${id}/approve`, {
        method: 'PATCH',
        body: JSON.stringify({ credit_limit_cents }),
      }, 'admin'),
    adminChangeCustomerLimit: (id: string, credit_limit_cents: number, reason: string) =>
      request(`/admin/customers/${id}/credit-limit`, {
        method: 'PATCH',
        body: JSON.stringify({ credit_limit_cents, reason }),
      }, 'admin'),
    adminAssignCustomerStaffRole: (id: string, body: { role_id: string; password?: string; mfa_code?: string }) =>
      request(`/admin/customers/${id}/staff-role`, { method: 'POST', body: JSON.stringify(body) }, 'admin'),
    adminBlockCustomer: (id: string, reason?: string) =>
      request(`/admin/customers/${id}/block`, {
        method: 'PATCH',
        body: JSON.stringify({ reason: reason ?? '' }),
      }, 'admin'),
    adminUnblockCustomer: (id: string) =>
      request(`/admin/customers/${id}/unblock`, { method: 'PATCH' }, 'admin'),
    adminListCollaboratorCategories: () =>
      request<{ items: CollaboratorCategory[] }>('/admin/collaborator-categories', {}, 'admin'),
    adminCreateCollaboratorCategory: (body: { name: string; slug?: string; margin_percent: number }) =>
      request<CollaboratorCategory>('/admin/collaborator-categories', {
        method: 'POST',
        body: JSON.stringify(body),
      }, 'admin'),
    adminUpdateCollaboratorCategory: (
      id: string,
      body: { name?: string; margin_percent?: number; active?: boolean },
    ) =>
      request<CollaboratorCategory>(`/admin/collaborator-categories/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(body),
      }, 'admin'),
    adminInventoryEntry: (body: {
      sku_id: string;
      quantity: number;
      reason?: string;
      note?: string;
      total_paid_cents?: number;
      other_expenses_cents?: number;
      unit_cost_cents?: number;
    }) =>
      request('/admin/inventory/entries', {
        method: 'POST',
        body: JSON.stringify({
          sku_id: body.sku_id,
          quantity: body.quantity,
          reason: body.reason ?? body.note ?? '',
          ...(body.total_paid_cents != null ? { total_paid_cents: body.total_paid_cents } : {}),
          ...(body.other_expenses_cents != null ? { other_expenses_cents: body.other_expenses_cents } : {}),
          ...(body.unit_cost_cents != null ? { unit_cost_cents: body.unit_cost_cents } : {}),
        }),
      }, 'admin'),
    adminInventoryBalances: () =>
      request<{ items: AdminInventoryBalance[] }>('/admin/inventory/balances', {}, 'admin'),
    adminInventoryMovements: (params?: {
      sku_id?: string;
      product_id?: string;
      limit?: number;
      offset?: number;
    }) => {
      const q = new URLSearchParams();
      if (params?.sku_id) q.set('sku_id', params.sku_id);
      if (params?.product_id) q.set('product_id', params.product_id);
      if (params?.limit != null) q.set('limit', String(params.limit));
      if (params?.offset != null) q.set('offset', String(params.offset));
      const qs = q.toString();
      return request<{ items: AdminInventoryMovement[]; total: number; limit: number; offset: number }>(
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
      total_paid_cents?: number;
      other_expenses_cents?: number;
      unit_cost_cents?: number;
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
      const headers: Record<string, string> = { 'X-App-Audience': 'admin' };
      const token = options.getAdminAccessToken?.();
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }
      const res = await fetch(`${baseUrl}/admin/products/${productId}/images`, {
        method: 'POST',
        credentials: 'omit',
        headers,
        body: form,
      });
      if (!res.ok) {
        if (res.status === 401) {
          options.onAdminUnauthorized?.();
        }
        const body = (await res.json().catch(() => ({}))) as ApiErrorBody;
        throw new ApiError({ code: body.code ?? 'ERROR', message: body.message ?? res.statusText });
      }
      return res.json() as Promise<{ image_url: string }>;
    },
    adminDeleteProductImage: (productId: string, imageId: string) =>
      request(`/admin/products/${productId}/images/${imageId}`, { method: 'DELETE' }, 'admin'),

    adminGetPricingSettings: () =>
      request<{ default_margin_percent: number }>('/admin/settings/pricing', {}, 'admin'),
    adminPatchPricingSettings: (default_margin_percent: number) =>
      request('/admin/settings/pricing', {
        method: 'PATCH',
        body: JSON.stringify({ default_margin_percent }),
      }, 'admin'),
    adminRepriceAllProducts: (margin_percent: number) =>
      request<{ products_updated: number }>('/admin/products/reprice-all', {
        method: 'POST',
        body: JSON.stringify({ margin_percent }),
      }, 'admin'),

    adminCloseBilling: (body: { year?: number; month?: number; reason: string }) =>
      request<{ closed_periods: number; year: number; month: number }>(
        '/admin/billing/close',
        { method: 'POST', body: JSON.stringify(body) },
        'admin',
      ),
    adminBillingSummary: () => request<AdminBillingSummary>('/admin/billing/summary', {}, 'admin'),
    adminListInvoices: (params?: {
      status?: string;
      year?: number;
      month?: number;
      search?: string;
      limit?: number;
      offset?: number;
    }) => {
      const q = new URLSearchParams();
      if (params?.status) q.set('status', params.status);
      if (params?.year != null) q.set('year', String(params.year));
      if (params?.month != null) q.set('month', String(params.month));
      if (params?.search?.trim()) q.set('search', params.search.trim());
      if (params?.limit != null) q.set('limit', String(params.limit));
      if (params?.offset != null) q.set('offset', String(params.offset));
      const qs = q.toString();
      return request<{ items: AdminInvoiceListItem[]; total: number; limit: number; offset: number }>(
        `/admin/billing/invoices${qs ? `?${qs}` : ''}`,
        {},
        'admin',
      );
    },
    adminGetInvoice: (id: string) => request<AdminInvoiceDetail>(`/admin/billing/invoices/${id}`, {}, 'admin'),
    adminAddInvoiceAdjustment: (
      id: string,
      body: { adjustment_type: 'credit' | 'debit'; amount_cents: number; reason: string },
    ) =>
      request<AdminInvoiceDetail>(`/admin/billing/invoices/${id}/adjustments`, {
        method: 'POST',
        body: JSON.stringify(body),
      }, 'admin'),
    adminBillingCalendar: (params?: { from?: string; to?: string }) => {
      const q = new URLSearchParams();
      if (params?.from) q.set('from', params.from);
      if (params?.to) q.set('to', params.to);
      const qs = q.toString();
      return request<{ items: BillingCalendarEntry[] }>(
        `/admin/billing/calendar${qs ? `?${qs}` : ''}`,
        {},
        'admin',
      );
    },
    adminUpsertBillingCalendar: (body: {
      date: string;
      name: string;
      scope?: string;
      is_business_day: boolean;
    }) =>
      request('/admin/billing/calendar', { method: 'PUT', body: JSON.stringify(body) }, 'admin'),
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

    acceptAdminInvitation: (body: { token: string; password: string; name?: string }) =>
      request('/auth/accept-invitation', { method: 'POST', body: JSON.stringify(body) }),

    adminListStaffUsers: () =>
      request<{ items: AdminStaffUser[] }>('/admin/users', {}, 'admin'),
    adminGetStaffUser: (id: string) =>
      request<{ user: AdminStaffUser; permissions: string[] }>(`/admin/users/${id}`, {}, 'admin'),
    adminListStaffRoles: () =>
      request<{ items: AdminStaffRole[] }>('/admin/roles', {}, 'admin'),
    adminCreateStaffInvitation: (body: { email: string; name: string; role_id: string }) =>
      request('/admin/users/invitations', { method: 'POST', body: JSON.stringify(body) }, 'admin'),
    adminRevokeStaffInvitation: (invitationId: string) =>
      request(`/admin/users/invitations/${invitationId}/revoke`, { method: 'POST' }, 'admin'),
    adminSetStaffUserRole: (id: string, body: { role_id: string; password?: string; mfa_code?: string }) =>
      request(`/admin/users/${id}/role`, { method: 'PATCH', body: JSON.stringify(body) }, 'admin'),
    adminSetStaffUserStatus: (id: string, body: { status: string; password?: string; mfa_code?: string }) =>
      request(`/admin/users/${id}/status`, { method: 'PATCH', body: JSON.stringify(body) }, 'admin'),
    adminRevokeStaffUserSessions: (id: string, body?: { password?: string; mfa_code?: string }) =>
      request(`/admin/users/${id}/sessions/revoke`, { method: 'POST', body: JSON.stringify(body ?? {}) }, 'admin'),

    adminListOrders: (params?: { status?: string; search?: string; limit?: number; offset?: number }) => {
      const q = new URLSearchParams();
      if (params?.status) q.set('status', params.status);
      if (params?.search) q.set('search', params.search);
      if (params?.limit) q.set('limit', String(params.limit));
      if (params?.offset) q.set('offset', String(params.offset));
      const qs = q.toString();
      return request<{ items: AdminOrderListItem[]; total: number }>(
        `/admin/orders${qs ? `?${qs}` : ''}`,
        {},
        'admin',
      );
    },
    adminGetOrder: (id: string) => request<AdminOrderDetail>(`/admin/orders/${id}`, {}, 'admin'),
    adminCancelOrder: (id: string) =>
      request<AdminOrderDetail>(`/admin/orders/${id}/cancel`, { method: 'POST' }, 'admin'),
    adminListAuditLogs: (params?: { action?: string; entity_type?: string; limit?: number; offset?: number }) => {
      const q = new URLSearchParams();
      if (params?.action) q.set('action', params.action);
      if (params?.entity_type) q.set('entity_type', params.entity_type);
      if (params?.limit) q.set('limit', String(params.limit));
      if (params?.offset) q.set('offset', String(params.offset));
      const qs = q.toString();
      return request<{ items: AuditLogEntry[]; total: number }>(
        `/admin/audit/logs${qs ? `?${qs}` : ''}`,
        {},
        'admin',
      );
    },
  };
}
