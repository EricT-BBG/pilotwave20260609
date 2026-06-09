<template>
  <section class="page">
    <header class="page-header">
      <div>
        <p class="eyebrow">{{ $t('Dashboard.RouterOverview') }}</p>
        <h1>{{ $t('System.Dashboard') }}</h1>
      </div>
    </header>

    <section class="panel">
      <div class="dashboard-filter-row">
        <label class="field">
          <span>{{ $t('Table.Namespace') }}</span>
          <select v-model="selectedNamespace" :disabled="!namespaceOptions.length" @change="updateNamespace">
            <option value="All">{{ $t('NamespaceInjection.AllNamespaces') }}</option>
            <option v-for="item in namespaceOptions" :key="item" :value="item">
              {{ item }}
            </option>
          </select>
        </label>

        <label class="field">
          <span>{{ $t('Gateway.SelectRouter') }}</span>
          <select v-model="selectedRouterKey" :disabled="!filteredRouters.length" @change="updateRouter">
            <option value="">{{ routerSelectPlaceholder }}</option>
            <option v-for="item in filteredRouters" :key="routerKey(item)" :value="routerKey(item)">
              {{ item.name }} / {{ item.namespace }}
            </option>
          </select>
        </label>
      </div>

      <div v-if="hasSelectedRouter" class="metric-grid">
        <div v-if="metricNotice" class="metrics-notice metric-grid-notice">
          <strong>{{ metricNotice.title }}</strong>
          <span>{{ metricNotice.message }}</span>
        </div>

        <article class="metric-card">
          <span>{{ $t('Router.ConnectionToday') }}</span>
          <strong>{{ successAvg }}%</strong>
          <div class="metric-bar"><span :style="{ width: metricWidth(successAvg) }"></span></div>
          <small>{{ totalRequest }} {{ $t('Router.RequestToday') }} / {{ failRequest }} {{ $t('Router.FailedToday') }}</small>
        </article>

        <article class="metric-card">
          <span>{{ $t('Router.ConnectionHour') }}</span>
          <strong>{{ successHourAvg }}%</strong>
          <div class="metric-bar"><span :style="{ width: metricWidth(successHourAvg) }"></span></div>
          <small>{{ totalHourRequest }} {{ $t('Router.RequestHour') }} / {{ failHourRequest }} {{ $t('Router.FailedHour') }}</small>
        </article>

        <article class="metric-card wide">
          <span>{{ $t('Router.SuccessRate') }}</span>
          <code>{{ metricPreview(successMetric) }}</code>
          <span>{{ $t('Router.Latency') }}</span>
          <code>{{ metricPreview(latencyMetric) }}</code>
          <span>OPS</span>
          <code>{{ metricPreview(opsMetric) }}</code>
        </article>
      </div>

      <div v-else class="empty-state empty-state-list">
        <div class="empty-illustration" aria-hidden="true">
          <span>RT</span>
        </div>
        <div class="empty-copy">
          <p class="eyebrow">{{ $t('ResourceList.ReadyForSetup') }}</p>
          <h2>{{ emptyStateTitle }}</h2>
          <p>{{ emptyStateMessage }}</p>
          <router-link class="primary-button" to="/new/router">
            {{ $t('Router.New') }}
          </router-link>
        </div>
      </div>
    </section>
  </section>
</template>

<script>
import moment from 'moment';
import { mapGetters } from 'vuex';

export default {
  name: 'Home',
  data() {
    return {
      selectedRouterKey: '',
      selectedNamespace: 'All',
      dName: '',
      dNamespace: '',
      metricError: '',
      metricsLoaded: false,
    };
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    async fetchData() {
      await this.$store.dispatch('Router_GetMenuItems', {
        page: 1,
        limit: -1,
        today: moment().format('YYYY-MM-DD'),
      });

      if (!this.filteredRouters.length) {
        this.selectedRouterKey = '';
        this.dName = '';
        this.dNamespace = '';
        return;
      }

      const first = this.filteredRouters[0];
      this.selectedRouterKey = this.routerKey(first);
      this.syncRouter(first);
      await this.fetchTime();
    },
    routerKey(router) {
      return `${router?.namespace || ''}/${router?.name || ''}`;
    },
    syncRouter(router) {
      this.dName = router?.name || '';
      this.dNamespace = router?.namespace || '';
    },
    async fetchTime() {
      if (!this.dName || !this.dNamespace) return;
      this.metricError = '';
      this.metricsLoaded = false;
      await this.$store.dispatch('Router_GetGrafana');

      if (!this.monitoringConfigured) {
        this.metricError = this.$t('Monitoring.NotConfigured');
        return;
      }

      const startTime = moment().startOf('day').format('YYYY-MM-DD HH:mm');
      const endTime = moment(startTime).add(1, 'days').format('YYYY-MM-DD HH:mm');
      const startHourTime = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm');
      const endHourTime = moment().format('YYYY-MM-DD HH:mm');

      const successResult = await this.$store.dispatch('Router_GetSuccessRate', {
        name: this.dName,
        namespace: this.dNamespace,
        startTime: moment(startTime).unix(),
        endTime: moment(endTime).unix(),
        interval: '1h',
      });
      const hourResult = await this.$store.dispatch('Router_GetHourSuccessRate', {
        name: this.dName,
        namespace: this.dNamespace,
        startTime: moment(startHourTime).unix(),
        endTime: moment(endHourTime).unix(),
        interval: '1h',
      });
      const latencyResult = await this.$store.dispatch('Router_GetLatency', {
        name: this.dName,
        namespace: this.dNamespace,
        percentage: 0.99,
        startTime: moment(startTime).unix(),
        endTime: moment(endTime).unix(),
        interval: '1h',
      });
      const opsResult = await this.$store.dispatch('Router_GetOPS', {
        name: this.dName,
        namespace: this.dNamespace,
        startTime: moment(startTime).unix(),
        endTime: moment(endTime).unix(),
        interval: '1h',
      });

      const failed = [successResult, hourResult, latencyResult, opsResult].find((result) => result?.ok === false);
      if (failed) {
        this.metricError = failed.message || this.$t('Monitoring.QueryFailed');
      }
      this.metricsLoaded = true;
    },
    async updateRouter() {
      const router = this.filteredRouters.find((item) => this.routerKey(item) === this.selectedRouterKey);
      if (!router) return;
      this.syncRouter(router);
      await this.fetchTime();
    },
    async updateNamespace() {
      const router = this.filteredRouters.find((item) => this.routerKey(item) === this.selectedRouterKey) || this.filteredRouters[0];
      if (!router) {
        this.selectedRouterKey = '';
        this.dName = '';
        this.dNamespace = '';
        return;
      }

      this.selectedRouterKey = this.routerKey(router);
      this.syncRouter(router);
      await this.fetchTime();
    },
    metricWidth(value) {
      const bounded = Math.max(0, Math.min(100, Number(value) || 0));
      return `${bounded}%`;
    },
    metricPreview(value) {
      if (!Array.isArray(value) || !value.length) return this.$t('Dashboard.NoData');
      return value.slice(-8).join(', ');
    },
  },
  computed: {
    ...mapGetters({
      routers: 'Router_GetMenuItems',
      language: 'Auth_GetLanguage',
      successAvg: 'Router_GetSuccessAvg',
      totalRequest: 'Router_GetTotalRequest',
      failRequest: 'Router_GetFailRequest',
      successHourAvg: 'Router_GetSuccessHourAvg',
      totalHourRequest: 'Router_GetTotalHourRequest',
      failHourRequest: 'Router_GetFailHourRequest',
      successMetric: 'Router_GetSuccessRate',
      latencyMetric: 'Router_GetLatency',
      opsMetric: 'Router_GetOPS',
      grafanaConfig: 'Router_GetGrafana',
    }),
    hasSelectedRouter() {
      return Boolean(this.dName && this.dNamespace);
    },
    namespaceOptions() {
      return [...new Set((this.routers || []).map((item) => item?.namespace).filter(Boolean))].sort();
    },
    filteredRouters() {
      if (this.selectedNamespace === 'All') return this.routers || [];
      return (this.routers || []).filter((item) => item?.namespace === this.selectedNamespace);
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
    },
    emptyStateTitle() {
      return this.language === 'tw' ? '目前尚未建立 VirtualService' : 'No VirtualServices have been created';
    },
    emptyStateMessage() {
      return this.language === 'tw'
        ? '建立第一個 VirtualService 後，這裡會顯示流量成功率、延遲與關聯資訊。'
        : 'Create the first VirtualService to start showing traffic success rate, latency, and association details.';
    },
    routerSelectPlaceholder() {
      return this.language === 'tw' ? '尚無 VirtualService 可選' : 'No VirtualService to select';
    },
  },
};
</script>
