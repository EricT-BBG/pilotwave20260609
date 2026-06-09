import { describe, expect, it } from 'vitest';
import { buildApiUrl, resolveApiBaseConfig } from '../runtime';

describe('runtime helpers', () => {
  it('uses global location safely when no explicit location is provided', () => {
    const originalLocation = globalThis.location;
    const location = {
      origin: 'https://pilotwave.test',
      protocol: 'https:',
      host: 'pilotwave.test'
    };

    globalThis.location = location;

    try {
      expect(resolveApiBaseConfig(undefined, { VITE_API_URL: '' })).toEqual({
        apiUrl: 'https://pilotwave.test',
        apiVersion: '/api/v1'
      });
    } finally {
      globalThis.location = originalLocation;
    }
  });

  it('falls back to window location when the API URL is empty', () => {
    const location = {
      protocol: 'https:',
      host: 'pilotwave.test'
    };
    const env = {
      VITE_API_URL: '',
      VITE_API_VERSION: '/api/v1'
    };

    expect(resolveApiBaseConfig(location, env)).toEqual({
      apiUrl: 'https://pilotwave.test',
      apiVersion: '/api/v1'
    });
  });

  it('builds a stable API URL with the configured version', () => {
    const location = {
      protocol: 'http:',
      host: 'localhost:8080'
    };

    const env = {
      VITE_API_URL: '//',
      VITE_API_VERSION: '/api/v1'
    };

    expect(buildApiUrl('/routers', location, env)).toBe('http://localhost:8080/api/v1/routers');
  });
});
