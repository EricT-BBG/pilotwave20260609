<template>
  <form class="form-stack" novalidate @submit.prevent="submit">
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

    <div v-if="status === 'update_success'" class="alert alert-success">
      {{ $t('Alert.Updated') }}
    </div>
    <div v-if="status === 'update_error'" class="alert alert-error">
      {{ $t('Alert.UpdateFailed') }}
    </div>

    <footer class="page-actions">
      <button class="secondary-button" type="button" @click="resetForm">
        {{ $t('Form.Cancel') }}
      </button>
      <button class="primary-button" type="submit">
        {{ $t('Form.Save') }}
      </button>
    </footer>
  </form>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'UserPwdSetting',
  data() {
    return {
      id: '',
      password: '',
      repeatPassword: '',
      submitted: false,
      touched: {
        password: false,
        repeatPassword: false,
      },
    };
  },
  mounted: async function() {
    this.$store.commit('User_ResetStatus');
    this.id = this.$route.params.id;
  },
  methods: {
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
    resetForm() {
      this.password = '';
      this.repeatPassword = '';
      this.submitted = false;
      this.touched = {
        password: false,
        repeatPassword: false,
      };
    },
    submit() {
      this.$store.commit('User_ResetStatus');
      this.submitted = true;
      this.markAllTouched();
      if (this.hasErrors) return;

      this.$store.dispatch('User_UpdatePwd', {
        id: this.id,
        password: this.password,
      });
    },
  },
  computed: {
    ...mapGetters({
      language: 'Auth_GetLanguage',
      status: 'User_GetStatus',
    }),
    hasErrors() {
      return [...this.passwordErrors, ...this.repeatPasswordErrors].length > 0;
    },
    passwordErrors() {
      if (!this.isTouched('password')) return [];
      return this.password ? [] : [this.$t('Form.Required')];
    },
    repeatPasswordErrors() {
      if (!this.isTouched('repeatPassword')) return [];
      return this.repeatPassword === this.password ? [] : [this.$t('Form.PwdNotMatch')];
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
