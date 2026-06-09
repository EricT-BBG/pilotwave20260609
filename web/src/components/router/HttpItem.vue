<template>
  <article class="panel router-http-card" :class="{ 'is-readonly': readonly }">
    <div v-if="!readonly" class="router-http-actions">
      <button
        type="button"
        class="danger-button"
        :disabled="httpItems.length === 1"
        @click="removeHttp()"
      >
        {{ $t('Router.RemoveHttpRoute') }}
      </button>
    </div>

    <section class="router-editor-section">
      <header class="router-editor-toolbar">
        <h2>{{ $t('Router.SetupPrefix') }}</h2>
        <button v-if="!readonly" type="button" class="icon-button" @click="addPrefix" :aria-label="$t('Router.AddPrefix')">
          +
        </button>
      </header>

      <div class="router-editor-stack">
        <PrefixItem
          v-for="(item, i) in http.prefixs"
          :key="'prefix-' + i"
          :index="i"
          :prefix="item"
          :httpIndex="index"
          :readonly="readonly"
        />
      </div>
    </section>

    <section v-if="router.protocol == 'http' || router.protocol == 'https'" class="router-editor-section">
      <button v-if="!readonly" type="button" class="secondary-button" @click="addHeader">
        {{ $t('Router.AddHeaderCheck') }}
      </button>
      <div class="router-editor-stack">
        <HeaderItem
          v-for="(item, i) in http.headers"
          :key="'header-' + i"
          :index="i"
          :httpIndex="index"
          :headerKey="item.key"
          :headerValue="item.value"
          :readonly="readonly"
        />
      </div>
    </section>

    <div class="router-editor-divider"></div>

    <section class="router-editor-grid">
      <label class="field full-span">
        <span>{{ $t('Router.Rewrite') }}</span>
        <input
          :value="http.rewrite"
          placeholder="/newcatalog"
          :disabled="readonly"
          @input="updateHttp('rewrite', $event.target.value)"
        >
      </label>

      <label class="field">
        <span>{{ $t('Router.FixedDelay') }} (Sec)</span>
        <input
          :value="http.fixedDelay"
          placeholder="7"
          :disabled="readonly"
          @input="updateHttp('fixedDelay', $event.target.value)"
        >
      </label>

      <label class="field">
        <span>{{ $t('Router.Timeout') }} (Sec)</span>
        <input
          :value="http.timeout"
          placeholder="60"
          type="number"
          :disabled="readonly"
          @input="updateHttp('timeout', $event.target.value)"
        >
      </label>
    </section>

    <section class="router-editor-section">
      <header class="router-editor-toolbar">
        <h2>{{ $t('Router.Destinations') }}</h2>
        <button v-if="!readonly" type="button" class="icon-button" @click="addDestination" :aria-label="$t('Router.AddDestination')">
          +
        </button>
      </header>

      <div class="router-editor-stack">
        <Destination
          v-for="(item, i) in http.destinations"
          :key="'destination-' + i"
          :index="i"
          :httpIndex="index"
          :host="item.host"
          :port="item.port"
          :weight="item.weight"
          :subset="item.subset"
          :readonly="readonly"
        />
      </div>
    </section>
  </article>
</template>

<script>
import { mapGetters } from 'vuex';
import Destination from './Destination.vue';
import HeaderItem from './HeaderItem.vue';
import PrefixItem from './PrefixItem.vue';

export default {
  name: 'RouterHttpItem',
  components: {
    PrefixItem,
    HeaderItem,
    Destination
  },
  mounted: async function() {  

  },
  props: {
    http: {
      type: Object,
      required: true,
    },
    index: {
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

    }
  },
  methods: {
    removeHttp () {
      if (this.readonly) return;
      if (this.httpItems.length === 1) return;

      this.$store.commit('Router_RemoveHttp', {
        httpIndex: this.index,
      })
    },
    addPrefix () {
      if (this.readonly) return;
      this.$store.commit('Router_AddPrefix', {
        httpIndex: this.index,
      })
    },
    addHeader () {
      if (this.readonly) return;
      this.$store.commit('Router_AddHeader', {
        httpIndex: this.index,
      })
    },
    updateHttp (key, value) {
      if (this.readonly) return;
      this.$store.commit('Router_UpdateHttp', {
        httpIndex: this.index,
        key: key,
        value: value
      })
    },
    addDestination () {
      if (this.readonly) return;
      this.$store.commit('Router_AddDestination', {
        httpIndex: this.index,
      })
    },
  },
  computed: {
    ...mapGetters({
      language: 'Auth_GetLanguage',
      router: 'Router_GetItem',
      httpItems: 'Router_GetHttpItems'
    }),
  }
}
</script>

<style scoped>
.router-http-card {
  background: #f8f6f1;
}

.router-http-card.is-readonly {
  background: #fff;
}

.router-http-actions {
  display: flex;
  justify-content: flex-end;
}

.router-http-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.router-editor-section,
.router-editor-stack {
  display: grid;
  gap: 12px;
}

.router-editor-toolbar {
  align-items: center;
  background: var(--pw-primary);
  border-radius: 16px;
  color: #fff;
  display: flex;
  justify-content: space-between;
  padding: 10px 14px;
}

.router-editor-toolbar h2 {
  font-size: 1rem;
  margin: 0;
}

.router-editor-divider {
  border-top: 1px solid var(--pw-border);
}

.router-editor-grid {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.full-span {
  grid-column: 1 / -1;
}

@media (max-width: 760px) {
  .router-editor-grid {
    grid-template-columns: 1fr;
  }
}
</style>
