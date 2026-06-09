<template>
  <ResourceListPage
    :title="$t('System.ServiceGateway')"
    :subtitle="$t('Gateway.Subtitle')"
    create-to="/new/gateway"
    :create-label="$t('Gateway.New')"
    :delete-label="$t('Gateway.Remove')"
    :search-label="$t('Form.Search')"
    :total-label="$t('Table.Total')"
    :name-label="$t('Table.Name')"
    :namespace-label="$t('Table.Namespace')"
    :cancel-label="$t('Form.Disagree')"
    :confirm-label="$t('Form.Agree')"
    :delete-error="status === 'delete_error' ? $t('Alert.RemoveFailed') : ''"
    :empty-title="$t('Gateway.EmptyTitle')"
    :empty-message="$t('Gateway.EmptyMessage')"
    :unavailable-title="$t('NamespaceInjection.IstioUnavailable')"
    :unavailable-message="istioUnavailableMessage"
    :create-disabled="isIstioUnavailable"
    :loading="loading"
    :columns="columns"
    :items="items"
    :meta="meta"
    :detail-url="detailUrl"
    @delete="deleteGateway"
  />
</template>

<script>
import { mapGetters } from 'vuex';
import ResourceListPage from '../../components/ResourceListPage.vue';

export default {
  name: 'Gateways',
  components: {
    ResourceListPage,
  },
  mounted() {
    this.$store.commit('Gateway_ResetStatus');
    this.fetchData();
  },
  data() {
    return {
      loading: false,
    };
  },
  methods: {
    async fetchData() {
      this.loading = true;
      try {
        await this.$store.dispatch('Gateway_GetItems', {
          page: 1,
          namespace: this.namespace || '',
          limit: -1,
        });
        await this.$store.dispatch('Gateway_GetMenuItems', {
          page: 1,
          namespace: this.namespace || '',
          limit: -1,
        });
        await this.$store.dispatch('Auth_GetNamespaces');
      } finally {
        this.loading = false;
      }
    },
    detailUrl(item) {
      return `/gateway/${item.name}?name=${item.name}&namespace=${item.namespace}`;
    },
    deleteGateway(item) {
      this.$store.dispatch('Gateway_DelItem', {
        name: item.name,
        namespace: item.namespace,
      });
    },
  },
  watch: {
    async status(val) {
      if (val === 'delete_success') await this.fetchData();
    },
    async namespace() {
      await this.fetchData();
    },
  },
  computed: {
    ...mapGetters({
      status: 'Gateway_GetStatus',
      meta: 'Gateway_GetMeta',
      items: 'Gateway_GetItems',
      namespace: 'Auth_GetNamespace',
      istioCapabilities: 'Auth_GetIstioCapabilities',
    }),
    isIstioUnavailable() {
      return this.istioCapabilities.installed === false || this.istioCapabilities.disabled === true;
    },
    istioUnavailableMessage() {
      const resource = this.$t('Gateway.ResourceName');

      if (this.istioCapabilities.disabled) {
        return this.istioCapabilities.message || this.$t('Istio.DisabledResource', { resource });
      }

      const missingCRDs = this.istioCapabilities.missingCRDs || [];
      if (missingCRDs.length) {
        return this.$t('Istio.MissingCRDsResource', { crds: missingCRDs.join(', '), resource });
      }

      return this.istioCapabilities.message || this.$t('Istio.NotInstalledResource', { resource });
    },
    columns() {
      return [
        { key: 'name', label: this.$t('Table.Name'), link: true },
        { key: 'hosts', label: this.$t('Table.Host') },
        { key: 'hostsCount', label: this.$t('Table.Servers'), align: 'center' },
        { key: 'ports', label: this.$t('Table.Port') },
        { key: 'namespace', label: this.$t('Table.Namespace') },
        { key: 'createdAt', label: this.$t('Table.CreateTime') },
      ];
    },
  },
};
</script>
