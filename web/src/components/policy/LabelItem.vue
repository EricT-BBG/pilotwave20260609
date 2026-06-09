<template>
  <div class="editor-row">
    <label class="field">
      <span>{{ $t('Policy.FieldName') }}</span>
      <input
        :value="labelKey"
        placeholder="app"
        @input="updateLabel('key', $event.target.value)"
      >
    </label>
    <label class="field">
      <span>{{ $t('Policy.FieldValue') }}</span>
      <input
        :value="labelValue"
        placeholder="httpbin"
        @input="updateLabel('value', $event.target.value)"
      >
    </label>
    <button class="danger-button compact-button" type="button" @click="removeLabel">
      Delete
    </button>
  </div>
</template>

<script>
export default {
  name: 'LabelItems',
  props: ['labelKey', 'labelValue', 'index'],
  methods: {
    removeLabel () {
      this.$store.commit('AuthPolicy_RemoveLabel', {
        index: this.index,
      })
    },
    updateLabel (key, value) {
      this.$store.commit('AuthPolicy_UpdateLabel', {
        index: this.index,
        key: key,
        value: value
      })
    },
  },
}
</script>

<style scoped>
.editor-row {
  align-items: end;
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
  margin-bottom: 12px;
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
