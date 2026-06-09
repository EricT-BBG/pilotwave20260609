<template>
  <section class="nested-card">
    <div class="section-header">
      <strong>{{ $t('Policy.To') }}</strong>
      <button class="danger-button compact-button" type="button" @click="removeTo">{{ $t('Form.Delete') }}</button>
    </div>

    <label v-for="field in visibleFields" :key="field.key" class="field">
      <span>{{ $t(field.label) }}</span>
      <input
        :data-testid="'authpolicy-to-' + field.key"
        :value="formatList($props[field.key])"
        placeholder="value-a, value-b"
        @input="updateTo(field.key, parseList($event.target.value))"
      >
    </label>

    <button class="secondary-button compact-button" type="button" @click="isMore = !isMore">
      {{ isMore ? 'Less' : $t('Policy.MoreSetting') }}
    </button>
  </section>
</template>

<script>
const fields = [
  { key: 'methods', label: 'Policy.Methods' },
  { key: 'paths', label: 'Policy.Paths' },
  { key: 'notMethods', label: 'Policy.NotMethods', advanced: true },
  { key: 'notPaths', label: 'Policy.NotPaths', advanced: true },
  { key: 'hosts', label: 'Policy.Hosts', advanced: true },
  { key: 'notHosts', label: 'Policy.NotHosts', advanced: true },
  { key: 'ports', label: 'Policy.Ports', advanced: true },
  { key: 'notPorts', label: 'Policy.NotPorts', advanced: true },
];

export default {
  name: 'AuthPolicyToItems',
  props: ['hosts', 'notHosts', 'ports', 'notPorts', 'methods', 'notMethods', 'paths', 'notPaths', 'ruleIndex', 'index'],
  data() {
    return {
      isMore: false
    }
  },
  methods: {
    removeTo () {
      this.$store.commit('AuthPolicy_RemoveTo', {
        ruleIndex: this.ruleIndex,
        index: this.index
      })
    },
    updateTo (key, value) {
      this.$store.commit('AuthPolicy_UpdateTo', {
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
