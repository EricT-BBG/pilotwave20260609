<template>
  <div class="router-header-row">
    <label class="field">
      <span>{{ $t('Router.FieldName') }}</span>
      <input
        placeholder="end-user"
        :value="headerKey"
        :disabled="readonly"
        @input="updateHeader('key', $event.target.value)"
      >
    </label>

    <label class="field">
      <span>{{ $t('Router.FieldValue') }}</span>
      <input
        placeholder="json"
        :value="headerValue"
        :disabled="readonly"
        @input="updateHeader('value', $event.target.value)"
      >
    </label>

    <button v-if="!readonly" type="button" class="danger-button compact-button" :aria-label="$t('Router.RemoveHeader')" @click="removeHeader()">
      ×
    </button>
  </div>
</template>

<script>

export default {
  name: 'HeaderItems',
  props: {
    headerKey: {
      type: String,
      default: '',
    },
    headerValue: {
      type: String,
      default: '',
    },
    index: {
      type: Number,
      required: true,
    },
    httpIndex: {
      type: Number,
      required: true,
    },
    readonly: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      rules: {
        required: value => !!value || 'Required.',
      }
    }
  },
  methods: {
    removeHeader () {
      if (this.readonly) return;
      this.$store.commit('Router_RemoveHeader', {
        httpIndex: this.httpIndex,
        index: this.index,
      })
    },
    updateHeader (key, value) {
      if (this.readonly) return;
      this.$store.commit('Router_UpdateHeader', {
        httpIndex: this.httpIndex,
        index: this.index,
        key: key,
        value: value
      })
    },
  }
}
</script>

<style scoped>
.router-header-row {
  align-items: end;
  display: grid;
  gap: 12px;
  grid-template-columns: 1fr 1fr auto;
}

.compact-button {
  min-height: 42px;
  padding: 0 12px;
}

@media (max-width: 760px) {
  .router-header-row {
    grid-template-columns: 1fr;
  }
}
</style>
