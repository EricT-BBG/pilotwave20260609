<template>
  <ResourceListPage
    :title="$t('System.ServiceRouter')"
    :subtitle="$t('Router.Subtitle')"
    create-to="/new/router"
    :create-label="$t('Router.New')"
    :delete-label="$t('Router.Remove')"
    :search-label="$t('Form.Search')"
    :total-label="$t('Table.Total')"
    :name-label="$t('Table.Name')"
    :namespace-label="$t('Table.Namespace')"
    :cancel-label="$t('Form.Disagree')"
    :confirm-label="$t('Form.Agree')"
    :delete-error="status === 'delete_error' ? $t('Alert.RemoveFailed') : ''"
    :empty-title="$t('Router.EmptyTitle')"
    :empty-message="$t('Router.EmptyMessage')"
    :unavailable-title="$t('NamespaceInjection.IstioUnavailable')"
    :unavailable-message="istioUnavailableMessage"
    :create-disabled="isIstioUnavailable"
    :loading="loading"
    :columns="columns"
    :items="items"
    :meta="meta"
    :detail-url="detailUrl"
    @delete="deleteRouter"
  />
</template>

<script>
import { mapGetters } from 'vuex';
import ResourceListPage from '../../components/ResourceListPage.vue';

export default {
  name: 'Routers',
  components: {
    ResourceListPage,
  },
  mounted() {
    this.$store.commit('Router_ResetStatus');
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
        await this.$store.dispatch('Router_GetItems', {
          page: 1,
          namespace: this.namespace || '',
          limit: -1,
        });
        await this.$store.dispatch('Router_GetMenuItems', {
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
      return `/router/${item.name}?name=${item.name}&namespace=${item.namespace}`;
    },
    deleteRouter(item) {
      this.$store.dispatch('Router_DelItem', {
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
      status: 'Router_GetStatus',
      meta: 'Router_GetMeta',
      items: 'Router_GetItems',
      namespace: 'Auth_GetNamespace',
      istioCapabilities: 'Auth_GetIstioCapabilities',
    }),
    isIstioUnavailable() {
      return this.istioCapabilities.installed === false || this.istioCapabilities.disabled === true;
    },
    istioUnavailableMessage() {
      const resource = this.$t('Router.ResourceName');

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
        { key: 'protocol', label: this.$t('Table.Protocol') },
        { key: 'namespace', label: this.$t('Table.Namespace') },
        { key: 'httpCount', label: this.$t('Table.RuleCount'), align: 'center' },
        { key: 'createdAt', label: this.$t('Table.CreateTime') },
      ];
    },
  },
};
</script>
