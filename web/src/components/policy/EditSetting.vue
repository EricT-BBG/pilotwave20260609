<template>
  <form class="form-stack" @submit.prevent="submit">
    <label class="field">
      <span>{{ $t('Policy.RuleName') }}*</span>
      <input v-model.trim="name" placeholder="allow-public-api" disabled @blur="touched.name = true">
      <small v-for="error in nameErrors" :key="error" class="field-error">{{ error }}</small>
    </label>

    <label class="field">
      <span>{{ $t('Table.Namespace') }}*</span>
      <select v-model="namespace" disabled @blur="touched.namespace = true">
        <option v-for="item in namespaces" :key="item" :value="item">{{ item }}</option>
      </select>
      <small v-for="error in spaceErrors" :key="error" class="field-error">{{ error }}</small>
    </label>

    <label class="field">
      <span>{{ $t('Policy.ListType') }}*</span>
      <select data-testid="authpolicy-edit-action" v-model="action" required @blur="touched.action = true">
        <option v-for="item in actions" :key="item.value" :value="item.value">{{ item.text }}</option>
      </select>
      <small v-for="error in actionErrors" :key="error" class="field-error">{{ error }}</small>
    </label>

    <section class="editor-section">
      <div class="section-header">
        <h2>{{ $t('Policy.NewLabel') }}</h2>
        <button class="secondary-button" type="button" @click="addLabel">+ {{ $t('Policy.NewLabel') }}</button>
      </div>
      <LabelItem
        v-for="(item, i) in labels"
        :key="'label-' + i"
        :index="i"
        :labelKey="item.key"
        :labelValue="item.value"
      />
      <div v-if="!labels.length" class="empty-state compact">{{ $t('Policy.NoLabelsConfigured') }}</div>
    </section>

    <section class="editor-section">
      <div class="section-header">
        <h2>{{ $t('Policy.SetupRule') }}</h2>
        <button class="secondary-button" type="button" @click="addRule">+ {{ $t('Policy.SetupRule') }}</button>
      </div>
      <RuleItem v-for="(item, index) in rules" :key="'rule-' + index" :rule="item" :ruleIndex="index" />
    </section>

    <div v-if="status === 'update_success'" class="alert alert-success">
      {{ $t('Alert.Updated') }}
    </div>
    <div v-if="status === 'update_error'" class="alert alert-error">
      {{ $t('Alert.UpdateFailed') }} {{ error_handle }}
    </div>

    <div class="page-actions form-actions">
      <button class="secondary-button" type="button" @click="goBack">{{ $t('Form.Cancel') }}</button>
      <button class="primary-button" data-testid="authpolicy-update-submit" type="submit">{{ $t('Form.Submit') }}</button>
    </div>
  </form>
</template>

<script>
import { mapGetters } from 'vuex';
import LabelItem from './LabelItem.vue';
import RuleItem from './RuleItem.vue';

const namePattern = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;

export default {
  name: 'EditAuthpolicy',
  components: {
    LabelItem,
    RuleItem,
  },
  mounted: async function() {
    this.$store.commit('AuthPolicy_ResetData');
    this.name = this.$route.query.name || '';
    this.namespace = this.$route.query.namespace || 'default';

    await this.fetchData();
  },
  data() {
    return {
      name: '',
      namespace: 'default',
      action: 'allow',
      submitted: false,
      touched: {
        name: false,
        namespace: false,
        action: false,
      },
    }
  },
  methods: {
    fetchData: async function() {
      const data = await this.$store.dispatch('AuthPolicy_GetItem', {
        name: this.name,
        namespace: this.namespace
      });

      if (!data) return;
      if (data.action) this.action = data.action.toLowerCase();
    },
    goBack () {
      window.scrollTo(0,0);
      this.$router.go(-1);
    },
    addRule () {
      this.$store.commit('AuthPolicy_AddRules')
    },
    addLabel () {
      this.$store.commit('AuthPolicy_AddLabels')
    },
    submit () {
      this.submitted = true;
      if (this.hasValidationErrors) return;

      this.$store.dispatch('AuthPolicy_UpdateItem', {
        name: this.name,
        namespace: this.namespace,
        action: this.action,
        labels: this.labels,
        rules: this.rules,
        resourceVersion: this.resourceVersion,
      })
    },
  },
  computed: {
    ...mapGetters({
      namespaces: 'Auth_GetNamespaces',
      language: 'Auth_GetLanguage',
      rules: 'AuthPolicy_GetRules',
      labels: 'AuthPolicy_GetLabels',
      resourceVersion: 'AuthPolicy_GetResourceVersion',
      status: 'AuthPolicy_GetStatus',
      error_handle: 'AuthPolicy_GetErrorHandle'
    }),
    actions() {
      return [
        { text: this.$t('Policy.Allow'), value: 'allow' },
        { text: this.$t('Policy.Deny'), value: 'deny' },
        { text: this.$t('Policy.Audit'), value: 'audit' },
      ];
    },
    hasValidationErrors () {
      return !this.name || !namePattern.test(this.name) || !this.namespace || !this.action;
    },
    nameErrors () {
      if (!this.submitted && !this.touched.name) return [];
      if (!this.name) return [this.$t('Form.Required')];
      if (!namePattern.test(this.name)) return [this.$t('Form.Valid')];
      return [];
    },
    spaceErrors () {
      if ((!this.submitted && !this.touched.namespace) || this.namespace) return [];
      return [this.$t('Form.Required')];
    },
    actionErrors () {
      if ((!this.submitted && !this.touched.action) || this.action) return [];
      return [this.$t('Form.Required')];
    },
  },
}
</script>

<style scoped>
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

.alert-success {
  background: #dcfce7;
  color: #166534;
}

.form-actions {
  border-top: 1px solid var(--pw-border);
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
