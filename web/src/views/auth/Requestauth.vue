<template>
  <ResourceListPage
    :title="$t('System.APIAuthentication')"
    :subtitle="$t('AuthRequest.Subtitle')"
    create-to="/new/requestauth"
    :create-label="$t('Auth.New')"
    :delete-label="$t('Auth.Remove')"
    :search-label="$t('Form.Search')"
    :total-label="$t('Table.Total')"
    :name-label="$t('Table.Name')"
    :namespace-label="$t('Table.Namespace')"
    :cancel-label="$t('Form.Disagree')"
    :confirm-label="$t('Form.Agree')"
    :delete-error="status === 'delete_error' ? $t('Alert.RemoveFailed') : ''"
    :empty-eyebrow="$t('ResourceList.ReadyForSetup')"
    :empty-title="$t('AuthRequest.EmptyTitle')"
    :empty-message="$t('AuthRequest.EmptyMessage')"
    :filtered-empty-eyebrow="$t('ResourceList.NoMatches')"
    :filtered-empty-title="$t('ResourceList.NoMatchingItems')"
    :filtered-empty-message="$t('ResourceList.TryDifferentKeyword')"
    :unavailable-title="$t('NamespaceInjection.IstioUnavailable')"
    :unavailable-message="istioUnavailableMessage"
    :create-disabled="isIstioUnavailable"
    :loading="loading"
    :columns="columns"
    :items="items"
    :meta="meta"
    :detail-url="detailUrl"
    @delete="deleteAuthRequest"
  />
</template>

<script>
import { mapGetters } from 'vuex';
import ResourceListPage from '../../components/ResourceListPage.vue';

export default {
  name: 'Requestauth',
  components: {
    ResourceListPage,
  },
  mounted() {
    this.$store.commit('AuthRequest_ResetStatus');
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
        await this.$store.dispatch('AuthRequest_GetItems', {
          page: 1,
          namespace: this.namespace || '',
          limit: -1,
        });
        await this.$store.dispatch('AuthRequest_GetMenuItems', {
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
      return `/requestauth/${item.id}?name=${item.name}&namespace=${item.namespace}`;
    },
    deleteAuthRequest(item) {
      this.$store.dispatch('AuthRequest_DelItem', {
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
      status: 'AuthRequest_GetStatus',
      meta: 'AuthRequest_GetMeta',
      items: 'AuthRequest_GetItems',
      namespace: 'Auth_GetNamespace',
      istioCapabilities: 'Auth_GetIstioCapabilities',
    }),
    isIstioUnavailable() {
      return this.istioCapabilities.installed === false || this.istioCapabilities.disabled === true;
    },
    istioUnavailableMessage() {
      const resource = this.$t('AuthRequest.ResourceName');

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
        { key: 'labels', label: this.$t('Table.Label') },
        { key: 'namespace', label: this.$t('Table.Namespace') },
        { key: 'ruleCount', label: this.$t('Table.RuleCount'), align: 'center' },
        { key: 'createdAt', label: this.$t('Table.CreateTime') },
      ];
    },
  },
};
</script>
