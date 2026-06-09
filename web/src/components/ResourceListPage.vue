<template>
  <section class="page">
    <header class="page-header">
      <div>
        <p class="eyebrow">{{ subtitle }}</p>
        <h1>{{ title }}</h1>
      </div>
      <div class="page-actions">
        <button
          v-if="createDisabled"
          class="secondary-button"
          data-testid="resource-create-disabled"
          type="button"
          disabled
          :title="unavailableMessage"
        >
          {{ createLabel }}
        </button>
        <router-link v-else class="primary-button" data-testid="resource-create" :to="createTo">{{ createLabel }}</router-link>
        <button v-if="selectedItem" class="danger-button" data-testid="resource-delete-open" type="button" @click="openDelete">
          {{ deleteLabel }}
        </button>
      </div>
    </header>

    <div v-if="createDisabled && unavailableMessage" class="alert">
      {{ unavailableMessage }}
    </div>

    <section class="panel">
      <label class="search-box">
        <span>{{ searchLabel }}</span>
        <input data-testid="resource-search" v-model.trim="condition" type="search" @input="filterItems" />
      </label>

      <div class="table-wrap">
        <table class="data-table">
          <colgroup>
            <col class="select-col" />
            <col class="index-col" />
            <col v-for="column in columns" :key="column.key" :style="columnWidthStyle(column)" />
          </colgroup>
          <thead>
            <tr>
              <th class="select-col"></th>
              <th class="index-col">#</th>
              <th
                v-for="column in columns"
                :key="column.key"
                :aria-sort="columnAriaSort(column)"
                :class="[
                  'sortable-col',
                  column.align === 'center' ? 'text-center' : '',
                  resizingColumn === column.key ? 'is-resizing' : '',
                ]"
              >
                <div class="column-header">
                  <button
                    class="sort-button"
                    type="button"
                    :data-testid="`resource-sort-${column.key}`"
                    @click="toggleSort(column)"
                  >
                    <span>{{ column.label }}</span>
                    <span v-if="sortIndicator(column)" class="sort-indicator" aria-hidden="true">
                      {{ sortIndicator(column) }}
                    </span>
                  </button>
                  <span
                    class="resize-handle"
                    role="separator"
                    aria-orientation="vertical"
                    :aria-label="`Resize ${column.label}`"
                    @click.stop
                    @mousedown="startColumnResize(column, $event)"
                  ></span>
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(item, index) in visibleItems"
              :key="rowKey(item)"
              class="resource-row"
              tabindex="0"
              @click="openDetail(item)"
              @keydown.enter="openDetail(item)"
            >
              <td class="select-col">
                <button
                  class="check-button"
                  :class="{ selected: isSelected(item) }"
                  type="button"
                  :data-testid="resourceTestId('select', item)"
                  @click.stop="toggleSelected(item)"
                >
                  <span aria-hidden="true">{{ isSelected(item) ? '✓' : '' }}</span>
                  <span class="sr-only">{{ isSelected(item) ? $t('Form.Selected') : $t('Form.Select') }}</span>
                </button>
              </td>
              <td class="index-col">{{ index + 1 }}</td>
              <td v-for="column in columns" :key="column.key" :class="column.align === 'center' ? 'text-center' : ''">
                <button
                  v-if="column.link"
                  class="link-button"
                  type="button"
                  @click.stop="openDetail(item)"
                >
                  {{ formatValue(item, column) }}
                </button>
                <span v-else-if="column.key === 'permissions'" class="chip-list">
                  <span v-for="permission in item.permissions || []" :key="permission" class="chip">{{ permission }}</span>
                </span>
                <span v-else>{{ formatValue(item, column) }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="shouldShowLoading" class="empty-state compact empty-state-list" data-testid="resource-loading">
        {{ computedLoadingMessage }}
      </div>

      <div v-else-if="shouldShowEmptyState" class="empty-state empty-state-list">
        <div class="empty-illustration" aria-hidden="true"></div>
        <div class="empty-copy">
          <p class="eyebrow">{{ computedEmptyEyebrow }}</p>
          <h2>{{ computedEmptyTitle }}</h2>
          <p>{{ computedEmptyMessage }}</p>
          <button v-if="!hasFilter && createDisabled" class="secondary-button" type="button" disabled :title="unavailableMessage">
            {{ createLabel }}
          </button>
          <router-link v-else-if="!hasFilter" class="primary-button" :to="createTo">
            {{ createLabel }}
          </router-link>
        </div>
      </div>

      <footer class="table-footer">
        <slot name="footer" />
        <span>{{ totalLabel }}: {{ meta?.total || visibleItems.length }}</span>
      </footer>
    </section>

    <div v-if="deleteOpen && selectedItem" class="modal-backdrop" @click.self="deleteOpen = false">
      <section class="modal-card small delete-dialog">
        <header class="modal-header delete-dialog-header">
          <div>
            <p class="eyebrow">{{ deleteLabel }}</p>
            <h2>{{ deleteLabel }}</h2>
          </div>
        </header>
        <dl class="delete-dialog-details">
          <div>
            <dt>{{ nameLabel }}</dt>
            <dd>{{ selectedItem.name || selectedItem.username }}</dd>
          </div>
          <div v-if="selectedItem.namespace">
            <dt>{{ namespaceLabel }}</dt>
            <dd>{{ selectedItem.namespace }}</dd>
          </div>
        </dl>
        <div v-if="deleteError" class="alert alert-error">{{ deleteError }}</div>
        <footer class="modal-actions">
          <button class="secondary-button" type="button" @click="deleteOpen = false">
            {{ cancelLabel }}
          </button>
          <button class="danger-button" data-testid="resource-delete-confirm" type="button" @click="$emit('delete', selectedItem)">
            {{ confirmLabel }}
          </button>
        </footer>
      </section>
    </div>
  </section>
</template>

<script>
import moment from 'moment';

export default {
  name: 'ResourceListPage',
  props: {
    title: { type: String, required: true },
    subtitle: { type: String, required: true },
    createTo: { type: String, required: true },
    createLabel: { type: String, required: true },
    deleteLabel: { type: String, required: true },
    searchLabel: { type: String, required: true },
    totalLabel: { type: String, required: true },
    nameLabel: { type: String, required: true },
    namespaceLabel: { type: String, required: true },
    cancelLabel: { type: String, required: true },
    confirmLabel: { type: String, required: true },
    deleteError: { type: String, default: '' },
    emptyEyebrow: { type: String, default: '' },
    emptyTitle: { type: String, default: '' },
    emptyMessage: { type: String, default: '' },
    unavailableTitle: { type: String, default: '' },
    unavailableMessage: { type: String, default: '' },
    createDisabled: { type: Boolean, default: false },
    loading: { type: Boolean, default: false },
    loadingMessage: { type: String, default: '' },
    filteredEmptyEyebrow: { type: String, default: '' },
    filteredEmptyTitle: { type: String, default: '' },
    filteredEmptyMessage: { type: String, default: '' },
    columns: { type: Array, required: true },
    items: { type: Array, default: () => [] },
    meta: { type: Object, default: () => ({}) },
    detailUrl: { type: Function, required: true },
    rowKey: {
      type: Function,
      default: (item) => {
        if (item.namespace && item.name) return `${item.namespace}/${item.name}`;
        return item.username || item.name || item.id || '';
      },
    },
  },
  emits: ['delete'],
  data() {
    return {
      condition: '',
      selectedItem: null,
      deleteOpen: false,
      sortKey: '',
      sortDirection: '',
      columnWidths: {},
      resizingColumn: null,
      resizeStartX: 0,
      resizeStartWidth: 0,
      resizeMoveHandler: null,
      resizeEndHandler: null,
    };
  },
  beforeUnmount() {
    this.removeColumnResizeListeners();
  },
  methods: {
    openDetail(item) {
      this.toUrl(this.detailUrl(item));
    },
    toUrl(url) {
      window.scrollTo(0, 0);
      if (url) this.$router.push(url);
    },
    formatValue(item, column) {
      if (column.format) return column.format(item[column.key], item);
      if (column.key === 'createdAt' && item[column.key]) return moment.unix(item[column.key]).format('MM/DD/YYYY, LT');
      const value = item[column.key];
      if (Array.isArray(value)) return value.join(', ');
      return value ?? '';
    },
    toggleSelected(item) {
      this.selectedItem = this.isSelected(item) ? null : item;
    },
    isSelected(item) {
      return this.selectedItem && this.rowKey(this.selectedItem) === this.rowKey(item);
    },
    resourceTestId(prefix, item) {
      const key = this.rowKey(item)
        .replace(/[^a-zA-Z0-9_-]+/g, '-')
        .replace(/^-+|-+$/g, '');
      return `resource-${prefix}-${key}`;
    },
    openDelete() {
      if (this.selectedItem) this.deleteOpen = true;
    },
    filterItems() {
      this.selectedItem = null;
    },
    toggleSort(column) {
      if (!column?.key) return;
      if (this.sortKey !== column.key) {
        this.sortKey = column.key;
        this.sortDirection = 'asc';
        return;
      }
      if (this.sortDirection === 'asc') {
        this.sortDirection = 'desc';
        return;
      }
      this.sortKey = '';
      this.sortDirection = '';
    },
    columnAriaSort(column) {
      if (this.sortKey !== column.key || !this.sortDirection) return 'none';
      return this.sortDirection === 'asc' ? 'ascending' : 'descending';
    },
    sortIndicator(column) {
      if (this.sortKey !== column.key) return '';
      if (this.sortDirection === 'asc') return '▲';
      if (this.sortDirection === 'desc') return '▼';
      return '';
    },
    sortValue(item) {
      const value = item?.[this.sortKey];
      if (Array.isArray(value)) return value.join(', ');
      return value;
    },
    compareItems(left, right) {
      const leftValue = this.sortValue(left);
      const rightValue = this.sortValue(right);
      const leftEmpty = leftValue === null || leftValue === undefined;
      const rightEmpty = rightValue === null || rightValue === undefined;
      if (leftEmpty && rightEmpty) return 0;
      if (leftEmpty) return 1;
      if (rightEmpty) return -1;

      let result;
      if (typeof leftValue === 'number' && typeof rightValue === 'number') {
        result = leftValue - rightValue;
      } else {
        result = String(leftValue).localeCompare(String(rightValue));
      }
      return this.sortDirection === 'desc' ? result * -1 : result;
    },
    columnWidthStyle(column) {
      const width = this.columnWidths[column.key];
      return width ? { width: `${width}px` } : null;
    },
    startColumnResize(column, event) {
      if (!column?.key) return;
      event.stopPropagation();
      event.preventDefault();

      this.removeColumnResizeListeners();
      const headerCell = event.currentTarget.closest('th');
      this.resizingColumn = column.key;
      this.resizeStartX = event.clientX;
      this.resizeStartWidth = this.columnWidths[column.key] || headerCell?.offsetWidth || 96;
      this.resizeMoveHandler = (moveEvent) => this.resizeColumn(moveEvent);
      this.resizeEndHandler = () => this.stopColumnResize();
      document.addEventListener('mousemove', this.resizeMoveHandler);
      document.addEventListener('mouseup', this.resizeEndHandler);
    },
    resizeColumn(event) {
      if (!this.resizingColumn) return;
      const nextWidth = Math.max(96, this.resizeStartWidth + event.clientX - this.resizeStartX);
      this.columnWidths = {
        ...this.columnWidths,
        [this.resizingColumn]: nextWidth,
      };
    },
    stopColumnResize() {
      this.removeColumnResizeListeners();
      this.resizingColumn = null;
    },
    removeColumnResizeListeners() {
      if (this.resizeMoveHandler) {
        document.removeEventListener('mousemove', this.resizeMoveHandler);
      }
      if (this.resizeEndHandler) {
        document.removeEventListener('mouseup', this.resizeEndHandler);
      }
      this.resizeMoveHandler = null;
      this.resizeEndHandler = null;
    },
  },
  computed: {
    hasFilter() {
      return Boolean(this.condition);
    },
    shouldShowLoading() {
      return this.loading && !this.visibleItems.length;
    },
    shouldShowEmptyState() {
      return !this.loading && !this.visibleItems.length;
    },
    computedLoadingMessage() {
      return this.loadingMessage || this.$t('Form.Loading');
    },
    computedEmptyEyebrow() {
      if (this.hasFilter) return this.filteredEmptyEyebrow || this.$t('ResourceList.NoMatches');
      return this.emptyEyebrow || this.$t('ResourceList.ReadyForSetup');
    },
    computedEmptyTitle() {
      if (this.hasFilter) return this.filteredEmptyTitle || this.$t('ResourceList.NoMatchingItems');
      if (this.createDisabled) return this.unavailableTitle || this.$t('NamespaceInjection.IstioUnavailable');
      return this.emptyTitle || this.$t('ResourceList.NoMatchingItems');
    },
    computedEmptyMessage() {
      if (this.hasFilter) return this.filteredEmptyMessage || this.$t('ResourceList.TryDifferentKeyword');
      if (this.createDisabled && this.unavailableMessage) return this.unavailableMessage;
      return this.emptyMessage || this.$t('ResourceList.TryDifferentKeyword');
    },
    visibleItems() {
      const sourceItems = this.items || [];
      let result = sourceItems;
      if (this.condition) {
        const term = this.condition.toLowerCase();
        result = sourceItems.filter((item) => JSON.stringify(item).toLowerCase().includes(term));
      }
      if (!this.sortKey || !this.sortDirection) return result;
      return [...result].sort((left, right) => this.compareItems(left, right));
    },
  },
};
</script>
