import { beforeEach, describe, expect, it, vi } from 'vitest';

const axiosMock = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
  defaults: {
    headers: {
      common: {},
      post: {}
    }
  }
};

vi.mock('axios', () => ({
  default: axiosMock
}));

const flushPromises = () => new Promise((resolve) => setTimeout(resolve, 0));

describe('RequestAuthentication action API contract', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetModules();
    global.sessionStorage = {
      getItem: vi.fn(() => 'token')
    };
  });

  it('sends create payload for request authentications', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: authRequestActions } = await import('../AuthRequest.js');
    const commit = vi.fn();
    const rules = [{ issuer: 'issuer-a', jwksUri: 'https://issuer-a/jwks', audiences: ['app'] }];
    const labels = [{ key: 'app', value: 'reviews' }];

    authRequestActions.AuthRequest_NewItem({ commit }, {
      name: 'jwt-reviews',
      namespace: 'default',
      rules,
      labels
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/security/requestauths'),
      {
        name: 'jwt-reviews',
        namespace: 'default',
        jwtRules: rules,
        selectorMatchLabels: labels
      }
    );
  });

  it('filters empty selector labels before creating request authentications', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: authRequestActions } = await import('../AuthRequest.js');
    const commit = vi.fn();
    const rules = [{ issuer: 'issuer-a', jwksUri: 'https://issuer-a/jwks', audiences: ['app'] }];

    authRequestActions.AuthRequest_NewItem({ commit }, {
      name: 'jwt-reviews',
      namespace: 'default',
      rules,
      labels: [
        { key: '', value: '' },
        { key: 'app', value: 'reviews' },
        { key: 'version', value: '' }
      ]
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/security/requestauths'),
      expect.objectContaining({
        selectorMatchLabels: [{ key: 'app', value: 'reviews' }]
      })
    );
  });

  it('sends update payload with resource version for request authentications', async () => {
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: authRequestActions } = await import('../AuthRequest.js');
    const commit = vi.fn();
    const rules = [{ issuer: 'issuer-a', jwksUri: 'https://issuer-a/jwks', audiences: ['app'] }];
    const labels = [{ key: 'app', value: 'reviews' }];

    authRequestActions.AuthRequest_UpdateItem({ commit }, {
      name: 'jwt-reviews',
      namespace: 'default',
      rules,
      labels,
      resourceVersion: 'rv-request'
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/security/requestauth/default/jwt-reviews'),
      {
        jwtRules: rules,
        selectorMatchLabels: labels,
        resourceversion: 'rv-request'
      }
    );
  });

  it('commits update_conflict when request authentication update returns HTTP 409', async () => {
    axiosMock.put.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { error: 'changed' }
      }
    });

    const { default: authRequestActions } = await import('../AuthRequest.js');
    const commit = vi.fn();

    authRequestActions.AuthRequest_UpdateItem({ commit }, {
      name: 'jwt-reviews',
      namespace: 'default',
      rules: [],
      labels: []
    });
    await flushPromises();

    expect(commit).toHaveBeenCalledWith('AuthRequest_SetStatus', {
      status: 'update_conflict',
      error_handle: 'changed'
    });
  });

  it('deletes request authentications by namespace and name', async () => {
    axiosMock.delete.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: authRequestActions } = await import('../AuthRequest.js');
    const commit = vi.fn();

    authRequestActions.AuthRequest_DelItem({ commit }, {
      name: 'jwt-reviews',
      namespace: 'default'
    });
    await flushPromises();

    expect(axiosMock.delete).toHaveBeenCalledWith(
      expect.stringContaining('/security/requestauth/default/jwt-reviews')
    );
  });
});
