<template>
  <section class="rule-card">
    <div class="section-header">
      <h2>{{ $t('Policy.Rule') }} {{ ruleIndex + 1 }}</h2>
      <button class="danger-button" type="button" :disabled="!ruleIndex" @click="removeRule">
        {{ $t('Policy.RemoveRule') }}
      </button>
    </div>

    <section class="nested-section">
      <div class="section-header">
        <strong>{{ $t('Policy.From') }}</strong>
        <button class="secondary-button compact-button" type="button" @click="addFrom">+ {{ $t('Policy.From') }}</button>
      </div>
      <FromItem
        v-for="(item, i) in rule.from"
        :key="'from-' + i"
        :principals="item.source.principals"
        :notPrincipals="item.source.notPrincipals"
        :requestPrincipals="item.source.requestPrincipals"
        :notRequestPrincipals="item.source.notRequestPrincipals"
        :namespaces="item.source.namespaces"
        :notNamespaces="item.source.notNamespaces"
        :ipBlocks="item.source.ipBlocks"
        :notIpBlocks="item.source.notIpBlocks"
        :remoteIpBlocks="item.source.remoteIpBlocks"
        :notRemoteIpBlocks="item.source.notRemoteIpBlocks"
        :ruleIndex="ruleIndex"
        :index="i"
      />
    </section>

    <section class="nested-section">
      <div class="section-header">
        <strong>{{ $t('Policy.To') }}</strong>
        <button class="secondary-button compact-button" type="button" @click="addTo">+ {{ $t('Policy.To') }}</button>
      </div>
      <ToItem
        v-for="(item, i) in rule.to"
        :key="'to-' + i"
        :hosts="item.operation.hosts"
        :notHosts="item.operation.notHosts"
        :ports="item.operation.ports"
        :notPorts="item.operation.notPorts"
        :methods="item.operation.methods"
        :notMethods="item.operation.notMethods"
        :paths="item.operation.paths"
        :notPaths="item.operation.notPaths"
        :ruleIndex="ruleIndex"
        :index="i"
      />
    </section>

    <section class="nested-section">
      <div class="section-header">
        <strong>{{ $t('Policy.When') }}</strong>
        <button class="secondary-button compact-button" type="button" @click="addWhen">+ {{ $t('Policy.When') }}</button>
      </div>
      <WhenItem
        v-for="(item, i) in rule.when"
        :key="'when-' + i"
        :wKey="item.key"
        :wValues="item.values"
        :notValues="item.notValues"
        :ruleIndex="ruleIndex"
        :index="i"
      />
      <div v-if="!rule.when.length" class="empty-state compact">
        {{ $t('Policy.NoWhenConditionsConfigured') }}
      </div>
    </section>
  </section>
</template>

<script>
import { mapGetters } from 'vuex';
import FromItem from './FromItem.vue';
import ToItem from './ToItem.vue';
import WhenItem from './WhenItem.vue';

export default {
  name: 'AuthPolicyRuleItems',
  components: {
    FromItem,
    ToItem,
    WhenItem,
  },
  props: ['rule', 'ruleIndex'],
  methods: {
    removeRule () {
      if (this.rules.length === 1) return;

      this.$store.commit('AuthPolicy_RemoveRule', {
        ruleIndex: this.ruleIndex
      })
    },
    addFrom () {
      this.$store.commit('AuthPolicy_AddFrom', {
        ruleIndex: this.ruleIndex
      })
    },
    addTo () {
      this.$store.commit('AuthPolicy_AddTo', {
        ruleIndex: this.ruleIndex
      })
    },
    addWhen () {
      this.$store.commit('AuthPolicy_AddWhen', {
        ruleIndex: this.ruleIndex
      })
    },
  },
  computed: {
    ...mapGetters({
      rules: 'AuthPolicy_GetRules'
    }),
  },
}
</script>

<style scoped>
.rule-card {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 18px;
  display: grid;
  gap: 16px;
  margin-bottom: 14px;
  padding: 16px;
}

.section-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.section-header h2 {
  font-size: 1.15rem;
  margin: 0;
}

.nested-section {
  display: grid;
  gap: 10px;
}

.compact-button {
  min-height: 36px;
}
</style>
