import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

import Navigation from '../Navigation.vue';

const currentDir = dirname(fileURLToPath(import.meta.url));
const navigationSource = readFileSync(resolve(currentDir, '../Navigation.vue'), 'utf8');

describe('Navigation actions', () => {
  it('exposes shell actions as left navigation actions', () => {
    expect(navigationSource).toContain('data-testid="nav-namespace-injection-open"');
    expect(navigationSource).toContain("$emit('open-namespace-injection')");
    expect(navigationSource).toContain('data-testid="nav-language-open"');
    expect(navigationSource).toContain("$emit('open-language-dialog')");
    expect(navigationSource).toContain('data-testid="nav-about-open"');
    expect(navigationSource).toContain("$emit('open-about-dialog')");
  });

  it('exposes TLS certificate management as a primary navigation item', () => {
    expect(navigationSource).toContain("to: '/tls-certificates'");
    expect(navigationSource).toContain("icon: 'TL'");
    expect(navigationSource).toContain("$t('System.TLSCertificates')");
  });

  it('uses Istio resource naming in primary navigation', () => {
    expect(navigationSource).toContain("{ to: '/gateways', icon: 'GW', label: this.$t('System.ServiceGateway') }");
    expect(navigationSource).toContain("{ to: '/routers', icon: 'VS', label: this.$t('System.ServiceRouter') }");
    expect(navigationSource).not.toContain("icon: 'RT'");
  });

  it('places TLS under policy and separates account monitoring and Istio actions into their own group', () => {
    const primaryItems = navigationSource.slice(
      navigationSource.indexOf('primaryItems()'),
      navigationSource.indexOf('isAdmin()')
    );

    expect(primaryItems.indexOf("to: '/authpolicies'")).toBeLessThan(
      primaryItems.indexOf("to: '/tls-certificates'")
    );

    const groupedActions = navigationSource.slice(
      navigationSource.indexOf('<div class="nav-separator"></div>'),
      navigationSource.indexOf('data-testid="nav-language-open"')
    );

    expect(groupedActions).toContain('to="/users?page=1"');
    expect(groupedActions).toContain("$t('System.GrafanaManagement')");
    expect(groupedActions).toContain('data-testid="nav-namespace-injection-open"');
    expect(groupedActions.match(/class="nav-separator"/g)).toHaveLength(2);
  });

  it('uses a wider centered monitoring settings dialog without top-right close', () => {
    expect(navigationSource).toContain('class="modal-card monitoring-dialog"');
    expect(navigationSource).toContain('class="monitoring-dialog-header"');
    expect(navigationSource).toContain('class="form-stack monitoring-form"');
    expect(navigationSource).toContain('class="modal-actions monitoring-dialog-actions"');
    expect(navigationSource).not.toContain('<button class="icon-button" type="button" @click="dialog = false">Close</button>');
  });

  it('closes the monitoring settings dialog with Escape', () => {
    expect(navigationSource).toContain("window.addEventListener('keydown', this.handleEscapeKey)");
    expect(navigationSource).toContain("window.removeEventListener('keydown', this.handleEscapeKey)");

    const state = {
      dialog: true,
    };

    Navigation.methods.handleEscapeKey.call(state, { key: 'Enter' });
    expect(state.dialog).toBe(true);

    Navigation.methods.handleEscapeKey.call(state, { key: 'Escape' });
    expect(state.dialog).toBe(false);
  });
});
