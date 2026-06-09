import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const srcRoot = resolve(import.meta.dirname, '../..');

const source = (path) => readFileSync(resolve(srcRoot, path), 'utf8');

describe('gateway selector labels UI contract', () => {
  it('lets users edit selector labels when creating a gateway', () => {
    const newGateway = source('views/gateway/NewGateway.vue');

    expect(newGateway).toContain('class="gateway-basic-grid"');
    expect(newGateway).toContain('SelectorLabelsEditor');
    expect(newGateway).toContain("key: 'istio'");
    expect(newGateway).toContain("value: 'ingressgateway'");
    expect(newGateway).toContain('selectorMatchLabels: this.selectorMatchLabels');
    expect(newGateway).toContain('placeholder="public-gateway"');
    expect(newGateway).toContain('namespace-option-header');
    expect(newGateway).toContain('namespaceInjectionStatusClass(item)');
    expect(newGateway).toContain('namespaceInjectionWarning');
    expect(newGateway).toContain('Auth_GetNamespaceDetails');
  });

  it('lets users edit selector labels when updating a gateway', () => {
    const editSetting = source('components/gateway/EditSetting.vue');

    expect(editSetting).toContain('class="gateway-basic-grid"');
    expect(editSetting).toContain('SelectorLabelsEditor');
    expect(editSetting).toContain('editableSelectorLabels');
    expect(editSetting).toContain('selectorMatchLabels: this.selectorMatchLabels');
    expect(editSetting).toContain('placeholder="public-gateway"');
  });

  it('opens selector labels in a dialog and keeps the form compact', () => {
    const selectorEditor = source('components/gateway/SelectorLabelsEditor.vue');

    expect(selectorEditor).toContain('selectorDialogOpen');
    expect(selectorEditor).toContain('draftSelectorLabels');
    expect(selectorEditor).toContain('selectorSummary');
    expect(selectorEditor).toContain('data-testid="gateway-selector-labels-dialog"');
    expect(selectorEditor).toContain('class="modal-card selector-labels-dialog"');
    expect(selectorEditor).toContain('openSelectorDialog');
    expect(selectorEditor).toContain('applySelectorLabels');
    expect(selectorEditor).toContain('selector-labels-table-head');
    expect(selectorEditor).toContain('draftSelectorRowErrors');
    expect(selectorEditor).toContain('selector-row-index');
    expect(selectorEditor).toContain("$t('Gateway.EditSelectorLabels')");
    expect(selectorEditor).not.toContain('v-if="!collapsed"');
  });

  it('keeps gateway server blocks compact', () => {
    const serverItem = source('components/gateway/ServerItem.vue');

    expect(serverItem).toContain('server-card compact-server-card');
    expect(serverItem).toContain('server-host-row');
    expect(serverItem).toContain('listener-toolbar');
  });
});
