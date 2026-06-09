<template>
  <section class="page router-setting-page" :class="{ 'is-readonly': readonly }">
    <header class="page-header router-setting-header">
      <div>
        <p class="eyebrow">{{ $t('Table.Protocol') }}</p>
        <h1 class="protocol-title">{{ router.protocol }}</h1>
      </div>
      <button v-if="!readonly" type="button" class="primary-button" @click="addSetting">
        {{ $t('Router.NewHttpRoute') }}
      </button>
    </header>

    <div class="router-route-list">
      <HttpItem
        v-for="(item, index) in httpItems"
        :key="'http-' + index"
        :http="item"
        :index="index"
        :readonly="readonly"
      />
    </div>

    <footer v-if="!readonly" class="panel router-setting-footer">
      <div v-show="status == 'update_success'" class="alert alert-success">
        {{ $t('Alert.Updated') }}
      </div>
      <div v-show="status == 'update_error'" class="alert alert-error">
        {{ $t('Alert.UpdateFailed') }} {{ errorHandle }}
      </div>
      <div v-show="status == 'update_conflict'" class="alert alert-error conflict-alert">
        <span>{{ errorHandle || $t('Router.RuleMappingConflict') }}</span>
        <button type="button" class="secondary-button" @click="reloadData">
          {{ $t('Form.Reload') }}
        </button>
      </div>
      <div class="page-actions router-setting-actions">
        <button type="button" class="secondary-button" @click="cancelEdit">
          {{ $t('Form.Cancel') }}
        </button>
        <button type="button" class="primary-button" @click="submit">
          {{ $t('Form.Submit') }}
        </button>
      </div>
    </footer>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';
import HttpItem from './HttpItem.vue';

export default {
  name: 'RouterSetting',
  components: {
    HttpItem,
  },
  props: {
    readonly: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['close-edit'],
  mounted: async function() {  
    this.$store.commit('Router_ResetStatus');
    this.name = this.$route.query.name;
    this.namespace = this.$route.query.namespace;

    await this.fetchRouter();
    await this.fetchData();
  },
  data() {
    return {
      name: '',
      namespace: '',
    }
  },
  methods: {
    fetchRouter: async function() {
      await this.$store.dispatch('Router_GetItem', {
        name: this.name,
        namespace: this.namespace
      });
    },
    fetchData: async function() {
      await this.$store.dispatch('Router_GetRule', {
        name: this.name,
        namespace: this.namespace
      });
    },
    reloadData: async function() {
      this.$store.commit('Router_ResetStatus');
      await this.fetchData();
    },
    cancelEdit: async function() {
      await this.fetchRouter();
      await this.fetchData();
      this.$emit('close-edit');
    },
    addSetting: async function() {
      if (this.readonly) return;
      await this.$store.commit('Router_AddHttp');
    },
    submit () {
      if (this.readonly) return;
      this.$store.commit('Router_ResetStatus');
      let data = {
        name: this.name,
        namespace: this.namespace,
        httpItems: this.httpItems,
        resourceVersion: this.ruleResourceVersion
      }

      this.$store.dispatch('Router_UpdateRules', data)
    },
  },
  computed: {
    ...mapGetters({
      language: 'Auth_GetLanguage',
      status: 'Router_GetStatus',
      errorHandle: 'Router_GetErrorHandle',
      router: 'Router_GetItem',
      httpItems: 'Router_GetHttpItems',
      ruleResourceVersion: 'Router_GetRuleResourceVersion'
    }),
  }
}
</script>

<style scoped>
.router-setting-page {
  gap: 18px;
}

.router-setting-page.is-readonly {
  gap: 14px;
}

.router-setting-header {
  align-items: flex-end;
}

.protocol-title {
  font-size: clamp(2rem, 6vw, 4rem);
  letter-spacing: -0.06em;
  margin: 0;
  text-transform: uppercase;
}

.router-route-list {
  display: grid;
  gap: 18px;
}

.router-setting-footer {
  border-top: 4px solid var(--pw-primary);
}

.alert-success {
  background: #dcfae6;
  color: #067647;
}

.router-setting-actions {
  justify-content: flex-end;
}

.conflict-alert {
  align-items: center;
  display: flex;
  justify-content: space-between;
}
</style>
