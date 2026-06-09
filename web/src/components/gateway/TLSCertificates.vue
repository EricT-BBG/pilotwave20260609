<template>
  <section class="tls-panel">
    <div class="section-heading">
      <div>
        <h2>{{ $t('Gateway.TLSCertificateStatus') }}</h2>
        <p>{{ $t('Gateway.TLSCertificates') }}</p>
      </div>
      <button class="secondary-button" type="button" @click="loadCertificates">
        {{ $t('Form.Reload') }}
      </button>
    </div>

    <div v-if="loading" class="empty-state compact">
      {{ $t('Gateway.TLSLoading') }}
    </div>

    <div v-else-if="!certificates.length" class="empty-state">
      <h3>{{ $t('Gateway.TLSNoCertificates') }}</h3>
      <p>{{ $t('Gateway.TLSNoCertificatesHelp') }}</p>
    </div>

    <div v-else class="certificate-grid">
      <article
        v-for="item in certificates"
        :key="certificateKey(item)"
        class="certificate-card"
        :class="`status-${item.status || 'unknown'}`"
      >
        <div class="certificate-card-header">
          <div>
            <strong>{{ primaryHost(item) }}</strong>
            <span class="certificate-meta-line">{{ item.protocol || 'TLS' }} : {{ item.port || '-' }}</span>
          </div>
          <span class="status-pill">{{ statusLabel(item) }}</span>
        </div>

        <dl class="certificate-fields certificate-field-grid">
          <div>
            <dt>{{ $t('Gateway.SecretName') }}</dt>
            <dd>{{ item.secretNamespace || '-' }} / {{ item.secretName || item.credentialName || '-' }}</dd>
          </div>
          <div>
            <dt>{{ $t('Gateway.TLSExpiresAt') }}</dt>
            <dd>{{ formatDate(item.notAfter) }}</dd>
          </div>
          <div>
            <dt>{{ $t('Gateway.TLSIssuer') }}</dt>
            <dd>{{ item.issuer || '-' }}</dd>
          </div>
          <div>
            <dt>{{ $t('Gateway.TLSSubject') }}</dt>
            <dd>{{ item.subject || '-' }}</dd>
          </div>
          <div>
            <dt>{{ $t('Gateway.TLSSAN') }}</dt>
            <dd>{{ (item.dnsNames || []).join(', ') || '-' }}</dd>
          </div>
          <div v-if="item.reason">
            <dt>{{ $t('Gateway.TLSReason') }}</dt>
            <dd>{{ item.reason }}</dd>
          </div>
          <div class="certificate-fingerprint-block">
            <dt>{{ $t('Gateway.TLSFingerprint') }}</dt>
            <dd class="fingerprint">{{ item.fingerprintSHA256 || '-' }}</dd>
          </div>
        </dl>

        <div class="certificate-card-footer">
          <span>{{ item.credentialName || item.secretName || '-' }}</span>
        </div>
      </article>
    </div>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'GatewayTLSCertificates',
  props: {
    name: {
      type: String,
      required: true,
    },
    namespace: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      loading: false,
    };
  },
  mounted() {
    this.loadCertificates();
  },
  methods: {
    async loadCertificates() {
      if (!this.name || !this.namespace) return;
      this.loading = true;
      await this.$store.dispatch('Gateway_GetTLSCertificates', {
        name: this.name,
        namespace: this.namespace,
      });
      this.loading = false;
    },
    certificateKey(item) {
      return [
        item.serverIndex,
        item.port,
        item.secretNamespace,
        item.secretName,
        item.credentialName,
      ].filter((value) => value !== undefined && value !== '').join(':');
    },
    primaryHost(item) {
      return item.hosts?.[0] || item.credentialName || this.$t('Gateway.TLSUnknownHost');
    },
    statusLabel(item) {
      const status = item.status || 'unknown';
      if (status === 'healthy') return this.$t('Gateway.TLSHealthy', { days: item.daysUntilExpiry });
      if (status === 'warning') return this.$t('Gateway.TLSWarning', { days: item.daysUntilExpiry });
      if (status === 'critical') return this.$t('Gateway.TLSCritical', { days: item.daysUntilExpiry });
      if (status === 'expired') return this.$t('Gateway.TLSExpired');
      if (status === 'missing') return this.$t('Gateway.TLSMissing');
      if (status === 'invalid') return this.$t('Gateway.TLSInvalid');
      return this.$t('Gateway.TLSUnknown');
    },
    formatDate(value) {
      if (!value) return '-';
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) return value;
      return date.toISOString().replace('.000Z', ' UTC');
    },
  },
  computed: {
    ...mapGetters({
      certificates: 'Gateway_GetTLSCertificates',
    }),
  },
  watch: {
    name() {
      this.loadCertificates();
    },
    namespace() {
      this.loadCertificates();
    },
  },
};
</script>

<style scoped>
.tls-panel {
  display: grid;
  gap: 18px;
}

.section-heading {
  align-items: center;
  display: flex;
  gap: 16px;
  justify-content: space-between;
}

.section-heading h2 {
  margin: 0;
}

.section-heading p {
  color: var(--pw-muted);
  margin: 6px 0 0;
}

.certificate-grid {
  display: grid;
  gap: 16px;
}

.certificate-card {
  background: var(--pw-surface);
  border: 1px solid var(--pw-border);
  border-left: 5px solid var(--pw-muted);
  border-radius: 14px;
  box-shadow: var(--pw-shadow-soft);
  display: grid;
  gap: 14px;
  padding: 16px 18px;
}

.certificate-card.status-healthy {
  border-left-color: #2f8f5b;
}

.certificate-card.status-warning {
  border-left-color: #d98d22;
}

.certificate-card.status-critical,
.certificate-card.status-expired,
.certificate-card.status-missing,
.certificate-card.status-invalid {
  border-left-color: #c7462d;
}

.certificate-card-header {
  align-items: center;
  border-bottom: 1px solid var(--pw-border);
  display: flex;
  gap: 16px;
  justify-content: space-between;
  padding-bottom: 12px;
}

.certificate-card-header strong {
  display: block;
  font-size: 1.1rem;
}

.certificate-meta-line {
  color: var(--pw-muted);
  display: block;
  margin-top: 4px;
}

.status-pill {
  background: var(--pw-surface-muted);
  border-radius: 999px;
  color: var(--pw-primary-strong);
  font-weight: 800;
  padding: 8px 12px;
  white-space: nowrap;
}

.certificate-fields,
.certificate-field-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 0;
}

.certificate-fields div {
  min-width: 0;
}

.certificate-fields dt {
  color: var(--pw-muted);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.certificate-fields dd {
  margin: 4px 0 0;
  overflow-wrap: anywhere;
}

.fingerprint {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #243142;
  display: block;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: clamp(0.72rem, 1vw, 0.86rem);
  line-height: 1.55;
  max-width: 100%;
  padding: 10px 12px;
  word-break: break-all;
}

.certificate-fingerprint-block {
  grid-column: 1 / -1;
}

.certificate-card-footer {
  color: var(--pw-muted);
  display: flex;
  font-size: 0.84rem;
  justify-content: flex-end;
}

@media (max-width: 980px) {
  .certificate-fields,
  .certificate-field-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 680px) {
  .section-heading,
  .certificate-card-header {
    align-items: stretch;
    flex-direction: column;
  }

  .certificate-fields,
  .certificate-field-grid {
    grid-template-columns: 1fr;
  }
}
</style>
