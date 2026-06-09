<template>
  <div class="router-prefix-row">
    <label class="field">
      <span>{{ $t('Router.Prefix') }}</span>
      <input
        placeholder="/wpcatalog"
        :value="prefix"
        :disabled="readonly"
        @input="updatePrefix($event.target.value)"
      >
    </label>

    <button
      v-if="!readonly"
      type="button"
      class="danger-button compact-button"
      :disabled="httpItems[httpIndex].prefixs.length === 1"
      :aria-label="$t('Router.RemovePrefix')"
      @click="removePrefix()"
    >
      ×
    </button>
  </div>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'PrefixItems',
  mounted() {

  },
  props: {
    prefix: {
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
      },
    }
  },
  methods: {
    removePrefix () {
      if (this.readonly) return;
      if (this.httpItems[this.httpIndex].prefixs.length === 1) return;

      this.$store.commit('Router_RemovePrefix', {
        httpIndex: this.httpIndex,
        index: this.index,
      })
    },
    updatePrefix (value) {
      if (this.readonly) return;
      this.$store.commit('Router_UpdatePrefix', {
        httpIndex: this.httpIndex,
        index: this.index,
        value: value
      })
    },
  },
  computed: {
    ...mapGetters({
      httpItems: 'Router_GetHttpItems'
    }),
  }
}
</script>

<style scoped>
.router-prefix-row {
  align-items: end;
  display: grid;
  gap: 12px;
  grid-template-columns: 1fr auto;
}

.compact-button {
  min-height: 42px;
  padding: 0 12px;
}

.compact-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

@media (max-width: 760px) {
  .router-prefix-row {
    grid-template-columns: 1fr;
  }
}
</style>
