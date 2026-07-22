import { describe, expect, it, vi, afterEach } from 'vitest';
import { createApiClient, ApiError } from './index';

describe('createApiClient unauthorized handling', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('invokes onAdminUnauthorized on 401 for admin audience', async () => {
    const onAdminUnauthorized = vi.fn();
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(JSON.stringify({ code: 'UNAUTHORIZED', message: 'Não autenticado' }), { status: 401 })),
    );

    const api = createApiClient('http://test/api/v1', {
      getAdminAccessToken: () => 'bad-token',
      onAdminUnauthorized,
    });

    await expect(api.adminListCustomers()).rejects.toBeInstanceOf(ApiError);
    expect(onAdminUnauthorized).toHaveBeenCalledOnce();
  });

  it('does not invoke handler when skipAdminUnauthorizedHandler is set', async () => {
    const onAdminUnauthorized = vi.fn();
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(JSON.stringify({ code: 'UNAUTHORIZED', message: 'x' }), { status: 401 })),
    );

    const api = createApiClient('http://test/api/v1', {
      getAdminAccessToken: () => 'bad-token',
      onAdminUnauthorized,
    });

    await expect(
      api.adminListCustomers({ skipAdminUnauthorizedHandler: true }),
    ).rejects.toBeInstanceOf(ApiError);
    expect(onAdminUnauthorized).not.toHaveBeenCalled();
  });

  it('throws ApiError with server code on 403', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(JSON.stringify({ code: 'FORBIDDEN', message: 'Permissão insuficiente' }), { status: 403 })),
    );

    const api = createApiClient('http://test/api/v1', {
      getAdminAccessToken: () => 'token',
    });

    try {
      await api.adminListCustomers();
      expect.unreachable('should throw');
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).code).toBe('FORBIDDEN');
    }
  });
});
