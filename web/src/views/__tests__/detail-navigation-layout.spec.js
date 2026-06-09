import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const srcRoot = resolve(import.meta.dirname, '../..');

const forbiddenDetailPatterns = [
  ['views/gateway/GatewayDetail.vue', "← {{ $t('Form.Cancel') }}"],
  ['views/router/RouterDetail.vue', "← {{ $t('Form.Cancel') }}"],
  ['views/auth/RequestauthDetail.vue', "← {{ $t('Form.Cancel') }}"],
  ['views/policy/AuthpolicyDetail.vue', "← {{ $t('Form.Cancel') }}"],
  ['views/user/UserDetail.vue', "{{ $t('Form.Cancel') }}"],
  ['views/gateway/NewGateway.vue', "← {{ $t('Form.Cancel') }}"],
  ['views/router/NewRouter.vue', "← {{ $t('Form.Cancel') }}"],
  ['views/auth/NewRequestauth.vue', "← {{ $t('Form.Cancel') }}"],
  ['views/policy/NewAuthpolicy.vue', "← {{ $t('Form.Cancel') }}"],
  ['views/user/NewUser.vue', "<button class=\"secondary-button\" type=\"button\" @click=\"goBack\">\n        {{ $t('Form.Cancel') }}\n      </button>\n    </header>"],
  ['views/gateway/GatewayDetail.vue', '<h1>{{ name }} <span>{{ namespace }}</span></h1>'],
  ['views/router/RouterDetail.vue', '<h1>{{ name }} <span>{{ namespace }}</span></h1>'],
  ['components/router/Cytoscape.vue', "label: 'Router\\n' + this.name"],
  ['components/router/Cytoscape.vue', "label: 'GATEWAY'"],
  ['components/router/Cytoscape.vue', "label: 'SERVICE'"],
  ['components/router/Cytoscape.vue', "label: '(Host)\\n'"],
  ['components/router/Cytoscape.vue', "label: '(Destination)\\n'"],
];

describe('detail navigation and relationship layout', () => {
  it('uses back-oriented navigation copy and separates resource metadata from titles', () => {
    const remaining = forbiddenDetailPatterns.filter(([file, pattern]) => {
      const source = readFileSync(resolve(srcRoot, file), 'utf8');
      return source.includes(pattern);
    });

    expect(remaining).toEqual([]);
  });

  it('uses compact inline headers for router and gateway detail pages', () => {
    const detailFiles = [
      ['views/router/RouterDetail.vue', 'Form.BackToRouters'],
      ['views/gateway/GatewayDetail.vue', 'Form.BackToGateways'],
      ['views/auth/RequestauthDetail.vue', 'Form.BackToAPIAuthentication'],
      ['views/policy/AuthpolicyDetail.vue', 'Form.BackToPolicies'],
      ['views/user/UserDetail.vue', 'Form.BackToUsers'],
    ];

    detailFiles.forEach(([file, backCopy]) => {
      const source = readFileSync(resolve(srcRoot, file), 'utf8');

      expect(source).toContain('detail-header--compact');
      expect(source).toContain('detail-header-title');
      expect(source).toContain('detail-title-card');
      if (file === 'views/user/UserDetail.vue') {
        expect(source).not.toContain('namespace-chip');
      } else {
        expect(source).toContain('namespace-chip');
      }
      expect(source).toContain(backCopy);
      expect(source).not.toContain('detail-meta');
      expect(source).not.toContain("Form.Cancel");
    });
  });

  it('uses breadcrumb-style compact headers for create pages', () => {
    const createFiles = [
      ['views/gateway/NewGateway.vue', 'Form.BackToGateways'],
      ['views/router/NewRouter.vue', 'Form.BackToRouters'],
      ['views/auth/NewRequestauth.vue', 'Form.BackToAPIAuthentication'],
      ['views/policy/NewAuthpolicy.vue', 'Form.BackToPolicies'],
      ['views/user/NewUser.vue', 'Form.BackToUsers'],
    ];

    createFiles.forEach(([file, backCopy]) => {
      const source = readFileSync(resolve(srcRoot, file), 'utf8');

      expect(source).toContain('detail-header--compact');
      expect(source).toContain('detail-title-card');
      expect(source).toContain('detail-resource-type');
      expect(source).toContain('detail-header-title');
      expect(source).toContain(backCopy);
    });
  });

  it('keeps TLS certificate detail headers from repeating eyebrow labels inside the detail card', () => {
    const source = readFileSync(resolve(srcRoot, 'views/gateway/TLSCertificates.vue'), 'utf8');
    const detailHeaderStart = source.indexOf('<header class="certificate-detail-header">');
    const detailHeaderEnd = source.indexOf('</header>', detailHeaderStart);
    const detailHeaderSource = source.slice(detailHeaderStart, detailHeaderEnd);

    expect(detailHeaderSource).toContain('certificate-detail-header');
    expect(detailHeaderSource).not.toContain('class="eyebrow"');
  });

  it('uses a structured detail header instead of loose inline title copy', () => {
    const appCss = readFileSync(resolve(srcRoot, 'styles/app.css'), 'utf8');

    expect(appCss).toContain('.detail-title-card');
    expect(appCss).toContain('grid-template-columns: auto minmax(0, 1fr) auto;');
    expect(appCss).toContain('box-shadow: 0 12px 28px rgba(25, 23, 20, 0.08);');
    expect(appCss).toContain('.detail-title-text');
    expect(appCss).toContain('@media (max-width: 900px)');
  });

  it('keeps gateway detail read-only until the header edit action is used', () => {
    const source = readFileSync(resolve(srcRoot, 'views/gateway/GatewayDetail.vue'), 'utf8');

    expect(source).toContain('editMode');
    expect(source).toContain('data-testid="gateway-edit-open"');
    expect(source).toContain("@click=\"openEditMode\"");
    expect(source).toContain("{{ $t('Form.Edit') }}");
    expect(source).toContain('v-if="showDetailEditButton"');
    expect(source).toContain("return this.selected !== 'router';");
    expect(source).toContain('class="detail-tab-toolbar"');
    expect(source).toContain('class="primary-button detail-edit-button"');
    expect(source).toContain("requestedTab === 'setting'");
    expect(source).toContain('readonlyTabs');
    expect(source).not.toContain('detail-header-actions');
    expect(source).not.toContain("{ key: 'setting', label: this.$t('Gateway.BasicSetting') }");
  });

  it('keeps router detail read-only until the header edit action is used', () => {
    const source = readFileSync(resolve(srcRoot, 'views/router/RouterDetail.vue'), 'utf8');

    expect(source).toContain('editMode');
    expect(source).toContain('routeRuleEditMode');
    expect(source).toContain('data-testid="router-edit-open"');
    expect(source).toContain("@click=\"openEditMode\"");
    expect(source).toContain("{{ $t('Form.Edit') }}");
    expect(source).toContain('v-if="showDetailEditButton"');
    expect(source).toContain("return this.selected !== 'gateway';");
    expect(source).toContain('class="detail-tab-toolbar"');
    expect(source).toContain('class="primary-button detail-edit-button"');
    expect(source).toContain("requestedTab === 'setting'");
    expect(source).toContain('readonlyTabs');
    expect(source).toContain(':readonly="!routeRuleEditMode"');
    expect(source).toContain('@close-edit="closeRouteRuleEdit"');
    expect(source).toContain("if (this.selected === 'router')");
    expect(source).not.toContain('detail-header-actions');
    expect(source).not.toContain("{ key: 'setting', label: this.$t('Router.BasicSetting') }");
  });

  it('keeps VirtualService route rules read-only until rule editing is requested', () => {
    const source = readFileSync(resolve(srcRoot, 'components/router/RouterSetting.vue'), 'utf8');
    const httpItemSource = readFileSync(resolve(srcRoot, 'components/router/HttpItem.vue'), 'utf8');

    expect(source).toContain('readonly');
    expect(source).toContain('emits: [\'close-edit\']');
    expect(source).toContain('v-if="!readonly"');
    expect(source).toContain(':readonly="readonly"');
    expect(source).toContain('cancelEdit');
    expect(source).toContain('Router.RuleMappingConflict');
    expect(source).not.toContain('Router rules changed in Kubernetes');
    expect(httpItemSource).toContain('readonly');
    expect(httpItemSource).toContain('if (this.readonly) return;');
  });

  it('keeps Gateway and VirtualService association tabs read-only until association edit is requested', () => {
    const files = [
      ['components/gateway/RouterSetting.vue', 'association-edit-open', 'Gateway.EditRouterAssociations'],
      ['components/router/GatewaySetting.vue', 'association-edit-open', 'Router.EditGatewayAssociations'],
    ];

    files.forEach(([file, testId, editCopy]) => {
      const source = readFileSync(resolve(srcRoot, file), 'utf8');

      expect(source).toContain('associationEditMode');
      expect(source).toContain(`data-testid="${testId}"`);
      expect(source).toContain(editCopy);
      expect(source).toContain('v-if="associationEditMode"');
      expect(source).toContain('association-readonly-toolbar');
      expect(source).toContain('openAssociationEdit');
      expect(source).toContain('closeAssociationEdit');
    });
  });

  it('uses compact gateway dashboard metrics instead of oversized rings and ambiguous series labels', () => {
    const source = readFileSync(resolve(srcRoot, 'components/gateway/Dashboard.vue'), 'utf8');

    expect(source).toContain('dashboard-card-header');
    expect(source).toContain('metric-list');
    expect(source).toContain('protocol-chip-list');
    expect(source).toContain('activeProtocolBreakdown');
    expect(source).toContain('dashboard-distribution-card');
    expect(source).toContain('distributionLabel');
    expect(source).toContain('Gateway.DistributionValues');
    expect(source).toContain('protocol-bar-list');
    expect(source).not.toContain('ring-progress');
    expect(source).not.toContain('spark-values');
    expect(source).not.toContain("values.join(' / ')");
    expect(source).not.toContain("Gateway.Min");
  });

  it('offers topology and list relationship views that can scale beyond one VirtualService', () => {
    const source = readFileSync(resolve(srcRoot, 'components/gateway/Cytoscape.vue'), 'utf8');

    expect(source).toContain("viewMode: 'topology'");
    expect(source).toContain('view-toggle');
    expect(source).toContain('topology-layout');
    expect(source).toContain('topology-source');
    expect(source).toContain('topology-route-connector');
    expect(source).toContain('topology-routes');
    expect(source).toContain('Gateway.RoutesToVirtualService');
    expect(source).toContain('relationship-layout');
    expect(source).toContain('relationship-list-grid');
    expect(source).toContain('compact-list');
    expect(source).toContain('endpoint-list');
    expect(source).not.toContain('relationship-map');
    expect(source).not.toContain('connector-left');
    expect(source).not.toContain('connector-right');
  });

  it('keeps Istio security detail pages read-only until edit is requested', () => {
    const files = [
      ['views/auth/RequestauthDetail.vue', 'requestauth-edit-open', 'Auth.BasicSetting'],
      ['views/policy/AuthpolicyDetail.vue', 'authpolicy-edit-open', 'Policy.BasicSetting'],
    ];

    files.forEach(([file, testId, copyKey]) => {
      const source = readFileSync(resolve(srcRoot, file), 'utf8');

      expect(source).toContain('editMode');
      expect(source).toContain(`data-testid="${testId}"`);
      expect(source).toContain('@click="openEditMode"');
      expect(source).toContain('v-if="editMode"');
      expect(source).toContain('class="detail-tab-toolbar"');
      expect(source).toContain('class="primary-button detail-edit-button"');
      expect(source).toContain('readonly-summary-grid');
      expect(source).toContain(copyKey);
      expect(source).not.toContain('<BasicSetting />');
      expect(source).not.toContain('detail-header-actions');
    });
  });

  it('keeps gateway port number fields constrained inside the listener grid', () => {
    const source = readFileSync(resolve(srcRoot, 'components/gateway/RouterItem.vue'), 'utf8');

    expect(source).toContain('minmax(0, 1fr) minmax(96px, 120px)');
    expect(source).toContain('.port-card .field {');
    expect(source).toContain('min-width: 0;');
    expect(source).toContain('.port-card input,');
    expect(source).toContain('width: 100%;');
  });
});
