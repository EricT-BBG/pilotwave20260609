import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it, vi } from 'vitest';

import Template from '../../components/Template.vue';
import Home from '../Home.vue';

const currentDir = dirname(fileURLToPath(import.meta.url));
const homeSource = readFileSync(resolve(currentDir, '../Home.vue'), 'utf8');

describe('dashboard namespace scope', () => {
  it('loads dashboard routers across all namespaces', async () => {
    const dispatch = vi.fn((action) => {
      if (action === 'Router_GetGrafana') {
        return Promise.resolve({ configured: false });
      }
      return Promise.resolve();
    });
    const context = {
      $store: { dispatch },
      $t: (key) => key,
      routers: [{ name: 'hello-tls', namespace: 'pilotwave-istio-smoke' }],
      filteredRouters: [{ name: 'hello-tls', namespace: 'pilotwave-istio-smoke' }],
      selectedRouterKey: '',
      selectedNamespace: 'All',
      dName: '',
      dNamespace: '',
      metricError: '',
      metricsLoaded: false,
      monitoringConfigured: false,
      routerKey: Home.methods.routerKey,
      syncRouter: Home.methods.syncRouter,
      fetchTime: Home.methods.fetchTime,
    };

    await Home.methods.fetchData.call(context);

    expect(dispatch).toHaveBeenCalledWith('Router_GetMenuItems', expect.not.objectContaining({
      namespace: expect.anything(),
    }));
    expect(context.selectedRouterKey).toBe('pilotwave-istio-smoke/hello-tls');
    expect(context.dNamespace).toBe('pilotwave-istio-smoke');
  });

  it('uses the selected router namespace for dashboard metrics', async () => {
    const dispatch = vi.fn((action) => {
      if (action === 'Router_GetGrafana') {
        return Promise.resolve({ configured: true });
      }
      return Promise.resolve({ ok: true });
    });
    const context = {
      $store: { dispatch },
      $t: (key) => key,
      dName: 'hello-tls',
      dNamespace: 'pilotwave-istio-smoke',
      metricError: '',
      metricsLoaded: false,
      monitoringConfigured: true,
    };

    await Home.methods.fetchTime.call(context);

    expect(dispatch).toHaveBeenCalledWith('Router_GetSuccessRate', expect.objectContaining({
      name: 'hello-tls',
      namespace: 'pilotwave-istio-smoke',
    }));
    expect(dispatch).toHaveBeenCalledWith('Router_GetLatency', expect.objectContaining({
      name: 'hello-tls',
      namespace: 'pilotwave-istio-smoke',
    }));
    expect(dispatch).toHaveBeenCalledWith('Router_GetOPS', expect.objectContaining({
      name: 'hello-tls',
      namespace: 'pilotwave-istio-smoke',
    }));
  });

  it('filters dashboard router choices by a local namespace selector', async () => {
    const context = {
      selectedNamespace: 'pilotwave-istio-smoke',
      selectedRouterKey: 'default/default-route',
      dName: '',
      dNamespace: '',
      filteredRouters: [{ name: 'hello-tls', namespace: 'pilotwave-istio-smoke' }],
      routerKey: Home.methods.routerKey,
      syncRouter: Home.methods.syncRouter,
      fetchTime: vi.fn(),
    };

    await Home.methods.updateNamespace.call(context);

    expect(context.selectedRouterKey).toBe('pilotwave-istio-smoke/hello-tls');
    expect(context.dNamespace).toBe('pilotwave-istio-smoke');
    expect(context.fetchTime).toHaveBeenCalled();
  });

  it('does not render a dashboard More Details link', () => {
    expect(homeSource).not.toContain('More Details');
    expect(homeSource).not.toContain('grafanaHref');
  });

  it('does not bind dashboard state to the global namespace getter or watcher', () => {
    expect(Home.computed.namespace).toBeUndefined();
    expect(Home.watch?.namespace).toBeUndefined();
  });

  it('hides the topbar namespace picker only on the dashboard route', () => {
    expect(Template.computed.showNamespacePicker.call({
      $route: { meta: { hideNamespacePicker: true } },
    })).toBe(false);
    expect(Template.computed.showNamespacePicker.call({
      $route: { meta: {} },
    })).toBe(true);
  });
});
