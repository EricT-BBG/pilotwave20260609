<template>
  <section class="dashboard-grid">
    <article class="metric-card dashboard-card">
      <div
        class="ring-progress"
        :style="{ '--progress': successAvg }"
        :aria-label="$t('Router.ConnectionToday')"
      >
        <strong>{{ successAvg }}%</strong>
      </div>

      <p class="dashboard-title">{{ $t('Router.ConnectionToday') }}</p>
      <div class="metric-tile primary">
        <strong>{{ totalRequest }}</strong>
        <span>{{ $t('Router.RequestToday') }}</span>
      </div>
      <div class="metric-tile">
        <strong>{{ failRequest }}</strong>
        <span>{{ $t('Router.FailedToday') }}</span>
      </div>

      <a
        v-if="grafanaConfig.host"
        class="primary-button dashboard-link"
        :href="grafanaHref"
        target="_blank"
        rel="noreferrer"
      >
        More Details
      </a>
    </article>

    <article class="metric-card dashboard-card">
      <div
        class="ring-progress"
        :style="{ '--progress': successHourAvg }"
        :aria-label="$t('Router.ConnectionHour')"
      >
        <strong>{{ successHourAvg }}%</strong>
      </div>

      <p class="dashboard-title">{{ $t('Router.ConnectionHour') }}</p>
      <div class="metric-tile primary">
        <strong>{{ totalHourRequest }}</strong>
        <span>{{ $t('Router.RequestHour') }}</span>
      </div>
      <div class="metric-tile">
        <strong>{{ failHourRequest }}</strong>
        <span>{{ $t('Router.FailedHour') }}</span>
      </div>
    </article>

    <section class="panel dashboard-chart-panel">
      <div v-if="metricNotice" class="metrics-notice">
        <strong>{{ metricNotice.title }}</strong>
        <span>{{ metricNotice.message }}</span>
      </div>

      <label class="field date-field">
        <span>{{ $t('Form.Date') }}</span>
        <input
          v-model="today"
          type="date"
          min="1950-01-01"
          :max="maxDate"
          @change="save(today)"
        >
      </label>

      <article class="spark-card">
        <p>{{ $t('Router.SuccessRate') }}</p>
        <svg class="sparkline" viewBox="0 0 100 36" role="img" :aria-label="$t('Router.SuccessRateTrend')">
          <polyline :points="sparklinePoints(successMetric)" />
        </svg>
        <div class="spark-values">{{ metricLabel(successMetric) }}</div>
      </article>

      <article class="spark-card warm">
        <p>{{ $t('Router.Latency') }}</p>
        <svg class="sparkline" viewBox="0 0 100 36" role="img" :aria-label="$t('Router.LatencyTrend')">
          <polyline :points="sparklinePoints(latencyMetric)" />
        </svg>
        <div class="spark-values">{{ metricLabel(latencyMetric) }}</div>
      </article>

      <article class="spark-card muted">
        <p>OPS</p>
        <svg class="sparkline" viewBox="0 0 100 36" role="img" :aria-label="$t('Router.OPSTrend')">
          <polyline :points="sparklinePoints(opsMetric)" />
        </svg>
        <div class="spark-values">{{ metricLabel(opsMetric) }}</div>
      </article>
    </section>
  </section>
</template>

<script>
import moment from 'moment';
import { mapGetters } from 'vuex';

export default {
  name: 'RouterDashboard',
  mounted: async function() {
    this.$store.commit('Router_ResetStatus');
    this.name = this.$route.query.name || this.$route.params.name || '';
    this.namespace = this.$route.query.namespace;

    if (this.name && this.namespace) {
      await this.fetchData();
    }
  },
  data() {
    return {
      name: '',
      namespace: '',
      today: moment().format('YYYY-MM-DD'),
      metricError: '',
      metricsLoaded: false
    }
  },
  methods: {
    fetchData: async function() {
      this.metricError = '';
      this.metricsLoaded = false;
      await this.$store.dispatch('Router_GetGrafana');

      if (!this.monitoringConfigured) {
        this.metricError = this.$t('Monitoring.NotConfigured');
        return;
      }

      let startTime = moment(this.today).startOf('day').format('YYYY-MM-DD HH:mm');
      let endTime = moment(startTime).add(1, 'days').format('YYYY-MM-DD HH:mm'); // 1 day
      let startHourTime = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm');
      let endHourTime = moment().format('YYYY-MM-DD HH:mm'); // 1 hour

      const successResult = await this.$store.dispatch('Router_GetSuccessRate', {
        name: this.name,
        namespace: this.namespace,
        startTime: moment(startTime).unix(),
        endTime: moment(endTime).unix(),
        interval: '1h'
      });
      const hourResult = await this.$store.dispatch('Router_GetHourSuccessRate', {
        name: this.name,
        namespace: this.namespace,
        startTime: moment(startHourTime).unix(),
        endTime: moment(endHourTime).unix(),
        interval: '1h'
      });
      const latencyResult = await this.$store.dispatch('Router_GetLatency', {
        name: this.name,
        namespace: this.namespace,
        percentage: 0.99,
        startTime: moment(startTime).unix(),
        endTime: moment(endTime).unix(),
        interval: '1h'
      });
      const opsResult = await this.$store.dispatch('Router_GetOPS', {
        name: this.name,
        namespace: this.namespace,
        startTime: moment(startTime).unix(),
        endTime: moment(endTime).unix(),
        interval: '1h'
      });

      const failed = [successResult, hourResult, latencyResult, opsResult].find((result) => result?.ok === false);
      if (failed) {
        this.metricError = failed.message || this.$t('Monitoring.QueryFailed');
      }
      this.metricsLoaded = true;
    },
    save: async function(date) {
      this.today = date;

      await this.fetchData();
    },
    normalizedSeries(series) {
      if (!Array.isArray(series) || !series.length) return [0];
      return series.map((value) => Number(value) || 0);
    },
    sparklinePoints(series) {
      const values = this.normalizedSeries(series);
      if (values.length === 1) return `0,18 100,18`;

      const min = Math.min(...values);
      const max = Math.max(...values);
      const range = max - min || 1;

      return values.map((value, index) => {
        const x = (index / (values.length - 1)) * 100;
        const y = 32 - ((value - min) / range) * 28;
        return `${x.toFixed(2)},${y.toFixed(2)}`;
      }).join(' ');
    },
    metricLabel(series) {
      const values = this.normalizedSeries(series);
      return values.slice(-8).join(' / ');
    },
  },
  watch: {
    '$route': async function() {
      this.name = this.$route.query.name || this.$route.params.name || '';
      this.namespace = this.$route.query.namespace;

      if (this.name && this.namespace) {
        await this.fetchData();
      }
    },
  },
  computed: {
    ...mapGetters({
      successAvg: 'Router_GetSuccessAvg',
      totalRequest: 'Router_GetTotalRequest',
      failRequest: 'Router_GetFailRequest',
      successHourAvg: 'Router_GetSuccessHourAvg',
      totalHourRequest: 'Router_GetTotalHourRequest',
      failHourRequest: 'Router_GetFailHourRequest',
      successMetric: 'Router_GetSuccessRate',
      latencyMetric: 'Router_GetLatency',
      opsMetric: 'Router_GetOPS',
      grafanaConfig: 'Router_GetGrafana'
    }),
    maxDate() {
      return moment().format('YYYY-MM-DD');
    },
    grafanaHref() {
      if (!this.grafanaConfig.host) return '';
      const protocol = this.grafanaConfig.isTls ? 'https://' : 'http://';
      return protocol + this.grafanaConfig.host + ':' + this.grafanaConfig.port;
    },
    monitoringConfigured() {
      return this.grafanaConfig?.configured === true && Boolean(this.grafanaConfig?.host && this.grafanaConfig?.port);
    },
    hasTelemetry() {
      const series = [
        this.successMetric,
        this.latencyMetric,
        this.opsMetric,
      ].flatMap((metric) => Array.isArray(metric) ? metric : []);

      return this.totalRequest > 0 || this.totalHourRequest > 0 || series.some((value) => Number(value) > 0);
    },
    metricNotice() {
      if (this.metricError) {
        return {
          title: this.$t('Monitoring.QueryFailed'),
          message: this.metricError,
        };
      }

      if (this.metricsLoaded && !this.hasTelemetry) {
        return {
          title: this.$t('Monitoring.NoTelemetry'),
          message: this.grafanaConfig.provider === 'prometheus'
            ? 'Check that Prometheus has Istio 1.7 metrics for this service.'
            : `Check Grafana datasource ${this.grafanaConfig.datasourceId || '1'} and Istio telemetry for this service.`,
        };
      }

      return null;
    }
  }
}
</script>

<style scoped>
.dashboard-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: minmax(220px, 1fr) minmax(220px, 1fr) minmax(320px, 2fr);
}

.dashboard-card {
  align-content: start;
}

.ring-progress {
  --progress: 0;
  align-items: center;
  aspect-ratio: 1;
  background:
    radial-gradient(circle, #fff 58%, transparent 59%),
    conic-gradient(var(--pw-primary) calc(var(--progress) * 1%), var(--pw-surface-muted) 0);
  border-radius: 999px;
  display: grid;
  justify-items: center;
  margin: 8px auto 18px;
  max-width: 190px;
  width: 100%;
}

.ring-progress strong {
  font-size: 2.4rem;
  letter-spacing: -0.07em;
}

.dashboard-title {
  color: var(--pw-muted);
  font-weight: 800;
  margin: 0;
  text-align: center;
}

.metric-tile {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 18px;
  display: grid;
  gap: 4px;
  justify-items: end;
  padding: 16px;
}

.metric-tile.primary {
  background: var(--pw-primary);
  color: #fff;
}

.metric-tile strong {
  font-size: 2.1rem;
  line-height: 1;
}

.dashboard-link {
  margin-top: 8px;
}

.dashboard-chart-panel {
  align-content: start;
}

.date-field {
  max-width: 240px;
}

.metrics-notice {
  background: #fff7ed;
  border: 1px solid #fed7aa;
  border-radius: 16px;
  color: #9a3412;
  display: grid;
  gap: 4px;
  padding: 14px 16px;
}

.metrics-notice span {
  color: #9a3412;
  font-size: 0.88rem;
}

.spark-card {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 18px;
  display: grid;
  gap: 8px;
  padding: 16px;
}

.spark-card p {
  color: var(--pw-muted);
  font-weight: 800;
  margin: 0;
}

.sparkline {
  background: linear-gradient(180deg, #fff, var(--pw-surface-muted));
  border-radius: 14px;
  height: 76px;
  width: 100%;
}

.sparkline polyline {
  fill: none;
  stroke: var(--pw-primary);
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2.8;
}

.spark-card.warm .sparkline polyline {
  stroke: var(--pw-accent);
}

.spark-card.muted .sparkline polyline {
  stroke: var(--pw-muted);
}

.spark-values {
  color: var(--pw-muted);
  font-size: 0.78rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1100px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}
</style>
