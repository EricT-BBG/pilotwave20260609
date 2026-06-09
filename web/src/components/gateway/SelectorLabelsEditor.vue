<template>
  <section class="selector-labels">
    <div class="selector-labels-header">
      <div>
        <strong>{{ $t('Gateway.SelectorLabels') }}</strong>
        <p>{{ selectorSummary }}</p>
      </div>
      <button class="secondary-button compact-button" type="button" @click="openSelectorDialog">
        {{ $t('Gateway.EditSelectorLabels') }}
      </button>
    </div>

    <small v-if="error" class="field-error">{{ error }}</small>

    <div
      v-if="selectorDialogOpen"
      class="modal-backdrop"
      data-testid="gateway-selector-labels-dialog"
      @click.self="closeSelectorDialog"
    >
      <section
        class="modal-card selector-labels-dialog"
        role="dialog"
        aria-modal="true"
        :aria-label="$t('Gateway.SelectorLabels')"
      >
        <header class="modal-header selector-dialog-header">
          <div>
            <h2>{{ $t('Gateway.SelectorLabels') }}</h2>
            <p>{{ $t('Gateway.SelectorLabelsHelp') }}</p>
          </div>
        </header>

        <div class="selector-labels-list">
          <div class="selector-labels-table-head" aria-hidden="true">
            <span>#</span>
            <span>{{ $t('Gateway.SelectorKey') }}</span>
            <span>{{ $t('Gateway.SelectorValue') }}</span>
            <span></span>
          </div>
          <div
            v-for="(item, index) in draftSelectorLabels"
            :key="'selector-draft-' + index"
            class="selector-label-row"
            :class="{ 'has-error': draftSelectorRowErrors[index] }"
          >
            <div class="selector-row-index">#{{ index + 1 }}</div>
            <label class="field selector-input-field">
              <span class="sr-only">{{ $t('Gateway.SelectorKey') }}</span>
              <input
                :data-testid="'gateway-selector-key-' + index"
                :value="item.key"
                placeholder="istio"
                type="text"
                @input="updateDraftLabel(index, 'key', $event.target.value)"
              />
            </label>
            <label class="field selector-input-field">
              <span class="sr-only">{{ $t('Gateway.SelectorValue') }}</span>
              <input
                :data-testid="'gateway-selector-value-' + index"
                :value="item.value"
                placeholder="ingressgateway"
                type="text"
                @input="updateDraftLabel(index, 'value', $event.target.value)"
              />
            </label>
            <button
              class="danger-button compact-button"
              type="button"
              :aria-label="$t('Gateway.RemoveSelectorLabel')"
              @click="removeDraftLabel(index)"
            >
              {{ $t('Form.Delete') }}
            </button>
            <small v-if="draftSelectorRowErrors[index]" class="field-error selector-row-error">
              {{ draftSelectorRowErrors[index] }}
            </small>
          </div>

          <div class="selector-add-row">
            <button
              class="secondary-button compact-button selector-add-button"
              type="button"
              @click="addDraftLabel"
            >
              + {{ $t('Gateway.AddSelectorLabel') }}
            </button>
          </div>
        </div>

        <footer class="modal-actions selector-dialog-actions">
          <button class="secondary-button" type="button" @click="closeSelectorDialog">
            {{ $t('Form.Cancel') }}
          </button>
          <button
            class="primary-button"
            type="button"
            :disabled="Boolean(draftSelectorError)"
            @click="applySelectorLabels"
          >
            {{ $t('Gateway.ApplySelectorLabels') }}
          </button>
        </footer>
      </section>
    </div>
  </section>
</template>

<script>
export default {
  name: 'SelectorLabelsEditor',
  props: {
    modelValue: {
      type: Array,
      default: () => [],
    },
    error: {
      type: String,
      default: '',
    },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      selectorDialogOpen: false,
      draftSelectorLabels: [],
    };
  },
  computed: {
    selectorSummary() {
      const summary = this.modelValue
        .filter((item) => item.key || item.value)
        .map((item) => `${item.key || '?'}=${item.value || '?'}`)
        .join(', ');
      return summary || this.$t('Gateway.SelectorLabelsHelp');
    },
    draftSelectorError() {
      return this.draftSelectorRowErrors.find(Boolean) || '';
    },
    draftSelectorRowErrors() {
      const keys = new Set();
      if (!this.draftSelectorLabels.length) return [this.$t('Form.Required')];

      return this.draftSelectorLabels.map((item) => {
        const key = String(item.key || '').trim();
        const value = String(item.value || '').trim();
        if (!key || !value) return this.$t('Form.Required');
        if (keys.has(key)) return this.$t('Form.Valid');
        keys.add(key);
        return '';
      });
    },
  },
  watch: {
    error(nextError) {
      if (nextError) this.openSelectorDialog();
    },
  },
  mounted() {
    document.addEventListener('keydown', this.handleEscape);
  },
  beforeUnmount() {
    document.removeEventListener('keydown', this.handleEscape);
  },
  methods: {
    cloneLabels(labels) {
      return (labels || []).map((item) => ({
        key: item.key || '',
        value: item.value || '',
      }));
    },
    openSelectorDialog() {
      const nextDraft = this.cloneLabels(this.modelValue);
      this.draftSelectorLabels = nextDraft.length ? nextDraft : [{ key: '', value: '' }];
      this.selectorDialogOpen = true;
    },
    closeSelectorDialog() {
      this.selectorDialogOpen = false;
    },
    handleEscape(event) {
      if (event.key === 'Escape') this.closeSelectorDialog();
    },
    addDraftLabel() {
      this.draftSelectorLabels = [
        ...this.draftSelectorLabels,
        { key: '', value: '' },
      ];
    },
    removeDraftLabel(index) {
      this.draftSelectorLabels = this.draftSelectorLabels.filter((_, itemIndex) => itemIndex !== index);
    },
    updateDraftLabel(index, field, value) {
      this.draftSelectorLabels = this.draftSelectorLabels.map((item, itemIndex) => {
        if (itemIndex !== index) return item;
        return {
          ...item,
          [field]: value,
        };
      });
    },
    applySelectorLabels() {
      if (this.draftSelectorError) return;

      const normalizedLabels = this.draftSelectorLabels.map((item) => ({
        key: String(item.key || '').trim(),
        value: String(item.value || '').trim(),
      }));
      this.$emit('update:modelValue', normalizedLabels);
      this.closeSelectorDialog();
    },
  },
};
</script>

<style scoped>
.selector-labels {
  border-top: 1px solid var(--pw-border);
  display: grid;
  gap: 12px;
  padding-top: 18px;
}

.selector-labels-header {
  align-items: flex-start;
  display: flex;
  gap: 16px;
  justify-content: space-between;
}

.selector-labels-header p {
  color: var(--pw-muted);
  font-size: 0.92rem;
  margin: 4px 0 0;
}

.selector-labels-list {
  display: grid;
  gap: 8px;
}

.selector-labels-table-head {
  color: var(--pw-muted);
  display: grid;
  font-size: 0.78rem;
  font-weight: 800;
  gap: 10px;
  grid-template-columns: 56px minmax(0, 1fr) minmax(0, 1fr) 92px;
  letter-spacing: 0.04em;
  padding: 0 10px;
}

.selector-label-row {
  align-items: center;
  background: var(--pw-surface-soft);
  border: 1px solid var(--pw-border);
  border-radius: 14px;
  display: grid;
  gap: 10px;
  grid-template-columns: 56px minmax(0, 1fr) minmax(0, 1fr) 92px;
  padding: 10px;
}

.selector-label-row.has-error {
  border-color: rgba(190, 39, 30, 0.7);
}

.selector-row-index {
  color: var(--pw-primary-strong);
  font-size: 0.85rem;
  font-weight: 900;
}

.selector-input-field input {
  min-height: 38px;
}

.selector-label-row .danger-button {
  justify-self: stretch;
  min-height: 38px;
  padding: 0 12px;
}

.selector-row-error {
  grid-column: 2 / -1;
}

.selector-add-row {
  display: flex;
  justify-content: flex-start;
  padding-top: 4px;
}

.selector-add-button {
  min-height: 38px;
  padding: 0 14px;
}

.selector-labels-dialog {
  max-height: calc(100vh - 48px);
  max-width: 840px;
  overflow: auto;
}

.selector-dialog-header {
  align-items: flex-start;
}

.selector-dialog-header p {
  color: var(--pw-muted);
  margin: 6px 0 0;
}

.selector-dialog-actions {
  border-top: 1px solid var(--pw-border);
  margin-top: 18px;
  padding-top: 16px;
}

@media (max-width: 760px) {
  .selector-labels-header,
  .selector-label-row {
    align-items: stretch;
    grid-template-columns: 1fr;
  }

  .selector-labels-table-head {
    display: none;
  }

  .selector-row-error {
    grid-column: auto;
  }

  .selector-labels-header {
    display: grid;
  }
}
</style>
