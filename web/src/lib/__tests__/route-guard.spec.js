import { describe, expect, it } from 'vitest';
import { resolveRouteAccess } from '../route-guard';

function createStorage(values = {}) {
  return {
    getItem(key) {
      return values[key] ?? null;
    }
  };
}

describe('route guard', () => {
  it('redirects anonymous users away from protected routes', () => {
    const result = resolveRouteAccess(
      { meta: { requiresAuth: true } },
      createStorage()
    );

    expect(result).toBe('/');
  });

  it('redirects non-admin users away from admin-only routes', () => {
    const result = resolveRouteAccess(
      { meta: { requiresAuth: true, requiresAdmin: true } },
      createStorage({
        accessToken: 'token',
        member: JSON.stringify({ permissions: ['viewer'] })
      })
    );

    expect(result).toBe('/dashboard');
  });

  it('allows admins into admin-only routes', () => {
    const result = resolveRouteAccess(
      { meta: { requiresAuth: true, requiresAdmin: true } },
      createStorage({
        accessToken: 'token',
        member: JSON.stringify({ permissions: ['admin'] })
      })
    );

    expect(result).toBe(true);
  });
});
