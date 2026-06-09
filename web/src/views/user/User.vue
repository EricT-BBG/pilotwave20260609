<template>
  <ResourceListPage
    :title="$t('System.AccountManagement')"
    :subtitle="$t('User.Subtitle')"
    create-to="/new/user"
    :create-label="$t('User.New')"
    :delete-label="$t('User.Remove')"
    :search-label="$t('Form.Search')"
    :total-label="$t('Table.Total')"
    :name-label="$t('User.Name')"
    :namespace-label="$t('Table.Namespace')"
    :cancel-label="$t('Form.Disagree')"
    :confirm-label="$t('Form.Agree')"
    :delete-error="status === 'delete_error' ? $t('Alert.RemoveFailed') : ''"
    :empty-title="$t('User.EmptyTitle')"
    :empty-message="$t('User.EmptyMessage')"
    :columns="columns"
    :items="items"
    :meta="meta"
    :loading="loading"
    :detail-url="detailUrl"
    :row-key="rowKey"
    @delete="deleteUser"
  >
    <template #footer>
      <button class="icon-button" type="button" @click="prePage">{{ $t('Form.Previous') }}</button>
      <button class="icon-button" type="button" @click="nextPage">{{ $t('Form.Next') }}</button>
      <span>{{ $t('Table.Page') }}: {{ page }}</span>
    </template>
  </ResourceListPage>
</template>

<script>
import { mapGetters } from 'vuex';
import ResourceListPage from '../../components/ResourceListPage.vue';

export default {
  name: 'Users',
  components: {
    ResourceListPage,
  },
  data() {
    return {
      page: 1,
      loading: false,
    };
  },
  mounted() {
    this.$store.commit('User_ResetStatus');
    this.fetchData();
  },
  methods: {
    async fetchData() {
      this.loading = true;
      try {
        this.page = Number(this.$route.query.page || 1);
        await this.$store.dispatch('User_GetItems', {
          page: this.page,
          search: this.$route.query.search || '',
          limit: 20,
        });
      } finally {
        this.loading = false;
      }
    },
    detailUrl(item) {
      return `/user/${item.id}`;
    },
    rowKey(item) {
      return item.id || item.username || item.email;
    },
    deleteUser(item) {
      this.$store.dispatch('User_DelItem', {
        id: item.id,
      });
    },
    async nextPage() {
      if (this.meta.limit * this.meta.page >= this.meta.total) return;
      const next = Number(this.$route.query.page || 1) + 1;
      await this.$router.push(`/users?page=${next}`);
      await this.fetchData();
    },
    async prePage() {
      const prev = Math.max(1, Number(this.$route.query.page || 1) - 1);
      if (prev === this.page) return;
      await this.$router.push(`/users?page=${prev}`);
      await this.fetchData();
    },
  },
  watch: {
    async status(val) {
      if (val === 'delete_success') await this.fetchData();
    },
    '$route.query': {
      async handler() {
        await this.fetchData();
      },
      deep: true,
    },
  },
  computed: {
    ...mapGetters({
      status: 'User_GetStatus',
      meta: 'User_GetMeta',
      items: 'User_GetItems',
    }),
    columns() {
      return [
        { key: 'name', label: this.$t('Table.Name'), link: true },
        { key: 'username', label: this.$t('Table.Account') },
        { key: 'email', label: this.$t('Table.Email') },
        { key: 'permissions', label: this.$t('Table.Role') },
        { key: 'createdAt', label: this.$t('Table.CreateTime') },
      ];
    },
  },
};
</script>
