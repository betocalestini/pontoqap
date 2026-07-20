import { useAuth } from './AuthProvider';

export function usePermissions(): string[] {
  return useAuth().permissions;
}

export function useHasPermission(code: string): boolean {
  return useAuth().permissions.includes(code);
}
