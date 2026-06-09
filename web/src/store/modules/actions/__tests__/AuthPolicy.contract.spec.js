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

describe('AuthorizationPolicy action API contract', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetModules();
    global.sessionStorage = {
      getItem: vi.fn(() => 'token')
    };
  });

  it('sends create payload for authorization policies', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: authPolicyActions } = await import('../AuthPolicy.js');
    const commit = vi.fn();
    const rules = [{ from: [], to: [], when: [] }];
    const labels = [{ key: 'app', value: 'reviews' }];

    authPolicyActions.AuthPolicy_NewItem({ commit }, {
      name: 'allow-reviews',
      namespace: 'default',
      action: 'allow',
      rules,
      labels
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/security/authpolicies'),
      {
        name: 'allow-reviews',
        namespace: 'default',
        action: 'allow',
        rules,
        selectorMatchLabels: labels
      }
    );
  });

  it('filters empty selector labels before creating authorization policies', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: authPolicyActions } = await import('../AuthPolicy.js');
    const commit = vi.fn();
    const rules = [{ from: [], to: [], when: [] }];

    authPolicyActions.AuthPolicy_NewItem({ commit }, {
      name: 'allow-reviews',
      namespace: 'default',
      action: 'allow',
      rules,
      labels: [
        { key: '', value: '' },
        { key: 'app', value: 'reviews' },
        { key: 'version', value: '' }
      ]
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/security/authpolicies'),
      expect.objectContaining({
        selectorMatchLabels: [{ key: 'app', value: 'reviews' }]
      })
    );
  });

  it('sends update payload with resource version for authorization policies', async () => {
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: authPolicyActions } = await import('../AuthPolicy.js');
    const commit = vi.fn();
    const rules = [{ from: [], to: [], when: [] }];
    const labels = [{ key: 'app', value: 'reviews' }];

    authPolicyActions.AuthPolicy_UpdateItem({ commit }, {
      name: 'allow-reviews',
      namespace: 'default',
      action: 'deny',
      rules,
      labels,
      resourceVersion: 'rv-policy'
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/security/authpolicy/default/allow-reviews'),
      {
        action: 'deny',
        rules,
        selectorMatchLabels: labels,
        resourceversion: 'rv-policy'
      }
    );
  });

  it('commits update_conflict when authorization policy update returns HTTP 409', async () => {
    axiosMock.put.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { error: 'changed' }
      }
    });

    const { default: authPolicyActions } = await import('../AuthPolicy.js');
    const commit = vi.fn();

    authPolicyActions.AuthPolicy_UpdateItem({ commit }, {
      name: 'allow-reviews',
      namespace: 'default',
      action: 'allow',
      rules: [],
      labels: []
    });
    await flushPromises();

    expect(commit).toHaveBeenCalledWith('AuthPolicy_SetStatus', {
      status: 'update_conflict',
      error_handle: 'changed'
    });
  });

  it('deletes authorization policies by namespace and name', async () => {
    axiosMock.delete.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: authPolicyActions } = await import('../AuthPolicy.js');
    const commit = vi.fn();

    authPolicyActions.AuthPolicy_DelItem({ commit }, {
      name: 'allow-reviews',
      namespace: 'default'
    });
    await flushPromises();

    expect(axiosMock.delete).toHaveBeenCalledWith(
      expect.stringContaining('/security/authpolicy/default/allow-reviews')
    );
  });
});
