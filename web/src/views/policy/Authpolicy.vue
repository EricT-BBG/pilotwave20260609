<template>
  <ResourceListPage
    :title="$t('System.BlackWhteList')"
    :subtitle="$t('Policy.Subtitle')"
    create-to="/new/authpolicy"
    :create-label="$t('Policy.New')"
    :delete-label="$t('Policy.Remove')"
    :search-label="$t('Form.Search')"
    :total-label="$t('Table.Total')"
    :name-label="$t('Table.Name')"
    :namespace-label="$t('Table.Namespace')"
    :cancel-label="$t('Form.Disagree')"
    :confirm-label="$t('Form.Agree')"
    :delete-error="status === 'delete_error' ? $t('Alert.RemoveFailed') : ''"
    :empty-title="$t('Policy.EmptyTitle')"
    :empty-message="$t('Policy.EmptyMessage')"
    :unavailable-title="$t('NamespaceInjection.IstioUnavailable')"
    :unavailable-message="istioUnavailableMessage"
    :create-disabled="isIstioUnavailable"
    :loading="loading"
    :columns="columns"
    :items="items"
    :meta="meta"
    :detail-url="detailUrl"
    @delete="deleteAuthPolicy"
  />
</template>

<script>
import { mapGetters } from 'vuex';
import ResourceListPage from '../../components/ResourceListPage.vue';

export default {
  name: 'Authpolicy',
  components: {
    ResourceListPage,
  },
  mounted() {
    this.$store.commit('AuthPolicy_ResetStatus');
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
        await this.$store.dispatch('AuthPolicy_GetItems', {
          page: 1,
          namespace: this.namespace || '',
          limit: -1,
        });
        await this.$store.dispatch('AuthPolicy_GetMenuItems', {
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
      return `/authpolicy/${item.id}?name=${item.name}&namespace=${item.namespace}`;
    },
    deleteAuthPolicy(item) {
      this.$store.dispatch('AuthPolicy_DelItem', {
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
      status: 'AuthPolicy_GetStatus',
      meta: 'AuthPolicy_GetMeta',
      items: 'AuthPolicy_GetItems',
      namespace: 'Auth_GetNamespace',
      istioCapabilities: 'Auth_GetIstioCapabilities',
    }),
    isIstioUnavailable() {
      return this.istioCapabilities.installed === false || this.istioCapabilities.disabled === true;
    },
    istioUnavailableMessage() {
      const resource = this.$t('Policy.ResourceName');

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
