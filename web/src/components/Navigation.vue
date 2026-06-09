<template>
  <aside class="side-nav">
    <button class="side-logo" type="button" @click="toUrl('/dashboard')">
      <img :src="pilotwaveLightImage" alt="Pilotwave" />
    </button>

    <nav class="nav-list" :aria-label="$t('Shell.MainNavigation')">
      <router-link v-for="item in primaryItems" :key="item.to" class="nav-item" :to="item.to">
        <span class="nav-icon">{{ item.icon }}</span>
        <span>{{ item.label }}</span>
      </router-link>

      <div class="nav-separator"></div>

      <router-link v-if="isAdmin" class="nav-item" to="/users?page=1">
        <span class="nav-icon">AC</span>
        <span>{{ $t('System.AccountManagement') }}</span>
      </router-link>

      <button v-if="isAdmin" class="nav-item nav-button" type="button" @click="dialog = true">
        <span class="nav-icon">MS</span>
        <span>{{ $t('System.GrafanaManagement') }}</span>
      </button>

      <button
        class="nav-item nav-button"
        data-testid="nav-namespace-injection-open"
        type="button"
        @click="$emit('open-namespace-injection')"
      >
        <span class="nav-icon">IN</span>
        <span>{{ $t('NamespaceInjection.Button') }}</span>
      </button>

      <div class="nav-separator"></div>

      <button
        class="nav-item nav-button"
        data-testid="nav-language-open"
        type="button"
        @click="$emit('open-language-dialog')"
      >
        <span class="nav-icon">LA</span>
        <span>{{ $t('Shell.Language') }}</span>
      </button>
      <button
        class="nav-item nav-button"
        data-testid="nav-about-open"
        type="button"
        @click="$emit('open-about-dialog')"
      >
        <span class="nav-icon">AB</span>
        <span>{{ $t('Shell.About') }}</span>
      </button>

      <div class="nav-separator"></div>

      <button class="nav-item nav-button" type="button" @click="toLink('//istio.io/latest/docs')">
        <span class="nav-icon">DO</span>
        <span>{{ $t('System.Document') }}</span>
      </button>
      <button class="nav-item nav-button" type="button" @click="toLink('//www.brobridge.com/contact.php')">
        <span class="nav-icon">SU</span>
        <span>{{ $t('System.Support') }}</span>
      </button>
    </nav>

    <div v-if="dialog" class="modal-backdrop" @click.self="dialog = false">
      <section class="modal-card monitoring-dialog">
        <header class="monitoring-dialog-header">
          <p class="eyebrow">{{ $t('Monitoring.Eyebrow') }}</p>
          <h2>{{ $t('Monitoring.Setting') }}</h2>
        </header>

        <form class="form-stack monitoring-form" @submit.prevent="submit">
          <label class="field">
            <span>{{ $t('Monitoring.Provider') }}*</span>
            <select v-model="provider">
              <option value="grafana">{{ $t('Monitoring.ProviderGrafana') }}</option>
              <option value="prometheus">{{ $t('Monitoring.ProviderPrometheus') }}</option>
            </select>
          </label>

          <label class="field">
            <span>{{ $t('Monitoring.Host') }}*</span>
            <input v-model.trim="host" placeholder="grafana.monitoring.svc.cluster.local" type="text" @blur="touched.host = true" />
            <small v-if="hostError" class="field-error">{{ hostError }}</small>
          </label>

          <label class="field">
            <span>{{ $t('Monitoring.ConnectionScheme') }}*</span>
            <select v-model="scheme">
              <option value="http">HTTP</option>
              <option value="https">HTTPS / TLS</option>
            </select>
            <small class="field-help">{{ $t('Monitoring.ConnectionSchemeHelp') }}</small>
          </label>

          <label class="field">
            <span>{{ $t('Monitoring.Port') }}*</span>
            <input v-model.trim="port" inputmode="numeric" type="text" @blur="touched.port = true" />
            <small v-if="portError" class="field-error">{{ portError }}</small>
          </label>

          <label v-if="provider === 'grafana'" class="field">
            <span>{{ $t('Monitoring.DatasourceID') }}*</span>
            <input v-model.trim="datasourceId" inputmode="numeric" type="text" @blur="touched.datasourceId = true" />
            <small v-if="datasourceIdError" class="field-error">{{ datasourceIdError }}</small>
          </label>

          <label class="field">
            <span>{{ $t('Monitoring.Token') }}</span>
            <input v-model.trim="token" type="password" @blur="touched.token = true" />
            <small class="field-help">{{ $t('Monitoring.TokenHelp') }}</small>
          </label>

          <div class="connection-preview monitoring-preview">
            {{ $t('Monitoring.ConnectionPreview') }} <strong>{{ monitoringURLPreview }}</strong>
          </div>

          <label v-if="scheme === 'https'" class="checkbox-field stacked monitoring-wide-field">
            <span>
              <input v-model="skipTlsVerify" type="checkbox" />
              {{ $t('Monitoring.SkipTLSVerify') }}
            </span>
            <small>{{ $t('Monitoring.SkipTLSVerifyHelp') }}</small>
          </label>

          <div v-if="testMessage" class="alert monitoring-wide-field" :class="testOk ? 'alert-success' : 'alert-error'">
            {{ testMessage }}
          </div>

          <div v-if="status === 'update_error'" class="alert alert-error monitoring-wide-field">
            {{ $t('Alert.UpdateFailed') }} {{ error_handle }}
          </div>

          <footer class="modal-actions monitoring-dialog-actions">
            <button class="secondary-button" type="button" @click="dialog = false">
              {{ $t('Form.Cancel') }}
            </button>
            <button class="secondary-button" type="button" :disabled="testing" @click="testConnection">
              {{ testing ? $t('Monitoring.Testing') : $t('Monitoring.TestConnection') }}
            </button>
            <button class="primary-button" type="submit">
              {{ $t('Form.Save') }}
            </button>
          </footer>
        </form>
      </section>
    </div>

  </aside>
</template>

<script>
import { mapGetters } from 'vuex';
import pilotwaveLightImage from '../assets/pilotwave_light.png';

export default {
  name: 'Navigation',
  emits: ['open-namespace-injection', 'open-language-dialog', 'open-about-dialog'],
  data() {
    return {
      pilotwaveLightImage,
      dialog: false,
      submitted: false,
      touched: {
        host: false,
        port: false,
        datasourceId: false,
      },
      id: '',
      provider: 'grafana',
      host: '',
      port: '',
      token: '',
      datasourceId: '1',
      scheme: 'http',
      skipTlsVerify: false,
      testing: false,
      testOk: false,
      testMessage: '',
    };
  },
  mounted() {
    this.fetchData();
    window.addEventListener('keydown', this.handleEscapeKey);
  },
  beforeUnmount() {
    window.removeEventListener('keydown', this.handleEscapeKey);
  },
  methods: {
    async fetchData() {
      await this.$store.dispatch('Router_GetMenuItems', {
        page: 1,
        limit: -1,
      });

      await this.fetchGrafana();
    },
    async fetchGrafana() {
      const data = await this.$store.dispatch('Router_GetGrafana');
      const config = data?.grafana || {};
      this.id = config.id || '';
      this.provider = config.provider || 'grafana';
      this.host = config.host || '';
      this.port = config.port || '';
      this.token = config.token || '';
      this.datasourceId = config.datasourceId || '1';
      this.scheme = config.isTls ? 'https' : 'http';
      this.skipTlsVerify = Boolean(config.skipTlsVerify);
    },
    toUrl(url) {
      window.scrollTo(0, 0);
      if (url) this.$router.push(url);
    },
    toLink(url) {
      if (url) location.href = url;
    },
    handleEscapeKey(event) {
      if (event.key !== 'Escape' || !this.dialog) return;
      this.dialog = false;
    },
    validateMonitoringForm() {
      this.submitted = true;
      this.touched.host = true;
      this.touched.port = true;
      this.touched.datasourceId = true;
      return !(this.hostError || this.portError || this.datasourceIdError);
    },
    async testConnection() {
      if (!this.validateMonitoringForm()) return;

      this.testing = true;
      this.testMessage = '';
      const result = await this.$store.dispatch('Router_TestGrafana', this.monitoringPayload);
      this.testOk = Boolean(result?.ok);
      this.testMessage = result?.message || this.$t('Monitoring.TestFailed');
      this.testing = false;
    },
    async submit() {
      if (!this.validateMonitoringForm()) return;

      this.$store.commit('Router_ResetStatus');
      await this.$store.dispatch('Router_UpdateGrafana', this.monitoringPayload);
    },
  },
  watch: {
    async status(val) {
      if (val === 'update_success') {
        await this.fetchGrafana();
        this.dialog = false;
      }
    },
  },
  computed: {
    ...mapGetters({
      status: 'Router_GetStatus',
      error_handle: 'Router_GetErrorHandle',
      userInfo: 'Auth_GetUserInfo',
    }),
    primaryItems() {
      return [
        { to: '/dashboard', icon: 'DB', label: this.$t('System.Dashboard') },
        { to: '/gateways', icon: 'GW', label: this.$t('System.ServiceGateway') },
        { to: '/routers', icon: 'VS', label: this.$t('System.ServiceRouter') },
        { to: '/requestauths', icon: 'AU', label: this.$t('System.APIAuthentication') },
        { to: '/authpolicies', icon: 'PL', label: this.$t('System.BlackWhteList') },
        { to: '/tls-certificates', icon: 'TL', label: this.$t('System.TLSCertificates') },
      ];
    },
    isAdmin() {
      return Array.isArray(this.userInfo?.permissions) && this.userInfo.permissions.includes('admin');
    },
    monitoringPayload() {
      return {
        id: this.id,
        provider: this.provider,
        host: this.host,
        port: String(this.port),
        token: this.token,
        datasourceId: this.provider === 'grafana' ? this.datasourceId : '',
        isTls: this.scheme === 'https',
        skipTlsVerify: this.scheme === 'https' && this.skipTlsVerify,
      };
    },
    monitoringURLPreview() {
      const host = this.host || '<host>';
      const port = this.port || '<port>';
      return `${this.scheme}://${host}:${port}`;
    },
    hostError() {
      if (!this.submitted && !this.touched.host) return '';
      return this.host ? '' : this.$t('Form.Required');
    },
    portError() {
      if (!this.submitted && !this.touched.port) return '';
      if (!this.port) return this.$t('Form.Required');
      return /^\d+$/.test(String(this.port)) ? '' : this.$t('Form.Number');
    },
    datasourceIdError() {
      if (this.provider !== 'grafana') return '';
      if (!this.submitted && !this.touched.datasourceId) return '';
      if (!this.datasourceId) return this.$t('Form.Required');
      return /^\d+$/.test(String(this.datasourceId)) ? '' : this.$t('Form.Number');
    },
  },
};
</script>

<style scoped>
.monitoring-dialog {
  max-height: min(86vh, 820px);
  max-width: min(860px, calc(100vw - 48px));
  overflow: auto;
  padding: 26px 30px 28px;
}

.monitoring-dialog-header {
  margin-bottom: 18px;
}

.monitoring-dialog-header h2 {
  font-size: 2rem;
  letter-spacing: 0;
  line-height: 1.08;
  margin: 2px 0 0;
}

.monitoring-form {
  display: grid;
  gap: 18px 20px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.monitoring-form .field {
  min-width: 0;
}

.monitoring-form .field input,
.monitoring-form .field select {
  min-height: 52px;
}

.monitoring-preview,
.monitoring-wide-field,
.monitoring-dialog-actions {
  grid-column: 1 / -1;
}

.monitoring-preview {
  border-radius: 10px;
  padding: 16px 18px;
}

.monitoring-dialog-actions {
  justify-content: center;
  margin-top: 4px;
}

.monitoring-dialog-actions .primary-button,
.monitoring-dialog-actions .secondary-button {
  min-width: 118px;
}

@media (max-width: 760px) {
  .monitoring-dialog {
    max-height: calc(100vh - 28px);
    max-width: calc(100vw - 28px);
    padding: 22px;
  }

  .monitoring-form {
    grid-template-columns: 1fr;
  }
}
</style>
