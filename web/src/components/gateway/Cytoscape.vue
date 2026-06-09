<template>
  <section class="relationship-panel" :aria-label="$t('Gateway.RelationshipSummary')">
    <header class="relationship-header">
      <div>
        <h2>{{ $t('Gateway.TrafficRelationship') }}</h2>
        <p>{{ $t('Gateway.TrafficRelationshipHelp', { namespace: namespace || '-' }) }}</p>
      </div>
      <div class="relationship-totals">
        <span>{{ $t('Gateway.AssociationRouting') }} <strong>{{ routerItems.length }}</strong></span>
        <span>{{ $t('Gateway.TLSEndpoints') }} <strong>{{ hostGroupCount }}</strong></span>
      </div>
    </header>

    <div class="view-toggle" role="tablist" :aria-label="$t('Gateway.RelationshipView')">
      <button
        type="button"
        :class="{ active: viewMode === 'topology' }"
        @click="viewMode = 'topology'"
      >
        {{ $t('Gateway.TopologyView') }}
      </button>
      <button
        type="button"
        :class="{ active: viewMode === 'list' }"
        @click="viewMode = 'list'"
      >
        {{ $t('Gateway.ListView') }}
      </button>
    </div>

    <div v-if="viewMode === 'topology'" class="topology-layout">
      <section class="topology-source">
        <article class="topology-node gateway-topology-node">
          <span>{{ $t('Gateway.Gateway') }}</span>
          <strong>{{ name || '-' }}</strong>
          <small>{{ namespace || '-' }}</small>
        </article>

        <div class="topology-listener-connector" aria-hidden="true"></div>

        <section class="topology-column endpoint-branch">
          <header>
            <span>{{ $t('Gateway.TLSEndpoints') }}</span>
            <strong>{{ hostGroupCount }}</strong>
          </header>
          <div v-if="hostGroups.length" class="topology-stack">
            <article
              v-for="group in hostGroups"
              :key="group.key"
              class="topology-node endpoint-node"
              :class="tlsStatusClass(group.tlsStatus)"
            >
              <span>{{ group.protocol }} / {{ group.port || '-' }}</span>
              <strong>{{ group.hosts.join(', ') || '-' }}</strong>
              <small v-if="group.isTls">{{ tlsStatusLabel(group.tlsStatus) }}</small>
            </article>
          </div>
          <div v-else class="relationship-empty compact">{{ $t('Gateway.NoHostsConfigured') }}</div>
        </section>
      </section>

      <div class="topology-route-connector" aria-hidden="true">
        <span>{{ $t('Gateway.RoutesToVirtualService') }}</span>
      </div>

      <section class="topology-column topology-routes">
        <header>
          <span>{{ $t('Gateway.AssociationRouting') }}</span>
          <strong>{{ routerItems.length }}</strong>
        </header>
        <div v-if="routerItems.length" class="topology-stack">
          <article v-for="router in routerItems" :key="routerKey(router)" class="topology-node vs-node">
            <span>{{ $t('Gateway.VirtualService') }}</span>
            <strong>{{ router.name || '-' }}</strong>
            <small>{{ router.namespace || namespace || '-' }}</small>
          </article>
        </div>
        <div v-else class="relationship-empty compact">{{ $t('Gateway.NoRoutersAssociated') }}</div>
      </section>
    </div>

    <div v-else class="relationship-layout">
      <article class="gateway-summary-card">
        <span class="node-kicker">{{ $t('Gateway.Gateway') }}</span>
        <h3>{{ name || '-' }}</h3>
        <dl class="gateway-summary-fields">
          <div>
            <dt>{{ $t('Table.Namespace') }}</dt>
            <dd>{{ namespace || '-' }}</dd>
          </div>
          <div>
            <dt>{{ $t('Gateway.Servers') }}</dt>
            <dd>{{ serverCount }}</dd>
          </div>
          <div>
            <dt>{{ $t('Gateway.VirtualService') }}</dt>
            <dd>{{ routerItems.length }}</dd>
          </div>
        </dl>
      </article>

      <div class="relationship-list-grid">
        <section class="relationship-section">
          <header class="section-title-row">
            <h3>{{ $t('Gateway.AssociationRouting') }}</h3>
            <strong>{{ routerItems.length }}</strong>
          </header>

          <div v-if="routerItems.length" class="compact-list">
            <article
              v-for="router in routerItems"
              :key="routerKey(router)"
              class="relation-row router-row"
            >
              <div>
                <span class="card-kicker">{{ $t('Gateway.VirtualService') }}</span>
                <strong class="primary-value">{{ router.name || '-' }}</strong>
              </div>
              <dl class="relation-fields">
                <div>
                  <dt>{{ $t('Table.Namespace') }}</dt>
                  <dd>{{ router.namespace || namespace || '-' }}</dd>
                </div>
                <div>
                  <dt>{{ $t('Gateway.Protocol') }}</dt>
                  <dd>{{ normalizeProtocol(router.protocol) }}</dd>
                </div>
                <div v-if="router.httpCount !== undefined">
                  <dt>{{ $t('Table.RuleCount') }}</dt>
                  <dd>{{ router.httpCount }}</dd>
                </div>
              </dl>
            </article>
          </div>

          <div v-else class="relationship-empty compact">
            {{ $t('Gateway.NoRoutersAssociated') }}
          </div>
        </section>

        <section class="relationship-section">
          <header class="section-title-row">
            <h3>{{ $t('Gateway.TLSEndpoints') }}</h3>
            <strong>{{ hostGroupCount }}</strong>
          </header>

          <div v-if="hostGroups.length" class="compact-list endpoint-list">
            <article
              v-for="group in hostGroups"
              :key="group.key"
              class="relation-row host-row"
              :class="tlsStatusClass(group.tlsStatus)"
            >
              <div class="endpoint-main">
                <span class="protocol-pill">{{ group.protocol }}</span>
                <div>
                  <span class="card-kicker">{{ $t('Gateway.ServerHost') }}</span>
                  <strong class="primary-value">{{ group.hosts.join(', ') || '-' }}</strong>
                </div>
              </div>

              <dl class="relation-fields host-fields">
                <div>
                  <dt>{{ $t('Gateway.Port') }}</dt>
                  <dd>{{ group.port || '-' }}</dd>
                </div>
                <div v-if="group.isTls">
                  <dt>{{ $t('Gateway.TLSMode') }}</dt>
                  <dd>{{ group.mode || 'SIMPLE' }}</dd>
                </div>
                <div v-if="group.isTls">
                  <dt>{{ $t('Gateway.SecretName') }}</dt>
                  <dd>{{ group.credentialname || '-' }}</dd>
                </div>
                <div v-if="group.isTls">
                  <dt>{{ $t('Gateway.TLSCertificateStatus') }}</dt>
                  <dd>{{ tlsStatusLabel(group.tlsStatus) }}</dd>
                </div>
              </dl>
            </article>
          </div>

          <div v-else class="relationship-empty compact">
            {{ $t('Gateway.NoHostsConfigured') }}
          </div>
        </section>
      </div>
    </div>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';

const TLS_PROTOCOLS = ['HTTPS', 'TLS'];

export default {
  name: 'GatewayCytoscape',
  mounted: async function() {
    await this.loadGatewayRelationship();
  },
  data() {
    return {
      name: '',
      namespace: '',
      loadingTLS: false,
      viewMode: 'topology',
    }
  },
  watch: {
    '$route': async function() {
      await this.loadGatewayRelationship();
    }
  },
  computed: {
    ...mapGetters({
      servers: 'Gateway_GetServers',
      routers: 'Router_GetItems',
      mappings: 'Gateway_GetMappings',
      tlsCertificates: 'Gateway_GetTLSCertificates',
    }),
    routerItems() {
      return this.mappings || [];
    },
    normalizedServers() {
      return (this.servers || []).filter((server) => {
        return (server.hosts || []).length || (server.ports || []).length;
      });
    },
    serverCount() {
      return this.normalizedServers.length;
    },
    hostGroups() {
      const groups = [];

      this.normalizedServers.forEach((server, serverIndex) => {
        const hosts = server.hosts || [];
        const ports = server.ports || [];

        ports.forEach((port, portIndex) => {
          const protocol = this.normalizeProtocol(port.protocol);
          const isTls = TLS_PROTOCOLS.includes(protocol);
          const tlsStatus = isTls ? this.findTLSStatus(serverIndex, port, hosts) : null;
          groups.push({
            key: [serverIndex, portIndex, port.port, protocol, hosts.join(',')].join(':'),
            hosts,
            protocol,
            port: port.port,
            mode: port.mode,
            credentialname: port.credentialname,
            isTls,
            tlsStatus,
          });
        });
      });

      return groups;
    },
    hostGroupCount() {
      return this.hostGroups.length;
    },
  },
  methods: {
    async loadGatewayRelationship() {
      this.$store.commit('Gateway_ResetStatus');
      this.name = this.$route.query.name || this.$route.params.name || '';
      this.namespace = this.$route.query.namespace;

      if (!this.name || !this.namespace) return;

      await this.fetchData();
      await this.fetchMapping();
      await this.fetchTLSCertificates();
    },
    fetchData: async function() {
      await this.$store.dispatch('Gateway_GetItem', {
        id: this.id,
        name: this.name,
        namespace: this.namespace
      });
    },
    fetchMapping: async function() {
      await this.$store.dispatch('Router_GetItems', {
        name: this.name,
        namespace: this.namespace,
        page: 1,
        limit: -1
      });

      await this.$store.dispatch('Gateway_GetMappings', {
        name: this.name,
        namespace: this.namespace,
        routers: this.routers
      });
    },
    async fetchTLSCertificates() {
      this.loadingTLS = true;
      await this.$store.dispatch('Gateway_GetTLSCertificates', {
        name: this.name,
        namespace: this.namespace,
      });
      this.loadingTLS = false;
    },
    routerKey(router) {
      return [router.namespace, router.name, router.protocol].filter(Boolean).join(':');
    },
    normalizeProtocol(protocol) {
      return String(protocol || '-').toUpperCase();
    },
    findTLSStatus(serverIndex, port, hosts) {
      const protocol = this.normalizeProtocol(port.protocol);
      return (this.tlsCertificates || []).find((item) => {
        const sameServer = Number(item.serverIndex) === Number(serverIndex);
        const samePort = Number(item.port) === Number(port.port);
        const sameProtocol = this.normalizeProtocol(item.protocol) === protocol;
        const sameCredential = item.credentialName && item.credentialName === port.credentialname;
        const sameHost = (item.hosts || []).some((host) => hosts.includes(host));
        return sameServer || (samePort && sameProtocol && (sameCredential || sameHost));
      });
    },
    tlsStatusLabel(status) {
      if (this.loadingTLS) return this.$t('Gateway.TLSLoading');
      if (!status) return this.$t('Gateway.TLSUnknown');
      if (status.status === 'healthy') return this.$t('Gateway.TLSHealthy', { days: status.daysUntilExpiry });
      if (status.status === 'warning') return this.$t('Gateway.TLSWarning', { days: status.daysUntilExpiry });
      if (status.status === 'critical') return this.$t('Gateway.TLSCritical', { days: status.daysUntilExpiry });
      if (status.status === 'expired') return this.$t('Gateway.TLSExpired');
      if (status.status === 'missing') return this.$t('Gateway.TLSMissing');
      if (status.status === 'invalid') return this.$t('Gateway.TLSInvalid');
      return this.$t('Gateway.TLSUnknown');
    },
    tlsStatusClass(status) {
      return {
        'tls-ok': status?.status === 'healthy',
        'tls-warning': ['warning', 'critical', 'expired'].includes(status?.status),
        'tls-error': ['missing', 'invalid'].includes(status?.status),
      };
    },
  },
}
</script>

<style scoped>
.relationship-panel {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  display: grid;
  gap: 18px;
  padding: 18px;
}

.relationship-header {
  align-items: center;
  border-bottom: 1px solid var(--pw-border);
  display: flex;
  gap: 16px;
  justify-content: space-between;
  padding-bottom: 16px;
}

.relationship-header h2 {
  font-size: 1.12rem;
  margin: 0;
}

.relationship-header p {
  color: var(--pw-muted);
  font-weight: 700;
  margin: 5px 0 0;
}

.relationship-totals {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.relationship-totals span {
  align-items: center;
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  color: var(--pw-muted);
  display: inline-flex;
  gap: 12px;
  min-height: 34px;
  padding: 5px 11px;
}

.relationship-totals strong {
  color: var(--pw-primary-strong);
  font-size: 1rem;
}

.view-toggle {
  align-items: center;
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  display: inline-flex;
  gap: 4px;
  justify-self: start;
  padding: 4px;
}

.view-toggle button {
  background: transparent;
  border: 0;
  border-radius: 999px;
  color: var(--pw-muted);
  cursor: pointer;
  font-weight: 800;
  min-height: 32px;
  padding: 6px 12px;
}

.view-toggle button.active {
  background: var(--pw-primary);
  color: #fff;
}

.topology-layout {
  align-items: start;
  display: grid;
  gap: 18px;
  grid-template-columns: minmax(280px, 360px) minmax(120px, 160px) minmax(280px, 1fr);
}

.topology-source {
  display: grid;
  gap: 0;
  min-width: 0;
}

.topology-column {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.topology-column header {
  align-items: center;
  border-bottom: 1px solid var(--pw-border);
  display: flex;
  gap: 12px;
  justify-content: space-between;
  padding-bottom: 8px;
}

.topology-column header span {
  color: var(--pw-muted);
  font-weight: 800;
}

.topology-column header strong {
  color: var(--pw-primary-strong);
  font-size: 1.18rem;
}

.topology-stack {
  display: grid;
  gap: 8px;
  max-height: 260px;
  overflow: auto;
  padding-right: 2px;
}

.topology-node {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-left: 4px solid #3f6f9f;
  border-radius: 8px;
  display: grid;
  gap: 5px;
  min-width: 0;
  padding: 12px;
}

.topology-node span,
.topology-node small {
  color: var(--pw-muted);
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.topology-node strong {
  font-size: 1rem;
  overflow-wrap: anywhere;
}

.gateway-topology-node {
  border-left-color: var(--pw-primary);
  box-shadow: 0 16px 32px rgba(25, 23, 20, 0.08);
  padding: 18px;
  position: relative;
}

.gateway-topology-node strong {
  font-size: 1.18rem;
}

.endpoint-node {
  border-left-color: #b57a20;
}

.endpoint-node.tls-ok {
  border-left-color: #2f8f5b;
}

.endpoint-node.tls-warning {
  border-left-color: #d98d22;
}

.endpoint-node.tls-error {
  border-left-color: #c7462d;
}

.topology-listener-connector {
  background: var(--pw-border);
  height: 30px;
  justify-self: center;
  width: 2px;
}

.topology-route-connector {
  align-items: center;
  display: grid;
  gap: 8px;
  justify-items: center;
  min-height: 126px;
  position: relative;
}

.topology-route-connector::before {
  background: var(--pw-border);
  content: "";
  height: 2px;
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 100%;
}

.topology-route-connector::after {
  border-bottom: 6px solid transparent;
  border-left: 10px solid var(--pw-border);
  border-top: 6px solid transparent;
  content: "";
  position: absolute;
  right: -1px;
  top: calc(50% - 6px);
}

.topology-route-connector span {
  background: var(--pw-surface);
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  color: var(--pw-muted);
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.04em;
  padding: 5px 10px;
  position: relative;
  z-index: 1;
}

.topology-routes,
.endpoint-branch {
  background: rgba(248, 246, 242, 0.48);
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  padding: 12px;
}

.endpoint-branch {
  border-left: 3px solid #b57a20;
}

.topology-routes {
  margin-top: 22px;
}

.relationship-layout {
  display: grid;
  gap: 16px;
  grid-template-columns: minmax(240px, 300px) minmax(0, 1fr);
}

.gateway-summary-card,
.relationship-section,
.relationship-empty {
  background: var(--pw-surface);
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  min-width: 0;
}

.gateway-summary-card {
  align-self: start;
  display: grid;
  gap: 16px;
  padding: 18px;
  position: sticky;
  top: 12px;
}

.node-kicker,
.card-kicker {
  color: var(--pw-muted);
  display: block;
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.gateway-summary-card h3 {
  font-size: 1.35rem;
  line-height: 1.2;
  margin: 0;
  overflow-wrap: anywhere;
}

.gateway-summary-fields,
.relation-fields {
  display: grid;
  gap: 10px;
  margin: 0;
}

.gateway-summary-fields {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.gateway-summary-fields div:first-child {
  grid-column: 1 / -1;
}

.relationship-list-grid {
  display: grid;
  gap: 14px;
  min-width: 0;
}

.relationship-section {
  display: grid;
  gap: 12px;
  padding: 14px;
}

.section-title-row {
  align-items: center;
  display: flex;
  gap: 12px;
  justify-content: space-between;
}

.section-title-row h3 {
  font-size: 1rem;
  margin: 0;
}

.section-title-row strong {
  color: var(--pw-primary-strong);
  font-size: 1.25rem;
}

.compact-list {
  display: grid;
  gap: 10px;
  max-height: 360px;
  overflow: auto;
  padding-right: 2px;
}

.relation-row {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-left: 4px solid #3f6f9f;
  border-radius: 8px;
  display: grid;
  gap: 12px;
  padding: 12px;
}

.host-row {
  border-left-color: #b57a20;
}

.host-row.tls-ok {
  border-left-color: #2f8f5b;
}

.host-row.tls-warning {
  border-left-color: #d98d22;
}

.host-row.tls-error {
  border-left-color: #c7462d;
}

.endpoint-main {
  align-items: center;
  display: flex;
  gap: 10px;
  min-width: 0;
}

.primary-value {
  color: var(--pw-text);
  display: block;
  font-size: 1.02rem;
  line-height: 1.25;
  overflow-wrap: anywhere;
}

.protocol-pill {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  color: var(--pw-primary-strong);
  flex: 0 0 auto;
  font-size: 0.76rem;
  font-weight: 800;
  line-height: 1;
  padding: 7px 9px;
  white-space: nowrap;
}

.relation-fields {
  grid-template-columns: repeat(auto-fit, minmax(110px, 1fr));
}

.gateway-summary-fields dt,
.relation-fields dt {
  color: var(--pw-muted);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.gateway-summary-fields dd,
.relation-fields dd {
  font-weight: 800;
  margin: 3px 0 0;
  overflow-wrap: anywhere;
}

.relationship-empty {
  color: var(--pw-muted);
  font-weight: 700;
  padding: 18px 14px;
  text-align: center;
}

@media (max-width: 1100px) {
  .relationship-header,
  .relationship-layout,
  .topology-layout {
    align-items: stretch;
  }

  .relationship-header {
    flex-direction: column;
  }

  .relationship-totals {
    justify-content: flex-start;
    width: 100%;
  }

  .relationship-layout {
    grid-template-columns: 1fr;
  }

  .topology-layout {
    grid-template-columns: 1fr;
  }

  .topology-route-connector {
    min-height: 44px;
  }

  .topology-route-connector::before {
    height: 44px;
    top: 0;
    transform: none;
    width: 2px;
  }

  .topology-route-connector::after {
    border-left: 6px solid transparent;
    border-right: 6px solid transparent;
    border-top: 10px solid var(--pw-border);
    bottom: -1px;
    right: auto;
  }

  .topology-routes {
    margin-top: 0;
  }

  .gateway-summary-card {
    position: static;
  }
}

@media (max-width: 640px) {
  .relationship-panel {
    padding: 14px;
  }

  .gateway-summary-fields,
  .relation-fields {
    grid-template-columns: 1fr;
  }

  .endpoint-main {
    align-items: flex-start;
  }

  .protocol-pill {
    margin-top: 2px;
  }
}
</style>
