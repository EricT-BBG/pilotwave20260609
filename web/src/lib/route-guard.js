import { hasAdminPermission, isAuthenticated, readStoredUser } from './session';

export function resolveRouteAccess(to, storage = sessionStorage) {
  if (!to.meta || !to.meta.requiresAuth) {
    return true;
  }

  if (!isAuthenticated(storage)) {
    return '/';
  }

  if (to.meta.requiresAdmin) {
    const user = readStoredUser(storage);
    if (!hasAdminPermission(user)) {
      return '/dashboard';
    }
  }

  return true;
}
