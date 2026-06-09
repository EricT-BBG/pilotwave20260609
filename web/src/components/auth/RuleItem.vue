<template>
  <section class="rule-card">
    <div class="editor-row">
      <label class="field">
        <span>{{ $t('Auth.Issuer') }}*</span>
        <input
          data-testid="requestauth-rule-issuer"
          :value="rule.issuer"
          placeholder="admin@secure.istio.io"
          required
          @input="updateRule('issuer', $event.target.value)"
        >
      </label>
      <label class="field">
        <span>{{ $t('Auth.JWKURI') }}</span>
        <input
          data-testid="requestauth-rule-jwks-uri"
          :value="rule.jwksUri"
          placeholder="https://raw.githubusercontent.com/istio/istio/release-1.7/security/tools/jwt/samples/jwks.json"
          @input="updateRule('jwksUri', $event.target.value)"
        >
      </label>
      <button class="danger-button compact-button" type="button" @click="removeRule">{{ $t('Form.Delete') }}</button>
    </div>
    <label class="field">
      <span>{{ $t('Auth.AUD') }}</span>
      <input
        data-testid="requestauth-rule-audiences"
        :value="formatList(rule.audiences)"
        placeholder="audience-a, audience-b"
        @input="updateRule('audiences', parseList($event.target.value))"
      >
    </label>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'AuthRequestRouterItems',
  props: ['rule', 'ruleIndex'],
  methods: {
    removeRule () {
      if (this.rules.length === 1) return;

      this.$store.commit('AuthRequest_RemoveRule', {
        ruleIndex: this.ruleIndex
      })
    },
    updateRule (key, value) {
      this.$store.commit('AuthRequest_UpdateRule', {
        ruleIndex: this.ruleIndex,
        key: key,
        value: value
      })
    },
    formatList(value) {
      return Array.isArray(value) ? value.join(', ') : '';
    },
    parseList(value) {
      return value.split(',').map(item => item.trim()).filter(Boolean);
    },
  },
  computed: {
    ...mapGetters({
      rules: 'AuthRequest_GetJwtRules'
    }),
  },
}
</script>

<style scoped>
.rule-card {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 18px;
  display: grid;
  gap: 12px;
  margin-bottom: 14px;
  padding: 16px;
}

.editor-row {
  align-items: end;
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr) auto;
}

.compact-button {
  min-height: 42px;
}

@media (max-width: 760px) {
  .editor-row {
    grid-template-columns: 1fr;
  }
}
</style>
