<template>
  <section class="page">
    <header class="page-header detail-header detail-header--compact">
      <div class="detail-header-row">
        <button class="secondary-button detail-back-button" type="button" @click="goBack">
          ← {{ $t('Form.BackToRouters') }}
        </button>
        <div class="detail-title-card">
          <div class="detail-title-text">
            <p class="eyebrow detail-resource-type">{{ $t('System.ServiceRouter') }}</p>
            <h1 class="detail-header-title">{{ $t('Router.New') }}</h1>
          </div>
        </div>
      </div>
    </header>

    <div class="panel">
      <div class="tab-strip" role="tablist">
        <button class="tab-button active" type="button">
          {{ $t('Router.BasicSetting') }}
        </button>
      </div>

      <form class="form-stack" @submit.prevent="submit">
        <label class="field">
          <span>{{ $t('Router.RouerName') }}*</span>
          <input
            data-testid="router-name"
            v-model.trim="name"
            placeholder="public-route"
            type="text"
            @blur="touch('name')"
          />
          <small v-if="errors.name" class="field-error">{{ errors.name }}</small>
        </label>

        <label class="field">
          <span>{{ $t('Table.Namespace') }}*</span>
          <select data-testid="router-namespace" v-model="namespace" @blur="touch('namespace')">
            <option v-for="item in namespaceOptions" :key="item" :value="item">
              {{ item }}
            </option>
          </select>
          <small v-if="errors.namespace" class="field-error">{{ errors.namespace }}</small>
        </label>

        <label class="field">
          <span>{{ $t('Table.Host') }}*</span>
          <input
            data-testid="router-hosts"
            v-model="hostsText"
            placeholder="localhost, api.example.local"
            type="text"
            @blur="touch('hosts')"
          />
          <small v-if="errors.hosts" class="field-error">{{ errors.hosts }}</small>
        </label>

        <label class="field">
          <span>{{ $t('Table.Protocol') }}*</span>
          <select data-testid="router-protocol" v-model="protocol" @blur="touch('protocol')">
            <option v-for="item in protocols" :key="item.value" :value="item.value">
              {{ item.name }}
            </option>
          </select>
          <small v-if="errors.protocol" class="field-error">{{ errors.protocol }}</small>
        </label>

        <section class="association-form">
          <div class="association-toolbar">
            <div class="association-summary">
              <span>{{ $t('Router.SelectGateway') }}</span>
              <strong>{{ $t('Form.Selected') }}: {{ selectedGateways.length }}</strong>
            </div>
          </div>

          <div v-if="gatewayOptions.length" class="association-list-wrap">
            <div class="association-list" data-testid="router-create-gateway-association-list">
              <label
                v-for="item in gatewayOptions"
                :key="item.value"
                class="association-row"
                :class="{ 'association-row--selected': selectedGateways.includes(item.value) }"
              >
                <input v-model="selectedGateways" type="checkbox" :value="item.value" />
                <span class="association-checkmark" aria-hidden="true">
                  {{ selectedGateways.includes(item.value) ? '✓' : '' }}
                </span>
                <span class="association-main">
                  <strong>{{ item.text }}</strong>
                  <small>{{ item.namespace }}</small>
                </span>
                <span v-if="selectedGateways.includes(item.value)" class="association-selected-badge">
                  {{ $t('Form.Selected') }}
                </span>
              </label>
            </div>
          </div>

          <div v-else class="empty-state compact">
            {{ $t('Router.NoGatewaysAssociated') }}
          </div>
        </section>

        <div v-if="status === 'create_error'" class="alert alert-error">
          {{ $t('Alert.CreateFailed') }} {{ errorHandle }}
        </div>

        <div class="page-actions form-actions">
          <button class="secondary-button" type="button" @click="goBack">
            {{ $t('Form.Cancel') }}
          </button>
          <button class="primary-button" data-testid="router-submit" type="submit">
            {{ $t('Form.Submit') }}
          </button>
        </div>
      </form>
    </div>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';

const dnsNamePattern = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;

export default {
  name: 'NewRouter',
  async mounted() {
    this.$store.commit('Router_ResetStatus');
    await this.$store.dispatch('Auth_GetNamespaces');
    await this.fetchGateways();
  },
  data() {
    return {
      touched: {},
      submitted: false,
      name: '',
      hostsText: 'localhost',
      namespace: 'default',
      protocol: 'http',
      selectedGateways: [],
    }
  },
  methods: {
    touch(field) {
      this.touched[field] = true;
    },
    goBack() {
      window.scrollTo(0, 0);
      this.$router.push('/routers');
    },
    async fetchGateways() {
      await this.$store.dispatch('Gateway_GetItems', {
        page: 1,
        limit: -1,
        namespace: this.namespace,
      });

      const validValues = new Set(this.gatewayOptions.map((item) => item.value));
      this.selectedGateways = this.selectedGateways.filter((item) => validValues.has(item));
    },
    selectedGatewayItems() {
      return this.selectedGateways
        .filter(Boolean)
        .map((item) => {
          const params = item.split(',');
          return {
            name: params[0],
            namespace: params[1],
          };
        });
    },
    submit() {
      this.submitted = true;
      this.touched = {
        name: true,
        namespace: true,
        hosts: true,
        protocol: true,
      };

      if (this.hasErrors) return;

      this.$store.dispatch('Router_NewItem', {
        name: this.name,
        protocol: this.protocol,
        namespace: this.namespace,
        hosts: this.hosts,
        gateways: this.selectedGatewayItems(),
      });
    },
  },
  watch: {
    status: async function(val) {
      if (val === 'create_success') {
        await this.$store.dispatch('Router_GetMenuItems', {
          page: 1,
          limit: -1
        });

        this.$router.push('/routers?page=1');
      }
    },
    namespace: async function() {
      await this.fetchGateways();
    },
  },
  computed: {
    ...mapGetters({
      protocols: 'Router_GetProtocols',
      namespaces: 'Auth_GetNamespaces',
      gateways: 'Gateway_GetItems',
      status: 'Router_GetStatus',
      errorHandle: 'Router_GetErrorHandle'
    }),
    namespaceOptions() {
      const namespaces = (this.namespaces || []).filter((item) => item && item !== 'All');
      return namespaces.length ? namespaces : ['default'];
    },
    hosts() {
      return this.hostsText
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean);
    },
    gatewayOptions() {
      return (this.gateways || []).filter((item) => item?.value);
    },
    errors() {
      const messages = {};
      const shouldShow = (field) => this.submitted || this.touched[field];

      if (shouldShow('name')) {
        if (!this.name) messages.name = this.$t('Form.Required');
        else if (!dnsNamePattern.test(this.name)) messages.name = this.$t('Form.Lowercase');
      }

      if (shouldShow('namespace') && !this.namespace) messages.namespace = this.$t('Form.Required');
      if (shouldShow('hosts') && !this.hosts.length) messages.hosts = this.$t('Form.Required');
      if (shouldShow('protocol') && !this.protocol) messages.protocol = this.$t('Form.Required');

      return messages;
    },
    hasErrors() {
      return Boolean(this.errors.name || this.errors.namespace || this.errors.hosts || this.errors.protocol);
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

.form-actions {
  border-top: 1px solid var(--pw-border);
  justify-content: flex-end;
  padding-top: 18px;
}

.association-form {
  border-top: 1px solid var(--pw-border);
  display: grid;
  gap: 14px;
  padding-top: 18px;
}

.association-toolbar {
  align-items: center;
  display: grid;
  gap: 16px;
}

.association-summary {
  align-items: center;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  display: flex;
  gap: 12px;
  justify-content: space-between;
  min-width: 0;
  padding: 10px 14px;
}

.association-summary span {
  color: #475569;
  font-size: 0.86rem;
  font-weight: 800;
}

.association-summary strong {
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 800;
  white-space: nowrap;
}

.association-list-wrap {
  min-width: 0;
}

.association-list {
  display: grid;
  gap: 8px;
  max-height: 220px;
  overflow: auto;
}

.association-row {
  align-items: center;
  background: #fff;
  border: 1px solid #d9e2ec;
  border-radius: 8px;
  cursor: pointer;
  display: grid;
  gap: 10px;
  grid-template-columns: auto auto minmax(0, 1fr) auto;
  min-height: 54px;
  padding: 10px 12px;
  transition: background-color 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
}

.association-row:hover {
  border-color: #9fb3c8;
}

.association-row--selected {
  background: #eef6ff;
  border-color: #2f80ed;
  box-shadow: 0 0 0 1px rgba(47, 128, 237, 0.12);
}

.association-row input {
  height: 16px;
  width: 16px;
}

.association-checkmark {
  align-items: center;
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  border-radius: 999px;
  color: #1769aa;
  display: inline-flex;
  font-size: 0.8rem;
  font-weight: 800;
  height: 22px;
  justify-content: center;
  width: 22px;
}

.association-row--selected .association-checkmark {
  background: #dbeafe;
  border-color: #2f80ed;
}

.association-main {
  display: grid;
  gap: 3px;
  min-width: 0;
}

.association-main strong,
.association-main small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.association-main strong {
  color: #102a43;
  font-size: 0.95rem;
}

.association-main small {
  color: #64748b;
  font-size: 0.78rem;
}

.association-selected-badge {
  background: #dbeafe;
  border: 1px solid #93c5fd;
  border-radius: 999px;
  color: #1d4ed8;
  font-size: 0.72rem;
  font-weight: 800;
  padding: 3px 8px;
  white-space: nowrap;
}

@media (max-width: 700px) {
  .association-summary {
    align-items: flex-start;
    flex-direction: column;
    gap: 4px;
  }

  .association-row {
    grid-template-columns: auto auto minmax(0, 1fr);
  }

  .association-selected-badge {
    grid-column: 3;
    justify-self: start;
  }
}
</style>
