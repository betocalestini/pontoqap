export type StoreUnauthorizedHandler = () => void;

let handler: StoreUnauthorizedHandler | null = null;

export function setStoreUnauthorizedHandler(h: StoreUnauthorizedHandler | null): void {
  handler = h;
}

export function notifyStoreUnauthorized(): void {
  handler?.();
}
