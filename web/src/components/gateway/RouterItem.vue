<template>
  <article class="port-card">
    <button
      v-if="canRemovePort"
      class="remove-port-link"
      :data-testid="`gateway-port-remove-${serverIndex}-${index}`"
      type="button"
      @click="removePort"
    >
      {{ $t('Gateway.RemovePort') }}
    </button>

    <div class="port-grid">
      <label class="field">
        <span>{{ $t('Gateway.Protocol') }}</span>
        <select
          data-testid="gateway-port-protocol"
          :data-server-index="serverIndex"
          :data-port-index="index"
          :value="protocol"
          @change="handleProtocolChange($event.target.value)"
        >
          <option v-for="item in protocols" :key="item.value" :value="item.value">
            {{ item.name }}
          </option>
        </select>
      </label>

      <label class="field">
        <span>{{ $t('Gateway.Port') }}</span>
        <input
          data-testid="gateway-port-number"
          :data-server-index="serverIndex"
          :data-port-index="index"
          :value="port"
          placeholder="80"
          type="number"
          @input="updatePort('port', $event.target.value)"
        />
      </label>
    </div>

    <div v-if="requiresTlsMaterial" class="tls-panel">
      <div class="tls-panel-header">
        <div>
          <strong>{{ $t('Gateway.TLSCredential') }}</strong>
          <p>{{ credentialSummaryText }}</p>
        </div>
        <div class="tls-action-stack">
          <button
            class="secondary-button"
            type="button"
            :aria-describedby="tlsConfigureHintId"
            :title="$t('Gateway.ConfigureTLSHint')"
            @click="openTLSDialog"
          >
            {{ $t('Gateway.ConfigureTLS') }}
          </button>
          <small :id="tlsConfigureHintId">{{ $t('Gateway.ConfigureTLSHint') }}</small>
        </div>
      </div>

      <dl class="tls-summary-grid">
        <div class="tls-summary-item">
          <dt>{{ $t('Gateway.TLSCredentialSource') }}</dt>
          <dd>{{ credentialModeLabel }}</dd>
        </div>
        <div class="tls-summary-item">
          <dt>{{ $t('Gateway.TLSMode') }}</dt>
          <dd>{{ normalizedMode }}</dd>
        </div>
        <div class="tls-summary-item wide">
          <dt>{{ credentialMode === 'existing' ? $t('Gateway.SecretName') : $t('Gateway.Certificate') }}</dt>
          <dd>{{ credentialDetailText }}</dd>
        </div>
      </dl>

      <small v-if="normalizedMode === 'MUTUAL' && credentialMode === 'upload' && !cacert" class="field-error">
        {{ $t('Gateway.CARequired') }}
      </small>

      <div v-if="tlsDialogOpen" class="dialog-backdrop" @click.self="closeTLSDialog">
        <section
          ref="tlsDialog"
          class="tls-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="$t('Gateway.TLSDialogTitle')"
          tabindex="-1"
          @keydown.esc.prevent.stop="closeTLSDialog"
          @keydown.enter="handleTLSDialogEnter"
          @wheel.stop
        >
          <header class="dialog-header">
            <div>
              <strong>{{ $t('Gateway.TLSDialogTitle') }}</strong>
              <p>{{ $t('Gateway.TLSDialogHelp') }}</p>
            </div>
          </header>

          <div class="dialog-body">
            <div v-if="tlsDialogError" class="tls-dialog-error">
              {{ tlsDialogError }}
            </div>

            <section v-if="currentTLSCertificate" class="current-certificate-card" data-testid="tls-current-certificate">
              <header>
                <div>
                  <span>{{ $t('Gateway.TLSCurrentCertificate') }}</span>
                  <strong>{{ currentTLSCertificate.hosts?.[0] || currentTLSCertificate.credentialName || currentTLSCertificate.secretName || '-' }}</strong>
                </div>
                <span class="status-pill">
                  {{ $t(statusLabelKey(currentTLSCertificate.status), { days: currentTLSCertificate.daysUntilExpiry }) }}
                </span>
              </header>
              <dl class="current-certificate-grid">
                <div>
                  <dt>{{ $t('Gateway.SecretName') }}</dt>
                  <dd>{{ currentCertificateSecret }}</dd>
                </div>
                <div>
                  <dt>{{ $t('Gateway.TLSExpiresAt') }}</dt>
                  <dd>{{ formatCertificateDate(currentTLSCertificate.notAfter) }}</dd>
                </div>
              </dl>
              <details class="current-certificate-details">
                <summary>{{ $t('Gateway.TLSCertificateDetail') }}</summary>
                <dl class="current-certificate-grid secondary">
                  <div>
                    <dt>{{ $t('Gateway.TLSIssuer') }}</dt>
                    <dd>{{ currentTLSCertificate.issuer || '-' }}</dd>
                  </div>
                  <div>
                    <dt>{{ $t('Gateway.TLSSAN') }}</dt>
                    <dd>{{ (currentTLSCertificate.dnsNames || currentTLSCertificate.hosts || []).join(', ') || '-' }}</dd>
                  </div>
                </dl>
                <details class="certificate-fingerprint-details">
                  <summary>{{ $t('Gateway.TLSFingerprint') }}</summary>
                  <p class="certificate-fingerprint">{{ currentTLSCertificate.fingerprintSHA256 || '-' }}</p>
                </details>
              </details>
            </section>

            <div class="tls-choice-row">
              <div class="credential-source-grid" role="radiogroup" :aria-label="$t('Gateway.TLSCredential')">
                <button
                  type="button"
                  class="credential-source-card"
                  :class="{ active: draftCredentialMode === 'upload' }"
                  role="radio"
                  :aria-checked="draftCredentialMode === 'upload'"
                  @click="setDraftCredentialMode('upload')"
                >
                  <span class="source-icon">↑</span>
                  <span>
                    <strong>{{ $t('Gateway.UploadPaste') }}</strong>
                    <small>{{ $t('Gateway.UploadPEMHelp') }}</small>
                  </span>
                </button>

                <button
                  type="button"
                  class="credential-source-card"
                  :class="{ active: draftCredentialMode === 'existing' }"
                  role="radio"
                  :aria-checked="draftCredentialMode === 'existing'"
                  @click="setDraftCredentialMode('existing')"
                >
                  <span class="source-icon">#</span>
                  <span>
                    <strong>{{ $t('Gateway.ExistingSecret') }}</strong>
                    <small>{{ $t('Gateway.ExistingSecretHelp') }}</small>
                  </span>
                </button>
              </div>

              <div class="tls-settings-grid">
                <label class="field">
                  <span>{{ $t('Gateway.TLSMode') }}</span>
                  <select :value="draftMode" @change="updateDraft('mode', $event.target.value)">
                    <option value="SIMPLE">{{ $t('Gateway.ServerTLS') }}</option>
                    <option value="MUTUAL">{{ $t('Gateway.MutualTLS') }}</option>
                  </select>
                </label>

                <label v-if="draftCredentialMode === 'existing'" class="field">
                  <span>{{ $t('Gateway.SecretName') }}</span>
                  <input
                    :value="tlsDraft.credentialname"
                    placeholder="pilotwave-app-tls"
                    type="text"
                    @input="updateDraftExistingSecret($event.target.value)"
                  />
                </label>
              </div>
            </div>

            <section v-if="draftCredentialMode === 'upload'" class="pem-material-panel">
              <div
                class="drop-zone"
                @dragover.prevent
                @drop.prevent="handleDrop"
              >
                <input ref="fileInput" multiple type="file" accept=".pem,.crt,.key" @change="handleFileInput" />
                <button class="secondary-button upload-action" type="button" @click="$refs.fileInput.click()">
                  <span>↑</span>
                  {{ $t('Gateway.UploadPEM') }}
                </button>
                <span>{{ $t('Gateway.PEMAutoDetectHelp') }}</span>
              </div>

              <div class="port-grid tls-grid">
                <label class="field">
                  <span>{{ $t('Gateway.Certificate') }}</span>
                  <textarea
                    :placeholder="$t('Gateway.CertificateContent')"
                    :value="tlsDraft.cert"
                    rows="3"
                    @input="updateDraftCertificate($event.target.value)"
                  ></textarea>
                </label>

                <label class="field">
                  <span>{{ $t('Gateway.PrivateKey') }}</span>
                  <textarea
                    :placeholder="$t('Gateway.PrivateKeyContent')"
                    :value="tlsDraft.pkey"
                    rows="3"
                    @input="updateDraftPrivateKey($event.target.value)"
                  ></textarea>
                </label>

                <label v-if="draftMode === 'MUTUAL'" class="field ca-field">
                  <span>{{ $t('Gateway.CACertificate') }}</span>
                  <textarea
                    :placeholder="$t('Gateway.CACertificateContent')"
                    :value="tlsDraft.cacert"
                    rows="3"
                    @input="updateDraft('cacert', $event.target.value)"
                  ></textarea>
                </label>
              </div>

              <dl class="credential-summary">
                <div>
                  <dt>{{ $t('Gateway.ParsedCertificateCount') }}</dt>
                  <dd>{{ recognizedCountLabel(draftTLSSummary.certificateCount) }}</dd>
                </div>
                <div>
                  <dt>{{ $t('Gateway.ParsedPrivateKey') }}</dt>
                  <dd>{{ draftTLSSummary.hasPrivateKey ? $t('Gateway.Present') : $t('Gateway.NotRecognized') }}</dd>
                </div>
                <div>
                  <dt>{{ $t('Gateway.ParsedCACertificateCount') }}</dt>
                  <dd>{{ recognizedCountLabel(draftTLSSummary.caCertificateCount) }}</dd>
                </div>
              </dl>
            </section>

            <small v-if="draftMode === 'MUTUAL' && draftCredentialMode === 'upload' && !tlsDraft.cacert" class="field-error">
              {{ $t('Gateway.CARequired') }}
            </small>
          </div>

          <footer class="dialog-footer">
            <button class="secondary-button" type="button" :disabled="tlsChecking" @click="closeTLSDialog">
              {{ $t('Form.Cancel') }}
            </button>
            <button class="primary-button" type="button" :disabled="tlsChecking" @click="confirmTLSDialog">
              {{ tlsChecking ? $t('Gateway.CheckingSecret') : $t('Gateway.Confirm') }}
            </button>
          </footer>
        </section>
      </div>
    </div>

    <div v-if="protocolConfirmOpen" class="dialog-backdrop" @click.self="cancelProtocolChange">
      <section class="confirm-dialog" role="dialog" aria-modal="true" :aria-label="$t('Gateway.TLSProtocolChangeTitle')">
        <header class="dialog-header">
          <div>
            <strong>{{ $t('Gateway.TLSProtocolChangeTitle') }}</strong>
            <p>{{ $t('Gateway.TLSProtocolChangeText') }}</p>
          </div>
        </header>
        <footer class="dialog-footer">
          <button class="secondary-button" type="button" @click="cancelProtocolChange">
            {{ $t('Form.Cancel') }}
          </button>
          <button class="danger-button" type="button" @click="confirmProtocolChange">
            {{ $t('Gateway.RemoveTLSAndContinue') }}
          </button>
        </footer>
      </section>
    </div>
  </article>
</template>

<script>
import { mapGetters } from 'vuex';
import { splitTLSPEM, summarizeTLSPEM } from '../../lib/pem';

export default {
  name: 'GatewayRouterItems',
  props: ['protocol', 'port', 'cert', 'pkey', 'cacert', 'credentialname', 'mode', 'name', 'hosts', 'serverIndex', 'index'],
  data() {
    return {
      credentialMode: this.credentialname && !this.cert && !this.pkey && !this.cacert ? 'existing' : 'upload',
      tlsDialogOpen: false,
      tlsDialogError: '',
      tlsChecking: false,
      protocolConfirmOpen: false,
      pendingProtocol: '',
      tlsDraft: {
        credentialMode: this.credentialname && !this.cert && !this.pkey && !this.cacert ? 'existing' : 'upload',
        mode: this.mode || 'SIMPLE',
        credentialname: this.credentialname || '',
        cert: this.cert || '',
        pkey: this.pkey || '',
        cacert: this.cacert || '',
      },
    }
  },
  methods: {
    newTLSDraft() {
      return {
        credentialMode: this.credentialname && !this.cert && !this.pkey && !this.cacert ? 'existing' : 'upload',
        mode: this.mode || 'SIMPLE',
        credentialname: this.credentialname || '',
        cert: this.cert || '',
        pkey: this.pkey || '',
        cacert: this.cacert || '',
      };
    },
    removePort() {
      if (!this.canRemovePort) return;
      this.$store.commit('Gateway_RemovePort', {
        serverIndex: this.serverIndex,
        index: this.index
      });
    },
    updatePort(key, value) {
      this.$store.commit('Gateway_UpdatePort', {
        serverIndex: this.serverIndex,
        index: this.index,
        key,
        value
      });
    },
    protocolRequiresTLS(value) {
      return value === 'HTTPS' || value === 'TLS';
    },
    handleProtocolChange(value) {
      if (this.protocolRequiresTLS(this.protocol) && !this.protocolRequiresTLS(value) && this.hasTLSConfiguration) {
        this.pendingProtocol = value;
        this.protocolConfirmOpen = true;
        return;
      }
      this.updatePort('protocol', value);
    },
    cancelProtocolChange() {
      this.pendingProtocol = '';
      this.protocolConfirmOpen = false;
    },
    confirmProtocolChange() {
      if (!this.pendingProtocol) return;
      this.updatePort('protocol', this.pendingProtocol);
      this.pendingProtocol = '';
      this.protocolConfirmOpen = false;
      this.credentialMode = 'upload';
      this.tlsDraft = this.newTLSDraft();
    },
    updateDraft(key, value) {
      this.tlsDialogError = '';
      this.tlsDraft = {
        ...this.tlsDraft,
        [key]: value,
      };
    },
    updateDraftCertificate(value) {
      const parts = splitTLSPEM(value);
      this.tlsDialogError = '';
      this.tlsDraft = {
        ...this.tlsDraft,
        cert: parts.cert || value,
        pkey: parts.pkey || this.tlsDraft.pkey,
      };
    },
    updateDraftPrivateKey(value) {
      const parts = splitTLSPEM(value);
      this.tlsDialogError = '';
      this.tlsDraft = {
        ...this.tlsDraft,
        cert: parts.cert ? [this.tlsDraft.cert, parts.cert].filter(Boolean).join('\n') : this.tlsDraft.cert,
        pkey: parts.pkey || value,
      };
    },
    updateDraftExistingSecret(value) {
      this.tlsDialogError = '';
      this.tlsDraft = {
        ...this.tlsDraft,
        credentialname: value,
        cert: '',
        pkey: '',
        cacert: '',
      };
    },
    setDraftCredentialMode(mode) {
      this.tlsDialogError = '';
      const next = {
        ...this.tlsDraft,
        credentialMode: mode,
        mode: this.tlsDraft.mode || 'SIMPLE',
      };
      if (mode === 'existing') {
        next.cert = '';
        next.pkey = '';
        next.cacert = '';
      } else {
        next.credentialname = '';
      }
      this.tlsDraft = next;
    },
    openTLSDialog() {
      this.tlsDraft = this.newTLSDraft();
      this.tlsDialogError = '';
      this.tlsDialogOpen = true;
      this.$nextTick(() => {
        this.$refs.tlsDialog?.focus();
      });
    },
    closeTLSDialog() {
      if (this.tlsChecking) return;
      this.tlsDialogOpen = false;
      this.tlsDialogError = '';
      this.tlsDraft = this.newTLSDraft();
    },
    async confirmTLSDialog() {
      this.tlsDialogError = '';
      const draft = {
        ...this.tlsDraft,
        mode: this.tlsDraft.mode || 'SIMPLE',
      };

      if (draft.credentialMode === 'existing') {
        const credentialName = String(draft.credentialname || '').trim();
        if (!credentialName) {
          this.tlsDialogError = this.$t('Gateway.SecretNameRequired');
          return;
        }

        this.tlsChecking = true;
        try {
          const result = await this.$store.dispatch('Gateway_CheckTLSSecretExists', {
            credentialname: credentialName,
          });
          if (!result.exists) {
            this.tlsDialogError = this.$t('Gateway.SecretNotFound', {
              namespace: result.secretNamespace || 'istio-system',
              name: result.secretName || credentialName,
            });
            return;
          }
        } catch (err) {
          this.tlsDialogError = err.response?.data?.error || this.$t('Gateway.SecretCheckFailed');
          return;
        } finally {
          this.tlsChecking = false;
        }

        this.updatePort('credentialname', credentialName);
        this.updatePort('mode', draft.mode);
        this.updatePort('cert', '');
        this.updatePort('pkey', '');
        this.updatePort('cacert', '');
        this.credentialMode = 'existing';
      } else {
        if (!String(draft.cert || '').trim()) {
          this.tlsDialogError = this.$t('Gateway.CertificateRequired');
          return;
        }
        if (!String(draft.pkey || '').trim()) {
          this.tlsDialogError = this.$t('Gateway.PrivateKeyRequired');
          return;
        }
        if (draft.mode === 'MUTUAL' && !String(draft.cacert || '').trim()) {
          this.tlsDialogError = this.$t('Gateway.CARequired');
          return;
        }

        this.updatePort('credentialname', '');
        this.updatePort('mode', draft.mode);
        this.updatePort('cert', draft.cert);
        this.updatePort('pkey', draft.pkey);
        this.updatePort('cacert', draft.cacert);
        this.credentialMode = 'upload';
      }

      this.tlsDialogOpen = false;
      this.tlsDialogError = '';
      this.tlsDraft = this.newTLSDraft();
    },
    handleTLSDialogEnter(event) {
      if (this.tlsChecking || event.isComposing) return;
      const tagName = event.target?.tagName?.toLowerCase();
      if (tagName === 'textarea' || tagName === 'select' || tagName === 'button' || tagName === 'summary') return;
      event.preventDefault();
      this.confirmTLSDialog();
    },
    handleDrop(event) {
      this.readFiles(event.dataTransfer.files);
    },
    handleFileInput(event) {
      this.readFiles(event.target.files);
      event.target.value = '';
    },
    readFiles(files) {
      Array.from(files || []).forEach((file) => {
        const reader = new FileReader();
        reader.onload = () => this.applyPEMText(reader.result || '');
        reader.readAsText(file);
      });
    },
    applyPEMText(value) {
      const parts = splitTLSPEM(value);
      this.tlsDialogError = '';
      this.tlsDraft = {
        ...this.tlsDraft,
        cert: parts.cert ? [this.tlsDraft.cert, parts.cert].filter(Boolean).join('\n') : this.tlsDraft.cert,
        pkey: parts.pkey || this.tlsDraft.pkey,
      };
    },
    resetCert() {
      this.credentialMode = 'upload';
      this.$store.commit('Gateway_ResetCert', {
        serverIndex: this.serverIndex,
        index: this.index,
      });
    },
    recognizedCountLabel(count) {
      return count > 0 ? String(count) : this.$t('Gateway.NotRecognized');
    },
    formatCertificateDate(value) {
      if (!value) return '-';
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) return value;
      return `${date.toISOString().slice(0, 16).replace('T', ' ')} UTC`;
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
  },
  computed: {
    ...mapGetters({
      servers: 'Gateway_GetServers',
      protocols: 'Gateway_GetProtocols',
      tlsCertificates: 'Gateway_GetTLSCertificates',
    }),
    requiresTlsMaterial() {
      return this.protocolRequiresTLS(this.protocol);
    },
    canRemovePort() {
      return (this.servers?.[this.serverIndex]?.ports || []).length > 1;
    },
    tlsConfigureHintId() {
      return `tls-configure-hint-${this.serverIndex ?? 0}-${this.index ?? 0}`;
    },
    hasTLSConfiguration() {
      return Boolean(this.credentialname || this.cert || this.pkey || this.cacert || this.mode);
    },
    normalizedMode() {
      return this.mode || 'SIMPLE';
    },
    draftCredentialMode() {
      return this.tlsDraft?.credentialMode || 'upload';
    },
    draftMode() {
      return this.tlsDraft?.mode || 'SIMPLE';
    },
    currentTLSCertificate() {
      const certificates = this.tlsCertificates || [];
      const serverIndex = Number(this.serverIndex);
      const port = Number(this.port);
      const credentialName = String(this.credentialname || '').trim();
      const hosts = this.hosts || [];

      return certificates.find((certificate) => {
        const certificateServerIndex = Number(certificate.serverIndex);
        const certificatePort = Number(certificate.port);
        const credentialNames = [
          certificate.credentialName,
          certificate.secretName,
          certificate.secretNamespace && certificate.secretName ? `${certificate.secretNamespace}/${certificate.secretName}` : '',
        ].filter(Boolean);
        const certificateHosts = certificate.hosts || certificate.dnsNames || [];
        const matchesServer = !Number.isNaN(certificateServerIndex) && certificateServerIndex === serverIndex;
        const matchesPort = !Number.isNaN(certificatePort) && certificatePort === port;
        const matchesCredential = credentialName && credentialNames.includes(credentialName);
        const matchesHost = hosts.some((host) => certificateHosts.includes(host));
        return matchesServer && matchesPort && (matchesCredential || matchesHost || !credentialName);
      }) || null;
    },
    currentCertificateSecret() {
      if (!this.currentTLSCertificate) return '-';
      return [
        this.currentTLSCertificate.secretNamespace,
        this.currentTLSCertificate.secretName || this.currentTLSCertificate.credentialName,
      ].filter(Boolean).join(' / ') || '-';
    },
    credentialModeLabel() {
      return this.credentialMode === 'existing'
        ? this.$t('Gateway.ExistingSecret')
        : this.$t('Gateway.UploadPaste');
    },
    credentialSummaryText() {
      if (this.credentialMode === 'existing') {
        return this.credentialname || this.$t('Gateway.SecretName');
      }
      return `${this.$t('Gateway.Certificate')}: ${this.tlsSummary.certificateCount} · ${this.$t('Gateway.PrivateKey')}: ${this.tlsSummary.hasPrivateKey ? this.$t('Gateway.Present') : this.$t('Gateway.Missing')}`;
    },
    credentialDetailText() {
      if (this.credentialMode === 'existing') {
        return this.credentialname || '-';
      }
      return `${this.tlsSummary.certificateCount}`;
    },
    tlsSummary() {
      return summarizeTLSPEM({
        cert: this.cert,
        pkey: this.pkey,
        cacert: this.cacert,
      });
    },
    draftTLSSummary() {
      return summarizeTLSPEM({
        cert: this.tlsDraft?.cert || '',
        pkey: this.tlsDraft?.pkey || '',
        cacert: this.tlsDraft?.cacert || '',
      });
    },
  },
  watch: {
    credentialname(value) {
      if (value && !this.cert && !this.pkey && !this.cacert) this.credentialMode = 'existing';
    },
    protocol(value) {
      if (value === 'HTTPS' || value === 'TLS') {
        if (!this.mode) this.updatePort('mode', 'SIMPLE');
      }
    },
  },
}
</script>

<style scoped>
.port-card {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  display: grid;
  gap: 10px;
  padding: 12px;
  position: relative;
}

.port-grid {
  align-items: center;
  display: grid;
  gap: 10px;
  grid-template-columns: minmax(0, 1fr) minmax(96px, 120px);
  min-width: 0;
}

.port-card .field {
  gap: 5px;
  min-width: 0;
}

.port-card .field span {
  font-size: 0.82rem;
}

.port-card input,
.port-card select {
  min-height: 42px;
  width: 100%;
}

.tls-grid {
  align-items: start;
  grid-template-columns: 1fr 1fr;
}

.tls-settings-grid {
  align-items: start;
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

.tls-panel {
  align-items: center;
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  display: grid;
  gap: 10px;
  grid-template-columns: minmax(0, 1fr) auto;
  padding: 10px;
}

.tls-panel-header {
  align-items: center;
  display: flex;
  gap: 12px;
  justify-content: space-between;
}

.tls-panel-header p {
  color: var(--pw-muted);
  font-size: 0.9rem;
  margin: 4px 0 0;
  overflow-wrap: anywhere;
}

.tls-action-stack {
  align-items: flex-end;
  display: grid;
  gap: 6px;
  justify-items: end;
  max-width: 260px;
}

.tls-action-stack small {
  color: var(--pw-muted);
  font-size: 12px;
  line-height: 1.4;
  text-align: right;
}

.tls-summary-grid {
  display: grid;
  gap: 8px;
  grid-column: 1 / -1;
  grid-template-columns: minmax(140px, 1fr) minmax(110px, 150px) minmax(180px, 2fr);
  margin: 0;
}

.tls-summary-item {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  min-width: 0;
  padding: 8px 10px;
}

.tls-summary-grid dt {
  color: var(--pw-muted);
  font-size: 0.78rem;
  font-weight: 800;
}

.tls-summary-grid dd {
  font-weight: 800;
  margin: 4px 0 0;
  overflow-wrap: anywhere;
}

.dialog-backdrop {
  align-items: center;
  background: rgba(15, 23, 42, 0.42);
  bottom: 0;
  display: flex;
  justify-content: center;
  left: 0;
  overscroll-behavior: contain;
  padding: 24px;
  position: fixed;
  right: 0;
  top: 0;
  z-index: 1000;
}

.tls-dialog {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  box-shadow: 0 24px 80px rgba(15, 23, 42, 0.22);
  display: grid;
  gap: 0;
  grid-template-rows: auto minmax(0, 1fr) auto;
  max-height: calc(100dvh - 32px);
  max-width: 900px;
  min-height: 0;
  overflow: hidden;
  width: min(900px, 100%);
}

.tls-dialog:focus {
  outline: none;
}

.confirm-dialog {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  box-shadow: 0 24px 80px rgba(15, 23, 42, 0.22);
  display: grid;
  max-width: 640px;
  overflow: hidden;
  width: min(640px, 100%);
}

.dialog-header,
.dialog-footer {
  align-items: center;
  display: flex;
  justify-content: space-between;
  padding: 16px 18px;
}

.dialog-header {
  border-bottom: 1px solid var(--pw-border);
}

.dialog-header strong {
  display: block;
  font-size: 1.1rem;
}

.dialog-header p {
  color: var(--pw-muted);
  margin: 4px 0 0;
}

.dialog-body {
  display: grid;
  gap: 12px;
  min-height: 0;
  overscroll-behavior: contain;
  overflow-y: auto;
  padding: 14px 16px;
}

.tls-dialog-error {
  background: #fff2ef;
  border: 1px solid #e0a195;
  border-radius: 8px;
  color: #9f2f1f;
  font-weight: 800;
  padding: 10px 12px;
}

.current-certificate-card {
  background:
    linear-gradient(135deg, rgba(58, 91, 217, 0.08), transparent 42%),
    var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  display: grid;
  gap: 8px;
  padding: 12px;
}

.current-certificate-card header {
  align-items: flex-start;
  display: flex;
  gap: 12px;
  justify-content: space-between;
}

.current-certificate-card header span:first-child,
.current-certificate-card p,
.current-certificate-grid dt {
  color: var(--pw-muted);
}

.current-certificate-card header span:first-child,
.current-certificate-grid dt {
  font-size: 0.74rem;
  font-weight: 900;
  text-transform: uppercase;
}

.current-certificate-card header strong {
  display: block;
  font-size: 1.02rem;
  margin-top: 4px;
  overflow-wrap: anywhere;
}

.current-certificate-card p {
  font-size: 0.86rem;
  margin: 0;
}

.current-certificate-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
}

.current-certificate-grid.secondary {
  margin-top: 8px;
}

.current-certificate-grid div {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  min-width: 0;
  padding: 8px 10px;
}

.current-certificate-grid dd {
  font-weight: 800;
  margin: 4px 0 0;
  overflow-wrap: anywhere;
}

.current-certificate-details summary,
.certificate-fingerprint-details summary {
  color: var(--pw-primary-strong);
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 900;
}

.certificate-fingerprint {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.82rem;
  line-height: 1.5;
  margin: 8px 0 0;
  overflow-wrap: anywhere;
  padding: 10px;
}

.status-pill {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  color: var(--pw-primary-strong);
  flex: 0 0 auto;
  font-weight: 900;
  padding: 8px 12px;
  white-space: nowrap;
}

.dialog-footer {
  border-top: 1px solid var(--pw-border);
  gap: 10px;
  justify-content: flex-end;
}

.icon-button {
  align-items: center;
  background: transparent;
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  color: var(--pw-muted);
  display: inline-flex;
  font-weight: 900;
  height: 36px;
  justify-content: center;
  width: 36px;
}

.tls-choice-row {
  align-items: start;
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1fr);
  min-width: 0;
}

.credential-source-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  min-width: 0;
}

.credential-source-card {
  align-items: center;
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  color: var(--pw-muted);
  display: flex;
  gap: 10px;
  min-height: 58px;
  min-width: 0;
  padding: 10px;
  text-align: left;
}

.credential-source-card.active {
  border-color: var(--pw-primary-strong);
  box-shadow: 0 0 0 3px rgba(15, 23, 42, 0.08);
  color: var(--pw-primary-strong);
}

.credential-source-card strong,
.credential-source-card small {
  display: block;
}

.credential-source-card strong {
  font-size: 0.95rem;
}

.credential-source-card small {
  color: var(--pw-muted);
  margin-top: 3px;
}

.credential-source-card span:last-child {
  min-width: 0;
}

.source-icon {
  align-items: center;
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  display: inline-flex;
  flex: 0 0 36px;
  font-weight: 900;
  height: 36px;
  justify-content: center;
  width: 36px;
}

.credential-source-card.active .source-icon {
  background: var(--pw-primary-strong);
  border-color: var(--pw-primary-strong);
  color: #fff;
}

.pem-material-panel {
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  display: grid;
  gap: 12px;
  padding: 12px;
}

.drop-zone {
  align-items: center;
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  display: flex;
  gap: 12px;
  padding: 10px 12px;
}

.drop-zone input {
  display: none;
}

.upload-action {
  align-items: center;
  display: inline-flex;
  gap: 8px;
}

.upload-action span {
  font-weight: 900;
}

.field textarea {
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  min-height: 112px;
  padding: 10px 12px;
  resize: vertical;
}

.ca-field {
  grid-column: 1 / -1;
}

.credential-summary {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 0;
}

.credential-summary div {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  padding: 8px 10px;
}

.credential-summary dt {
  color: var(--pw-muted);
  font-size: 0.78rem;
}

.credential-summary dd {
  font-weight: 800;
  margin: 4px 0 0;
}

.remove-port-link {
  background: transparent;
  border: 0;
  color: var(--pw-error);
  font-weight: 800;
  justify-self: end;
  padding: 0;
  text-decoration: underline;
  text-underline-offset: 3px;
  white-space: nowrap;
}

@media (max-width: 900px) {
  .port-grid,
  .tls-grid,
  .tls-settings-grid,
  .tls-choice-row,
  .credential-source-grid,
  .tls-summary-grid,
  .credential-summary,
  .current-certificate-grid {
    grid-template-columns: 1fr;
  }

  .tls-panel {
    grid-template-columns: 1fr;
  }

  .tls-panel-header,
  .current-certificate-card header {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
  }

  .dialog-backdrop {
    align-items: stretch;
    padding: 12px;
  }

  .tls-dialog {
    max-height: calc(100vh - 24px);
  }
}
</style>
