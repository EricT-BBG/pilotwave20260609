<template>
  <div class="editor-row">
    <label class="field">
      <span>{{ $t('Auth.FieldName') }}</span>
      <input
        :value="labelKey"
        placeholder="app"
        @input="updateLabel('key', $event.target.value)"
      >
    </label>
    <label class="field">
      <span>{{ $t('Auth.FieldValue') }}</span>
      <input
        :value="labelValue"
        placeholder="httpbin"
        @input="updateLabel('value', $event.target.value)"
      >
    </label>
    <button class="danger-button compact-button" type="button" @click="removeLabel">{{ $t('Form.Delete') }}</button>
  </div>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'LabelItems',
  props: ['labelKey', 'labelValue', 'index'],
  methods: {
    removeLabel () {
      if (this.labels.length === 1) return;

      this.$store.commit('AuthRequest_RemoveLabel', {
        index: this.index,
      })
    },
    updateLabel (key, value) {
      this.$store.commit('AuthRequest_UpdateLabel', {
        index: this.index,
        key: key,
        value: value
      })
    },
  },
  computed: {
    ...mapGetters({
      labels: 'AuthRequest_GetLabels',
    }),
  }
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
