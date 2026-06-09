import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

import Template from '../Template.vue';

const currentDir = dirname(fileURLToPath(import.meta.url));
const templateSource = readFileSync(resolve(currentDir, '../Template.vue'), 'utf8');
const resourceListSource = readFileSync(resolve(currentDir, '../ResourceListPage.vue'), 'utf8');
const gatewayRouterItemSource = readFileSync(resolve(currentDir, '../gateway/RouterItem.vue'), 'utf8');
const appCssSource = readFileSync(resolve(currentDir, '../../styles/app.css'), 'utf8');
const twLangSource = readFileSync(resolve(currentDir, '../../plugins/translate/tw/lang.js'), 'utf8');
const enLangSource = readFileSync(resolve(currentDir, '../../plugins/translate/en/lang.js'), 'utf8');

function templateLogic(overrides = {}) {
  const state = {
    injectionTargetNamespace: '',
    injectionDialog: true,
    injectionMode: 'disabled',
    injectionRevision: '',
    injectionMessage: '',
    injectionOk: false,
    injectionConfirmOpen: false,
    confirmedLabelKeys: [],
    injectionDraftsByNamespace: {},
    namespaceDetails: [
      {
        name: 'pilotwave-istio-smoke',
        labels: {},
        istioInjection: {
          mode: 'disabled',
          revision: '',
        },
      },
      {
        name: 'pilotwave-other',
        labels: {
          'istio-injection': 'enabled',
        },
        istioInjection: {
          mode: 'enabled',
          revision: '',
        },
      },
    ],
    istioCapabilities: {
      revisions: ['pilotwave-smoke-rev'],
      revisionTags: [],
    },
    $t: (key, params = {}) => (params.revision ? `${key}:${params.revision}` : key),
    ...overrides,
  };

  for (const [name, method] of Object.entries(Template.methods)) {
    state[name] = method.bind(state);
  }

  for (const name of [
    'injectionTargetDetail',
    'istioRevisionOptions',
    'hasIstioRevisionOptions',
    'targetCurrentLabels',
    'targetNextLabels',
    'labelDiffItems',
    'changedLabelDiffItems',
    'hasLabelChanges',
  ]) {
    Object.defineProperty(state, name, {
      configurable: true,
      get: () => Template.computed[name].call(state),
    });
  }

  return state;
}

describe('Template topbar menus', () => {
  it('keeps the top-right action limited to sign out', () => {
    expect(templateSource).toContain('data-testid="topbar-logout-open"');
    expect(templateSource).not.toContain('data-testid="topbar-management-menu-open"');
    expect(templateSource).not.toContain('data-testid="topbar-management-menu"');
    expect(templateSource).toContain('data-testid="language-dialog"');
    expect(templateSource).toContain('data-testid="logout-confirm-dialog"');
    expect(templateSource).toContain('data-testid="about-version-dialog"');
    expect(templateSource).toContain('class="modal-actions language-dialog-actions"');
    expect(templateSource).not.toContain('data-testid="topbar-ops-menu-open"');
    expect(templateSource).not.toContain('data-testid="topbar-account-menu-open"');
    expect(templateSource).not.toContain('class="language-select"');
    expect(templateSource).not.toContain('data-testid="language-dialog-close"');
  });

  it('keeps topbar actions pinned to the right when namespace picker is hidden', () => {
    expect(appCssSource).toMatch(/\.topbar-actions\s*\{[^}]*margin-left:\s*auto;/);
  });

  it('uses an integrated namespace menu with reload and injection status', () => {
    expect(templateSource).toContain('data-testid="namespace-menu-open"');
    expect(templateSource).toContain('data-testid="namespace-menu"');
    expect(templateSource).toContain('data-testid="namespace-refresh"');
    expect(templateSource).toContain('namespace-status-badge');
    expect(templateSource).toContain('selectedNamespaceBadge');
    expect(templateSource).toContain('namespaceMenuItems');
    expect(templateSource).not.toContain('<select v-model="namespace">');
    expect(templateSource).not.toContain('namespace-refresh-button');
  });

  it('keeps detail headers visually separated from back actions', () => {
    expect(appCssSource).toMatch(/\.page-header\.detail-header:not\(\.detail-header--compact\) > div\s*\{[^}]*gap:\s*16px;/);
    expect(appCssSource).toMatch(/\.page-header\.detail-header--compact \.detail-header-row\s*\{[^}]*display:\s*grid;/);
    expect(appCssSource).toMatch(/\.detail-title-card\s*\{[^}]*box-shadow:\s*0 12px 28px rgba\(25,\s*23,\s*20,\s*0\.08\);/);
    expect(appCssSource).toMatch(/\.detail-title-text\s*\{[^}]*align-items:\s*baseline;/);
  });

  it('keeps page titles compact across shared pages', () => {
    expect(appCssSource).toMatch(/\.page-header h1\s*\{[^}]*font-size:\s*clamp\(1\.35rem,\s*2vw,\s*2rem\);/);
    expect(appCssSource).toMatch(/\.page-header h1\s*\{[^}]*letter-spacing:\s*0;/);
    expect(appCssSource).toMatch(/\.page-header\.detail-header--compact \.detail-resource-type::after\s*\{[^}]*content:\s*"\/";/);
  });

  it('keeps namespace injection copy focused on the apply confirmation step', () => {
    expect(templateSource.match(/\$t\('NamespaceInjection\.RestartWarning'\)/g)).toHaveLength(1);
    expect(twLangSource).not.toContain('預設注入（建議）');
    expect(enLangSource).not.toContain('Default injection (recommended)');
  });

  it('keeps About focused on compact version information', () => {
    expect(templateSource).toContain('class="about-summary"');
    expect(templateSource).toContain('class="about-version-grid"');
    expect(templateSource).toContain('data-testid="about-dialog-close"');
    expect(templateSource).toContain('{{ buildInfo.version }}');
    expect(templateSource).toContain('{{ buildInfo.buildLabel }}');
    expect(templateSource).not.toContain('BROBRIDGE Cloud API Management');
    expect(templateSource).not.toContain('Copyright ©');
    expect(templateSource).not.toContain('https://www.brobridge.com/');
    expect(templateSource).not.toContain('compact-about-meta');
  });

  it('does not duplicate top-right close controls when dialogs already have footer actions', () => {
    expect(resourceListSource).toContain('class="modal-actions"');
    expect(resourceListSource).toContain('class="modal-card small delete-dialog"');
    expect(resourceListSource).not.toContain('<button class="icon-button" type="button" @click="deleteOpen = false">{{ $t(\'Form.Close\') }}</button>');

    expect(gatewayRouterItemSource).toContain('class="dialog-footer"');
    expect(gatewayRouterItemSource).not.toContain('@click="closeTLSDialog">x</button>');
    expect(gatewayRouterItemSource).not.toContain('@click="cancelProtocolChange">x</button>');

    expect(templateSource).toContain('data-testid="about-dialog-close"');
  });

  it('opens dialogs directly from shell actions', () => {
    const state = {
      languageDialog: false,
      logoutConfirmDialog: false,
      aboutDialog: false,
      pendingLanguage: '',
      language: 'tw',
    };

    Template.methods.openLanguageDialog.call(state);
    expect(state.languageDialog).toBe(true);
    expect(state.pendingLanguage).toBe('tw');

    Template.methods.openLogoutConfirmDialog.call(state);
    expect(state.logoutConfirmDialog).toBe(true);

    Template.methods.openAboutDialog.call(state);
    expect(state.aboutDialog).toBe(true);
  });

  it('closes management dialogs with Escape', () => {
    const event = { key: 'Escape' };
    const state = {
      languageDialog: true,
      logoutConfirmDialog: true,
      aboutDialog: true,
    };

    Template.methods.handleEscapeKey.call(state, event);

    expect(state.languageDialog).toBe(false);
    expect(state.logoutConfirmDialog).toBe(false);
    expect(state.aboutDialog).toBe(false);
  });

  it('shows and clears a global API unavailable alert', () => {
    expect(templateSource).toContain('data-testid="api-error-alert"');
    expect(templateSource).toContain("window.addEventListener('pilotwave-api-error', this.handleApiError)");
    expect(templateSource).toContain("window.removeEventListener('pilotwave-api-error', this.handleApiError)");

    const state = {
      apiErrorMessage: '',
      $t: (key) => key,
    };

    Template.methods.handleApiError.call(state, {
      detail: {
        message: 'Alert.ApiUnavailable',
      },
    });
    expect(state.apiErrorMessage).toBe('Alert.ApiUnavailable');

    Template.methods.clearApiError.call(state);
    expect(state.apiErrorMessage).toBe('');
  });

  it('keeps namespace injection drafts when switching namespace rows', () => {
    const state = templateLogic();

    state.selectInjectionNamespace('pilotwave-istio-smoke');
    state.injectionMode = 'revision';
    state.injectionRevision = 'pilotwave-smoke-rev';

    state.selectInjectionNamespace('pilotwave-other');
    expect(state.injectionMode).toBe('enabled');

    state.selectInjectionNamespace('pilotwave-istio-smoke');
    expect(state.injectionMode).toBe('revision');
    expect(state.injectionRevision).toBe('pilotwave-smoke-rev');

    state.injectionMode = 'enabled';
    state.selectInjectionNamespace('pilotwave-istio-smoke');
    expect(state.injectionMode).toBe('enabled');
  });

  it('marks namespace rows with unsaved injection drafts', () => {
    expect(templateSource).toContain('data-testid="namespace-injection-unsaved-badge"');

    const state = templateLogic();
    state.selectInjectionNamespace('pilotwave-istio-smoke');
    expect(state.hasUnsavedInjectionDraft('pilotwave-istio-smoke')).toBe(false);

    state.injectionMode = 'enabled';
    expect(state.hasUnsavedInjectionDraft('pilotwave-istio-smoke')).toBe(true);
    expect(state.hasUnsavedInjectionDraft('pilotwave-other')).toBe(false);

    state.selectInjectionNamespace('pilotwave-other');
    expect(state.hasUnsavedInjectionDraft('pilotwave-istio-smoke')).toBe(true);
  });
});
