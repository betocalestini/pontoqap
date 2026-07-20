import * as SecureStore from 'expo-secure-store';

const SESSION_KEY = 'store_session_token';

/** BK-1303 — armazenamento seguro de sessão (token opaco para evolução futura). */
export async function saveSessionToken(token: string) {
  await SecureStore.setItemAsync(SESSION_KEY, token);
}

export async function loadSessionToken() {
  return SecureStore.getItemAsync(SESSION_KEY);
}

export async function clearSessionToken() {
  await SecureStore.deleteItemAsync(SESSION_KEY);
}
