import { createApiClient } from '@store/api-client';
import {
  clearAdminAccessToken,
  getAdminAccessToken,
} from './auth/token';

function onAdminUnauthorized() {
  clearAdminAccessToken();
  if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/login')) {
    window.location.replace('/login');
  }
}

export const api = createApiClient('/api/v1', {
  getAdminAccessToken,
  onAdminUnauthorized,
});
