export function readStoredUser(storage = sessionStorage) {
  const raw = storage.getItem('member');
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function hasAdminPermission(user) {
  return Boolean(user && Array.isArray(user.permissions) && user.permissions.includes('admin'));
}

export function isAuthenticated(storage = sessionStorage) {
  return Boolean(storage.getItem('accessToken'));
}
