import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it, vi } from 'vitest';

import GatewayRouterSetting from '../gateway/RouterSetting.vue';
import RouterGatewaySetting from '../router/GatewaySetting.vue';

const currentDir = dirname(fileURLToPath(import.meta.url));
const routerGatewaySource = readFileSync(resolve(currentDir, '../router/GatewaySetting.vue'), 'utf8');
const gatewayRouterSource = readFileSync(resolve(currentDir, '../gateway/RouterSetting.vue'), 'utf8');

describe('Router and Gateway association checkbox lists', () => {
  it('renders router gateway association as a checkbox list with selected state affordances', () => {
    expect(routerGatewaySource).not.toContain('v-model="selectedGateways" multiple');
    expect(routerGatewaySource).toContain('class="association-list"');
    expect(routerGatewaySource).toContain('data-testid="router-gateway-association-list"');
    expect(routerGatewaySource).toContain('type="checkbox"');
    expect(routerGatewaySource).toContain('association-row--selected');
    expect(routerGatewaySource).toContain("$t('Form.Selected')");
    expect(routerGatewaySource).toContain('selectedGateways.length');
  });

  it('renders gateway router association as a checkbox list with selected state affordances', () => {
    expect(gatewayRouterSource).not.toContain('v-model="selectedRouters" multiple');
    expect(gatewayRouterSource).toContain('class="association-list"');
    expect(gatewayRouterSource).toContain('data-testid="gateway-router-association-list"');
    expect(gatewayRouterSource).toContain('type="checkbox"');
    expect(gatewayRouterSource).toContain('association-row--selected');
    expect(gatewayRouterSource).toContain("$t('Form.Selected')");
    expect(gatewayRouterSource).toContain('selectedRouters.length');
  });

  it('separates association filters and actions from the full-width checkbox list', () => {
    [routerGatewaySource, gatewayRouterSource].forEach((source) => {
      expect(source).toContain('class="association-toolbar"');
      expect(source).toContain('class="association-delta-summary"');
      expect(source).toContain('class="association-list-wrap"');
      expect(source).toContain('grid-template-columns: minmax(220px, 320px) minmax(0, 1fr) auto auto;');
      expect(source).not.toContain('grid-template-columns: minmax(160px, 220px) minmax(260px, 1fr) auto;');
    });
  });

  it('labels existing, added, and removed association changes before submit', () => {
    [routerGatewaySource, gatewayRouterSource].forEach((source) => {
      expect(source).toContain('initialSelected');
      expect(source).toContain('associationState');
      expect(source).toContain('associationBadge');
      expect(source).toContain("$t('Form.Existing')");
      expect(source).toContain("$t('Form.Added')");
      expect(source).toContain("$t('Form.Removed')");
      expect(source).toContain('association-row--added');
      expect(source).toContain('association-row--removed');
    });
  });

  it('keeps association checkbox editing behind an explicit edit action', () => {
    [routerGatewaySource, gatewayRouterSource].forEach((source) => {
      expect(source).toContain('associationEditMode: false');
      expect(source).toContain('openAssociationEdit');
      expect(source).toContain('closeAssociationEdit');
      expect(source).toContain('association-readonly-toolbar');
      expect(source).toContain('v-if="associationEditMode"');
      expect(source).not.toContain('<span>{{ $t(\'Form.Selected\') }}: {{ mappings.length }}</span>');
    });
  });

  it('keeps router gateway submit payload based on selected gateway values', () => {
    const dispatch = vi.fn();
    const commit = vi.fn();
    const state = {
      name: 'reviews',
      selectedGateways: ['edge-gw,istio-system', 'internal-gw,default'],
      mappingResourceVersion: '123',
      $route: { query: { namespace: 'bookinfo' } },
      $store: { commit, dispatch },
    };

    RouterGatewaySetting.methods.submit.call(state);

    expect(commit).toHaveBeenCalledWith('Router_ResetStatus');
    expect(dispatch).toHaveBeenCalledWith('Router_MappingGateways', {
      name: 'reviews',
      namespace: 'bookinfo',
      gateways: [
        { name: 'edge-gw', namespace: 'istio-system' },
        { name: 'internal-gw', namespace: 'default' },
      ],
      resourceVersion: '123',
    });
  });

  it('preselects router gateway mappings returned without option values', async () => {
    const state = {
      name: 'reviews',
      selectedGateways: [],
      gateways: [],
      $route: { query: { namespace: 'bookinfo' } },
      $store: {
        dispatch: vi.fn()
          .mockResolvedValueOnce([])
          .mockResolvedValueOnce([
            { name: 'hello-gateway', namespace: 'pilotwave-istio-smoke' },
          ]),
      },
    };
    state.mappingValue = RouterGatewaySetting.methods.mappingValue.bind(state);

    await RouterGatewaySetting.methods.fetchMapping.call(state);

    expect(state.selectedGateways).toEqual(['hello-gateway,pilotwave-istio-smoke']);
  });

  it('keeps gateway router submit payload based on selected router values', () => {
    const dispatch = vi.fn();
    const commit = vi.fn();
    const state = {
      name: 'edge-gw',
      selectedRouters: ['reviews,bookinfo', 'ratings,bookinfo'],
      mappingResourceVersions: { reviews: '321' },
      $route: { query: { namespace: 'istio-system' } },
      $store: { commit, dispatch },
    };

    GatewayRouterSetting.methods.submit.call(state);

    expect(commit).toHaveBeenCalledWith('Gateway_ResetStatus');
    expect(dispatch).toHaveBeenCalledWith('Gateway_MappingRouters', {
      name: 'edge-gw',
      namespace: 'istio-system',
      routers: [
        { name: 'reviews', namespace: 'bookinfo' },
        { name: 'ratings', namespace: 'bookinfo' },
      ],
      resourceVersions: { reviews: '321' },
    });
  });

  it('preselects gateway router mappings returned without option values', async () => {
    const state = {
      name: 'edge-gw',
      selectedRouters: [],
      routers: [],
      $route: { query: { namespace: 'istio-system' } },
      $store: {
        dispatch: vi.fn()
          .mockResolvedValueOnce([])
          .mockResolvedValueOnce([
            { name: 'hello-tls', namespace: 'pilotwave-istio-smoke' },
          ]),
      },
    };
    state.mappingValue = GatewayRouterSetting.methods.mappingValue.bind(state);

    await GatewayRouterSetting.methods.fetchMapping.call(state);

    expect(state.selectedRouters).toEqual(['hello-tls,pilotwave-istio-smoke']);
  });
});
