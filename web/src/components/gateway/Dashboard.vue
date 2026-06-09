<template>
  <section class="dashboard-grid">
    <article class="metric-card dashboard-card">
      <header class="dashboard-card-header">
        <div>
          <h3>{{ $t('Gateway.TLSReadyPorts') }}</h3>
          <p>{{ securePorts }} / {{ totalPorts }}</p>
        </div>
        <strong>{{ tlsReadyPercent }}%</strong>
      </header>

      <dl class="metric-list">
        <div>
          <dt>{{ $t('Gateway.DeclaredPorts') }}</dt>
          <dd>{{ totalPorts }}</dd>
        </div>
        <div>
          <dt>{{ $t('Gateway.HTTPSTLSPorts') }}</dt>
          <dd>{{ securePorts }}</dd>
        </div>
      </dl>
    </article>

    <article class="metric-card dashboard-card">
      <header class="dashboard-card-header">
        <div>
          <h3>{{ $t('Gateway.ServersWithHosts') }}</h3>
          <p>{{ populatedServers }} / {{ totalServers }}</p>
        </div>
        <strong>{{ serverCoverage }}%</strong>
      </header>

      <dl class="metric-list">
        <div>
          <dt>{{ $t('Table.Servers') }}</dt>
          <dd>{{ totalServers }}</dd>
        </div>
        <div>
          <dt>{{ $t('Gateway.ServerHost') }}</dt>
          <dd>{{ totalHosts }}</dd>
        </div>
      </dl>
    </article>

    <section class="metric-card dashboard-card protocol-card">
      <header class="dashboard-card-header">
        <div>
          <h3>{{ $t('Gateway.ProtocolsInUse') }}</h3>
          <p>{{ protocolTotal }} {{ $t('Gateway.DeclaredPorts') }}</p>
        </div>
      </header>

      <div class="protocol-chip-list">
        <span
          v-for="item in activeProtocolBreakdown"
          :key="item.protocol"
          class="protocol-chip"
        >
          <strong>{{ item.protocol }}</strong>
          <em>{{ item.count }}</em>
        </span>
        <span v-if="!activeProtocolBreakdown.length" class="protocol-empty">-</span>
      </div>

      <dl class="metric-list secondary">
        <div>
          <dt>{{ $t('Gateway.HostsPerServer') }}</dt>
          <dd>{{ averageHostsPerServer }}</dd>
        </div>
        <div>
          <dt>{{ $t('Gateway.PortsPerServer') }}</dt>
          <dd>{{ averagePortsPerServer }}</dd>
        </div>
      </dl>
    </section>

    <section class="metric-card dashboard-distribution-card">
      <header class="dashboard-card-header compact">
        <div>
          <h3>{{ $t('Gateway.CurrentDistribution') }}</h3>
          <p>{{ $t('Gateway.CurrentDistributionHelp') }}</p>
        </div>
      </header>

      <div class="distribution-grid">
        <article class="distribution-chart">
          <div class="distribution-chart-title">
            <strong>{{ $t('Gateway.HostsPerServer') }}</strong>
            <span>{{ distributionLabel(hostSpreadSeries) }}</span>
          </div>
          <svg class="sparkline" viewBox="0 0 100 34" role="img" :aria-label="$t('Gateway.HostsPerServer')">
            <polyline :points="sparklinePoints(hostSpreadSeries)" />
          </svg>
        </article>

        <article class="distribution-chart warm">
          <div class="distribution-chart-title">
            <strong>{{ $t('Gateway.PortsPerServer') }}</strong>
            <span>{{ distributionLabel(portSpreadSeries) }}</span>
          </div>
          <svg class="sparkline" viewBox="0 0 100 34" role="img" :aria-label="$t('Gateway.PortsPerServer')">
            <polyline :points="sparklinePoints(portSpreadSeries)" />
          </svg>
        </article>

        <article class="distribution-chart protocol-bars">
          <div class="distribution-chart-title">
            <strong>{{ $t('Gateway.ProtocolsInUse') }}</strong>
            <span>{{ protocolTotal }} {{ $t('Gateway.DeclaredPorts') }}</span>
          </div>
          <div class="protocol-bar-list">
            <div v-for="item in protocolBreakdown" :key="item.protocol" class="protocol-bar-row">
              <span>{{ item.protocol }}</span>
              <div class="protocol-bar-track">
                <i :style="{ width: protocolBarWidth(item.count) }"></i>
              </div>
              <strong>{{ item.count }}</strong>
            </div>
          </div>
        </article>
      </div>
    </section>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'GatewayDashboard',
  mounted: async function() {
    this.name = this.$route.query.name || this.$route.params.name || '';
    this.namespace = this.$route.query.namespace || '';

    if (this.name && this.namespace) {
      await this.fetchData();
    }
  },
  data() {
    return {
      name: '',
      namespace: ''
    }
  },
  watch: {
    '$route': async function() {
      this.name = this.$route.query.name || this.$route.params.name || '';
      this.namespace = this.$route.query.namespace || '';

      if (this.name && this.namespace) {
        await this.fetchData();
      }
    }
  },
  computed: {
    ...mapGetters({
      gateway: 'Gateway_GetItem',
      servers: 'Gateway_GetServers',
    }),
    totalServers() {
      return this.servers.length;
    },
    totalHosts() {
      return this.servers.reduce((sum, server) => sum + (server.hosts?.length || 0), 0);
    },
    totalPorts() {
      return this.servers.reduce((sum, server) => sum + (server.ports?.length || 0), 0);
    },
    securePorts() {
      return this.servers.reduce((sum, server) => {
        return sum + (server.ports || []).filter((port) => {
          return ['HTTPS', 'TLS'].includes(String(port.protocol || '').toUpperCase());
        }).length;
      }, 0);
    },
    tlsReadyPercent() {
      if (!this.totalPorts) return 0;
      return Math.round((this.securePorts / this.totalPorts) * 100);
    },
    serverCoverage() {
      if (!this.totalServers) return 0;
      return Math.round((this.populatedServers / this.totalServers) * 100);
    },
    populatedServers() {
      return this.servers.filter((server) => (server.hosts || []).length > 0).length;
    },
    hostSpreadSeries() {
      return this.withFallbackSeries(this.servers.map((server) => (server.hosts || []).length));
    },
    portSpreadSeries() {
      return this.withFallbackSeries(this.servers.map((server) => (server.ports || []).length));
    },
    protocolBreakdown() {
      const order = ['HTTP', 'HTTP2', 'HTTPS', 'GRPC', 'TCP', 'UDP', 'TLS'];
      return order.map((protocol) => {
        const count = this.servers.reduce((sum, server) => {
          return sum + (server.ports || []).filter((port) => String(port.protocol || '').toUpperCase() === protocol).length;
        }, 0);
        return { protocol, count };
      });
    },
    activeProtocolBreakdown() {
      return this.protocolBreakdown.filter((item) => item.count > 0);
    },
    protocolTotal() {
      return this.protocolBreakdown.reduce((sum, item) => sum + item.count, 0);
    },
    averageHostsPerServer() {
      if (!this.totalServers) return '0';
      return (this.totalHosts / this.totalServers).toFixed(1);
    },
    averagePortsPerServer() {
      if (!this.totalServers) return '0';
      return (this.totalPorts / this.totalServers).toFixed(1);
    }
  },
  methods: {
    fetchData: async function() {
      await this.$store.dispatch('Gateway_GetItem', {
        name: this.name,
        namespace: this.namespace
      });
    },
    withFallbackSeries(series) {
      return series.length ? series : [0];
    },
    normalizedSeries(series) {
      if (!Array.isArray(series) || !series.length) return [0];
      return series.map((value) => Number(value) || 0);
    },
    sparklinePoints(series) {
      const values = this.normalizedSeries(series);
      if (values.length === 1) return '0,17 100,17';

      const min = Math.min(...values);
      const max = Math.max(...values);
      const range = max - min || 1;

      return values.map((value, index) => {
        const x = (index / (values.length - 1)) * 100;
        const y = 30 - ((value - min) / range) * 26;
        return `${x.toFixed(2)},${y.toFixed(2)}`;
      }).join(' ');
    },
    distributionLabel(series) {
      const values = this.normalizedSeries(series);
      return this.$t('Gateway.DistributionValues', { values: values.join(', ') });
    },
    protocolBarWidth(count) {
      const max = Math.max(...this.protocolBreakdown.map((item) => item.count), 1);
      return `${Math.max((count / max) * 100, count > 0 ? 8 : 0)}%`;
    },
  }
}
</script>

<style scoped>
.dashboard-grid {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.dashboard-card {
  align-content: start;
  border-radius: 10px;
  display: grid;
  gap: 16px;
  padding: 18px;
}

.dashboard-card-header {
  align-items: start;
  border-bottom: 1px solid var(--pw-border);
  display: flex;
  gap: 14px;
  justify-content: space-between;
  padding-bottom: 14px;
}

.dashboard-card-header h3 {
  font-size: 1rem;
  line-height: 1.3;
  margin: 0;
}

.dashboard-card-header p {
  color: var(--pw-muted);
  font-size: 0.86rem;
  font-weight: 700;
  margin: 5px 0 0;
}

.dashboard-card-header > strong {
  color: var(--pw-primary-strong);
  font-size: 1.6rem;
  line-height: 1;
}

.dashboard-card-header.compact {
  padding-bottom: 12px;
}

.metric-list {
  display: grid;
  gap: 10px;
  margin: 0;
}

.metric-list div {
  align-items: center;
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  display: flex;
  gap: 12px;
  justify-content: space-between;
  min-height: 50px;
  padding: 10px 12px;
}

.metric-list dt {
  color: var(--pw-muted);
  font-size: 0.84rem;
  font-weight: 800;
}

.metric-list dd {
  font-size: 1.28rem;
  font-weight: 900;
  margin: 0;
}

.metric-list.secondary {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.metric-list.secondary div {
  align-items: flex-start;
  display: grid;
}

.protocol-card {
  gap: 14px;
}

.protocol-chip-list {
  align-content: start;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  min-height: 42px;
}

.protocol-chip {
  align-items: center;
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  display: inline-flex;
  gap: 8px;
  min-height: 34px;
  padding: 6px 10px;
}

.protocol-chip strong {
  font-size: 0.82rem;
}

.protocol-chip em {
  background: var(--pw-primary);
  border-radius: 999px;
  color: #fff;
  font-style: normal;
  font-weight: 900;
  min-width: 24px;
  padding: 2px 7px;
  text-align: center;
}

.protocol-empty {
  color: var(--pw-muted);
  font-weight: 800;
}

.dashboard-distribution-card {
  grid-column: 1 / -1;
  gap: 14px;
}

.distribution-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.distribution-chart {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  display: grid;
  gap: 10px;
  padding: 12px;
}

.distribution-chart-title {
  align-items: start;
  display: grid;
  gap: 4px;
}

.distribution-chart-title strong {
  font-size: 0.92rem;
}

.distribution-chart-title span {
  color: var(--pw-muted);
  font-size: 0.78rem;
  font-weight: 700;
}

.sparkline {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  height: 58px;
  width: 100%;
}

.sparkline polyline {
  fill: none;
  stroke: var(--pw-primary);
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2.6;
}

.distribution-chart.warm .sparkline polyline {
  stroke: var(--pw-accent);
}

.protocol-bar-list {
  display: grid;
  gap: 7px;
}

.protocol-bar-row {
  align-items: center;
  display: grid;
  gap: 8px;
  grid-template-columns: 52px minmax(0, 1fr) 28px;
}

.protocol-bar-row span,
.protocol-bar-row strong {
  font-size: 0.78rem;
  font-weight: 800;
}

.protocol-bar-track {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  height: 10px;
  overflow: hidden;
}

.protocol-bar-track i {
  background: var(--pw-primary);
  display: block;
  height: 100%;
}

@media (max-width: 1100px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }

  .distribution-grid {
    grid-template-columns: 1fr;
  }
}
</style>
