<template>
  <section class="page">
    <header class="page-header">
      <div>
        <p class="eyebrow">{{ $t('Gateway.TLSCertificatesEyebrow') }}</p>
        <h1>{{ isDetailMode ? $t('Gateway.TLSCertificateDetail') : $t('System.TLSCertificates') }}</h1>
      </div>
      <div class="page-actions">
        <button v-if="isDetailMode" class="secondary-button" type="button" @click="backToList">
          {{ $t('Form.BackToTLSCertificates') }}
        </button>
        <button class="secondary-button" type="button" :disabled="loading" @click="fetchData">
          {{ $t('Form.Reload') }}
        </button>
      </div>
    </header>

    <section class="panel">
      <div v-if="loading" class="empty-state compact">
        {{ $t('Gateway.TLSLoading') }}
      </div>

      <div v-else-if="loadError" class="alert alert-error">
        {{ loadError }}
      </div>

      <article v-else-if="selectedCertificate" class="certificate-detail" data-testid="tls-certificate-detail">
        <header class="certificate-detail-header">
          <div>
            <h2>{{ selectedCertificate.name }}</h2>
            <p>{{ selectedCertificate.gatewayNamespace }} / {{ selectedCertificate.gatewayName }}</p>
          </div>
          <span class="status-pill">{{ $t(selectedCertificate.statusLabel, { days: selectedCertificate.daysUntilExpiry }) }}</span>
        </header>

        <div class="certificate-summary-grid">
          <section class="certificate-summary-card">
            <span>{{ $t('Gateway.TLSExpiresAt') }}</span>
            <strong>{{ selectedCertificate.expiresAt }}</strong>
          </section>
          <section class="certificate-summary-card">
            <span>{{ $t('Table.Protocol') }}</span>
            <strong>{{ selectedCertificate.protocol || 'TLS' }} / {{ selectedCertificate.port || '-' }}</strong>
          </section>
          <section class="certificate-summary-card">
            <span>{{ $t('Gateway.SecretName') }}</span>
            <strong>{{ selectedCertificate.secret }}</strong>
          </section>
        </div>

        <div class="certificate-info-layout">
          <section class="certificate-info-card">
            <h3>{{ $t('System.ServiceGateway') }}</h3>
            <dl class="certificate-detail-grid">
              <div>
                <dt>{{ $t('Table.Name') }}</dt>
                <dd>{{ selectedCertificate.gatewayName }}</dd>
              </div>
              <div>
                <dt>{{ $t('Table.Namespace') }}</dt>
                <dd>{{ selectedCertificate.gatewayNamespace }}</dd>
              </div>
            </dl>
          </section>

          <section class="certificate-info-card">
            <h3>{{ $t('Gateway.TLSCertificateStatus') }}</h3>
            <dl class="certificate-detail-grid">
              <div>
                <dt>{{ $t('Gateway.TLSIssuer') }}</dt>
                <dd>{{ selectedCertificate.issuer || '-' }}</dd>
              </div>
              <div>
                <dt>{{ $t('Gateway.TLSSubject') }}</dt>
                <dd>{{ selectedCertificate.subject || '-' }}</dd>
              </div>
              <div>
                <dt>{{ $t('Gateway.TLSSAN') }}</dt>
                <dd>{{ (selectedCertificate.dnsNames || []).join(', ') || '-' }}</dd>
              </div>
              <div v-if="selectedCertificate.reason">
                <dt>{{ $t('Gateway.TLSReason') }}</dt>
                <dd>{{ selectedCertificate.reason }}</dd>
              </div>
            </dl>
          </section>

          <section class="certificate-info-card wide">
            <h3>{{ $t('Gateway.TLSFingerprint') }}</h3>
            <p class="fingerprint">{{ selectedCertificate.fingerprintSHA256 || '-' }}</p>
          </section>
        </div>

        <footer class="certificate-detail-actions">
          <button class="primary-button" type="button" @click="openUpdate(selectedCertificate)">
            {{ $t('Gateway.TLSRenewOrUpdate') }}
          </button>
          <button class="secondary-button" type="button" @click="openStatus(selectedCertificate)">
            {{ $t('Gateway.TLSViewStatus') }}
          </button>
        </footer>
      </article>

      <div v-else-if="isDetailMode" class="empty-state empty-state-list">
        <div class="empty-illustration" aria-hidden="true"></div>
        <div class="empty-copy">
          <p class="eyebrow">{{ $t('ResourceList.NoMatches') }}</p>
          <h2>{{ $t('Gateway.TLSCertificateNotFound') }}</h2>
          <p>{{ $t('Gateway.TLSCertificateNotFoundHelp') }}</p>
          <button class="primary-button" type="button" @click="backToList">
            {{ $t('Form.BackToTLSCertificates') }}
          </button>
        </div>
      </div>

      <div v-else-if="!certificates.length" class="empty-state empty-state-list">
        <div class="empty-illustration" aria-hidden="true"></div>
        <div class="empty-copy">
          <p class="eyebrow">{{ $t('ResourceList.ReadyForSetup') }}</p>
          <h2>{{ $t('Gateway.TLSNoCertificates') }}</h2>
          <p>{{ $t('Gateway.TLSNoCertificatesListHelp') }}</p>
          <router-link class="primary-button" to="/gateways">
            {{ $t('System.ServiceGateway') }}
          </router-link>
        </div>
      </div>

      <div v-else class="table-wrap">
        <table class="data-table tls-certificate-table">
          <thead>
            <tr>
              <th class="index-col">#</th>
              <th>{{ $t('Table.Name') }}</th>
              <th>{{ $t('Gateway.SecretName') }}</th>
              <th>{{ $t('System.ServiceGateway') }}</th>
              <th>{{ $t('Gateway.TLSExpiresAt') }}</th>
              <th>{{ $t('Gateway.TLSCertificateStatus') }}</th>
              <th class="text-center">{{ $t('Form.Action') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(item, index) in certificates"
              :key="item.id"
              class="resource-row"
              data-testid="tls-certificate-row"
              tabindex="0"
              @click="openDetail(item)"
              @keydown.enter="openDetail(item)"
            >
              <td class="index-col">{{ index + 1 }}</td>
              <td>
                <button class="link-button" type="button" @click.stop="openDetail(item)">
                  {{ item.name }}
                </button>
              </td>
              <td>{{ item.secret }}</td>
              <td>{{ item.gatewayNamespace }} / {{ item.gatewayName }}</td>
              <td>{{ item.expiresAt }}</td>
              <td>
                <span class="status-pill">{{ $t(item.statusLabel, { days: item.daysUntilExpiry }) }}</span>
              </td>
              <td class="text-center">
                <button class="secondary-button" type="button" @click.stop="openUpdate(item)">
                  {{ $t('Gateway.TLSRenewOrUpdate') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <footer class="table-footer">
        <span>{{ $t('Table.Total') }}: {{ certificates.length }}</span>
      </footer>
    </section>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'TLSCertificates',
  data() {
    return {
      certificates: [],
      loading: false,
      loadError: '',
    };
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    async fetchData() {
      this.loading = true;
      this.loadError = '';

      try {
        const gateways = await this.$store.dispatch('Gateway_GetItems', {
          page: 1,
          namespace: this.namespace === 'All' ? '' : this.namespace || '',
          limit: -1,
        });

        const certificateGroups = await Promise.all(
          (gateways || []).map(async (gateway) => {
            const certificates = await this.$store.dispatch('Gateway_GetTLSCertificates', {
              name: gateway.name,
              namespace: gateway.namespace,
            });
            return certificates.map((certificate) => this.decorateCertificate(certificate, gateway));
          })
        );

        this.certificates = certificateGroups.flat();
      } catch (err) {
        this.certificates = [];
        this.loadError = err?.message || this.$t('Alert.LoadFailed');
      } finally {
        this.loading = false;
      }
    },
    decorateCertificate(certificate, gateway) {
      const gatewayName = gateway.name || '';
      const gatewayNamespace = gateway.namespace || '';
      const name = certificate.hosts?.[0] || certificate.credentialName || certificate.secretName || this.$t('Gateway.TLSUnknownHost');
      const secretName = certificate.secretName || certificate.credentialName || '-';
      const secretNamespace = certificate.secretNamespace || gatewayNamespace || '-';

      return {
        ...certificate,
        id: [
          gatewayNamespace,
          gatewayName,
          certificate.serverIndex,
          certificate.port,
          secretNamespace,
          secretName,
        ].join(':'),
        gatewayName,
        gatewayNamespace,
        name,
        secret: `${secretNamespace} / ${secretName}`,
        expiresAt: this.formatDate(certificate.notAfter),
        statusLabel: this.statusLabelKey(certificate.status),
        detailTo: `/tls-certificates/${encodeURIComponent([
          gatewayNamespace,
          gatewayName,
          certificate.serverIndex,
          certificate.port,
          secretNamespace,
          secretName,
        ].join(':'))}`,
        updateTo: this.gatewayUrl(gateway, 'setting'),
        statusTo: this.gatewayUrl(gateway, 'tls'),
      };
    },
    gatewayUrl(gateway, tab) {
      const name = encodeURIComponent(gateway.name || '');
      const namespace = encodeURIComponent(gateway.namespace || '');
      return `/gateway/${name}?name=${name}&namespace=${namespace}&tab=${tab}`;
    },
    openDetail(item) {
      this.$router.push(item.detailTo);
    },
    openUpdate(item) {
      this.$router.push(item.updateTo);
    },
    openStatus(item) {
      this.$router.push(item.statusTo);
    },
    backToList() {
      this.$router.push('/tls-certificates');
    },
    statusLabelKey(status) {
      if (status === 'healthy') return 'Gateway.TLSHealthy';
      if (status === 'warning') return 'Gateway.TLSWarning';
      if (status === 'critical') return 'Gateway.TLSCritical';
      if (status === 'expired') return 'Gateway.TLSExpired';
      if (status === 'missing') return 'Gateway.TLSMissing';
      if (status === 'invalid') return 'Gateway.TLSInvalid';
      return 'Gateway.TLSUnknown';
    },
    formatDate(value) {
      if (!value) return '-';
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) return value;
      return `${date.toISOString().slice(0, 16).replace('T', ' ')} UTC`;
    },
  },
  watch: {
    async namespace() {
      await this.fetchData();
    },
  },
  computed: {
    ...mapGetters({
      namespace: 'Auth_GetNamespace',
    }),
    selectedCertificateId() {
      const id = this.$route?.params?.id || '';
      return decodeURIComponent(id);
    },
    isDetailMode() {
      return Boolean(this.selectedCertificateId);
    },
    selectedCertificate() {
      const id = decodeURIComponent(this.$route?.params?.id || '');
      if (!id) return null;
      return this.certificates.find((item) => item.id === id) || null;
    },
  },
};
</script>

<style scoped>
.tls-certificate-table td {
  vertical-align: middle;
}

.certificate-detail {
  align-items: start;
  background:
    linear-gradient(135deg, rgba(58, 91, 217, 0.08), transparent 34%),
    var(--pw-surface);
  border: 1px solid var(--pw-border);
  border-radius: 18px;
  display: grid;
  gap: 18px;
  padding: 24px;
}

.certificate-detail-header {
  align-items: start;
  display: flex;
  gap: 16px;
  justify-content: space-between;
}

.certificate-detail-header p:not(.eyebrow) {
  color: var(--pw-muted);
  font-weight: 800;
  margin: 8px 0 0;
}

.certificate-detail h2 {
  font-size: clamp(1.65rem, 3vw, 2.4rem);
  letter-spacing: 0;
  margin: 4px 0 0;
}

.certificate-summary-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.certificate-summary-card,
.certificate-info-card {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 14px;
}

.certificate-summary-card {
  display: grid;
  gap: 6px;
  min-width: 0;
  padding: 16px;
}

.certificate-summary-card span {
  color: var(--pw-muted);
  font-size: 0.74rem;
  font-weight: 900;
  text-transform: uppercase;
}

.certificate-summary-card strong {
  font-size: 1.05rem;
  overflow-wrap: anywhere;
}

.certificate-info-layout {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.certificate-info-card {
  min-width: 0;
  padding: 18px;
}

.certificate-info-card.wide {
  grid-column: 1 / -1;
}

.certificate-info-card h3 {
  font-size: 1rem;
  letter-spacing: 0;
  margin: 0 0 14px;
}

.certificate-detail-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
}

.certificate-detail-grid div {
  min-width: 0;
}

.certificate-detail-grid dt {
  color: var(--pw-muted);
  font-size: 0.74rem;
  font-weight: 900;
  text-transform: uppercase;
}

.certificate-detail-grid dd {
  font-weight: 700;
  margin: 4px 0 0;
  overflow-wrap: anywhere;
}

.certificate-detail-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.status-pill {
  background: var(--pw-surface-muted);
  border-radius: 999px;
  color: var(--pw-primary-strong);
  font-weight: 800;
  padding: 8px 12px;
  white-space: nowrap;
}

.fingerprint {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.85rem;
  line-height: 1.6;
  margin: 0;
  overflow-wrap: anywhere;
  padding: 14px;
}

@media (max-width: 900px) {
  .certificate-summary-grid,
  .certificate-info-layout,
  .certificate-detail-grid {
    grid-template-columns: 1fr;
  }

  .certificate-detail-header,
  .certificate-detail-actions {
    display: grid;
  }
}
</style>
