<template>
  <article class="server-card compact-server-card">
    <div class="server-header">
      <strong>{{ $t('Gateway.Hosts') }} {{ index + 1 }}*</strong>
      <button
        class="danger-button"
        :data-testid="`gateway-server-remove-${index}`"
        :disabled="!canRemoveServer"
        type="button"
        @click="removeServer"
      >
        {{ $t('Gateway.DeleteHostBlock') }}
      </button>
    </div>

    <label class="field server-host-row">
      <span>{{ $t('Gateway.Hosts') }}</span>
      <input
        data-testid="gateway-server-hosts"
        :data-server-index="index"
        :value="hostsText"
        placeholder="example.local, *.example.local"
        type="text"
        @input="updateHostText($event.target.value)"
      />
    </label>

    <div class="section-toolbar listener-toolbar">
      <div>
        <strong>{{ $t('Gateway.ListenerRules') }}</strong>
        <p>{{ $t('Gateway.ListenerRulesHelp') }}</p>
      </div>
      <button class="secondary-button" :data-testid="`gateway-server-add-port-${index}`" type="button" @click="addPort">
        + {{ $t('Gateway.AddPort') }}
      </button>
    </div>

    <div class="port-list">
      <RouterItem
        v-for="(item, i) in server.ports"
        :key="'port' + index + '-' + i"
        :port="item.port"
        :protocol="item.protocol"
        :credentialname="item.credentialname"
        :mode="item.mode"
        :name="item.name"
        :cert="item.cert"
        :pkey="item.pkey"
        :cacert="item.cacert"
        :hosts="server.hosts"
        :server-index="index"
        :index="i"
      />
    </div>
  </article>
</template>

<script>
import { mapGetters } from 'vuex';
import RouterItem from './RouterItem.vue';

export default {
  name: 'GatewayServerItem',
  components: {
    RouterItem,
  },
  props: [ 'server', 'index' ],
  methods: {
    updateHostText(value) {
      this.$store.commit('Gateway_UpdateHosts', {
        hosts: value
          .split(',')
          .map((item) => item.trim())
          .filter(Boolean),
        serverIndex: this.index,
      });
    },
    removeServer() {
      if (this.servers.length === 1) return;
      this.$store.commit('Gateway_RemoveServer', {
        serverIndex: this.index,
      });
    },
    addPort() {
      this.$store.commit('Gateway_AddPorts', {
        serverIndex: this.index,
      });
    },
  },
  computed: {
    ...mapGetters({
      servers: 'Gateway_GetServers',
    }),
    hostsText() {
      return this.server.hosts?.join(', ') || '';
    },
    canRemoveServer() {
      return this.servers.length > 1;
    },
  },
}
</script>

<style scoped>
.server-card {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  display: grid;
  gap: 10px;
  padding: 12px;
}

.server-header,
.section-toolbar {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.section-toolbar p {
  color: var(--pw-muted);
  margin: 4px 0 0;
}

.server-header {
  border-bottom: 1px solid var(--pw-border);
  padding-bottom: 8px;
}

.port-list {
  display: grid;
  gap: 8px;
}

.listener-toolbar {
  padding-top: 4px;
}

.server-host-row {
  margin: 0;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}
</style>
