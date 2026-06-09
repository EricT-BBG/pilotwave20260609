<template>
  <section class="relationship-panel" :aria-label="$t('Router.AssociationInfo')">
    <div class="relationship-map">
      <aside class="relationship-column gateways-column">
        <div class="column-heading">
          <span class="column-kicker">{{ $t('Router.ConnectGateway') }}</span>
          <strong>{{ gatewayItems.length }}</strong>
        </div>

        <div v-if="gatewayItems.length" class="item-list">
          <article
            v-for="gateway in gatewayItems"
            :key="gatewayKey(gateway)"
            class="relation-card gateway-card"
          >
            <div class="card-heading">
              <span class="card-kicker">{{ $t('Gateway.Gateway') }}</span>
              <strong class="primary-value">{{ gateway.name || '-' }}</strong>
            </div>
            <dl class="relation-fields">
              <div>
                <dt>{{ $t('Table.Namespace') }}</dt>
                <dd>{{ gateway.namespace || '-' }}</dd>
              </div>
              <div>
                <dt>{{ $t('Table.Hosts') }}</dt>
                <dd>{{ gatewayHosts(gateway).join(', ') || '-' }}</dd>
              </div>
              <div>
                <dt>{{ $t('Table.Port') }}</dt>
                <dd>{{ gateway.ports || '-' }}</dd>
              </div>
            </dl>
          </article>
        </div>

        <div v-else class="relationship-empty compact">
          {{ $t('Router.NoGatewaysAssociated') }}
        </div>
      </aside>

      <div class="connector connector-left" aria-hidden="true">
        <span>{{ $t('Router.RoutesToLabel') }}</span>
      </div>

      <article class="router-node">
        <span class="node-kicker">{{ $t('Router.ResourceName') }}</span>
        <h3>{{ name || '-' }}</h3>
        <dl>
          <div>
            <dt>{{ $t('Table.Namespace') }}</dt>
            <dd>{{ namespace || '-' }}</dd>
          </div>
          <div>
            <dt>{{ $t('Table.Hosts') }}</dt>
            <dd>{{ routerHosts.length }}</dd>
          </div>
          <div>
            <dt>{{ $t('Table.RuleCount') }}</dt>
            <dd>{{ routeCount }}</dd>
          </div>
          <div>
            <dt>{{ $t('Router.Destinations') }}</dt>
            <dd>{{ destinationCount }}</dd>
          </div>
        </dl>
      </article>

      <div class="connector connector-right" aria-hidden="true">
        <span>{{ $t('Router.ForwardsToLabel') }}</span>
      </div>

      <aside class="relationship-column destinations-column">
        <div class="column-heading">
          <span class="column-kicker">{{ $t('Router.HostsLabel') }} / {{ $t('Router.Destinations') }}</span>
          <strong>{{ destinationGroupCount }}</strong>
        </div>

        <div v-if="destinationGroups.length" class="item-list">
          <article
            v-for="group in destinationGroups"
            :key="group.key"
            class="relation-card destination-card"
          >
            <div class="card-heading">
              <span class="card-kicker">{{ group.label }}</span>
              <strong class="primary-value">{{ group.primary }}</strong>
            </div>
            <dl class="relation-fields destination-fields">
              <div v-if="group.namespace">
                <dt>{{ $t('Table.Namespace') }}</dt>
                <dd>{{ group.namespace }}</dd>
              </div>
              <div v-if="group.prefixs.length">
                <dt>{{ $t('Router.Prefix') }}</dt>
                <dd>{{ group.prefixs.join(', ') }}</dd>
              </div>
              <div v-if="group.port">
                <dt>{{ $t('Router.Port') }}</dt>
                <dd>{{ group.port }}</dd>
              </div>
              <div v-if="group.weight !== ''">
                <dt>{{ $t('Router.Weight') }}</dt>
                <dd>{{ group.weight }}</dd>
              </div>
              <div v-if="group.subset">
                <dt>{{ $t('Router.Subset') }}</dt>
                <dd>{{ group.subset }}</dd>
              </div>
            </dl>
          </article>
        </div>

        <div v-else class="relationship-empty compact">
          {{ $t('Router.NoDestinationsConfigured') }}
        </div>
      </aside>
    </div>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'RouterCytoscape',
  props: [ 'dName', 'dNamespace' ],
  mounted: async function() {
    await this.loadRouterRelationship();
  },
  data () {
    return {
      name: '',
      namespace: '',
    }
  },
  methods: {
    async loadRouterRelationship() {
      this.$store.commit('Router_ResetStatus');
      this.name = this.$route.query.name || this.$route.params.name || this.dName;
      this.namespace = this.$route.query.namespace || this.dNamespace;

      if (!this.name || !this.namespace) return;
      await this.fetchData();
      await this.fetchMapping();
      await this.fetchRule();
    },
    fetchRule: async function() {
      await this.$store.dispatch('Router_GetRule', {
        name: this.name,
        namespace: this.namespace
      });
    },
    fetchData: async function() {
      await this.$store.dispatch('Router_GetItem', {
        name: this.name,
        namespace: this.namespace
      });
    },
    fetchMapping: async function() {
      await this.$store.dispatch('Gateway_GetItems', {
        page: 1,
        limit: -1,
        namespace: '',
      });

      await this.$store.dispatch('Router_GetMappings', {
        name: this.name,
        namespace: this.namespace,
        gateways: this.gateways
      });
    },
    gatewayKey(gateway) {
      return [gateway.namespace, gateway.name].filter(Boolean).join(':');
    },
    gatewayHosts(gateway) {
      if (Array.isArray(gateway.hostNames)) return gateway.hostNames.filter(Boolean);
      if (!gateway.hosts) return [];
      return String(gateway.hosts).split(/\s*;\s*|\s*,\s*/).filter(Boolean);
    },
    parseServiceHost(host) {
      const value = String(host || '').trim();
      const parts = value.split('.');
      const serviceIndex = parts.indexOf('svc');
      if (serviceIndex === 2 && parts[0] && parts[1]) {
        return {
          name: parts[0],
          namespace: parts[1],
        };
      }

      return {
        name: value || '-',
        namespace: '',
      };
    },
    destinationKey(routeIndex, destinationIndex, destination) {
      return [
        routeIndex,
        destinationIndex,
        destination.host,
        destination.port,
        destination.subset,
      ].filter((item) => item !== undefined && item !== null).join(':');
    },
  },
  watch: {
    '$route': async function() {
      await this.loadRouterRelationship();
    },
    dName: async function() {
      await this.loadRouterRelationship();
    },
    dNamespace: async function() {
      await this.loadRouterRelationship();
    },
  },
  computed: {
    ...mapGetters({
      gateways: 'Gateway_GetItems',
      router: 'Router_GetItem',
      mappings: 'Router_GetMappings',
      httpItems: 'Router_GetHttpItems'
    }),
    gatewayItems() {
      return this.mappings || [];
    },
    routerHosts() {
      return this.router?.hosts || [];
    },
    routeCount() {
      return (this.httpItems || []).length;
    },
    destinationGroups() {
      const hostCards = this.routerHosts.map((host, index) => {
        const parsed = this.parseServiceHost(host);
        return {
          key: ['host', index, host].join(':'),
          label: this.$t('Router.Host'),
          primary: parsed.name,
          namespace: parsed.namespace,
          prefixs: [],
          port: '',
          weight: '',
          subset: '',
        };
      });

      const destinationCards = [];
      (this.httpItems || []).forEach((item, routeIndex) => {
        (item.destinations || []).forEach((destination, destinationIndex) => {
          const parsed = this.parseServiceHost(destination.host);
          destinationCards.push({
            key: ['destination', this.destinationKey(routeIndex, destinationIndex, destination)].join(':'),
            label: this.$t('Router.Destination'),
            primary: parsed.name,
            namespace: parsed.namespace,
            prefixs: (item.prefixs || []).filter(Boolean),
            port: destination.port || '',
            weight: destination.weight !== undefined && destination.weight !== null ? destination.weight : '',
            subset: destination.subset || '',
          });
        });
      });

      return hostCards.concat(destinationCards);
    },
    destinationCount() {
      return (this.httpItems || []).reduce((sum, item) => {
        return sum + (item.destinations || []).filter((destination) => destination.host).length;
      }, 0);
    },
    destinationGroupCount() {
      return this.destinationGroups.length;
    },
  },
}
</script>

<style scoped>
.relationship-panel {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  padding: 20px;
}

.relationship-map {
  align-items: center;
  display: grid;
  gap: 18px;
  grid-template-columns: minmax(220px, 1fr) minmax(72px, 96px) minmax(240px, 300px) minmax(72px, 96px) minmax(260px, 1.25fr);
  min-height: 420px;
}

.relationship-column {
  align-self: stretch;
  display: grid;
  gap: 14px;
  grid-template-rows: auto 1fr;
  min-width: 0;
}

.column-heading {
  align-items: end;
  border-bottom: 1px solid var(--pw-border);
  display: flex;
  justify-content: space-between;
  padding-bottom: 10px;
}

.column-heading strong {
  color: var(--pw-primary-strong);
  font-size: 1.4rem;
}

.column-kicker,
.node-kicker {
  color: var(--pw-muted);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.item-list {
  align-content: center;
  display: grid;
  gap: 12px;
}

.relation-card,
.router-node,
.relationship-empty {
  background: var(--pw-surface);
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  min-width: 0;
}

.relation-card {
  box-shadow: var(--pw-shadow-soft);
  display: grid;
  gap: 10px;
  padding: 14px;
}

.gateway-card {
  border-left: 5px solid #3f6f9f;
}

.destination-card {
  border-left: 5px solid #7c6a43;
}

.card-heading {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.card-kicker {
  color: var(--pw-muted);
  display: block;
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.primary-value {
  color: var(--pw-text);
  display: block;
  font-size: 1.08rem;
  line-height: 1.25;
  overflow-wrap: anywhere;
}

.router-node {
  box-shadow: var(--pw-shadow);
  display: grid;
  gap: 16px;
  justify-self: stretch;
  padding: 22px;
  position: relative;
  z-index: 1;
}

.router-node h3 {
  font-size: 1.55rem;
  line-height: 1.2;
  margin: 0;
  overflow-wrap: anywhere;
}

.router-node dl,
.relation-fields {
  display: grid;
  gap: 10px;
  margin: 0;
}

.router-node dl {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.router-node dl div:first-child {
  grid-column: 1 / -1;
}

.router-node dt,
.relation-fields dt {
  color: var(--pw-muted);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.router-node dd,
.relation-fields dd {
  font-weight: 800;
  margin: 3px 0 0;
  overflow-wrap: anywhere;
}

.relation-fields {
  grid-template-columns: repeat(auto-fit, minmax(96px, 1fr));
}

.gateway-card .relation-fields div:first-child {
  grid-column: 1 / -1;
}

.destination-fields {
  grid-template-columns: repeat(auto-fit, minmax(112px, 1fr));
}

.connector {
  align-items: center;
  color: var(--pw-muted);
  display: grid;
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  min-height: 48px;
  position: relative;
  text-align: center;
  text-transform: uppercase;
}

.connector::before {
  background: var(--pw-border);
  content: "";
  height: 2px;
  left: 0;
  position: absolute;
  right: 0;
  top: 50%;
}

.connector::after {
  border-bottom: 5px solid transparent;
  border-left: 8px solid var(--pw-border);
  border-top: 5px solid transparent;
  content: "";
  position: absolute;
  right: 0;
  top: calc(50% - 5px);
}

.connector-left::after {
  left: 0;
  right: auto;
  transform: rotate(180deg);
}

.connector span {
  background: #fff;
  justify-self: center;
  padding: 0 8px;
  position: relative;
  z-index: 1;
}

.relationship-empty {
  align-self: center;
  color: var(--pw-muted);
  font-weight: 700;
  padding: 28px 16px;
  text-align: center;
}

@media (max-width: 1100px) {
  .relationship-map {
    grid-template-columns: 1fr;
    min-height: 0;
  }

  .relationship-column {
    grid-template-rows: auto auto;
  }

  .gateways-column {
    order: 1;
  }

  .connector-left {
    order: 2;
  }

  .router-node {
    order: 3;
  }

  .connector-right {
    order: 4;
  }

  .destinations-column {
    order: 5;
  }

  .connector {
    min-height: 40px;
  }

  .connector::before {
    bottom: 0;
    height: auto;
    left: 50%;
    right: auto;
    top: 0;
    width: 2px;
  }

  .connector::after {
    border-left: 5px solid transparent;
    border-right: 5px solid transparent;
    border-top: 8px solid var(--pw-border);
    bottom: 0;
    left: calc(50% - 5px);
    right: auto;
    top: auto;
    transform: none;
  }
}

@media (max-width: 640px) {
  .relationship-panel {
    padding: 14px;
  }

  .router-node dl,
  .destination-fields {
    grid-template-columns: 1fr;
  }
}
</style>
