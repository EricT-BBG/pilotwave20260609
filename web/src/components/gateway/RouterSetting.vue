<template>
  <section class="form-stack">
    <div v-if="!associationEditMode" class="association-readonly-toolbar">
      <div>
        <strong>{{ $t('Gateway.AssociationRouting') }}</strong>
      </div>
      <button
        class="primary-button association-edit-button"
        data-testid="association-edit-open"
        type="button"
        @click="openAssociationEdit"
      >
        {{ $t('Gateway.EditRouterAssociations') }}
      </button>
    </div>

    <div v-if="associationEditMode" class="association-form">
      <div class="association-toolbar">
        <label class="field association-namespace">
          <span>{{ $t('Table.Namespace') }}</span>
          <select v-model="namespace" @change="fetchRouter">
            <option value="All">{{ $t('NamespaceInjection.AllNamespaces') }}</option>
            <option v-for="item in namespaces" :key="item" :value="item">
              {{ item }}
            </option>
          </select>
        </label>

        <div class="association-delta-summary">
          <span>{{ $t('Form.Selected') }} {{ selectedRouters.length }}</span>
          <span>{{ $t('Form.Added') }} {{ addedRouters.length }}</span>
          <span>{{ $t('Form.Removed') }} {{ removedRouters.length }}</span>
        </div>

        <button class="primary-button association-submit" type="button" @click="submit">
          {{ $t('Gateway.UpdateRouter') }}
        </button>
        <button class="secondary-button association-submit" type="button" @click="closeAssociationEdit">
          {{ $t('Form.Cancel') }}
        </button>
      </div>

      <div class="association-list-wrap">
        <div class="association-list" data-testid="gateway-router-association-list">
          <label
            v-for="item in routers"
            :key="item.value"
            class="association-row"
            :class="associationRowClass(item.value)"
          >
            <input v-model="selectedRouters" type="checkbox" :value="item.value" />
            <span class="association-checkmark" aria-hidden="true">
              {{ selectedRouters.includes(item.value) ? '✓' : '' }}
            </span>
            <span class="association-main">
              <strong>{{ item.text }}</strong>
              <small>{{ item.namespace }}</small>
            </span>
            <span v-if="associationBadge(item.value)" class="association-selected-badge" :class="'state-' + associationState(item.value)">
              {{ associationBadge(item.value) }}
            </span>
          </label>
        </div>
      </div>
    </div>

    <div v-if="status === 'update_conflict'" class="alert alert-error conflict-alert">
      <span>{{ errorHandle || $t('Gateway.RouterMappingConflict') }}</span>
      <button class="secondary-button" type="button" @click="reloadMapping">
        {{ $t('Form.Reload') }}
      </button>
    </div>

    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th v-for="header in headers" :key="header.value" :class="'text-' + header.align">
              {{ header.text }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in mappings" :key="item.name + ':' + item.namespace">
            <td>
              <button
                class="link-button"
                type="button"
                @click="toUrl('/router/' + item.name + '?name=' + item.name + '&namespace=' + item.namespace)"
              >
                {{ item.name }}
              </button>
            </td>
            <td>{{ item.protocol }}</td>
            <td>{{ item.namespace }}</td>
            <td class="text-center">{{ item.httpCount }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="!mappings.length" class="empty-state compact">
      {{ $t('Gateway.NoRoutersAssociated') }}
    </div>

    <div class="table-footer">{{ $t('Table.Total') }}: {{ mappings.length }}</div>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'GatewayRouterSetting',
  mounted: async function() {
    this.$store.commit('Gateway_ResetStatus');
    this.name = this.$route.query.name || this.$route.params.name || '';
    await this.$store.dispatch('Auth_GetNamespaces');

    await this.fetchMapping();
  },
  data() {
    return {
      name: '',
      namespace: 'All',
      selectedRouters: [],
      initialSelectedRouters: [],
      associationEditMode: false,
    }
  },
  methods: {
    openAssociationEdit() {
      this.associationEditMode = true;
    },
    closeAssociationEdit() {
      this.associationEditMode = false;
      this.selectedRouters = (this.mappings || []).map((item) => this.mappingValue(item)).filter(Boolean);
      this.initialSelectedRouters = [...this.selectedRouters];
    },
    fetchRouter: async function() {
      await this.$store.dispatch('Router_GetItems', {
        namespace: this.namespace || '',
        page: 1,
        limit: -1
      });
    },
    fetchMapping: async function() {
      await this.$store.dispatch('Router_GetItems', {
        namespace: '',
        page: 1,
        limit: -1
      });

      const routers = await this.$store.dispatch('Gateway_GetMappings', {
        name: this.name,
        namespace: this.$route.query.namespace,
        routers: this.routers
      });

      this.selectedRouters = (routers || []).map((item) => this.mappingValue(item)).filter(Boolean);
      this.initialSelectedRouters = [...this.selectedRouters];
    },
    mappingValue(item) {
      if (item?.value) return item.value;
      if (!item?.name || !item?.namespace) return '';
      return item.name + ',' + item.namespace;
    },
    isInitiallySelected(value) {
      return this.initialSelectedRouters.includes(value);
    },
    isSelected(value) {
      return this.selectedRouters.includes(value);
    },
    associationState(value) {
      if (this.isInitiallySelected(value) && this.isSelected(value)) return 'existing';
      if (!this.isInitiallySelected(value) && this.isSelected(value)) return 'added';
      if (this.isInitiallySelected(value) && !this.isSelected(value)) return 'removed';
      return 'none';
    },
    associationBadge(value) {
      const state = this.associationState(value);
      if (state === 'existing') return this.$t('Form.Existing');
      if (state === 'added') return this.$t('Form.Added');
      if (state === 'removed') return this.$t('Form.Removed');
      return '';
    },
    associationRowClass(value) {
      return {
        'association-row--selected': this.isSelected(value),
        'association-row--existing': this.associationState(value) === 'existing',
        'association-row--added': this.associationState(value) === 'added',
        'association-row--removed': this.associationState(value) === 'removed',
      };
    },
    submit() {
      this.$store.commit('Gateway_ResetStatus');
      const routers = this.selectedRouters
        .filter(Boolean)
        .map((item) => {
          const params = item.split(',');
          return {
            name: params[0],
            namespace: params[1]
          };
        });

      this.$store.dispatch('Gateway_MappingRouters', {
        name: this.name,
        namespace: this.$route.query.namespace,
        routers,
        resourceVersions: this.mappingResourceVersions
      });
    },
    reloadMapping: async function() {
      this.$store.commit('Gateway_ResetStatus');
      await this.fetchMapping();
    },
    toUrl(url) {
      window.scrollTo(0, 0);
      if (url) this.$router.push(url);
    },
  },
  watch: {
    status: async function(val) {
      if (val === 'update_success' || val === 'delete_success') {
        this.selectedRouters = [];
        this.associationEditMode = false;
        await this.fetchMapping();
      }
    },
  },
  computed: {
    ...mapGetters({
      namespaces: 'Auth_GetNamespaces',
      status:'Gateway_GetStatus',
      errorHandle:'Gateway_GetErrorHandle',
      routers:'Router_GetItems',
      mappings:'Gateway_GetMappings',
      mappingResourceVersions:'Gateway_GetMappingResourceVersions',
    }),
    headers() {
      return [
        { text: this.$t('Table.Name'), align: 'left', value: 'name' },
        { text: this.$t('Table.Protocol'), align: 'left', value: 'protocol' },
        { text: this.$t('Table.Namespace'), align: 'left', value: 'namespace' },
        { text: this.$t('Table.RuleCount'), align: 'center', value: 'httpCount' },
      ];
    },
    addedRouters() {
      return this.selectedRouters.filter((value) => !this.initialSelectedRouters.includes(value));
    },
    removedRouters() {
      return this.initialSelectedRouters.filter((value) => !this.selectedRouters.includes(value));
    },
  }
}
</script>

<style scoped>
.association-form {
  display: grid;
  gap: 14px;
}

.association-readonly-toolbar {
  align-items: center;
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  display: flex;
  gap: 16px;
  justify-content: space-between;
  padding: 14px 16px;
}

.association-readonly-toolbar div {
  display: grid;
  gap: 4px;
}

.association-readonly-toolbar span {
  color: var(--pw-muted);
  font-size: 0.88rem;
  font-weight: 700;
}

.association-edit-button {
  min-height: 42px;
  white-space: nowrap;
}

.conflict-alert {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.association-toolbar {
  align-items: end;
  display: grid;
  gap: 16px;
  grid-template-columns: minmax(220px, 320px) minmax(0, 1fr) auto auto;
}

.association-namespace {
  min-width: 0;
}

.association-delta-summary {
  align-items: center;
  align-self: stretch;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-start;
  min-width: 0;
}

.association-delta-summary span {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  color: #475569;
  font-size: 0.86rem;
  font-weight: 800;
  padding: 6px 10px;
  white-space: nowrap;
}

.association-submit {
  align-self: center;
  min-height: 42px;
  padding: 0 16px;
  white-space: nowrap;
}

.association-list-wrap {
  min-width: 0;
}

.association-list {
  display: grid;
  gap: 8px;
  max-height: 220px;
  overflow: auto;
}

.association-row {
  align-items: center;
  background: #fff;
  border: 1px solid #d9e2ec;
  border-radius: 8px;
  cursor: pointer;
  display: grid;
  gap: 10px;
  grid-template-columns: auto auto minmax(0, 1fr) auto;
  min-height: 54px;
  padding: 10px 12px;
  transition: background-color 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
}

.association-row:hover {
  border-color: #9fb3c8;
}

.association-row--selected {
  background: #eef6ff;
  border-color: #2f80ed;
  box-shadow: 0 0 0 1px rgba(47, 128, 237, 0.12);
}

.association-row--removed {
  background: #fff7ed;
  border-color: #f59e0b;
}

.association-row--added {
  background: #ecfdf5;
  border-color: #10b981;
}

.association-row input {
  height: 16px;
  width: 16px;
}

.association-checkmark {
  align-items: center;
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  border-radius: 999px;
  color: #1769aa;
  display: inline-flex;
  font-size: 0.8rem;
  font-weight: 800;
  height: 22px;
  justify-content: center;
  width: 22px;
}

.association-row--selected .association-checkmark {
  background: #dbeafe;
  border-color: #2f80ed;
}

.association-main {
  display: grid;
  gap: 3px;
  min-width: 0;
}

.association-main strong,
.association-main small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.association-main strong {
  color: #102a43;
  font-size: 0.95rem;
}

.association-main small {
  color: #64748b;
  font-size: 0.78rem;
}

.association-selected-badge {
  background: #dbeafe;
  border: 1px solid #93c5fd;
  border-radius: 999px;
  color: #1d4ed8;
  font-size: 0.82rem;
  font-weight: 900;
  padding: 6px 10px;
  white-space: nowrap;
}

.association-selected-badge.state-added {
  background: #d1fae5;
  border-color: #6ee7b7;
  color: #047857;
}

.association-selected-badge.state-removed {
  background: #ffedd5;
  border-color: #fdba74;
  color: #c2410c;
}

.association-selected-badge.state-existing {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #475569;
}

@media (max-width: 900px) {
  .association-toolbar {
    grid-template-columns: 1fr;
  }

  .association-readonly-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

}
</style>
