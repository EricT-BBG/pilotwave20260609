<template>
  <section class="page">
    <header class="page-header detail-header detail-header--compact">
      <div class="detail-header-row">
        <button class="secondary-button detail-back-button" type="button" @click="goBack">
          ← {{ $t('Form.BackToRouters') }}
        </button>
        <div class="detail-title-card">
          <div class="detail-title-text">
            <p class="eyebrow detail-resource-type">{{ $t('System.ServiceRouter') }}</p>
            <h1 class="detail-header-title">{{ name }}</h1>
          </div>
        </div>
        <span class="namespace-chip">
          <span class="namespace-chip-label">{{ $t('Table.Namespace') }}</span>
          <span class="namespace-chip-value">{{ namespace || '-' }}</span>
        </span>
      </div>
    </header>

    <div v-if="!contentReady" class="panel">
      <div class="empty-state compact">{{ $t('Router.LoadingDetail') }}</div>
    </div>

    <div v-else class="panel">
      <div v-if="!editMode && !routeRuleEditMode" class="detail-tab-toolbar">
        <div class="tab-strip" role="tablist">
          <button
            v-for="item in readonlyTabs"
            :key="item.key"
            class="tab-button"
            :class="{ active: selected === item.key }"
            :data-testid="`router-tab-${item.key}`"
            type="button"
            @click="selectTab(item.key)"
          >
            {{ item.label }}
          </button>
        </div>
        <button
          v-if="showDetailEditButton"
          class="primary-button detail-edit-button"
          data-testid="router-edit-open"
          type="button"
          @click="openEditMode"
        >
          {{ $t('Form.Edit') }}
        </button>
      </div>

      <div class="tab-panel">
        <BasicSetting v-if="editMode" :key="detailKey + ':basic'" />
        <template v-else>
          <Cytoscape v-if="selected === 'information'" :key="detailKey + ':cy'" />
          <Dashboard v-else-if="selected === 'dashboard'" :key="detailKey + ':dashboard'" />
          <RouterSetting
            v-else-if="selected === 'router'"
            :key="detailKey + ':router'"
            :readonly="!routeRuleEditMode"
            @close-edit="closeRouteRuleEdit"
          />
          <GatewaySetting v-else-if="selected === 'gateway'" :key="detailKey + ':gateway'" />
        </template>
      </div>
    </div>
  </section>
</template>

<script>
import { defineAsyncComponent } from 'vue';
import { mapGetters } from 'vuex';

export default {
  name: 'RouterDetail',
  components: {
    Cytoscape: defineAsyncComponent(() => import('../../components/router/Cytoscape.vue')),
    Dashboard: defineAsyncComponent(() => import('../../components/router/Dashboard.vue')),
    BasicSetting: defineAsyncComponent(() => import('../../components/router/EditSetting.vue')),
    RouterSetting: defineAsyncComponent(() => import('../../components/router/RouterSetting.vue')),
    GatewaySetting: defineAsyncComponent(() => import('../../components/router/GatewaySetting.vue'))
  },
  mounted: async function() {
    await this.initializeView();
  },
  data() {
    return {
      selected: 'information',
      name: '',
      namespace: '',
      editMode: false,
      routeRuleEditMode: false,
      contentReady: false
    }
  },
  methods: {
    initializeView: async function() {
      this.$store.commit('Router_ResetStatus');
      this.contentReady = false;

      await this.syncRouteContext();
      this.resetTabs();

      this.contentReady = true;
    },
    resolveNamespace: async function(name) {
      const routers = await this.$store.dispatch('Router_GetItems', {
        page: 1,
        limit: -1,
        namespace: ''
      });
      const match = routers.find((item) => item.name === name);
      return match?.namespace || '';
    },
    syncRouteContext: async function() {
      const name = this.$route.query.name || this.$route.params.name || '';
      let namespace = this.$route.query.namespace || '';

      if (!name) {
        this.name = '';
        this.namespace = '';
        return;
      }

      if (!namespace) {
        namespace = await this.resolveNamespace(name);
      }

      this.name = name;
      this.namespace = namespace;

      if (name !== this.$route.query.name || namespace !== this.$route.query.namespace) {
        await this.$router.replace({
          name: 'RouterDetail',
          params: { name },
          query: {
            ...this.$route.query,
            name,
            namespace
          }
        });
      }
    },
    resetTabs() {
      this.routeRuleEditMode = false;
      const requestedTab = this.$route.query.tab || 'information';
      if (requestedTab === 'setting') {
        this.selected = 'setting';
        this.editMode = true;
        return;
      }

      const allowedTabs = this.readonlyTabs.map((item) => item.key);
      this.selected = allowedTabs.includes(requestedTab) ? requestedTab : 'information';
      this.editMode = false;
    },
    selectTab(tab) {
      this.$store.commit('Router_ResetStatus');
      this.editMode = false;
      this.routeRuleEditMode = false;
      this.selected = tab;
    },
    openEditMode() {
      this.$store.commit('Router_ResetStatus');
      if (this.selected === 'router') {
        this.editMode = false;
        this.routeRuleEditMode = true;
        return;
      }

      this.routeRuleEditMode = false;
      this.selected = 'setting';
      this.editMode = true;
    },
    closeRouteRuleEdit() {
      this.$store.commit('Router_ResetStatus');
      this.routeRuleEditMode = false;
    },
    goBack() {
      window.scrollTo(0, 0);
      this.$router.push('/routers');
    },
  },
  watch: {
    '$route': async function() {
      await this.initializeView();
    },
  },
  computed: {
    ...mapGetters({
      language: 'Auth_GetLanguage',
    }),
    readonlyTabs() {
      return [
        { key: 'information', label: this.$t('Router.AssociationInfo') },
        { key: 'dashboard', label: this.$t('Router.Overview') },
        { key: 'router', label: this.$t('Router.RuleSetting') },
        { key: 'gateway', label: this.$t('Router.ConnectGateway') },
      ];
    },
    showDetailEditButton() {
      return this.selected !== 'gateway';
    },
    detailKey() {
      return [this.name, this.namespace].filter(Boolean).join(':');
    }
  }
}
</script>

<style scoped>
.tab-strip {
  border-bottom: 1px solid var(--pw-border);
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

.detail-tab-toolbar {
  align-items: center;
  border-bottom: 1px solid var(--pw-border);
  display: flex;
  gap: 16px;
  justify-content: space-between;
}

.detail-tab-toolbar .tab-strip {
  border-bottom: 0;
}

.detail-edit-button {
  min-height: 44px;
  min-width: 92px;
  white-space: nowrap;
}

.tab-button {
  background: transparent;
  border: 0;
  border-bottom: 3px solid transparent;
  color: var(--pw-muted);
  font-weight: 800;
  padding: 12px 4px;
}

.tab-button.active {
  border-color: var(--pw-accent);
  color: var(--pw-primary-strong);
}

.tab-panel {
  min-height: 360px;
}

@media (max-width: 760px) {
  .detail-tab-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .detail-edit-button {
    width: 100%;
  }
}
</style>
