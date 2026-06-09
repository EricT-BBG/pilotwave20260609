import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, expect, it, vi } from 'vitest';

import GatewayDetail from '../gateway/GatewayDetail.vue';
import TLSCertificates from '../gateway/TLSCertificates.vue';
import GatewayEditSetting from '../../components/gateway/EditSetting.vue';
import RouterItem from '../../components/gateway/RouterItem.vue';

const srcRoot = resolve(import.meta.dirname, '../..');
const routerSource = readFileSync(resolve(srcRoot, 'router/index.js'), 'utf8');
const gatewayEditSource = readFileSync(resolve(srcRoot, 'components/gateway/EditSetting.vue'), 'utf8');
const routerItemSource = readFileSync(resolve(srcRoot, 'components/gateway/RouterItem.vue'), 'utf8');

describe('TLS certificate management navigation', () => {
  it('registers a protected TLS certificates list route', () => {
    expect(routerSource).toContain("path: 'tls-certificates'");
    expect(routerSource).toContain("name: 'TLSCertificates'");
    expect(routerSource).toContain("views/gateway/TLSCertificates.vue");
    expect(routerSource).toContain("path: 'tls-certificates/:id'");
    expect(routerSource).toContain("name: 'TLSCertificateDetail'");
  });

  it('opens gateway detail on a requested tab for certificate updates', () => {
    const context = {
      $route: { query: { tab: 'setting' } },
      selected: 'information',
      tabs: [
        { key: 'information' },
        { key: 'setting' },
      ],
    };

    GatewayDetail.methods.resetTabs.call(context);

    expect(context.selected).toBe('setting');
  });
});

describe('TLS certificates list page', () => {
  const gateway = {
    name: 'edge-gw',
    namespace: 'istio-system',
  };

  const certificate = {
    serverIndex: 0,
    port: 443,
    protocol: 'HTTPS',
    hosts: ['app.example.com'],
    secretNamespace: 'istio-system',
    secretName: 'edge-cert',
    credentialName: 'edge-cert',
    status: 'warning',
    daysUntilExpiry: 12,
    notAfter: '2026-06-01T00:00:00Z',
    issuer: 'Example CA',
    subject: 'CN=app.example.com',
    dnsNames: ['app.example.com'],
    fingerprintSHA256: 'AA:BB',
  };

  it('aggregates gateway certificate metadata into list rows', async () => {
    const dispatch = vi.fn((action) => {
      if (action === 'Gateway_GetItems') return Promise.resolve([gateway]);
      if (action === 'Gateway_GetTLSCertificates') return Promise.resolve([certificate]);
      return Promise.resolve([]);
    });
    const context = {
      $store: { dispatch },
      namespace: 'All',
      certificates: [],
      loading: false,
      loadError: '',
      ...TLSCertificates.methods,
      $t: (key) => key,
    };

    await TLSCertificates.methods.fetchData.call(context);

    expect(dispatch).toHaveBeenCalledWith('Gateway_GetItems', {
      page: 1,
      namespace: '',
      limit: -1,
    });
    expect(dispatch).toHaveBeenCalledWith('Gateway_GetTLSCertificates', gateway);
    expect(context.certificates).toEqual([
      expect.objectContaining({
        gatewayName: 'edge-gw',
        gatewayNamespace: 'istio-system',
        name: 'app.example.com',
        secret: 'istio-system / edge-cert',
        expiresAt: '2026-06-01 00:00 UTC',
        statusLabel: 'Gateway.TLSWarning',
        detailTo: '/tls-certificates/istio-system%3Aedge-gw%3A0%3A443%3Aistio-system%3Aedge-cert',
        updateTo: '/gateway/edge-gw?name=edge-gw&namespace=istio-system&tab=setting',
      }),
    ]);
  });

  it('uses a table list and opens certificate detail from row clicks', () => {
    const source = readFileSync(resolve(srcRoot, 'views/gateway/TLSCertificates.vue'), 'utf8');
    expect(source).toContain('class="data-table tls-certificate-table"');
    expect(source).toContain('<th class="index-col">#</th>');
    expect(source).toContain('<td class="index-col">{{ index + 1 }}</td>');
    expect(source).toContain('data-testid="tls-certificate-row"');
    expect(source).toContain('@click="openDetail(item)"');
    expect(source).toContain('selectedCertificate');
    expect(source).toContain('data-testid="tls-certificate-detail"');
    expect(source).toContain('class="certificate-summary-grid"');
    expect(source).toContain('.certificate-detail-actions');
    expect(source).toContain('justify-content: center;');
  });

  it('opens detail and update routes from the selected certificate', () => {
    const context = {
      $router: { push: vi.fn() },
    };

    TLSCertificates.methods.openDetail.call(context, {
      detailTo: '/tls-certificates/istio-system%3Aedge-gw%3A0%3A443%3Aistio-system%3Aedge-cert',
    });

    expect(context.$router.push).toHaveBeenCalledWith('/tls-certificates/istio-system%3Aedge-gw%3A0%3A443%3Aistio-system%3Aedge-cert');

    TLSCertificates.methods.openUpdate.call(context, {
      updateTo: '/gateway/edge-gw?name=edge-gw&namespace=istio-system&tab=setting',
    });

    expect(context.$router.push).toHaveBeenCalledWith('/gateway/edge-gw?name=edge-gw&namespace=istio-system&tab=setting');
  });

  it('loads current certificate status for the gateway update TLS dialog', async () => {
    const dispatch = vi.fn(() => Promise.resolve([]));
    const context = {
      name: 'edge-gw',
      namespace: 'istio-system',
      $store: { dispatch },
    };

    await GatewayEditSetting.methods.fetchData.call(context);

    expect(dispatch).toHaveBeenCalledWith('Gateway_GetItem', {
      name: 'edge-gw',
      namespace: 'istio-system',
    });
    expect(dispatch).toHaveBeenCalledWith('Gateway_GetTLSCertificates', {
      name: 'edge-gw',
      namespace: 'istio-system',
    });
    expect(gatewayEditSource).toContain('Gateway_GetTLSCertificates');
  });

  it('shows the matched current certificate inside the TLS update dialog', () => {
    const context = {
      tlsCertificates: [certificate],
      serverIndex: 0,
      port: 443,
      credentialname: 'edge-cert',
      hosts: ['app.example.com'],
    };

    expect(RouterItem.computed.currentTLSCertificate.call(context)).toMatchObject({
      credentialName: 'edge-cert',
      notAfter: '2026-06-01T00:00:00Z',
      fingerprintSHA256: 'AA:BB',
    });
    expect(routerItemSource).toContain('data-testid="tls-current-certificate"');
    expect(routerItemSource).toContain("{{ $t('Gateway.TLSCurrentCertificate') }}");
    expect(routerItemSource).toContain('class="certificate-fingerprint-details"');
    expect(routerItemSource).toContain('@keydown.esc.prevent.stop="closeTLSDialog"');
    expect(routerItemSource).toContain('@keydown.enter="handleTLSDialogEnter"');
    expect(routerItemSource).toContain('@input="updateDraftPrivateKey($event.target.value)"');
    expect(routerItemSource).toContain('grid-template-rows: auto minmax(0, 1fr) auto;');
    expect(routerItemSource).toContain('overscroll-behavior: contain;');
  });

  it('uses compact readable TLS certificate cards in gateway detail', () => {
    const source = readFileSync(resolve(srcRoot, 'components/gateway/TLSCertificates.vue'), 'utf8');

    expect(source).toContain('certificate-meta-line');
    expect(source).toContain('certificate-field-grid');
    expect(source).toContain('certificate-fingerprint-block');
    expect(source).toContain('certificate-card-footer');
    expect(source).toContain("<p>{{ $t('Gateway.TLSCertificates') }}</p>");
    expect(source).not.toContain("<p class=\"eyebrow\">{{ $t('Gateway.TLSCertificates') }}</p>");
    expect(source).toContain('grid-column: 1 / -1;');
    expect(source).toContain('font-size: clamp(0.72rem, 1vw, 0.86rem);');
  });

  it('selects a certificate detail by route id after refresh', async () => {
    const dispatch = vi.fn((action) => {
      if (action === 'Gateway_GetItems') return Promise.resolve([gateway]);
      if (action === 'Gateway_GetTLSCertificates') return Promise.resolve([certificate]);
      return Promise.resolve([]);
    });
    const context = {
      $route: {
        params: {
          id: 'istio-system:edge-gw:0:443:istio-system:edge-cert',
        },
      },
      $store: { dispatch },
      namespace: 'All',
      certificates: [],
      loading: false,
      loadError: '',
      ...TLSCertificates.methods,
      $t: (key) => key,
    };

    await TLSCertificates.methods.fetchData.call(context);

    expect(TLSCertificates.computed.selectedCertificate.call(context)).toMatchObject({
      gatewayName: 'edge-gw',
      secret: 'istio-system / edge-cert',
      subject: 'CN=app.example.com',
    });
  });
});
