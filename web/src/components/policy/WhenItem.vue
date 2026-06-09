<template>
  <section class="nested-card">
    <div class="editor-row">
      <label class="field">
        <span>{{ $t('Policy.Key') }}</span>
        <input
          :value="wKey"
          placeholder="testing@secure.istio.io"
          @input="updateWhen('key', $event.target.value)"
        >
      </label>
      <button class="danger-button compact-button" type="button" @click="removeWhen">
        Delete
      </button>
    </div>
    <label class="field">
      <span>{{ $t('Policy.Values') }}</span>
      <input
        :value="formatList(wValues)"
        placeholder="value-a, value-b"
        @input="updateWhen('values', parseList($event.target.value))"
      >
    </label>
    <label class="field">
      <span>{{ $t('Policy.NotValues') }}</span>
      <input
        :value="formatList(notValues)"
        placeholder="value-a, value-b"
        @input="updateWhen('notValues', parseList($event.target.value))"
      >
    </label>
  </section>
</template>

<script>
export default {
  name: 'AuthPolicyWhenItems',
  props: ['wKey', 'wValues', 'notValues', 'ruleIndex', 'index'],
  methods: {
    removeWhen () {
      this.$store.commit('AuthPolicy_RemoveWhen', {
        ruleIndex: this.ruleIndex,
        index: this.index
      })
    },
    updateWhen (key, value) {
      this.$store.commit('AuthPolicy_UpdateWhen', {
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

.editor-row {
  align-items: end;
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1fr) auto;
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
