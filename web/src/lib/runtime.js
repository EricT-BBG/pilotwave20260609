export function getEnvValue(primaryKey, fallbackKey = '', env = import.meta.env || {}) {
  const primary = env[primaryKey];
  if (primary !== undefined && primary !== '') {
    return primary;
  }

  if (!fallbackKey) {
    return '';
  }

  const fallback = env[fallbackKey];
  if (fallback !== undefined && fallback !== '') {
    return fallback;
  }

  return '';
}

function getDefaultLocation() {
  if (typeof globalThis !== 'undefined' && globalThis.location) {
    return globalThis.location;
  }

  return {
    origin: '',
    protocol: 'http:',
    host: 'localhost'
  };
}

function getLocationOrigin(location) {
  if (location?.origin) {
    return location.origin;
  }

  if (location?.protocol && location?.host) {
    return `${location.protocol}//${location.host}`;
  }

  return '';
}

export function resolveApiBaseConfig(location = getDefaultLocation(), env = import.meta.env || {}) {
  let apiUrl = getEnvValue('VITE_API_URL', 'VUE_APP_API_URL', env);
  const apiVersion = getEnvValue('VITE_API_VERSION', 'VUE_APP_API_VERSION', env) || '/api/v1';

  if (apiUrl === '//' || apiUrl === '') {
    apiUrl = getLocationOrigin(location);
  }

  return {
    apiUrl,
    apiVersion
  };
}

export function buildApiUrl(pathname, location = getDefaultLocation(), env = import.meta.env || {}) {
  const { apiUrl, apiVersion } = resolveApiBaseConfig(location, env);
  return `${apiUrl}${apiVersion}${pathname}`;
}
