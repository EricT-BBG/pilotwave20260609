<template>
  <section class="page">
    <header class="page-header detail-header detail-header--compact">
      <div class="detail-header-row">
        <button class="secondary-button detail-back-button" type="button" @click="goBack">
          ← {{ $t('Form.BackToUsers') }}
        </button>
        <div class="detail-title-card">
          <div class="detail-title-text">
            <p class="eyebrow detail-resource-type">{{ $t('System.AccountManagement') }}</p>
            <h1 class="detail-header-title">{{ displayUsername }}</h1>
          </div>
        </div>
      </div>
    </header>

    <section class="panel">
      <nav class="detail-tabs" aria-label="User settings">
        <button
          class="secondary-button"
          :class="{ active: tab === 'setting' }"
          type="button"
          @click="tab = 'setting'"
        >
          {{ $t('User.BasicSetting') }}
        </button>
        <button
          class="secondary-button"
          :class="{ active: tab === 'password' }"
          type="button"
          @click="tab = 'password'"
        >
          {{ $t('User.PwdUpdate') }}
        </button>
      </nav>

      <BasicSetting v-if="tab === 'setting'" />
      <PasswordSetting v-else />
    </section>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';
import BasicSetting from '../../components/user/EditSetting.vue';
import PasswordSetting from '../../components/user/PasswordSetting.vue';

export default {
  name: 'UserDetail',
  components: {
    BasicSetting,
    PasswordSetting,
  },
  data() {
    return {
      tab: 'setting',
      id: '',
    };
  },
  mounted: async function() {
    this.$store.commit('User_ResetStatus');
    this.id = this.$route.params.id;
    await this.fetchData();
  },
  methods: {
    fetchData: async function() {
      await this.$store.dispatch('User_GetItem', {
        id: this.id,
      });
    },
    goBack() {
      window.scrollTo(0, 0);
      this.$router.go(-1);
    },
  },
  computed: {
    ...mapGetters({
      language: 'Auth_GetLanguage',
      status: 'User_GetStatus',
      user: 'User_GetItem',
    }),
    displayUsername() {
      return this.user?.username || this.id;
    },
  },
};
</script>

<style scoped>
.detail-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.detail-tabs .active {
  background: var(--pw-primary-strong);
  border-color: var(--pw-primary-strong);
  color: #fff;
}
</style>
