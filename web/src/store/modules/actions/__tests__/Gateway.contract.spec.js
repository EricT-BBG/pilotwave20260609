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

describe('Gateway action API contract', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetModules();
    global.sessionStorage = {
      getItem: vi.fn(() => 'token')
    };
    global.window = {
      btoa: (value) => Buffer.from(value, 'binary').toString('base64')
    };
  });

  it('sends create payload with base64 TLS material', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_NewItem({ commit }, {
      name: 'gw-a',
      namespace: 'default',
      servers: [{
        hosts: ['app.example.local'],
        ports: [{
          protocol: 'TLS',
          port: 15443,
          mode: 'MUTUAL',
          cert: 'plain-cert',
          pkey: 'plain-key',
          cacert: 'plain-ca'
        }]
      }]
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/gateways'),
      {
        name: 'gw-a',
        namespace: 'default',
        servers: [{
          hosts: ['app.example.local'],
          ports: [{
            protocol: 'TLS',
            port: 15443,
            mode: 'MUTUAL',
            cert: 'cGxhaW4tY2VydA==',
            pkey: 'cGxhaW4ta2V5',
            cacert: 'cGxhaW4tY2E='
          }]
        }]
      }
    );
  });

  it('sends selector labels when creating a gateway', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_NewItem({ commit }, {
      name: 'gw-custom',
      namespace: 'default',
      selectorMatchLabels: {
        app: 'custom-ingress'
      },
      servers: [{
        hosts: ['app.example.local'],
        ports: [{
          protocol: 'HTTP',
          port: 80
        }]
      }]
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/gateways'),
      expect.objectContaining({
        selectormatchlabels: {
          app: 'custom-ingress'
        }
      })
    );
  });

  it('leaves non-TLS gateway material unencoded when creating a gateway', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_NewItem({ commit }, {
      name: 'gw-http',
      namespace: 'default',
      servers: [{
        hosts: ['app.example.local'],
        ports: [{
          protocol: 'HTTP',
          port: 80,
          cert: 'plain-cert',
          pkey: 'plain-key'
        }]
      }]
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/gateways'),
      expect.objectContaining({
        servers: [{
          hosts: ['app.example.local'],
          ports: [{
            protocol: 'HTTP',
            port: 80,
            cert: 'plain-cert',
            pkey: 'plain-key'
          }]
        }]
      })
    );
  });

  it('injects namespace query param when loading gateways for a namespace', async () => {
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: {
        gateways: [],
        meta: { page: 1, limit: 20, total: 0 }
      }
    });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    await gatewayActions.Gateway_GetItems({ commit }, {
      namespace: 'bookinfo',
      page: 2,
      limit: 50
    });

    expect(axiosMock.get).toHaveBeenCalledWith(
      expect.stringContaining('/gateways'),
      {
        params: {
          page: 2,
          limit: 50,
          namespace: 'bookinfo'
        }
      }
    );
  });

  it('omits namespace query param when loading gateways across all namespaces', async () => {
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: {
        gateways: [],
        meta: { page: 1, limit: 20, total: 0 }
      }
    });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    await gatewayActions.Gateway_GetItems({ commit }, {
      namespace: 'All'
    });

    expect(axiosMock.get).toHaveBeenCalledWith(
      expect.stringContaining('/gateways'),
      {
        params: {
          page: 1,
          limit: 20
        }
      }
    );
  });

  it('sends resource version, selector labels, and base64 TLS material when updating a gateway', async () => {
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_UpdateItem({ commit }, {
      name: 'gw-a',
      namespace: 'default',
      resourceVersion: '123',
      selectorMatchLabels: {
        istio: 'ingressgateway',
        app: 'custom-ingress'
      },
      servers: [{
        hosts: ['app.example.local'],
        ports: [{
          protocol: 'HTTPS',
          port: 443,
          mode: 'SIMPLE',
          cert: 'plain-cert',
          pkey: 'plain-key'
        }]
      }]
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/gateway/default/gw-a'),
      {
        servers: [{
          hosts: ['app.example.local'],
          ports: [{
            protocol: 'HTTPS',
            port: 443,
            mode: 'SIMPLE',
            cert: 'cGxhaW4tY2VydA==',
            pkey: 'cGxhaW4ta2V5'
          }]
        }],
        selectormatchlabels: {
          istio: 'ingressgateway',
          app: 'custom-ingress'
        },
        resourceversion: '123'
      }
    );
  });

  it('sends router mappings with virtual service resource versions', async () => {
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_MappingRouters({ commit }, {
      name: 'gw-a',
      namespace: 'default',
      routers: [{ name: 'router-a', namespace: 'default' }],
      resourceVersions: {
        'default/router-a': 'rv-router-a'
      }
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/gateway/default/gw-a/routers'),
      {
        routers: [{ name: 'router-a', namespace: 'default' }],
        resourceversions: {
          'default/router-a': 'rv-router-a'
        }
      }
    );
  });

  it('blocks mTLS uploads without a CA bundle', async () => {
    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_NewItem({ commit }, {
      name: 'gw-mtls',
      namespace: 'default',
      servers: [{
        hosts: ['app.example.local'],
        ports: [{
          protocol: 'HTTPS',
          port: 443,
          mode: 'MUTUAL',
          cert: 'plain-cert',
          pkey: 'plain-key'
        }]
      }]
    });
    await flushPromises();

    expect(axiosMock.post).not.toHaveBeenCalled();
    expect(commit).toHaveBeenCalledWith('Gateway_SetStatus', {
      status: 'create_error',
      error_handle: 'Port 443 requires a CA bundle for mTLS.'
    });
  });

  it('blocks TLS uploads without a private key', async () => {
    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_NewItem({ commit }, {
      name: 'gw-missing-key',
      namespace: 'default',
      servers: [{
        hosts: ['app.example.local'],
        ports: [{
          protocol: 'HTTPS',
          port: 443,
          mode: 'SIMPLE',
          cert: 'plain-cert'
        }]
      }]
    });
    await flushPromises();

    expect(axiosMock.post).not.toHaveBeenCalled();
    expect(commit).toHaveBeenCalledWith('Gateway_SetStatus', {
      status: 'create_error',
      error_handle: 'Port 443 requires a private key.'
    });
  });

  it('sends existing Secret TLS ports without PEM material', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_NewItem({ commit }, {
      name: 'gw-existing',
      namespace: 'default',
      servers: [{
        hosts: ['app.example.local'],
        ports: [{
          protocol: 'HTTPS',
          port: 443,
          mode: 'MUTUAL',
          credentialname: 'existing-mtls',
          cert: '',
          pkey: '',
          cacert: ''
        }]
      }]
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/gateways'),
      {
        name: 'gw-existing',
        namespace: 'default',
        servers: [{
          hosts: ['app.example.local'],
          ports: [{
            protocol: 'HTTPS',
            port: 443,
            mode: 'MUTUAL',
            credentialname: 'existing-mtls',
            cert: '',
            pkey: '',
            cacert: ''
          }]
        }]
      }
    );
  });

  it('loads gateway TLS certificate status without private key material', async () => {
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: {
        certificates: [{
          credentialName: 'pilotwave-app-tls',
          secretNamespace: 'istio-system',
          secretName: 'pilotwave-app-tls',
          status: 'warning',
          daysUntilExpiry: 18,
          notAfter: '2026-06-10T00:00:00Z',
          fingerprintSHA256: 'AA:BB',
        }]
      }
    });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    const certificates = await gatewayActions.Gateway_GetTLSCertificates({ commit }, {
      name: 'gw-a',
      namespace: 'default',
    });

    expect(axiosMock.get).toHaveBeenCalledWith(
      expect.stringContaining('/gateway/default/gw-a/tls-certificates')
    );
    expect(certificates[0]).toMatchObject({
      credentialName: 'pilotwave-app-tls',
      secretNamespace: 'istio-system',
      secretName: 'pilotwave-app-tls',
    });
    expect(certificates[0]).not.toHaveProperty('cert');
    expect(certificates[0]).not.toHaveProperty('pkey');
    expect(certificates[0]).not.toHaveProperty('privateKey');
    expect(certificates[0]).not.toHaveProperty('serverCertificate');
    expect(commit).toHaveBeenCalledWith('Gateway_GetTLSCertificates', {
      certificates,
    });
  });

  it('checks whether an existing TLS Secret exists', async () => {
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: {
        exists: true,
        secretNamespace: 'istio-system',
        secretName: 'existing-mtls'
      }
    });

    const { default: gatewayActions } = await import('../Gateway.js');

    const result = await gatewayActions.Gateway_CheckTLSSecretExists({}, {
      credentialname: 'existing-mtls'
    });

    expect(axiosMock.get).toHaveBeenCalledWith(
      expect.stringContaining('/gateway/tls-secret/exists'),
      {
        params: {
          credentialname: 'existing-mtls'
        }
      }
    );
    expect(result).toMatchObject({
      exists: true,
      secretNamespace: 'istio-system',
      secretName: 'existing-mtls'
    });
  });

  it('commits update_conflict when gateway update returns HTTP 409', async () => {
    axiosMock.put.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { error: 'changed' }
      }
    });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_UpdateItem({ commit }, {
      name: 'gw-a',
      namespace: 'default',
      servers: []
    });
    await flushPromises();

    expect(commit).toHaveBeenCalledWith('Gateway_SetStatus', {
      status: 'update_conflict',
      error_handle: 'changed'
    });
  });

  it('commits update_conflict when gateway mapping returns HTTP 409', async () => {
    axiosMock.put.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { error: 'changed' }
      }
    });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_MappingRouters({ commit }, {
      name: 'gw-a',
      namespace: 'default',
      routers: []
    });
    await flushPromises();

    expect(commit).toHaveBeenCalledWith('Gateway_SetStatus', {
      status: 'update_conflict',
      error_handle: 'changed'
    });
  });

  it('deletes gateways by namespace and name', async () => {
    axiosMock.delete.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: gatewayActions } = await import('../Gateway.js');
    const commit = vi.fn();

    gatewayActions.Gateway_DelItem({ commit }, {
      name: 'gw-a',
      namespace: 'default'
    });
    await flushPromises();

    expect(axiosMock.delete).toHaveBeenCalledWith(
      expect.stringContaining('/gateway/default/gw-a')
    );
  });
});
