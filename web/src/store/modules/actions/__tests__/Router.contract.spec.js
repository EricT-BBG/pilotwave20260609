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

describe('Router action API contract', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetModules();
    global.sessionStorage = {
      getItem: vi.fn(() => 'token')
    };
  });

  it('sends router create payload', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    routerActions.Router_NewItem({ commit }, {
      name: 'rt-a',
      namespace: 'default',
      protocol: 'http',
      hosts: ['app.example.local']
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/routers'),
      {
        name: 'rt-a',
        namespace: 'default',
        protocol: 'http',
        hosts: ['app.example.local']
      }
    );
  });

  it('creates router gateway mappings after router create when gateways are selected', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    routerActions.Router_NewItem({ commit }, {
      name: 'rt-a',
      namespace: 'default',
      protocol: 'http',
      hosts: ['app.example.local'],
      gateways: [{ name: 'gw-a', namespace: 'istio-system' }]
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/router/default/rt-a/gateways'),
      {
        gateways: [{ name: 'gw-a', namespace: 'istio-system' }],
        resourceversion: ''
      }
    );
    expect(commit).toHaveBeenCalledWith('Router_SetStatus', {
      status: 'create_success',
      error_handle: ''
    });
  });

  it('injects namespace query param when loading routers for a namespace', async () => {
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: {
        routers: [],
        meta: { page: 1, limit: 20, total: 0 }
      }
    });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    await routerActions.Router_GetItems({ commit }, {
      namespace: 'bookinfo',
      page: 3,
      limit: 10
    });

    expect(axiosMock.get).toHaveBeenCalledWith(
      expect.stringContaining('/routers'),
      {
        params: {
          page: 3,
          limit: 10,
          namespace: 'bookinfo'
        }
      }
    );
  });

  it('omits namespace query param when loading routers across all namespaces', async () => {
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: {
        routers: [],
        meta: { page: 1, limit: 20, total: 0 }
      }
    });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    await routerActions.Router_GetItems({ commit }, {
      namespace: 'All'
    });

    expect(axiosMock.get).toHaveBeenCalledWith(
      expect.stringContaining('/routers'),
      {
        params: {
          page: 1,
          limit: 20
        }
      }
    );
  });

  it('coalesces duplicate in-flight success rate requests', async () => {
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: {
        metrics: [{ timestamp: 1, value: 100 }],
        successRate: 100,
        totalSuccessReqests: 10,
        totalReqests: 10
      }
    });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();
    const payload = {
      name: 'hello-tls',
      namespace: 'pilotwave-istio-smoke',
      startTime: 1779465600,
      endTime: 1779552000,
      interval: '1h'
    };

    const [first, second] = await Promise.all([
      routerActions.Router_GetSuccessRate({ commit }, payload),
      routerActions.Router_GetSuccessRate({ commit }, payload)
    ]);

    expect(axiosMock.get).toHaveBeenCalledTimes(1);
    expect(first).toEqual(second);
    expect(commit).toHaveBeenCalledTimes(2);
  });

  it('sends router update payload with resource version', async () => {
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    routerActions.Router_UpdateItem({ commit }, {
      name: 'rt-a',
      namespace: 'default',
      protocol: 'http',
      hosts: ['app.example.local'],
      resourceVersion: 'rv-router'
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/router/default/rt-a'),
      {
        protocol: 'http',
        hosts: ['app.example.local'],
        resourceversion: 'rv-router'
      }
    );
  });

  it('sends router rules payload with resource version', async () => {
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();
    const httpItems = [{
      prefixs: ['/api'],
      headers: [],
      rewrite: '',
      fixedDelay: 0,
      timeout: 30,
      destinations: [{
        host: 'svc.default.svc.cluster.local',
        port: 80,
        weight: 100,
        subset: ''
      }]
    }];

    routerActions.Router_UpdateRules({ commit }, {
      name: 'rt-a',
      namespace: 'default',
      httpItems,
      resourceVersion: 'rv-rules'
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/router/default/rt-a/rules'),
      {
        name: 'rt-a',
        namespace: 'default',
        https: httpItems,
        resourceversion: 'rv-rules'
      }
    );
  });

  it('sends gateway mappings payload with resource version', async () => {
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    routerActions.Router_MappingGateways({ commit }, {
      name: 'rt-a',
      namespace: 'default',
      gateways: [{ name: 'gw-a', namespace: 'istio-system' }],
      resourceVersion: 'rv-mapping'
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/router/default/rt-a/gateways'),
      {
        gateways: [{ name: 'gw-a', namespace: 'istio-system' }],
        resourceversion: 'rv-mapping'
      }
    );
  });

  it('commits update_conflict when router update returns HTTP 409', async () => {
    axiosMock.put.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { error: 'changed' }
      }
    });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    routerActions.Router_UpdateItem({ commit }, {
      name: 'rt-a',
      namespace: 'default',
      protocol: 'http',
      hosts: []
    });
    await flushPromises();

    expect(commit).toHaveBeenCalledWith('Router_SetStatus', {
      status: 'update_conflict',
      error_handle: 'changed'
    });
  });

  it('commits update_conflict when router rules return HTTP 409', async () => {
    axiosMock.put.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { error: 'changed' }
      }
    });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    routerActions.Router_UpdateRules({ commit }, {
      name: 'rt-a',
      namespace: 'default',
      httpItems: []
    });
    await flushPromises();

    expect(commit).toHaveBeenCalledWith('Router_SetStatus', {
      status: 'update_conflict',
      error_handle: 'changed'
    });
  });

  it('commits update_conflict when router mapping returns HTTP 409', async () => {
    axiosMock.put.mockRejectedValueOnce({
      response: {
        status: 409,
        data: { error: 'changed' }
      }
    });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    routerActions.Router_MappingGateways({ commit }, {
      name: 'rt-a',
      namespace: 'default',
      gateways: []
    });
    await flushPromises();

    expect(commit).toHaveBeenCalledWith('Router_SetStatus', {
      status: 'update_conflict',
      error_handle: 'changed'
    });
  });

  it('deletes routers by namespace and name', async () => {
    axiosMock.delete.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    routerActions.Router_DelItem({ commit }, {
      name: 'rt-a',
      namespace: 'default'
    });
    await flushPromises();

    expect(axiosMock.delete).toHaveBeenCalledWith(
      expect.stringContaining('/router/default/rt-a')
    );
  });

  it('deletes router gateway mappings by namespace and router name', async () => {
    axiosMock.delete.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    routerActions.Router_DelMapping({ commit }, {
      name: 'rt-a',
      namespace: 'default',
      gatewayId: 'istio-system/gw-a'
    });
    await flushPromises();

    expect(axiosMock.delete).toHaveBeenCalledWith(
      expect.stringContaining('/router/default/rt-a/gateways'),
      {
        params: {
          gatewayId: 'istio-system/gw-a'
        }
      }
    );
  });

  it('sends monitoring source TLS verification contract', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: { id: 'grafana-id' } });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    await routerActions.Router_UpdateGrafana({ commit }, {
      id: 'grafana-id',
      provider: 'grafana',
      host: 'grafana.monitoring.svc',
      port: '443',
      token: 'token',
      datasourceId: '1',
      isTls: true,
      skipTlsVerify: true
    });

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/grafanas'),
      {
        id: 'grafana-id',
        provider: 'grafana',
        host: 'grafana.monitoring.svc',
        port: '443',
        token: 'token',
        datasourceId: '1',
        isTls: true,
        skipTlsVerify: true
      }
    );
  });

  it('uses a short timeout for monitoring source tests', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: { ok: true, message: 'ok' } });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    await routerActions.Router_TestGrafana({ commit }, {
      provider: 'grafana',
      host: 'grafana.monitoring.svc',
      port: '80',
      token: '',
      datasourceId: '1',
      isTls: false,
      skipTlsVerify: false
    });

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/monitoring/test'),
      {
        provider: 'grafana',
        host: 'grafana.monitoring.svc',
        port: '80',
        token: '',
        datasourceId: '1',
        isTls: false,
        skipTlsVerify: false
      },
      { timeout: 8000 }
    );
  });

  it('preserves monitoring configured flag from grafana response', async () => {
    axiosMock.get.mockResolvedValueOnce({
      status: 200,
      data: {
        configured: false,
        provider: 'grafana',
        host: '192.168.1.196',
        port: '80',
        datasourceId: '1'
      }
    });

    const { default: routerActions } = await import('../Router.js');
    const commit = vi.fn();

    const result = await routerActions.Router_GetGrafana({ commit });

    expect(result.grafana.configured).toBe(false);
    expect(commit).toHaveBeenCalledWith('Router_GetGrafana', {
      grafana: {
        configured: false,
        provider: 'grafana',
        host: '192.168.1.196',
        port: '80',
        datasourceId: '1'
      }
    });
  });
});
