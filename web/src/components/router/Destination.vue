<template>
  <article class="router-destination-card">
    <label class="field host-field">
      <span>{{ $t('Router.Host') }}*</span>
      <input
        :value="host"
        placeholder="userapi.prod.svc.cluster"
        :disabled="readonly"
        @input="updateDestination('host', $event.target.value)"
      >
    </label>

    <label class="field">
      <span>{{ $t('Router.Port') }}*</span>
      <input
        :value="port"
        placeholder="80"
        type="number"
        :disabled="readonly"
        @input="updateDestination('port', $event.target.value)"
      >
    </label>

    <label class="field">
      <span>{{ $t('Router.Weight') }}*</span>
      <input
        :value="weight"
        placeholder="90"
        type="number"
        :disabled="readonly"
        @input="updateDestination('weight', $event.target.value)"
      >
    </label>

    <label class="field">
      <span>{{ $t('Router.Subset') }}</span>
      <input
        :value="subset"
        placeholder="v2"
        :disabled="readonly"
        @input="updateDestination('subset', $event.target.value)"
      >
    </label>

    <button
      v-if="!readonly"
      type="button"
      class="danger-button compact-button"
      :disabled="httpItems[httpIndex].destinations.length === 1"
      :aria-label="$t('Router.RemoveDestination')"
      @click="removeDestination()"
    >
      ×
    </button>
  </article>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'Destination',
  props: {
    host: {
      type: String,
      default: '',
    },
    port: {
      type: [String, Number],
      default: '',
    },
    weight: {
      type: [String, Number],
      default: '',
    },
    subset: {
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
    removeDestination () {
      if (this.readonly) return;
      if (this.httpItems[this.httpIndex].destinations.length === 1) return;

      this.$store.commit('Router_RemoveDestination', {
        httpIndex: this.httpIndex,
        index: this.index,
      })
    },
    updateDestination (key, value) {
      if (this.readonly) return;
      this.$store.commit('Router_UpdateDestination', {
        httpIndex: this.httpIndex,
        index: this.index,
        key: key,
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
.router-destination-card {
  align-items: end;
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 18px;
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(240px, 2fr) repeat(3, minmax(90px, 1fr)) auto;
  padding: 14px;
}

.compact-button {
  min-height: 42px;
  padding: 0 12px;
}

.compact-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

@media (max-width: 980px) {
  .router-destination-card {
    grid-template-columns: 1fr 1fr;
  }

  .host-field {
    grid-column: 1 / -1;
  }
}

@media (max-width: 640px) {
  .router-destination-card {
    grid-template-columns: 1fr;
  }
}
</style>
