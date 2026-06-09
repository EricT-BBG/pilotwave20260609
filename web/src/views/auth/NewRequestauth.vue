<template>
  <section class="page">
    <header class="page-header detail-header detail-header--compact">
      <div class="detail-header-row">
        <button class="secondary-button detail-back-button" type="button" @click="goBack">← {{ $t('Form.BackToAPIAuthentication') }}</button>
        <div class="detail-title-card">
          <div class="detail-title-text">
            <p class="eyebrow detail-resource-type">{{ $t('System.APIAuthentication') }}</p>
            <h1 class="detail-header-title">{{ $t('Auth.New') }}</h1>
          </div>
        </div>
      </div>
    </header>

    <form class="panel form-stack" @submit.prevent="submit">
      <div class="tab-strip">
        <button class="tab-button active" type="button">{{ $t('Auth.BasicSetting') }}</button>
      </div>

      <label class="field">
        <span>{{ $t('Auth.RuleName') }}*</span>
        <input data-testid="requestauth-name" v-model.trim="name" placeholder="jwt-auth" required @blur="touched.name = true">
        <small v-for="error in nameErrors" :key="error" class="field-error">{{ error }}</small>
      </label>

      <label class="field">
        <span>{{ $t('Table.Namespace') }}*</span>
        <select data-testid="requestauth-namespace" v-model="namespace" required @blur="touched.namespace = true">
          <option v-for="item in namespaceOptions" :key="item" :value="item">{{ item }}</option>
        </select>
        <small v-for="error in spaceErrors" :key="error" class="field-error">{{ error }}</small>
      </label>

      <section class="editor-section">
        <div class="section-header">
          <h2>{{ $t('Auth.NewLabel') }}</h2>
          <button class="secondary-button" type="button" @click="addLabel">+ {{ $t('Auth.NewLabel') }}</button>
        </div>
        <LabelItem
          v-for="(item, i) in labels"
          :key="'label-' + i"
          :index="i"
          :labelKey="item.key"
          :labelValue="item.value"
        />
      </section>

      <section class="editor-section">
        <div class="section-header">
          <h2>{{ $t('Auth.SetupRule') }}</h2>
          <button class="secondary-button" type="button" @click="addRule">+ {{ $t('Auth.SetupRule') }}</button>
        </div>
        <RuleItem v-for="(item, index) in rules" :key="'rule-' + index" :rule="item" :ruleIndex="index" />
      </section>

      <div v-if="status === 'create_error'" class="alert alert-error">
        {{ $t('Alert.CreateFailed') }} {{ error_handle }}
      </div>

      <div class="page-actions form-actions">
        <button class="secondary-button" type="button" @click="goBack">{{ $t('Form.Cancel') }}</button>
        <button class="primary-button" data-testid="requestauth-submit" type="submit">{{ $t('Form.Submit') }}</button>
      </div>
    </form>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';
import LabelItem from '../../components/auth/LabelItem.vue';
import RuleItem from '../../components/auth/RuleItem.vue';

const namePattern = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;

export default {
  name: 'NewRequestauth',
  components: {
    LabelItem,
    RuleItem,
  },
  mounted() {
    this.$store.commit('AuthRequest_ResetData');
  },
  data() {
    return {
      name: '',
      namespace: 'default',
      submitted: false,
      touched: {
        name: false,
        namespace: false,
      },
    }
  },
  methods: {
    goBack () {
      window.scrollTo(0,0);
      this.$router.push('/requestauths');
    },
    addRule () {
      this.$store.commit('AuthRequest_AddRules')
    },
    addLabel () {
      this.$store.commit('AuthRequest_AddLabels')
    },
    submit () {
      this.submitted = true;
      if (this.hasValidationErrors) return;

      this.$store.dispatch('AuthRequest_NewItem', {
        name: this.name,
        namespace: this.namespace,
        labels: this.labels,
        rules: this.rules,
      })
    },
  },
  watch: {
    status (val) {
      if (val === 'create_success') this.$router.push('/requestauths')
    }
  },
  computed: {
    ...mapGetters({
      namespaces: 'Auth_GetNamespaces',
      language: 'Auth_GetLanguage',
      rules: 'AuthRequest_GetJwtRules',
      labels: 'AuthRequest_GetLabels',
      status: 'AuthRequest_GetStatus',
      error_handle: 'AuthRequest_GetErrorHandle'
    }),
    shouldShowNameErrors () {
      return this.submitted || this.touched.name;
    },
    shouldShowNamespaceErrors () {
      return this.submitted || this.touched.namespace;
    },
    hasValidationErrors () {
      return !this.name || !namePattern.test(this.name) || !this.namespace;
    },
    namespaceOptions() {
      const namespaces = (this.namespaces || []).filter((item) => item && item !== 'All');
      return namespaces.length ? namespaces : ['default'];
    },
    nameErrors () {
      if (!this.shouldShowNameErrors) return [];
      if (!this.name) return [this.$t('Form.Required')];
      if (!namePattern.test(this.name)) return [this.$t('Form.Lowercase')];
      return [];
    },
    spaceErrors () {
      if (!this.shouldShowNamespaceErrors || this.namespace) return [];
      return [this.$t('Form.Required')];
    },
  },
}
</script>

<style scoped>
.tab-strip {
  border-bottom: 1px solid var(--pw-border);
}

.tab-button {
  background: var(--pw-primary);
  border: 0;
  border-radius: 14px 14px 0 0;
  color: #fff;
  font-weight: 800;
  min-height: 42px;
  padding: 0 18px;
}

.editor-section {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 18px;
  display: grid;
  gap: 12px;
  padding: 16px;
}

.section-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.section-header h2 {
  font-size: 1rem;
  margin: 0;
}

.form-actions {
  border-top: 1px solid var(--pw-border);
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
