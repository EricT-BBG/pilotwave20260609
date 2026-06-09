<template>
  <div class="auth-page">
    <header class="auth-header">
      <button class="brand-button" type="button" @click="toUrl('/')">
        <img :src="logoImage" alt="Brobridge" />
      </button>
    </header>

    <main class="auth-layout">
      <section class="auth-card">
        <img class="auth-product-logo" :src="pilotwaveDarkImage" alt="Pilotwave" />
        <p class="eyebrow">BROBRIDGE Cloud API Management</p>
        <h1>{{ $t('System.SigninAccount') }}</h1>

        <form class="form-stack" @submit.prevent="submit">
          <label class="field">
            <span>{{ $t('Form.Account') }}*</span>
            <input
              data-testid="login-account"
              v-model.trim="account"
              autocomplete="username"
              type="text"
              @blur="touched.account = true"
            />
            <small v-if="accountError" class="field-error">{{ accountError }}</small>
          </label>

          <label class="field">
            <span>{{ $t('Form.Password') }}*</span>
            <input
              data-testid="login-password"
              v-model.trim="password"
              autocomplete="current-password"
              type="password"
              @blur="touched.password = true"
              @keydown.enter.prevent="submit"
            />
            <small v-if="passwordError" class="field-error">{{ passwordError }}</small>
          </label>

          <div class="language-actions">
            <button type="button" class="text-button" @click="switchLang('tw')">繁中</button>
            <button type="button" class="text-button" @click="switchLang('en')">English</button>
          </div>

          <div v-if="status === 'signin_error'" class="alert alert-error">
            {{ $t('Alert.ConfirmPw') }}
          </div>

          <button class="primary-button full-width" data-testid="login-submit" type="submit">
            {{ $t('Form.Signin') }}
          </button>
        </form>

        <p class="auth-build-info">
          Pilotwave {{ buildInfo.version }} · Build {{ buildInfo.buildLabel }}
        </p>
      </section>
    </main>
  </div>
</template>

<script>
import { mapGetters } from 'vuex';
import logoImage from '../assets/logo.png';
import pilotwaveDarkImage from '../assets/pilotwave_dark.png';
import { buildInfo } from '../lib/buildInfo';
import { normalizeLocale, resolveLocale, SUPPORTED_LOCALES } from '../lib/locale';

export default {
  name: 'Landingpage',
  data() {
    return {
      logoImage,
      pilotwaveDarkImage,
      buildInfo,
      account: '',
      password: '',
      submitted: false,
      touched: {
        account: false,
        password: false,
      },
    };
  },
  mounted() {
    this.$store.commit('Auth_ResetStatus');
    this.applyLocale(this.language || resolveLocale());

    if (sessionStorage.getItem('accessToken')) {
      this.$router.replace('/dashboard');
    }
  },
  methods: {
    applyLocale(newLang) {
      const normalizedLang = normalizeLocale(newLang);
      const targetLang = SUPPORTED_LOCALES.includes(normalizedLang) ? normalizedLang : resolveLocale();
      if (this.$i18n?.locale && typeof this.$i18n.locale === 'object' && 'value' in this.$i18n.locale) {
        this.$i18n.locale.value = targetLang;
      } else if (this.$i18n) {
        this.$i18n.locale = targetLang;
      }

      this.$store.commit('Auth_SetLanguage', {
        lang: targetLang,
      });
    },
    toUrl(url) {
      window.scrollTo(0, 0);
      if (url) this.$router.push(url);
    },
    submit() {
      this.submitted = true;
      this.touched.account = true;
      this.touched.password = true;

      if (this.accountError || this.passwordError) return;

      this.$store.commit('Auth_ResetStatus');
      this.$store.dispatch('Auth_Signin', {
        account: this.account,
        password: this.password,
      });
    },
    switchLang(newLang) {
      this.applyLocale(newLang);
    },
  },
  watch: {
    status(val) {
      if (val === 'signin_success') this.$router.push('/dashboard');
    },
  },
  computed: {
    ...mapGetters({
      language: 'Auth_GetLanguage',
      status: 'Auth_GetStatus',
    }),
    accountError() {
      if (!this.submitted && !this.touched.account) return '';
      return this.account ? '' : this.$t('Form.Required');
    },
    passwordError() {
      if (!this.submitted && !this.touched.password) return '';
      return this.password ? '' : this.$t('Form.Required');
    },
  },
};
</script>
