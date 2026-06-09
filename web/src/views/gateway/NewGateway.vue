<template>
  <section class="page">
    <header class="page-header detail-header detail-header--compact">
      <div class="detail-header-row">
        <button class="secondary-button detail-back-button" type="button" @click="goBack">
          ← {{ $t('Form.BackToGateways') }}
        </button>
        <div class="detail-title-card">
          <div class="detail-title-text">
            <p class="eyebrow detail-resource-type">{{ $t('System.ServiceGateway') }}</p>
            <h1 class="detail-header-title">{{ $t('Gateway.New') }}</h1>
          </div>
        </div>
      </div>
    </header>

    <div class="panel">
      <div class="tab-strip" role="tablist">
        <button class="tab-button active" type="button">
          {{ $t('Gateway.BasicSetting') }}
        </button>
      </div>

      <form class="form-stack" @submit.prevent="submit">
        <div class="gateway-basic-grid">
          <label class="field">
            <span>{{ $t('Gateway.GatewayName') }}*</span>
            <input
              data-testid="gateway-name"
              v-model.trim="name"
              placeholder="public-gateway"
              type="text"
              @blur="touch('name')"
            />
            <small v-if="errors.name" class="field-error">{{ errors.name }}</small>
          </label>

          <div class="field gateway-namespace-field">
            <span>{{ $t('Table.Namespace') }}*</span>
            <button
              class="namespace-select-button"
              data-testid="gateway-namespace"
              type="button"
              @blur="touch('namespace')"
              @click="toggleNamespaceMenu"
            >
              <span class="namespace-select-name">{{ namespace || $t('NamespaceInjection.ChooseNamespace') }}</span>
              <span
                v-if="namespace"
                class="namespace-status-pill"
                :class="namespaceInjectionStatusClass(namespace)"
              >
                {{ namespaceInjectionStatusLabel(namespace) }}
              </span>
              <span aria-hidden="true">⌄</span>
            </button>
            <div v-if="namespaceMenuOpen" class="namespace-option-panel" data-testid="gateway-namespace-options">
              <div class="namespace-option-header" aria-hidden="true">
                <span>{{ $t('Table.Namespace') }}</span>
                <span>{{ $t('Gateway.InjectionStatus') }}</span>
              </div>
              <button
                v-for="item in namespaceOptions"
                :key="item"
                class="namespace-option-row"
                :class="{ selected: item === namespace }"
                type="button"
                @mousedown.prevent
                @click="selectNamespace(item)"
              >
                <span class="namespace-option-name">{{ item }}</span>
                <span class="namespace-status-pill" :class="namespaceInjectionStatusClass(item)">
                  {{ namespaceInjectionStatusLabel(item) }}
                </span>
              </button>
            </div>
            <small v-if="errors.namespace" class="field-error">{{ errors.namespace }}</small>
            <small
              v-else-if="namespace"
              class="field-help namespace-injection-hint"
              :class="{ warning: namespaceInjectionWarning }"
            >
              {{ namespaceInjectionStatusLabel(namespace) }}
            </small>
          </div>
        </div>

        <div v-if="showNamespaceInjectionWarning" class="alert namespace-injection-alert">
          {{ namespaceInjectionWarning }}
        </div>

        <SelectorLabelsEditor
          v-model="selectorLabels"
          :error="errors.selectorLabels"
        />

        <div class="section-toolbar">
          <strong>{{ $t('Gateway.ProtocolPortSetting') }}</strong>
          <button class="secondary-button" data-testid="gateway-add-server" type="button" @click="addServer">
            + {{ $t('Gateway.NewServer') }}
          </button>
        </div>

        <ServerItem
          v-for="(item, index) in servers"
          :key="'server-' + index"
          :server="item"
          :index="index"
        />

        <div v-if="status === 'create_error'" class="alert alert-error">
          {{ $t('Alert.CreateFailed') }}, {{ errorHandle }}
        </div>

        <div class="page-actions form-actions">
          <button class="secondary-button" type="button" @click="goBack">
            {{ $t('Form.Cancel') }}
          </button>
          <button class="primary-button" data-testid="gateway-submit" type="submit">
            {{ $t('Form.Submit') }}
          </button>
        </div>
      </form>
    </div>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';
import ServerItem from '../../components/gateway/ServerItem.vue';
import SelectorLabelsEditor from '../../components/gateway/SelectorLabelsEditor.vue';

const dnsNamePattern = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;

export default {
  name: 'NewGateway',
  components: {
    SelectorLabelsEditor,
    ServerItem,
  },
  async mounted() {
    this.$store.commit('Gateway_ResetStatus');
    this.$store.commit('Gateway_ResetData');
    await this.$store.dispatch('Auth_GetNamespaces');
  },
  data() {
    return {
      touched: {},
      submitted: false,
      name: '',
      namespace: 'default',
      selectorLabels: [
        { key: 'istio', value: 'ingressgateway' },
      ],
      namespaceMenuOpen: false,
      namespaceWarningAcknowledged: false,
      showNamespaceInjectionWarning: false,
    }
  },
  methods: {
    touch(field) {
      this.touched[field] = true;
    },
    goBack() {
      window.scrollTo(0, 0);
      this.$router.push('/gateways');
    },
    addServer() {
      this.$store.commit('Gateway_AddServers');
    },
    toggleNamespaceMenu() {
      this.namespaceMenuOpen = !this.namespaceMenuOpen;
    },
    selectNamespace(name) {
      this.namespace = name;
      this.namespaceMenuOpen = false;
      this.touch('namespace');
      this.resetNamespaceWarning();
    },
    resetNamespaceWarning() {
      this.namespaceWarningAcknowledged = false;
      this.showNamespaceInjectionWarning = false;
    },
    namespaceDetail(name) {
      return this.namespaceDetailMap[name] || null;
    },
    namespaceInjectionMode(name) {
      const injection = this.namespaceDetail(name)?.istioInjection || {};
      return injection.mode || injection.status || 'disabled';
    },
    namespaceInjectionStatusLabel(name) {
      const detail = this.namespaceDetail(name);
      if (!detail) return this.$t('NamespaceInjection.StatusOff');
      const injection = detail.istioInjection || {};
      const mode = injection.mode || injection.status || 'disabled';
      const revision = injection.revision || '';
      if (mode === 'enabled') return this.$t('NamespaceInjection.StatusDefault');
      if (mode === 'revision') {
        return revision
          ? this.$t('NamespaceInjection.StatusRevisionValue', { revision })
          : this.$t('NamespaceInjection.StatusRevision');
      }
      return this.$t('NamespaceInjection.StatusOff');
    },
    namespaceInjectionStatusClass(name) {
      const mode = this.namespaceInjectionMode(name);
      if (mode === 'enabled') return 'enabled';
      if (mode === 'revision') return 'revision';
      return 'disabled';
    },
    submit() {
      this.submitted = true;
      this.touched = {
        name: true,
        namespace: true,
      };

      if (this.hasErrors) return;
      if (this.namespaceInjectionWarning && !this.namespaceWarningAcknowledged) {
        this.namespaceWarningAcknowledged = true;
        this.showNamespaceInjectionWarning = true;
        return;
      }

      this.$store.dispatch('Gateway_NewItem', {
        name: this.name,
        namespace: this.namespace,
        servers: this.servers,
        selectorMatchLabels: this.selectorMatchLabels,
      });
    },
  },
  watch: {
    status: async function(val) {
      if (val === 'create_success') {
        this.$router.push('/gateways');
      }
    }
  },
  computed: {
    ...mapGetters({
      namespaces: 'Auth_GetNamespaces',
      namespaceDetails: 'Auth_GetNamespaceDetails',
      servers: 'Gateway_GetServers',
      status: 'Gateway_GetStatus',
      errorHandle: 'Gateway_GetErrorHandle'
    }),
    namespaceOptions() {
      const namespaces = (this.namespaces || []).filter((item) => item && item !== 'All');
      return namespaces.length ? namespaces : ['default'];
    },
    namespaceDetailMap() {
      return (this.namespaceDetails || []).reduce((items, item) => {
        if (item?.name) items[item.name] = item;
        return items;
      }, {});
    },
    namespaceIsInjected() {
      const mode = this.namespaceInjectionMode(this.namespace);
      return mode === 'enabled' || mode === 'revision';
    },
    namespaceInjectionWarning() {
      if (!this.namespace || this.namespaceIsInjected) return '';
      return this.$t('Gateway.NamespaceInjectionWarning', { namespace: this.namespace });
    },
    errors() {
      const messages = {};
      const shouldShow = (field) => this.submitted || this.touched[field];

      if (shouldShow('name')) {
        if (!this.name) messages.name = this.$t('Form.Required');
        else if (!dnsNamePattern.test(this.name)) messages.name = this.$t('Form.Valid');
      }

      if (shouldShow('namespace') && !this.namespace) messages.namespace = this.$t('Form.Required');
      if (this.submitted) {
        const keys = new Set();
        if (!this.selectorLabels.length) messages.selectorLabels = this.$t('Form.Required');
        for (const item of this.selectorLabels) {
          const key = String(item.key || '').trim();
          const value = String(item.value || '').trim();
          if (!key || !value) messages.selectorLabels = this.$t('Form.Required');
          if (key && keys.has(key)) messages.selectorLabels = this.$t('Form.Valid');
          keys.add(key);
        }
      }

      return messages;
    },
    selectorMatchLabels() {
      return this.selectorLabels.reduce((labels, item) => {
        const key = String(item.key || '').trim();
        const value = String(item.value || '').trim();
        if (key && value) labels[key] = value;
        return labels;
      }, {});
    },
    hasErrors() {
      return Boolean(this.errors.name || this.errors.namespace || this.errors.selectorLabels);
    },
  },
}
</script>

<style scoped>
.detail-header {
  align-items: flex-start;
}

.tab-strip {
  border-bottom: 1px solid var(--pw-border);
  display: flex;
  gap: 8px;
}

.tab-button {
  background: transparent;
  border: 0;
  border-bottom: 3px solid transparent;
  color: var(--pw-muted);
  font-weight: 800;
  padding: 12px 4px;
}

.tab-button.active {
  border-color: var(--pw-accent);
  color: var(--pw-primary-strong);
}

.section-toolbar {
  align-items: center;
  border-top: 1px solid var(--pw-border);
  display: flex;
  justify-content: space-between;
  padding-top: 18px;
}

.gateway-basic-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: minmax(260px, 520px) minmax(240px, 360px);
}

.gateway-namespace-field {
  position: relative;
}

.namespace-select-button {
  align-items: center;
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  color: var(--pw-primary-strong);
  display: grid;
  gap: 8px;
  grid-template-columns: minmax(0, 1fr) auto auto;
  min-height: 42px;
  padding: 0 12px;
  text-align: left;
  width: 100%;
}

.namespace-select-button:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.16);
  outline: none;
}

.namespace-select-name {
  font-weight: 800;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.namespace-option-panel {
  background: #fff;
  border: 1px solid rgba(216, 210, 198, 0.95);
  border-radius: 16px;
  box-shadow: 0 22px 48px rgba(25, 23, 20, 0.18);
  display: grid;
  left: 0;
  max-height: 320px;
  overflow: auto;
  padding: 8px;
  position: absolute;
  right: 0;
  top: calc(100% + 8px);
  z-index: 20;
}

.namespace-option-header,
.namespace-option-row {
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1fr) minmax(130px, auto);
}

.namespace-option-header {
  color: var(--pw-muted);
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.04em;
  padding: 8px 10px;
}

.namespace-option-row {
  align-items: center;
  background: transparent;
  border: 0;
  border-radius: 10px;
  color: var(--pw-primary-strong);
  min-height: 42px;
  padding: 8px 10px;
  text-align: left;
}

.namespace-option-row:hover,
.namespace-option-row.selected {
  background: var(--pw-surface-soft);
}

.namespace-option-name {
  font-weight: 800;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.namespace-status-pill {
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  font-size: 0.72rem;
  font-weight: 900;
  justify-self: end;
  padding: 4px 8px;
  white-space: nowrap;
}

.namespace-status-pill.enabled {
  background: #dcfce7;
  border-color: #86efac;
  color: #166534;
}

.namespace-status-pill.revision {
  background: #dbeafe;
  border-color: #93c5fd;
  color: #1d4ed8;
}

.namespace-status-pill.disabled {
  background: #fef2f2;
  border-color: #fecaca;
  color: #991b1b;
}

.namespace-injection-alert {
  border-left: 4px solid var(--pw-warning, #c98a20);
}

.namespace-injection-hint.warning {
  color: var(--pw-error);
  font-weight: 700;
}

.form-actions {
  border-top: 1px solid var(--pw-border);
  justify-content: flex-end;
  padding-top: 18px;
}

@media (max-width: 760px) {
  .gateway-basic-grid {
    grid-template-columns: 1fr;
  }
}
</style>
