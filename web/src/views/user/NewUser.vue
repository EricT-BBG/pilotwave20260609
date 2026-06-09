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
            <h1 class="detail-header-title">{{ $t('User.New') }}</h1>
          </div>
        </div>
      </div>
    </header>

    <section class="panel">
      <p class="eyebrow">{{ $t('User.BasicSetting') }}</p>

      <form class="form-stack" novalidate @submit.prevent="submit">
        <label class="field">
          <span>{{ $t('User.Name') }}*</span>
          <input
            v-model.trim="name"
            type="text"
            placeholder="admin"
            required
            @blur="touch('name')"
          />
          <small v-for="error in nameErrors" :key="error" class="field-error">{{ error }}</small>
        </label>

        <label class="field">
          <span>{{ $t('User.Username') }}*</span>
          <input
            v-model.trim="username"
            type="text"
            placeholder="administrator"
            required
            @blur="touch('username')"
          />
          <small v-for="error in usernameErrors" :key="error" class="field-error">{{ error }}</small>
        </label>

        <label class="field">
          <span>{{ $t('Form.Password') }}*</span>
          <input v-model="password" type="password" required @blur="touch('password')" />
          <small v-for="error in passwordErrors" :key="error" class="field-error">{{ error }}</small>
        </label>

        <label class="field">
          <span>{{ $t('Form.ConfirmPw') }}*</span>
          <input v-model="repeatPassword" type="password" required @blur="touch('repeatPassword')" />
          <small v-for="error in repeatPasswordErrors" :key="error" class="field-error">{{ error }}</small>
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

        <label class="checkbox-field">
          <input v-model="isAdmin" type="checkbox" />
          <span>{{ $t('User.Admin') }}</span>
        </label>

        <div v-if="status === 'create_error'" class="alert alert-error">
          {{ $t('Alert.CreateFailed') }}
        </div>

        <footer class="page-actions">
          <button class="secondary-button" type="button" @click="goBack">
            {{ $t('Form.Cancel') }}
          </button>
          <button class="primary-button" type="submit">
            {{ $t('Form.Submit') }}
          </button>
        </footer>
      </form>
    </section>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'NewUser',
  data() {
    return {
      name: '',
      username: '',
      password: '',
      repeatPassword: '',
      email: '',
      isAdmin: false,
      submitted: false,
      touched: {
        name: false,
        username: false,
        password: false,
        repeatPassword: false,
        email: false,
      },
    };
  },
  methods: {
    goBack() {
      window.scrollTo(0, 0);
      this.$router.go(-1);
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
      this.submitted = true;
      this.markAllTouched();
      if (this.hasErrors) return;

      const permissions = [];
      if (this.isAdmin) permissions.push('admin');

      this.$store.dispatch('User_NewItem', {
        name: this.name,
        username: this.username,
        password: this.password,
        email: this.email,
        permissions,
      });
    },
  },
  watch: {
    status(val) {
      if (val === 'create_success') this.$router.push('/users?page=1');
    },
  },
  computed: {
    ...mapGetters({
      language: 'Auth_GetLanguage',
      status: 'User_GetStatus',
    }),
    hasErrors() {
      return [
        ...this.nameErrors,
        ...this.usernameErrors,
        ...this.passwordErrors,
        ...this.repeatPasswordErrors,
        ...this.emailErrors,
      ].length > 0;
    },
    nameErrors() {
      if (!this.isTouched('name')) return [];
      return this.name ? [] : [this.$t('Form.Required')];
    },
    usernameErrors() {
      if (!this.isTouched('username')) return [];
      return this.username ? [] : [this.$t('Form.Required')];
    },
    passwordErrors() {
      if (!this.isTouched('password')) return [];
      return this.password ? [] : [this.$t('Form.Required')];
    },
    repeatPasswordErrors() {
      if (!this.isTouched('repeatPassword')) return [];
      return this.repeatPassword === this.password ? [] : [this.$t('Form.PwdNotMatch')];
    },
    emailErrors() {
      if (!this.isTouched('email')) return [];
      if (!this.email) return [this.$t('Form.Required')];
      return this.isEmail(this.email) ? [] : [this.$t('Form.EmailValid')];
    },
  },
};
</script>
