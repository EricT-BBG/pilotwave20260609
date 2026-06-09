<template>
  <section class="page">
    <header class="page-header detail-header detail-header--compact">
      <div class="detail-header-row">
        <button class="secondary-button detail-back-button" type="button" @click="goBack">
          ← {{ $t('Form.BackToPolicies') }}
        </button>
        <div class="detail-title-card">
          <div class="detail-title-text">
            <p class="eyebrow detail-resource-type">{{ $t('System.BlackWhteList') }}</p>
            <h1 class="detail-header-title">{{ name }}</h1>
          </div>
        </div>
        <span class="namespace-chip">
          <span class="namespace-chip-label">{{ $t('Table.Namespace') }}</span>
          <span class="namespace-chip-value">{{ namespace || '-' }}</span>
        </span>
      </div>
    </header>

    <div class="panel">
      <div v-if="!editMode" class="detail-tab-toolbar">
        <div class="tab-strip">
          <button class="tab-button active" type="button">{{ $t('Policy.BasicSetting') }}</button>
        </div>
        <button class="primary-button detail-edit-button" data-testid="authpolicy-edit-open" type="button" @click="openEditMode">
          {{ $t('Form.Edit') }}
        </button>
      </div>
      <BasicSetting v-if="editMode" />
      <div v-else class="readonly-summary">
        <div class="readonly-summary-grid">
          <article>
            <span>{{ $t('Table.Namespace') }}</span>
            <strong>{{ namespace || '-' }}</strong>
          </article>
          <article>
            <span>{{ $t('Policy.ListType') }}</span>
            <strong>{{ policyAction }}</strong>
          </article>
          <article>
            <span>{{ $t('Table.Label') }}</span>
            <strong>{{ labels.length }}</strong>
          </article>
          <article>
            <span>{{ $t('Table.RuleCount') }}</span>
            <strong>{{ rules.length }}</strong>
          </article>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';
import BasicSetting from '../../components/policy/EditSetting.vue';

export default {
  name: 'AuthPolicyDetail',
  components: {
    BasicSetting,
  },
  mounted: async function() {
    this.$store.commit('AuthPolicy_ResetStatus');
    this.name = this.$route.query.name || '';
    this.namespace = this.$route.query.namespace || '';
    await this.fetchData();
  },
  data () {
    return {
      name: '',
      namespace: '',
      editMode: false
    }
  },
  methods: {
    fetchData: async function() {
      if (!this.name || !this.namespace) return;
      await this.$store.dispatch('AuthPolicy_GetItem', {
        name: this.name,
        namespace: this.namespace
      });
    },
    openEditMode() {
      this.$store.commit('AuthPolicy_ResetStatus');
      this.editMode = true;
    },
    goBack () {
      window.scrollTo(0,0);
      this.$router.push('/authpolicies');
    },
  },
  computed: {
    ...mapGetters({
      language: 'Auth_GetLanguage',
      status: 'AuthPolicy_GetStatus',
      policy: 'AuthPolicy_GetItem',
      rules: 'AuthPolicy_GetRules',
      labels: 'AuthPolicy_GetLabels',
    }),
    policyAction() {
      return this.policy?.action || '-';
    },
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

.tab-button {
  background: transparent;
  border: 0;
  border-bottom: 3px solid var(--pw-accent);
  color: var(--pw-primary-strong);
  font-weight: 800;
  min-height: 42px;
  padding: 0 4px;
}

.detail-tab-toolbar {
  align-items: center;
  border-bottom: 1px solid var(--pw-border);
  display: flex;
  gap: 16px;
  justify-content: space-between;
  margin-bottom: 18px;
}

.detail-tab-toolbar .tab-strip {
  border-bottom: 0;
}

.detail-edit-button {
  min-height: 44px;
  min-width: 92px;
  white-space: nowrap;
}

.readonly-summary {
  display: grid;
  gap: 18px;
}

.readonly-summary-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.readonly-summary-grid article {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 12px;
  display: grid;
  gap: 8px;
  padding: 16px;
}

.readonly-summary-grid span {
  color: var(--pw-muted);
  font-size: 0.9rem;
  font-weight: 700;
}

.readonly-summary-grid strong {
  color: var(--pw-primary-strong);
  font-size: 1.2rem;
  text-transform: uppercase;
}

@media (max-width: 900px) {
  .readonly-summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .readonly-summary-grid {
    grid-template-columns: 1fr;
  }

  .detail-tab-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .detail-edit-button {
    width: 100%;
  }
}
</style>
