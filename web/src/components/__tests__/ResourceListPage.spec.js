import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it, vi } from 'vitest';

import ResourceListPage from '../ResourceListPage.vue';

const currentDir = dirname(fileURLToPath(import.meta.url));
const resourceListSource = readFileSync(resolve(currentDir, '../ResourceListPage.vue'), 'utf8');
const requestauthSource = readFileSync(resolve(currentDir, '../../views/auth/Requestauth.vue'), 'utf8');
const appCssSource = readFileSync(resolve(currentDir, '../../styles/app.css'), 'utf8');

const baseProps = {
  title: 'Gateway',
  subtitle: 'Istio gateway service',
  createTo: '/new/gateway',
  createLabel: 'New Gateway',
  deleteLabel: 'Remove',
  searchLabel: 'Search',
  totalLabel: 'Total',
  nameLabel: 'Name',
  namespaceLabel: 'Namespace',
  cancelLabel: 'Cancel',
  confirmLabel: 'Confirm',
  columns: [{ key: 'name', label: 'Name' }],
  items: [],
  detailUrl: (item) => `/gateway/${item.name}`,
};

const makeVm = (overrides = {}) => ({
  ...baseProps,
  ...ResourceListPage.methods,
  columns: [
    { key: 'name', label: 'Name' },
    { key: 'createdAt', label: 'Created' },
    { key: 'permissions', label: 'Permissions' },
  ],
  items: [
    { name: 'bravo', createdAt: 30, permissions: ['write'] },
    { name: 'alpha', createdAt: 200, permissions: ['read'] },
    { name: 'charlie', createdAt: 100, permissions: ['admin'] },
  ],
  condition: '',
  sortKey: '',
  sortDirection: '',
  columnWidths: {},
  resizingColumn: null,
  resizeStartX: 0,
  resizeStartWidth: 0,
  $t: (key) => key,
  ...overrides,
});

const visibleNames = (vm) => ResourceListPage.computed.visibleItems.call(vm).map((item) => item.name);

describe('ResourceListPage unavailable create state', () => {
  it('shows loading copy instead of empty-state copy while data is loading', () => {
    const vm = {
      ...baseProps,
      loading: true,
      condition: '',
      visibleItems: [],
      $t: (key) => key,
    };

    expect(ResourceListPage.computed.shouldShowLoading.call(vm)).toBe(true);
    expect(ResourceListPage.computed.shouldShowEmptyState.call(vm)).toBe(false);
    expect(ResourceListPage.computed.computedLoadingMessage.call(vm)).toBe('Form.Loading');
  });

  it('expects resource list views to pass loading state into the shared list', () => {
    const listViews = [
      '../../views/gateway/Gateway.vue',
      '../../views/router/Router.vue',
      '../../views/auth/Requestauth.vue',
      '../../views/policy/Authpolicy.vue',
      '../../views/user/User.vue',
    ];

    for (const file of listViews) {
      const source = readFileSync(resolve(currentDir, file), 'utf8');
      expect(source).toContain(':loading="loading"');
      expect(source).toContain('loading: false');
      expect(source).toContain('this.loading = true');
      expect(source).toContain('this.loading = false');
    }
  });

  it('uses unavailable empty-state copy when create is disabled', () => {
    const vm = {
      ...baseProps,
      createDisabled: true,
      unavailableTitle: 'Istio unavailable',
      unavailableMessage: 'Istio CRDs are missing. Gateway creation is disabled.',
      condition: '',
    };

    expect(ResourceListPage.computed.computedEmptyTitle.call(vm)).toBe('Istio unavailable');
    expect(ResourceListPage.computed.computedEmptyMessage.call(vm)).toBe('Istio CRDs are missing. Gateway creation is disabled.');
  });

  it('uses translated empty-state copy without generated initials', () => {
    expect(resourceListSource).not.toContain('{{ emptyInitials }}');
    expect(resourceListSource).not.toContain('emptyInitials()');
    expect(resourceListSource).toContain('emptyEyebrow');
    expect(resourceListSource).toContain('filteredEmptyEyebrow');

    expect(requestauthSource).toContain(":subtitle=\"$t('AuthRequest.Subtitle')\"");
    expect(requestauthSource).toContain(":empty-eyebrow=\"$t('ResourceList.ReadyForSetup')\"");
    expect(requestauthSource).toContain(":filtered-empty-eyebrow=\"$t('ResourceList.NoMatches')\"");
    expect(requestauthSource).toContain(":empty-title=\"$t('AuthRequest.EmptyTitle')\"");
    expect(requestauthSource).toContain(":empty-message=\"$t('AuthRequest.EmptyMessage')\"");
    expect(requestauthSource).toContain(":filtered-empty-title=\"$t('ResourceList.NoMatchingItems')\"");
    expect(requestauthSource).toContain(":filtered-empty-message=\"$t('ResourceList.TryDifferentKeyword')\"");
    expect(requestauthSource).not.toContain('No API authentication rules yet');
    expect(requestauthSource).not.toContain('RequestAuthentication resources validate JWT issuers before traffic reaches protected services.');
  });
});

describe('ResourceListPage table interactions', () => {
  it('keeps the incoming item order when no sort is active', () => {
    const vm = makeVm();

    expect(visibleNames(vm)).toEqual(['bravo', 'alpha', 'charlie']);
  });

  it('renders a fixed row number column for list scanning', () => {
    expect(resourceListSource).toContain('<th class="index-col">#</th>');
    expect(resourceListSource).toContain('<td class="index-col">{{ index + 1 }}</td>');
    expect(appCssSource).toMatch(/\.data-table th\s*\{[^}]*font-size:\s*0\.9rem;/);
    expect(appCssSource).toMatch(/\.data-table th\s*\{[^}]*padding:\s*14px 12px;/);
    expect(appCssSource).toMatch(/\.data-table th\s*\{[^}]*text-align:\s*center;/);
    expect(appCssSource).toMatch(/\.data-table th\.sortable-col\s*\{[^}]*padding:\s*0;/);
  });

  it('cycles name sorting through ascending, descending, and reset', () => {
    const vm = makeVm();

    ResourceListPage.methods.toggleSort.call(vm, vm.columns[0]);
    expect(vm.sortKey).toBe('name');
    expect(vm.sortDirection).toBe('asc');
    expect(visibleNames(vm)).toEqual(['alpha', 'bravo', 'charlie']);

    ResourceListPage.methods.toggleSort.call(vm, vm.columns[0]);
    expect(vm.sortDirection).toBe('desc');
    expect(visibleNames(vm)).toEqual(['charlie', 'bravo', 'alpha']);

    ResourceListPage.methods.toggleSort.call(vm, vm.columns[0]);
    expect(vm.sortKey).toBe('');
    expect(vm.sortDirection).toBe('');
    expect(visibleNames(vm)).toEqual(['bravo', 'alpha', 'charlie']);
  });

  it('sorts numeric columns by their raw number values', () => {
    const vm = makeVm();

    ResourceListPage.methods.toggleSort.call(vm, vm.columns[1]);

    expect(visibleNames(vm)).toEqual(['bravo', 'charlie', 'alpha']);
  });

  it('sorts only the filtered rows after search is applied', () => {
    const vm = makeVm({ condition: 'a' });

    ResourceListPage.methods.toggleSort.call(vm, vm.columns[0]);

    expect(visibleNames(vm)).toEqual(['alpha', 'bravo', 'charlie']);
    expect(ResourceListPage.computed.visibleItems.call(vm)).toHaveLength(3);

    vm.condition = 'al';
    expect(visibleNames(vm)).toEqual(['alpha']);
  });

  it('tracks resized column widths without changing selection or delete behavior', () => {
    const listeners = {};
    const stopPropagation = vi.fn();
    const preventDefault = vi.fn();
    const addEventListener = vi.fn((event, handler) => {
      listeners[event] = handler;
    });
    const removeEventListener = vi.fn();
    vi.stubGlobal('document', { addEventListener, removeEventListener });
    const vm = makeVm();

    ResourceListPage.methods.startColumnResize.call(vm, vm.columns[0], {
      clientX: 100,
      currentTarget: {
        closest: () => ({
          offsetWidth: 140,
        }),
      },
      stopPropagation,
      preventDefault,
    });

    listeners.mousemove({ clientX: 72 });

    expect(stopPropagation).toHaveBeenCalled();
    expect(preventDefault).toHaveBeenCalled();
    expect(vm.columnWidths.name).toBe(112);
    expect(vm.selectedItem).toBeUndefined();
    expect(vm.deleteOpen).toBeUndefined();

    listeners.mouseup();
    expect(vm.resizingColumn).toBeNull();
    expect(removeEventListener).toHaveBeenCalledWith('mousemove', listeners.mousemove);
    expect(removeEventListener).toHaveBeenCalledWith('mouseup', listeners.mouseup);

    vi.unstubAllGlobals();
  });
});
