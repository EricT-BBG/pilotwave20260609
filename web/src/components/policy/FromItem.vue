<template>
  <section class="nested-card">
    <div class="section-header">
      <strong>{{ $t('Policy.From') }}</strong>
      <button class="danger-button compact-button" type="button" @click="removeFrom">{{ $t('Form.Delete') }}</button>
    </div>

    <label v-for="field in visibleFields" :key="field.key" class="field">
      <span>{{ $t(field.label) }}</span>
      <input
        :data-testid="'authpolicy-from-' + field.key"
        :value="formatList($props[field.key])"
        placeholder="value-a, value-b"
        @input="updateFrom(field.key, parseList($event.target.value))"
      >
    </label>

    <button class="secondary-button compact-button" type="button" @click="isMore = !isMore">
      {{ isMore ? 'Less' : $t('Policy.MoreSetting') }}
    </button>
  </section>
</template>

<script>
const fields = [
  { key: 'ipBlocks', label: 'Policy.IPBlocks' },
  { key: 'requestPrincipals', label: 'Policy.RequestPrincipals' },
  { key: 'notIpBlocks', label: 'Policy.NotIPBlocks', advanced: true },
  { key: 'notRequestPrincipals', label: 'Policy.NotRequestPrincipals', advanced: true },
  { key: 'principals', label: 'Policy.Principals', advanced: true },
  { key: 'notPrincipals', label: 'Policy.NotPrincipals', advanced: true },
  { key: 'namespaces', label: 'Policy.Namespaces', advanced: true },
  { key: 'notNamespaces', label: 'Policy.NotNamespaces', advanced: true },
  { key: 'remoteIpBlocks', label: 'Policy.RemoteIPBlocks', advanced: true },
  { key: 'notRemoteIpBlocks', label: 'Policy.NotRemoteIPBlocks', advanced: true },
];

export default {
  name: 'AuthPolicyFromItems',
  props: ['principals', 'notPrincipals', 'requestPrincipals', 'notRequestPrincipals', 'namespaces', 'notNamespaces', 'ipBlocks', 'notIpBlocks', 'remoteIpBlocks', 'notRemoteIpBlocks', 'ruleIndex', 'index'],
  data() {
    return {
      isMore: false
    }
  },
  methods: {
    removeFrom () {
      this.$store.commit('AuthPolicy_RemoveFrom', {
        ruleIndex: this.ruleIndex,
        index: this.index
      })
    },
    updateFrom (key, value) {
      this.$store.commit('AuthPolicy_UpdateFrom', {
        ruleIndex: this.ruleIndex,
        index: this.index,
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
    visibleFields() {
      return fields.filter(field => this.isMore || !field.advanced);
    },
  },
}
</script>

<style scoped>
.nested-card {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 16px;
  display: grid;
  gap: 12px;
  margin-bottom: 12px;
  padding: 14px;
}

.section-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.compact-button {
  min-height: 36px;
}
</style>
