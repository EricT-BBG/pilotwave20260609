<template>
  <form class="form-stack" @submit.prevent="submit">
    <div class="gateway-basic-grid">
      <label class="field">
        <span>{{ $t('Gateway.GatewayName') }}*</span>
        <input v-model.trim="name" disabled placeholder="public-gateway" type="text" />
      </label>

      <label class="field">
        <span>{{ $t('Table.Namespace') }}*</span>
        <input v-model.trim="namespace" disabled type="text" />
      </label>
    </div>

    <SelectorLabelsEditor
      v-model="editableSelectorLabels"
      :error="selectorLabelsError"
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

    <div v-if="status === 'update_success'" class="alert alert-success">
      {{ $t('Alert.Updated') }}
    </div>
    <div v-if="status === 'update_error'" class="alert alert-error">
      {{ $t('Alert.UpdateFailed') }} {{ errorHandle }}
    </div>
    <div v-if="status === 'update_conflict'" class="alert alert-error conflict-alert">
      <span>{{ errorHandle || 'This Gateway changed in Kubernetes. Reload before submitting again.' }}</span>
      <button class="secondary-button" type="button" @click="reloadData">
        Reload
      </button>
    </div>

    <div class="page-actions form-actions">
      <button class="secondary-button" type="button" @click="goBack">
        {{ $t('Form.Cancel') }}
      </button>
      <button class="primary-button" data-testid="gateway-update-submit" type="submit">
        {{ $t('Form.Submit') }}
      </button>
    </div>
  </form>
</template>

<script>
import { mapGetters } from 'vuex';
import SelectorLabelsEditor from './SelectorLabelsEditor.vue';
import ServerItem from './ServerItem.vue';

export default {
  name: 'GatewayEditSetting',
  components: {
    SelectorLabelsEditor,
    ServerItem,
  },
  mounted: async function() {
    this.$store.commit('Gateway_ResetStatus');
    this.name = this.$route.query.name;
    this.namespace = this.$route.query.namespace;

    await this.fetchData();
  },
  data() {
    return {
      name: '',
      namespace: 'default',
      submitted: false,
      editableSelectorLabels: [
        { key: 'istio', value: 'ingressgateway' },
      ],
    }
  },
  methods: {
    fetchData: async function() {
      await this.$store.dispatch('Gateway_GetItem', {
        name: this.name,
        namespace: this.namespace
      });
      await this.$store.dispatch('Gateway_GetTLSCertificates', {
        name: this.name,
        namespace: this.namespace
      });
    },
    reloadData: async function() {
      this.$store.commit('Gateway_ResetStatus');
      await this.fetchData();
    },
    goBack() {
      window.scrollTo(0, 0);
      this.$router.go(-1);
    },
    addServer() {
      this.$store.commit('Gateway_AddServers');
    },
    submit() {
      this.submitted = true;
      if (this.selectorLabelsError) return;

      this.$store.commit('Gateway_ResetStatus');
      this.$store.dispatch('Gateway_UpdateItem', {
        name: this.name,
        namespace: this.namespace,
        servers: this.servers,
        selectorMatchLabels: this.selectorMatchLabels,
        resourceVersion: this.resourceVersion,
      });
    },
    selectorRowsFromMap(labels) {
      return Object.entries(labels || {}).map(([key, value]) => ({ key, value }));
    },
  },
  watch: {
    status: async function(val) {
      if (val === 'update_success') await this.fetchData();
    },
    currentGatewaySelectorMatchLabels: {
      immediate: true,
      handler(labels) {
        const rows = this.selectorRowsFromMap(labels);
        this.editableSelectorLabels = rows.length ? rows : [{ key: 'istio', value: 'ingressgateway' }];
      },
    },
  },
  computed: {
    ...mapGetters({
      servers: 'Gateway_GetServers',
      currentGatewaySelectorMatchLabels: 'Gateway_GetSelectorMatchLabels',
      resourceVersion: 'Gateway_GetResourceVersion',
      status: 'Gateway_GetStatus',
      errorHandle: 'Gateway_GetErrorHandle'
    }),
    selectorMatchLabels() {
      return this.editableSelectorLabels.reduce((labels, item) => {
        const key = String(item.key || '').trim();
        const value = String(item.value || '').trim();
        if (key && value) labels[key] = value;
        return labels;
      }, {});
    },
    selectorLabelsError() {
      if (!this.submitted) return '';

      const keys = new Set();
      if (!this.editableSelectorLabels.length) return this.$t('Form.Required');
      for (const item of this.editableSelectorLabels) {
        const key = String(item.key || '').trim();
        const value = String(item.value || '').trim();
        if (!key || !value) return this.$t('Form.Required');
        if (keys.has(key)) return this.$t('Form.Valid');
        keys.add(key);
      }

      return '';
    },
  },
}
</script>

<style scoped>
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

.alert-success {
  background: #dcfce7;
  color: #166534;
}

.form-actions {
  border-top: 1px solid var(--pw-border);
  justify-content: flex-end;
  padding-top: 18px;
}

.conflict-alert {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

@media (max-width: 760px) {
  .gateway-basic-grid {
    grid-template-columns: 1fr;
  }
}
</style>
