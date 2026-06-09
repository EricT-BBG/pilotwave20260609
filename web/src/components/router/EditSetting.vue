<template>
  <form class="form-stack" @submit.prevent="submit">
    <label class="field">
      <span>{{ $t('Router.RouerName') }}*</span>
      <input v-model.trim="name" disabled placeholder="public-route" type="text" />
    </label>

    <label class="field">
      <span>{{ $t('Table.Namespace') }}*</span>
      <input v-model.trim="namespace" disabled type="text" />
    </label>

    <label class="field">
      <span>{{ $t('Table.Host') }}*</span>
      <input
        data-testid="router-edit-hosts"
        v-model="hostsText"
        placeholder="localhost, api.example.local"
        type="text"
        @blur="touch('hosts')"
      />
      <small v-if="errors.hosts" class="field-error">{{ errors.hosts }}</small>
    </label>

    <label class="field">
      <span>{{ $t('Table.Protocol') }}*</span>
      <select v-model="protocol" @blur="touch('protocol')">
        <option v-for="item in protocols" :key="item.value" :value="item.value">
          {{ item.name }}
        </option>
      </select>
      <small v-if="errors.protocol" class="field-error">{{ errors.protocol }}</small>
    </label>

    <div v-if="status === 'update_success'" class="alert alert-success">
      {{ $t('Alert.Updated') }}
    </div>
    <div v-if="status === 'update_error'" class="alert alert-error">
      {{ $t('Alert.UpdateFailed') }} {{ errorHandle }}
    </div>
    <div v-if="status === 'update_conflict'" class="alert alert-error conflict-alert">
      <span>{{ errorHandle || 'This Router changed in Kubernetes. Reload before submitting again.' }}</span>
      <button class="secondary-button" type="button" @click="reloadData">
        Reload
      </button>
    </div>

    <div class="page-actions form-actions">
      <button class="secondary-button" type="button" @click="fetchData">
        {{ $t('Form.Cancel') }}
      </button>
      <button class="primary-button" data-testid="router-update-submit" type="submit">
        {{ $t('Form.Submit') }}
      </button>
    </div>
  </form>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  name: 'EditRouter',
  mounted: async function() {
    this.$store.commit('Router_ResetStatus');
    this.name = this.$route.query.name;
    this.namespace = this.$route.query.namespace;
    await this.fetchData();
  },
  data() {
    return {
      touched: {},
      submitted: false,
      name: '',
      namespace: '',
      protocol: 'http',
      hostsText: '',
    }
  },
  methods: {
    touch(field) {
      this.touched[field] = true;
    },
    fetchData: async function() {
      let data = await this.$store.dispatch('Router_GetItem', {
        name: this.name,
        namespace: this.namespace
      });

      try {
        data = JSON.parse(JSON.stringify(data));
      } catch (err) {
        console.log(err);
      }

      if (!data) return;

      this.protocol = data.protocol;
      this.hostsText = (data.hosts || []).join(', ');
      this.submitted = false;
      this.touched = {};
    },
    reloadData: async function() {
      this.$store.commit('Router_ResetStatus');
      await this.fetchData();
    },
    submit() {
      this.$store.commit('Router_ResetStatus');
      this.submitted = true;
      this.touched = {
        protocol: true,
        hosts: true,
      };

      if (this.hasErrors) return;

      this.$store.dispatch('Router_UpdateItem', {
        name: this.name,
        protocol: this.protocol,
        namespace: this.namespace,
        hosts: this.hosts,
        resourceVersion: this.router.resourceversion,
      });
    },
  },
  computed: {
    ...mapGetters({
      protocols: 'Router_GetProtocols',
      status: 'Router_GetStatus',
      errorHandle: 'Router_GetErrorHandle',
      router: 'Router_GetItem',
    }),
    hosts() {
      return this.hostsText
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean);
    },
    errors() {
      const messages = {};
      const shouldShow = (field) => this.submitted || this.touched[field];

      if (shouldShow('hosts') && !this.hosts.length) messages.hosts = this.$t('Form.Required');
      if (shouldShow('protocol') && !this.protocol) messages.protocol = this.$t('Form.Required');

      return messages;
    },
    hasErrors() {
      return Boolean(this.errors.hosts || this.errors.protocol);
    },
  },
}
</script>

<style scoped>
.alert-success {
  background: #dcfce7;
  color: #166534;
}

.form-actions {
  border-top: 1px solid var(--pw-border);
  justify-content: flex-end;
  padding-top: 18px;
}

.conflict-alert {
  align-items: center;
  display: flex;
  justify-content: space-between;
}
</style>
