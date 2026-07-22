import { createApiClient } from '@store/api-client';
import { clearStoreAccessToken, getStoreAccessToken } from './auth/token';
import { notifyStoreUnauthorized } from './auth/storeUnauthorized';

export const api = createApiClient('/api/v1', {
  getStoreAccessToken,
  onStoreUnauthorized: () => {
    clearStoreAccessToken();
    notifyStoreUnauthorized();
  },
});
