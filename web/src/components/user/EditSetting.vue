<template>
  <form v-if="loaded" class="form-stack" novalidate @submit.prevent="submit">
    <label class="field">
      <span>{{ $t('User.Name') }}*</span>
      <input v-model.trim="name" type="text" placeholder="admin" required @blur="touch('name')" />
      <small v-for="error in nameErrors" :key="error" class="field-error">{{ error }}</small>
    </label>

    <label class="field">
      <span>{{ $t('User.Email') }}*</span>
      <input
        v-model.trim="email"
        type="email"
        placeholder="user@brobridge.com"
        required
        @blur="touch('email')"
      />
      <small v-for="error in emailErrors" :key="error" class="field-error">{{ error }}</small>
    </label>

    <label v-if="canManageAdmin" class="checkbox-field">
      <input v-model="isAdmin" type="checkbox" />
      <span>{{ $t('User.Admin') }}</span>
    </label>

    <div v-if="status === 'update_success'" class="alert alert-success">
      {{ $t('Alert.Updated') }}
    </div>
    <div v-if="status === 'update_error'" class="alert alert-error">
      {{ $t('Alert.UpdateFailed') }}
    </div>

    <footer class="page-actions">
      <button class="secondary-button" type="button" @click="fetchData">
        {{ $t('Form.Cancel') }}
      </button>
      <button class="primary-button" type="submit">
        {{ $t('Form.Save') }}
      </button>
    </footer>
  </form>

  <div v-else class="empty-state compact">
    <p>{{ $t('User.LoadingSettings') }}</p>
  </div>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'UserEditSetting',
  data() {
    return {
      id: '',
      loaded: false,
      name: '',
      email: '',
      permissions: [],
      isAdmin: false,
      submitted: false,
      touched: {
        name: false,
        email: false,
      },
    };
  },
  mounted: async function() {
    this.$store.commit('User_ResetStatus');
    this.id = this.$route.params.id;
    await this.fetchData();
  },
  methods: {
    fetchData: async function() {
      this.submitted = false;
      this.touched = {
        name: false,
        email: false,
      };

      let data = await this.$store.dispatch('User_GetItem', {
        id: this.id,
      });

      try {
        data = JSON.parse(JSON.stringify(data));
      } catch (err) {
        console.log(err);
      }

      if (!data) return;

      this.name = data.name || '';
      this.email = data.email || '';
      this.permissions = data.permissions || [];
      this.isAdmin = this.permissions.indexOf('admin') >= 0;
      this.loaded = true;
    },
    touch(field) {
      this.touched[field] = true;
    },
    markAllTouched() {
      Object.keys(this.touched).forEach((field) => {
        this.touched[field] = true;
      });
    },
    isTouched(field) {
      return this.submitted || this.touched[field];
    },
    isEmail(value) {
      return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
    },
    submit() {
      this.$store.commit('User_ResetStatus');
      this.submitted = true;
      this.markAllTouched();
      if (this.hasErrors) return;

      const permissions = [];
      if (this.isAdmin) permissions.push('admin');

      this.$store.dispatch('User_UpdateItem', {
        id: this.id,
        name: this.name,
        email: this.email,
        permissions,
      });
    },
  },
  computed: {
    ...mapGetters({
      language: 'Auth_GetLanguage',
      status: 'User_GetStatus',
      userInfo: 'Auth_GetUserInfo',
    }),
    canManageAdmin() {
      return (this.userInfo?.permissions || []).indexOf('admin') >= 0;
    },
    hasErrors() {
      return [...this.nameErrors, ...this.emailErrors].length > 0;
    },
    nameErrors() {
      if (!this.isTouched('name')) return [];
      return this.name ? [] : [this.$t('Form.Required')];
    },
    emailErrors() {
      if (!this.isTouched('email')) return [];
      if (!this.email) return [this.$t('Form.Required')];
      return this.isEmail(this.email) ? [] : [this.$t('Form.EmailValid')];
    },
  },
};
</script>

<style scoped>
.alert-success {
  background: #dcfce7;
  color: #166534;
}
</style>
