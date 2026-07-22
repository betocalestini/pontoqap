import { describe, expect, it, vi, afterEach } from 'vitest';
import { createApiClient } from './index';

describe('createApiClient request headers', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('keeps Authorization when init supplies custom headers (checkout Idempotency-Key)', async () => {
    const fetchMock = vi.fn(async (_url: string, init?: RequestInit) => {
      const h = init?.headers as Record<string, string>;
      expect(h.Authorization).toBe('Bearer store-token');
      expect(h['Idempotency-Key']).toBeTruthy();
      expect(h['X-App-Audience']).toBe('store');
      return new Response(JSON.stringify({ order_number: 'ORD-1' }), { status: 201 });
    });
    vi.stubGlobal('fetch', fetchMock);

    const api = createApiClient('http://test/api/v1', {
      getStoreAccessToken: () => 'store-token',
    });

    await api.checkout({ skipStoreUnauthorizedHandler: true });

    expect(fetchMock).toHaveBeenCalledOnce();
  });
});
