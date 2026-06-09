import { beforeEach, describe, expect, it, vi } from 'vitest';

const axiosMock = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
  patch: vi.fn(),
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

describe('Auth action API contract', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetModules();
    global.sessionStorage = {
      clear: vi.fn(),
      getItem: vi.fn(() => 'token'),
      setItem: vi.fn()
    };
  });

  it('trims credentials and stores token when sign-in succeeds', async () => {
    axiosMock.post.mockResolvedValueOnce({
      status: 200,
      data: {
        token: 'jwt-token',
        username: 'admin'
      }
    });

    const { default: authActions } = await import('../Auth.js');
    const commit = vi.fn();

    authActions.Auth_Signin({ commit }, {
      account: ' admin ',
      password: ' admin '
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/auth/signin'),
      {
        username: 'admin',
        password: 'admin'
      },
      {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        }
      }
    );
    expect(sessionStorage.clear).toHaveBeenCalled();
    expect(sessionStorage.setItem).toHaveBeenCalledWith('accessToken', 'jwt-token');
    expect(commit).toHaveBeenCalledWith('Auth_SetStatus', {
      status: 'signin_success',
      error_handle: ''
    });
  });

  it('loads namespaces with authentication header', async () => {
    const items = [
      { name: 'default', istioInjection: 'enabled', istioRevision: '' },
      { name: 'istio-system', istioInjection: 'disabled', istioRevision: '' }
    ];
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: {
        namespaces: ['default', 'istio-system'],
        items
      }
    });

    const { default: authActions } = await import('../Auth.js');
    const commit = vi.fn();

    const namespaces = await authActions.Auth_GetNamespaces({ commit });

    expect(axiosMock.defaults.headers.common.Authentication).toBe('token');
    expect(axiosMock.get).toHaveBeenCalledWith(expect.stringContaining('/namespaces'));
    expect(commit).toHaveBeenCalledWith('Auth_GetNamespaces', {
      namespaces: items
    });
    expect(namespaces).toEqual(['default', 'istio-system']);
  });

  it('loads cluster capabilities with authentication header', async () => {
    const capabilities = {
      istio: {
        installed: false,
        disabled: false,
        missingCRDs: ['gateways.networking.istio.io'],
        availableCRDs: [],
        message: 'Istio CRDs are missing'
      }
    };
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: capabilities
    });

    const { default: authActions } = await import('../Auth.js');
    const commit = vi.fn();

    const result = await authActions.Auth_GetClusterCapabilities({ commit });

    expect(axiosMock.defaults.headers.common.Authentication).toBe('token');
    expect(axiosMock.get).toHaveBeenCalledWith(expect.stringContaining('/cluster/capabilities'));
    expect(commit).toHaveBeenCalledWith('Auth_SetClusterCapabilities', capabilities);
    expect(result).toEqual(capabilities);
  });

  it('updates namespace Istio injection labels with JSON patch contract', async () => {
    axiosMock.patch.mockResolvedValueOnce({
      status: 200,
      data: {
        namespace: { name: 'demo', istioInjection: 'enabled' }
      }
    });

    const { default: authActions } = await import('../Auth.js');

    const result = await authActions.Auth_UpdateNamespaceInjection({}, {
      name: 'demo',
      mode: 'revision',
      revision: '1-22-0'
    });

    expect(axiosMock.patch).toHaveBeenCalledWith(
      expect.stringContaining('/namespace/demo/istio-injection'),
      {
        mode: 'revision',
        revision: '1-22-0',
        allowSystemNamespace: false
      },
      {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        }
      }
    );
    expect(result.namespace.name).toBe('demo');
  });

  it('URL-encodes namespace names when updating Istio injection', async () => {
    axiosMock.patch.mockResolvedValueOnce({
      status: 200,
      data: {
        namespace: { name: 'team/a', istioInjection: 'enabled' }
      }
    });

    const { default: authActions } = await import('../Auth.js');

    await authActions.Auth_UpdateNamespaceInjection({}, {
      name: 'team/a',
      mode: 'enabled'
    });

    expect(axiosMock.patch).toHaveBeenCalledWith(
      expect.stringContaining('/namespace/team%2Fa/istio-injection'),
      expect.objectContaining({
        mode: 'enabled',
        revision: '',
        allowSystemNamespace: false
      }),
      expect.any(Object)
    );
  });

  it('sets authentication header when updating namespace Istio injection', async () => {
    axiosMock.patch.mockResolvedValueOnce({
      status: 200,
      data: {
        namespace: { name: 'demo', istioInjection: 'enabled' }
      }
    });

    const { default: authActions } = await import('../Auth.js');

    await authActions.Auth_UpdateNamespaceInjection({}, {
      name: 'demo',
      mode: 'enabled'
    });

    expect(axiosMock.defaults.headers.common.Authentication).toBe('token');
  });

  it('clears namespace Istio revision when injection mode is not revision', async () => {
    axiosMock.patch.mockResolvedValueOnce({
      status: 200,
      data: {
        namespace: { name: 'demo', istioInjection: 'disabled', istioRevision: '' }
      }
    });

    const { default: authActions } = await import('../Auth.js');

    await authActions.Auth_UpdateNamespaceInjection({}, {
      name: 'demo',
      mode: 'disabled',
      revision: 'should-be-cleared'
    });

    expect(axiosMock.patch).toHaveBeenCalledWith(
      expect.stringContaining('/namespace/demo/istio-injection'),
      {
        mode: 'disabled',
        revision: '',
        allowSystemNamespace: false
      },
      expect.any(Object)
    );
  });

  it('sends explicit system namespace confirmation when requested', async () => {
    axiosMock.patch.mockResolvedValueOnce({
      status: 200,
      data: {
        namespace: { name: 'kube-system', systemNamespace: true }
      }
    });

    const { default: authActions } = await import('../Auth.js');

    await authActions.Auth_UpdateNamespaceInjection({}, {
      name: 'kube-system',
      mode: 'enabled',
      allowSystemNamespace: true
    });

    expect(axiosMock.patch).toHaveBeenCalledWith(
      expect.stringContaining('/namespace/kube-system/istio-injection'),
      {
        mode: 'enabled',
        revision: '',
        allowSystemNamespace: true
      },
      expect.any(Object)
    );
  });
});
