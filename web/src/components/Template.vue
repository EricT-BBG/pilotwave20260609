<template>
  <div class="app-shell">
    <Navigation
      @open-about-dialog="openAboutDialog"
      @open-language-dialog="openLanguageDialog"
      @open-namespace-injection="openInjectionDialog"
    />

    <div class="main-shell">
      <header class="topbar">
        <div v-if="showNamespacePicker" class="namespace-picker" @click.stop>
          <span class="namespace-picker-label">{{ $t('Table.Namespace') }}</span>
          <div class="namespace-menu-wrap">
            <button
              class="namespace-menu-trigger"
              data-testid="namespace-menu-open"
              type="button"
              :aria-expanded="namespaceMenuOpen ? 'true' : 'false'"
              @click="toggleNamespaceMenu"
            >
              <span class="namespace-menu-current">
                <strong>{{ selectedNamespaceLabel }}</strong>
                <small v-if="selectedNamespaceStatusLabel">{{ selectedNamespaceStatusLabel }}</small>
              </span>
              <span
                v-if="selectedNamespaceBadge"
                class="namespace-status-badge"
                :class="selectedNamespaceBadgeClass"
                data-testid="namespace-current-status"
              >
                {{ selectedNamespaceBadge }}
              </span>
              <span class="namespace-menu-chevron">⌄</span>
            </button>

            <div v-if="namespaceMenuOpen" class="namespace-menu" data-testid="namespace-menu">
              <div class="namespace-menu-toolbar">
                <span>{{ $t('NamespacePicker.Title') }}</span>
                <button
                  class="text-button namespace-menu-reload"
                  data-testid="namespace-refresh"
                  type="button"
                  :disabled="refreshingNamespaces"
                  @click="reloadNamespacesFromMenu"
                >
                  {{ refreshingNamespaces ? $t('Form.Loading') : $t('Form.Reload') }}
                </button>
              </div>

              <div class="namespace-menu-list">
                <button
                  v-for="item in namespaceMenuItems"
                  :key="item.value"
                  class="namespace-menu-option"
                  :class="{ selected: item.value === namespace }"
                  :data-testid="`namespace-menu-option-${item.value}`"
                  type="button"
                  @click="selectNamespaceFromMenu(item.value)"
                >
                  <span class="namespace-menu-option-main">
                    <strong>{{ item.label }}</strong>
                    <small v-if="item.description">{{ item.description }}</small>
                  </span>
                  <span
                    v-if="item.badge"
                    class="namespace-status-badge"
                    :class="item.badgeClass"
                    :data-testid="`namespace-menu-status-${item.value}`"
                  >
                    {{ item.badge }}
                  </span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="topbar-actions">
          <button
            class="secondary-button topbar-logout-button"
            data-testid="topbar-logout-open"
            type="button"
            @click="openLogoutConfirmDialog"
          >
            {{ $t('Form.Signout') }}
          </button>
        </div>
      </header>

      <div v-if="isIstioUnavailable" class="alert" data-testid="istio-capability-warning">
        {{ istioUnavailableMessage }}
      </div>

      <div v-if="apiErrorMessage" class="alert alert-error api-error-alert" data-testid="api-error-alert">
        <span>{{ $t(apiErrorMessage) }}</span>
        <button class="secondary-button" type="button" @click="clearApiError">
          {{ $t('Form.Close') }}
        </button>
      </div>

      <main class="content-shell">
        <router-view />
      </main>

      <div v-if="languageDialog" class="modal-backdrop management-backdrop" data-testid="language-dialog" @click.self="languageDialog = false">
        <section class="modal-card management-dialog">
          <header class="modal-header">
            <div>
              <p class="eyebrow">{{ $t('Shell.Language') }}</p>
              <h2>{{ $t('Shell.LanguageTitle') }}</h2>
            </div>
          </header>

          <div class="language-choice-list">
            <button
              class="language-choice"
              :class="{ selected: pendingLanguage === 'tw' }"
              type="button"
              @click="pendingLanguage = 'tw'"
            >
              <strong>{{ $t('Shell.LanguageChineseName') }}</strong>
              <span>{{ $t('Shell.LanguageChineseHelp') }}</span>
            </button>
            <button
              class="language-choice"
              :class="{ selected: pendingLanguage === 'en' }"
              type="button"
              @click="pendingLanguage = 'en'"
            >
              <strong>{{ $t('Shell.LanguageEnglishName') }}</strong>
              <span>{{ $t('Shell.LanguageEnglishHelp') }}</span>
            </button>
          </div>

          <footer class="modal-actions language-dialog-actions">
            <button class="secondary-button" type="button" @click="languageDialog = false">
              {{ $t('Form.Cancel') }}
            </button>
            <button class="primary-button" type="button" @click="saveLanguage">
              {{ $t('Form.Save') }}
            </button>
          </footer>
        </section>
      </div>

      <div v-if="logoutConfirmDialog" class="modal-backdrop management-backdrop" data-testid="logout-confirm-dialog" @click.self="logoutConfirmDialog = false">
        <section class="modal-card management-dialog compact">
          <header class="modal-header">
            <div>
              <p class="eyebrow">{{ $t('Shell.SignOut') }}</p>
              <h2>{{ $t('Shell.SignOutTitle') }}</h2>
            </div>
          </header>

          <p class="dialog-copy">{{ $t('Shell.SignOutHelp') }}</p>

          <footer class="modal-actions">
            <button class="secondary-button" type="button" @click="logoutConfirmDialog = false">
              {{ $t('Form.Cancel') }}
            </button>
            <button class="danger-button" type="button" @click="confirmSignout">
              {{ $t('Form.Signout') }}
            </button>
          </footer>
        </section>
      </div>

      <div v-if="aboutDialog" class="modal-backdrop management-backdrop" data-testid="about-version-dialog" @click.self="aboutDialog = false">
        <section class="modal-card management-dialog compact-about-dialog">
          <button
            class="about-close-button"
            data-testid="about-dialog-close"
            type="button"
            @click="aboutDialog = false"
          >
            {{ $t('Form.Close') }}
          </button>

          <div class="about-summary">
            <div class="about-product-mark">P</div>
            <div>
              <p class="eyebrow">{{ $t('Shell.About') }}</p>
              <h2>Pilotwave</h2>
            </div>
          </div>

          <dl class="about-version-grid">
            <div>
              <dt>{{ $t('Shell.Version') }}</dt>
              <dd>{{ buildInfo.version }}</dd>
            </div>
            <div>
              <dt>{{ $t('Shell.BuildTime') }}</dt>
              <dd>{{ buildInfo.buildLabel }}</dd>
            </div>
          </dl>
        </section>
      </div>

      <div v-if="injectionDialog" class="modal-backdrop" @click.self="closeInjectionDialog">
        <section class="modal-card namespace-injection-modal">
          <header class="modal-header">
            <div>
              <p class="eyebrow">{{ $t('NamespaceInjection.Eyebrow') }}</p>
              <h2>{{ $t('NamespaceInjection.Title') }}</h2>
            </div>
          </header>

          <form class="namespace-injection-form" @submit.prevent="openInjectionConfirmation">
            <div class="namespace-injection-layout">
              <section class="namespace-injection-panel">
                <div class="namespace-injection-panel-title">
                  <h3>{{ $t('NamespaceInjection.ChooseNamespace') }}</h3>
                  <span>{{ filteredNamespaceCountText }}</span>
                </div>

                <label class="namespace-injection-search">
                  <span>{{ $t('NamespaceInjection.SearchNamespaces') }}</span>
                  <input
                    v-model.trim="namespaceInjectionSearch"
                    :placeholder="$t('NamespaceInjection.SearchPlaceholder')"
                    data-testid="namespace-injection-search"
                    type="search"
                  />
                </label>

                <div class="namespace-injection-list" data-testid="namespace-injection-list">
                  <button
                    v-for="item in filteredNamespaceDetails"
                    :key="item.name"
                    class="namespace-injection-row"
                    :data-testid="`namespace-injection-row-${item.name}`"
                    :class="{ selected: item.name === injectionTargetNamespace, system: isSystemNamespaceDetail(item) }"
                    type="button"
                    @click="selectInjectionNamespace(item.name)"
                  >
                    <span class="namespace-injection-row-main">
                      <span class="namespace-injection-name">{{ item.name }}</span>
                      <span v-if="item.istioInjection?.revision" class="namespace-injection-revision">
                        {{ item.istioInjection.revision }}
                      </span>
                    </span>
                    <span class="namespace-injection-row-badges">
                      <span v-if="isSystemNamespaceDetail(item)" class="namespace-injection-badge system">
                        {{ $t('NamespaceInjection.SystemNamespaceShort') }}
                      </span>
                      <span
                        class="namespace-injection-badge"
                        :class="injectionStatusClass(item)"
                      >
                        {{ formatInjectionStatus(item) }}
                      </span>
                      <span
                        v-if="hasUnsavedInjectionDraft(item.name)"
                        class="namespace-injection-badge pending"
                        data-testid="namespace-injection-unsaved-badge"
                      >
                        {{ $t('NamespaceInjection.UnsavedBadge') }}
                      </span>
                    </span>
                  </button>
                  <div v-if="!filteredNamespaceDetails.length" class="empty-state namespace-empty-state">
                    <strong>{{ namespaceEmptyTitle }}</strong>
                    <p>{{ namespaceEmptyText }}</p>
                    <button class="secondary-button" type="button" :disabled="refreshingNamespaces" @click="reload">
                      {{ $t('Form.Reload') }}
                    </button>
                  </div>
                </div>
              </section>

              <section class="namespace-injection-panel namespace-injection-editor">
                <div class="namespace-injection-panel-title">
                  <h3>{{ $t('NamespaceInjection.ChooseMode') }}</h3>
                  <span>{{ injectionTargetNamespace || $t('NamespaceInjection.NoNamespaceSelected') }}</span>
                </div>

                <div v-if="!injectionTargetNamespace" class="empty-state namespace-editor-empty" data-testid="namespace-injection-editor-empty">
                  <strong>{{ $t('NamespaceInjection.SelectNamespaceTitle') }}</strong>
                  <p>{{ $t('NamespaceInjection.SelectNamespaceText') }}</p>
                </div>

                <template v-else>
                  <p class="namespace-injection-explain">
                    {{ $t('NamespaceInjection.Explain') }}
                  </p>

                  <div class="namespace-current-state">
                    <span>{{ $t('NamespaceInjection.CurrentState') }}</span>
                    <strong>{{ formatInjectionStatus(injectionTargetDetail) }}</strong>
                  </div>

                  <div v-if="selectedTargetIsSystem" class="alert namespace-system-risk" data-testid="namespace-injection-system-risk">
                    <strong>{{ $t('NamespaceInjection.SystemRiskTitle') }}</strong>
                    <span>{{ $t('NamespaceInjection.SystemRiskText') }}</span>
                  </div>

                  <div class="namespace-injection-options">
                    <label
                      v-for="option in injectionModeOptions"
                      :key="option.value"
                      class="namespace-injection-option"
                      :class="{
                        selected: injectionMode === option.value,
                        disabled: option.disabled,
                        warning: option.value === 'disabled'
                      }"
                    >
                      <input
                        v-model="injectionMode"
                        :disabled="option.disabled"
                        :value="option.value"
                        name="namespace-injection-mode"
                        type="radio"
                      />
                      <span>
                        <strong>{{ option.label }}</strong>
                        <small>{{ option.description }}</small>
                      </span>
                    </label>
                  </div>

                  <label v-if="injectionMode === 'revision'" class="field">
                    <span>{{ $t('NamespaceInjection.DetectedRevision') }}</span>
                    <select
                      v-model="injectionRevision"
                      data-testid="namespace-injection-revision"
                      :disabled="!hasIstioRevisionOptions"
                    >
                      <option value="" disabled>{{ $t('NamespaceInjection.SelectRevision') }}</option>
                      <option v-for="item in istioRevisionOptions" :key="item.value" :value="item.value">
                        {{ item.label }}
                      </option>
                    </select>
                    <small class="field-help">{{ $t('NamespaceInjection.RevisionHelp') }}</small>
                  </label>

                  <div v-if="hasLabelChanges" class="namespace-unsaved-change" data-testid="namespace-injection-unsaved-change">
                    {{ $t('NamespaceInjection.UnsavedChanges') }}
                  </div>
                </template>
              </section>
            </div>

            <footer class="namespace-injection-footer">
              <div class="namespace-injection-footer-copy">
                <div v-if="injectionMessage" class="alert" :class="injectionOk ? 'namespace-success' : 'alert-error'">
                  {{ injectionMessage }}
                </div>

                <div v-if="applyDisabledReason" class="namespace-apply-reason">
                  {{ applyDisabledReason }}
                </div>
              </div>

              <div class="modal-actions">
                <button class="secondary-button" type="button" @click="closeInjectionDialog">
                  {{ $t('Form.Cancel') }}
                </button>
                <button
                  class="primary-button"
                  data-testid="namespace-injection-save"
                  type="submit"
                  :disabled="!canSaveInjection || savingInjection"
                >
                  {{ injectionApplyButtonLabel }}
                </button>
              </div>
            </footer>
          </form>
        </section>
      </div>

      <div
        v-if="injectionConfirmOpen"
        class="modal-backdrop namespace-confirm-backdrop"
        data-testid="namespace-injection-confirm-dialog"
        @click.self="closeInjectionConfirmation"
      >
        <section class="modal-card namespace-confirm-dialog">
          <header class="modal-header">
            <div>
              <p class="eyebrow">{{ injectionTargetNamespace }}</p>
              <h2>{{ $t('NamespaceInjection.ConfirmTitle') }}</h2>
            </div>
          </header>

          <div class="namespace-confirm-body">
            <p>{{ $t('NamespaceInjection.ConfirmText') }}</p>

            <div class="namespace-confirm-summary">
              <span>{{ $t('NamespaceInjection.CurrentState') }}</span>
              <strong>{{ formatInjectionStatus(injectionTargetDetail) }}</strong>
              <span>{{ $t('NamespaceInjection.ConfirmTargetMode') }}</span>
              <strong>{{ selectedInjectionModeLabel }}</strong>
            </div>

            <div class="namespace-confirm-list" data-testid="namespace-injection-confirm-list">
              <label
                v-for="item in changedLabelDiffItems"
                :key="item.key"
                class="namespace-confirm-row"
                :data-testid="`namespace-injection-confirm-label-${toTestIdKey(item.key)}`"
              >
                <input
                  v-model="confirmedLabelKeys"
                  :value="item.key"
                  type="checkbox"
                />
                <span>
                  <strong>{{ item.key }}</strong>
                  <small>
                    <code>{{ item.from }}</code>
                    <span>→</span>
                    <code>{{ item.to }}</code>
                  </small>
                </span>
              </label>
            </div>

            <div v-if="selectedTargetIsSystem" class="alert namespace-system-risk">
              <strong>{{ $t('NamespaceInjection.SystemRiskTitle') }}</strong>
              <span>{{ $t('NamespaceInjection.SystemRiskText') }}</span>
            </div>

            <div v-if="!allConfirmationLabelsChecked" class="namespace-apply-reason">
              {{ $t('NamespaceInjection.ConfirmSelectAll') }}
            </div>

            <div class="alert namespace-warning">
              {{ $t('NamespaceInjection.RestartWarning') }}
            </div>
          </div>

          <footer class="modal-actions namespace-confirm-actions">
            <button class="secondary-button" type="button" @click="closeInjectionConfirmation">
              {{ $t('NamespaceInjection.Discard') }}
            </button>
            <button
              class="primary-button"
              data-testid="namespace-injection-confirm-apply"
              type="button"
              :disabled="savingInjection || !allConfirmationLabelsChecked"
              @click="saveInjection"
            >
              {{ savingInjection ? $t('Form.Saving') : $t('NamespaceInjection.ConfirmApply') }}
            </button>
          </footer>
        </section>
      </div>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex';
import Navigation from './Navigation.vue';
import { buildInfo } from '../lib/buildInfo';
import { normalizeLocale, resolveLocale, SUPPORTED_LOCALES } from '../lib/locale';

export default {
  name: 'Template',
  components: {
    Navigation,
  },
  mounted() {
    this.applyLocale(this.language || resolveLocale());
    this.fetchData();
    window.addEventListener('keydown', this.handleEscapeKey);
    window.addEventListener('pilotwave-api-error', this.handleApiError);
    document.addEventListener('click', this.closeNamespaceMenu);
  },
  beforeUnmount() {
    window.removeEventListener('keydown', this.handleEscapeKey);
    window.removeEventListener('pilotwave-api-error', this.handleApiError);
    document.removeEventListener('click', this.closeNamespaceMenu);
  },
  data() {
    return {
      refreshingNamespaces: false,
      namespaceMenuOpen: false,
      injectionDialog: false,
      injectionMode: 'disabled',
      injectionRevision: '',
      injectionTargetNamespace: '',
      namespaceInjectionSearch: '',
      savingInjection: false,
      injectionMessage: '',
      injectionOk: false,
      injectionConfirmOpen: false,
      confirmedLabelKeys: [],
      injectionDraftsByNamespace: {},
      languageDialog: false,
      logoutConfirmDialog: false,
      aboutDialog: false,
      pendingLanguage: '',
      apiErrorMessage: '',
      buildInfo,
    };
  },
  methods: {
    applyLocale(newLang) {
      const normalizedLang = normalizeLocale(newLang);
      const targetLang = SUPPORTED_LOCALES.includes(normalizedLang) ? normalizedLang : resolveLocale();
      if (this.$i18n?.locale && typeof this.$i18n.locale === 'object' && 'value' in this.$i18n.locale) {
        this.$i18n.locale.value = targetLang;
      } else if (this.$i18n) {
        this.$i18n.locale = targetLang;
      }

      this.$store.commit('Auth_SetLanguage', {
        lang: targetLang,
      });
    },
    async fetchData() {
      this.refreshingNamespaces = true;
      try {
        const [namespaces] = await Promise.all([
          this.$store.dispatch('Auth_GetNamespaces'),
          this.$store.dispatch('Auth_GetClusterCapabilities'),
        ]);
        if (Array.isArray(namespaces) && namespaces.length && !this.namespaceOptions.includes(this.namespace)) {
          this.$store.commit('Auth_SetNamespace', 'All');
        }
        if (this.injectionDialog) {
          this.pruneInjectionDrafts();
        }
      } finally {
        this.refreshingNamespaces = false;
      }
    },
    async reload() {
      await this.fetchData();
    },
    async reloadNamespacesFromMenu() {
      await this.reload();
    },
    toggleNamespaceMenu() {
      this.namespaceMenuOpen = !this.namespaceMenuOpen;
    },
    closeNamespaceMenu() {
      this.namespaceMenuOpen = false;
    },
    selectNamespaceFromMenu(value) {
      this.namespace = value || 'All';
      this.closeNamespaceMenu();
    },
    openLanguageDialog() {
      this.pendingLanguage = normalizeLocale(this.language) || resolveLocale();
      this.languageDialog = true;
    },
    saveLanguage() {
      this.switchLang(this.pendingLanguage || 'en');
      this.languageDialog = false;
    },
    openLogoutConfirmDialog() {
      this.logoutConfirmDialog = true;
    },
    confirmSignout() {
      this.logoutConfirmDialog = false;
      this.signout();
    },
    openAboutDialog() {
      this.aboutDialog = true;
    },
    handleEscapeKey(event) {
      if (event.key !== 'Escape') return;
      if (this.injectionConfirmOpen) {
        this.closeInjectionConfirmation();
        return;
      }
      if (this.injectionDialog) {
        this.closeInjectionDialog();
        return;
      }
      this.namespaceMenuOpen = false;
      this.languageDialog = false;
      this.logoutConfirmDialog = false;
      this.aboutDialog = false;
    },
    handleApiError(event) {
      this.apiErrorMessage = event?.detail?.message || 'Alert.ApiUnavailable';
    },
    clearApiError() {
      this.apiErrorMessage = '';
    },
    formatInjectionStatus(item) {
      const status = item?.istioInjection?.mode || item?.istioInjection?.status || 'disabled';
      if (status === 'enabled') return this.$t('NamespaceInjection.StatusDefault');
      if (status === 'revision') return this.$t('NamespaceInjection.StatusRevision');
      return this.$t('NamespaceInjection.StatusOff');
    },
    injectionStatusClass(item) {
      const status = item?.istioInjection?.mode || item?.istioInjection?.status || 'disabled';
      return {
        enabled: status === 'enabled',
        revision: status === 'revision',
      };
    },
    formatNamespacePickerStatus(item) {
      const mode = item?.istioInjection?.mode || item?.istioInjection?.status || 'disabled';
      const revision = item?.istioInjection?.revision || '';
      if (mode === 'enabled') return this.$t('NamespacePicker.Injected');
      if (mode === 'revision') {
        return revision
          ? this.$t('NamespacePicker.RevisionBadge', { revision })
          : this.$t('NamespaceInjection.StatusRevision');
      }
      return this.$t('NamespacePicker.NotInjected');
    },
    isSystemNamespaceDetail(item) {
      return item?.systemNamespace === true;
    },
    getNamespaceDetail(name) {
      return (this.namespaceDetails || []).find((item) => item.name === name);
    },
    openInjectionDialog() {
      if (this.isIstioUnavailable) return;
      if (this.namespace && this.namespace !== 'All') {
        this.injectionTargetNamespace = this.namespace;
      } else if (this.injectionTargetNamespace && !this.getNamespaceDetail(this.injectionTargetNamespace)) {
        this.injectionTargetNamespace = '';
      }
      this.hydrateInjectionForm();
      this.injectionMessage = '';
      this.injectionOk = false;
      this.injectionDialog = true;
    },
    closeInjectionDialog() {
      this.injectionDialog = false;
      this.closeInjectionConfirmation();
      this.injectionDraftsByNamespace = {};
    },
    selectInjectionNamespace(name) {
      this.persistCurrentInjectionDraft();
      this.injectionTargetNamespace = name;
      this.hydrateInjectionForm();
      this.injectionMessage = '';
      this.injectionOk = false;
      this.closeInjectionConfirmation();
    },
    currentInjectionDraft() {
      return {
        mode: this.injectionMode,
        revision: this.injectionMode === 'revision' ? this.injectionRevision : '',
      };
    },
    persistCurrentInjectionDraft() {
      if (!this.injectionDialog || !this.injectionTargetNamespace) return;
      if (!this.getNamespaceDetail(this.injectionTargetNamespace)) return;
      this.injectionDraftsByNamespace = {
        ...this.injectionDraftsByNamespace,
        [this.injectionTargetNamespace]: this.currentInjectionDraft(),
      };
    },
    normalizeInjectionDraft(injection = {}) {
      const mode = injection.mode || 'disabled';
      if (mode === 'revision' && !this.isKnownIstioRevision(injection.revision)) {
        return {
          mode: this.hasIstioRevisionOptions ? 'revision' : 'disabled',
          revision: this.hasIstioRevisionOptions ? this.istioRevisionOptions[0].value : '',
        };
      }

      return {
        mode: mode === 'revision' && !this.hasIstioRevisionOptions ? 'disabled' : mode,
        revision: mode === 'revision' ? injection.revision || '' : '',
      };
    },
    clusterInjectionDraft(name) {
      return this.normalizeInjectionDraft(this.getNamespaceDetail(name)?.istioInjection || {});
    },
    setInjectionFormDraft(draft) {
      this.injectionMode = draft.mode;
      this.injectionRevision = draft.mode === 'revision' ? draft.revision : '';
    },
    pruneInjectionDrafts() {
      const existing = new Set((this.namespaceDetails || []).map((item) => item?.name).filter(Boolean));
      const drafts = Object.entries(this.injectionDraftsByNamespace).filter(([name]) => existing.has(name));
      this.injectionDraftsByNamespace = Object.fromEntries(drafts);
      if (this.injectionTargetNamespace && !existing.has(this.injectionTargetNamespace)) {
        this.injectionTargetNamespace = '';
        this.injectionMode = 'disabled';
        this.injectionRevision = '';
      }
    },
    hydrateInjectionForm() {
      const name = this.injectionTargetNamespace;
      if (!name || !this.getNamespaceDetail(name)) {
        this.injectionMode = 'disabled';
        this.injectionRevision = '';
        return;
      }

      const draft = this.injectionDraftsByNamespace[name] || this.clusterInjectionDraft(name);
      this.injectionDraftsByNamespace = {
        ...this.injectionDraftsByNamespace,
        [name]: draft,
      };
      this.setInjectionFormDraft(draft);
    },
    isKnownIstioRevision(revision) {
      if (!revision) return false;
      return this.istioRevisionOptions.some((item) => item.value === revision);
    },
    openInjectionConfirmation() {
      if (!this.canSaveInjection) return;
      this.confirmedLabelKeys = [];
      this.injectionConfirmOpen = true;
    },
    closeInjectionConfirmation() {
      this.injectionConfirmOpen = false;
      this.confirmedLabelKeys = [];
    },
    toTestIdKey(key) {
      return String(key || '').replace(/[^a-zA-Z0-9_-]+/g, '-');
    },
    hasUnsavedInjectionDraft(name) {
      const draft = name === this.injectionTargetNamespace
        ? this.currentInjectionDraft()
        : this.injectionDraftsByNamespace[name];
      if (!draft) return false;

      const current = this.clusterInjectionDraft(name);
      return draft.mode !== current.mode || draft.revision !== current.revision;
    },
    async saveInjection() {
      if (!this.canSaveInjection) return;
      if (!this.allConfirmationLabelsChecked) return;
      const isSystemNamespace = this.isSystemNamespaceDetail(this.injectionTargetDetail);

      this.savingInjection = true;
      this.injectionMessage = '';
      this.injectionOk = false;
      try {
        await this.$store.dispatch('Auth_UpdateNamespaceInjection', {
          name: this.injectionTargetNamespace,
          mode: this.injectionMode,
          revision: this.injectionRevision,
          allowSystemNamespace: isSystemNamespace,
        });
        await this.fetchData();
        this.injectionDraftsByNamespace = {
          ...this.injectionDraftsByNamespace,
          [this.injectionTargetNamespace]: this.clusterInjectionDraft(this.injectionTargetNamespace),
        };
        this.hydrateInjectionForm();
        this.injectionOk = true;
        this.injectionMessage = this.$t('NamespaceInjection.Success');
        this.closeInjectionConfirmation();
      } catch (err) {
        this.injectionOk = false;
        this.injectionMessage = err?.response?.data || this.$t('Alert.UpdateFailed');
      } finally {
        this.savingInjection = false;
      }
    },
    toUrl(url) {
      window.scrollTo(0, 0);
      if (url) this.$router.push(url);
    },
    openUserSettings() {
      if (this.userId) {
        this.toUrl(`/user/${this.userId}`);
      }
    },
    switchLang(newLang) {
      this.applyLocale(newLang);
    },
    signout() {
      const locale = sessionStorage.getItem('locale');
      sessionStorage.clear();
      if (locale) {
        sessionStorage.setItem('locale', locale);
        this.$store.commit('Auth_SetLanguage', {
          lang: locale,
        });
      }
      this.toUrl('/');
    },
  },
  computed: {
    ...mapGetters({
      namespaces: 'Auth_GetNamespaces',
      namespaceDetails: 'Auth_GetNamespaceDetails',
      userInfo: 'Auth_GetUserInfo',
      language: 'Auth_GetLanguage',
      istioCapabilities: 'Auth_GetIstioCapabilities',
    }),
    namespaceOptions() {
      return (this.namespaces || []).filter((item) => item && item !== 'All');
    },
    selectedNamespaceDetail() {
      if (!this.namespace || this.namespace === 'All') return null;
      return this.getNamespaceDetail(this.namespace);
    },
    selectedNamespaceLabel() {
      if (!this.namespace || this.namespace === 'All') return this.$t('NamespaceInjection.AllNamespaces');
      return this.namespace;
    },
    selectedNamespaceStatusLabel() {
      if (!this.namespace || this.namespace === 'All') return '';
      const revision = this.selectedNamespaceDetail?.istioInjection?.revision || '';
      return revision ? this.$t('NamespacePicker.RevisionHelp', { revision }) : '';
    },
    selectedNamespaceBadge() {
      if (!this.selectedNamespaceDetail) return '';
      return this.formatNamespacePickerStatus(this.selectedNamespaceDetail);
    },
    selectedNamespaceBadgeClass() {
      return this.injectionStatusClass(this.selectedNamespaceDetail);
    },
    namespaceMenuItems() {
      const items = [{
        value: 'All',
        label: this.$t('NamespaceInjection.AllNamespaces'),
        description: '',
        badge: '',
        badgeClass: {},
      }];

      for (const item of this.namespaceDetails || []) {
        if (!item?.name) continue;
        items.push({
          value: item.name,
          label: item.name,
          description: '',
          badge: this.formatNamespacePickerStatus(item),
          badgeClass: this.injectionStatusClass(item),
        });
      }

      return items;
    },
    filteredNamespaceDetails() {
      const search = (this.namespaceInjectionSearch || '').trim().toLowerCase();
      return (this.namespaceDetails || []).filter((item) => {
        if (!item?.name) return false;
        const mode = item?.istioInjection?.mode || item?.istioInjection?.status || 'disabled';
        const revision = item?.istioInjection?.revision || '';

        if (!search) return true;
        return [item.name, mode, revision].some((value) => String(value || '').toLowerCase().includes(search));
      });
    },
    namespaceEmptyTitle() {
      if (!(this.namespaceDetails || []).length) return this.$t('NamespaceInjection.NoNamespacesTitle');
      return this.$t('NamespaceInjection.NoMatchingNamespacesTitle');
    },
    namespaceEmptyText() {
      if (!(this.namespaceDetails || []).length) return this.$t('NamespaceInjection.NoNamespaces');
      if (this.namespaceInjectionSearch) return this.$t('NamespaceInjection.NoSearchMatches');
      return this.$t('NamespaceInjection.NoFilterMatches');
    },
    userId() {
      return this.userInfo?.uid || '';
    },
    injectionTargetDetail() {
      return this.getNamespaceDetail(this.injectionTargetNamespace);
    },
    isIstioUnavailable() {
      return this.istioCapabilities.installed === false || this.istioCapabilities.disabled === true;
    },
    istioRevisionOptions() {
      const options = [];
      const seen = new Set();

      for (const tag of this.istioCapabilities.revisionTags || []) {
        if (!tag?.name || seen.has(tag.name)) continue;
        seen.add(tag.name);
        options.push({
          value: tag.name,
          label: tag.revision
            ? this.$t('NamespaceInjection.RevisionTagOption', { tag: tag.name, revision: tag.revision })
            : this.$t('NamespaceInjection.RevisionTagOnlyOption', { tag: tag.name }),
        });
      }

      for (const revision of this.istioCapabilities.revisions || []) {
        if (!revision || seen.has(revision)) continue;
        seen.add(revision);
        options.push({
          value: revision,
          label: this.$t('NamespaceInjection.RevisionOption', { revision }),
        });
      }

      return options;
    },
    hasIstioRevisionOptions() {
      return this.istioRevisionOptions.length > 0;
    },
    injectionModeOptions() {
      return [
        {
          value: 'disabled',
          label: this.$t('NamespaceInjection.ModeOff'),
          description: this.$t('NamespaceInjection.ModeOffDescription'),
          disabled: false,
        },
        {
          value: 'enabled',
          label: this.$t('NamespaceInjection.ModeDefault'),
          description: this.$t('NamespaceInjection.ModeDefaultDescription'),
          disabled: false,
        },
        {
          value: 'revision',
          label: this.$t('NamespaceInjection.ModeRevision'),
          description: this.hasIstioRevisionOptions
            ? this.$t('NamespaceInjection.ModeRevisionDescription')
            : this.$t('NamespaceInjection.ModeRevisionUnavailable'),
          disabled: !this.hasIstioRevisionOptions,
        },
      ];
    },
    istioUnavailableMessage() {
      if (this.istioCapabilities.disabled) {
        return this.istioCapabilities.message || this.$t('NamespaceInjection.IstioDisabled');
      }

      if (this.istioCapabilities.installed === false) {
        const missingCRDs = this.istioCapabilities.missingCRDs || [];
        if (missingCRDs.length) {
          return this.$t('NamespaceInjection.IstioMissingCRDs', { crds: missingCRDs.join(', ') });
        }
        return this.istioCapabilities.message || this.$t('NamespaceInjection.IstioNotInstalled');
      }

      return '';
    },
    canSaveInjection() {
      return !this.applyDisabledReason;
    },
    applyDisabledReason() {
      if (this.isIstioUnavailable) return this.$t('NamespaceInjection.DisabledIstioUnavailable');
      if (!this.injectionTargetNamespace) return this.$t('NamespaceInjection.DisabledNoNamespace');
      if (this.injectionMode === 'revision' && !this.hasIstioRevisionOptions) {
        return this.$t('NamespaceInjection.DisabledRevisionUnavailable');
      }
      if (this.injectionMode === 'revision' && !this.isKnownIstioRevision(this.injectionRevision)) {
        return this.$t('NamespaceInjection.DisabledRevisionRequired');
      }
      if (!['disabled', 'enabled', 'revision'].includes(this.injectionMode)) {
        return this.$t('NamespaceInjection.DisabledInvalidMode');
      }
      if (!this.hasLabelChanges) return this.$t('NamespaceInjection.DisabledNoChange');
      return '';
    },
    selectedTargetIsSystem() {
      return this.isSystemNamespaceDetail(this.injectionTargetDetail);
    },
    targetCurrentLabels() {
      const labels = this.injectionTargetDetail?.labels || this.injectionTargetDetail?.raw?.metadata?.labels || {};
      const mode = this.injectionTargetDetail?.istioInjection?.mode || 'disabled';
      const revision = this.injectionTargetDetail?.istioInjection?.revision || '';
      return {
        'istio-injection': labels['istio-injection'] || (mode === 'enabled' ? 'enabled' : ''),
        'istio.io/rev': labels['istio.io/rev'] || (mode === 'revision' ? revision : ''),
      };
    },
    targetNextLabels() {
      if (this.injectionMode === 'enabled') {
        return {
          'istio-injection': 'enabled',
          'istio.io/rev': '',
        };
      }
      if (this.injectionMode === 'revision') {
        return {
          'istio-injection': '',
          'istio.io/rev': this.injectionRevision || '',
        };
      }
      return {
        'istio-injection': 'disabled',
        'istio.io/rev': '',
      };
    },
    labelDiffItems() {
      const empty = this.$t('NamespaceInjection.LabelUnset');
      const current = this.targetCurrentLabels;
      const next = this.targetNextLabels;
      return ['istio-injection', 'istio.io/rev'].map((key) => ({
        key,
        from: current[key] || empty,
        to: next[key] || empty,
        changed: current[key] !== next[key],
      }));
    },
    changedLabelDiffItems() {
      return this.labelDiffItems.filter((item) => item.changed);
    },
    allConfirmationLabelsChecked() {
      if (!this.changedLabelDiffItems.length) return false;
      return this.changedLabelDiffItems.every((item) => this.confirmedLabelKeys.includes(item.key));
    },
    selectedInjectionModeLabel() {
      const option = this.injectionModeOptions.find((item) => item.value === this.injectionMode);
      return option?.label || this.$t('NamespaceInjection.NoNamespaceSelected');
    },
    injectionApplyButtonLabel() {
      if (this.savingInjection) return this.$t('Form.Saving');
      if (this.injectionTargetNamespace && !this.hasLabelChanges) return this.$t('NamespaceInjection.DisabledNoChange');
      return this.$t('NamespaceInjection.Apply');
    },
    hasLabelChanges() {
      return this.changedLabelDiffItems.length > 0;
    },
    injectionPreviewText() {
      if (!this.injectionTargetNamespace) return '';
      if (this.injectionMode === 'enabled') {
        return this.$t('NamespaceInjection.PreviewDefault', { namespace: this.injectionTargetNamespace });
      }
      if (this.injectionMode === 'revision') {
        const revision = this.injectionRevision || '<revision>';
        return this.$t('NamespaceInjection.PreviewRevision', { namespace: this.injectionTargetNamespace, revision });
      }
      return this.$t('NamespaceInjection.PreviewOff', { namespace: this.injectionTargetNamespace });
    },
    filteredNamespaceCountText() {
      return this.$t('NamespaceInjection.NamespaceCountFilteredVisible', {
        count: this.filteredNamespaceDetails.length,
        total: (this.namespaceDetails || []).length,
      });
    },
    showNamespacePicker() {
      return this.$route?.meta?.hideNamespacePicker !== true;
    },
    namespace: {
      get() {
        return this.$store.state.Auth.namespace;
      },
      set(value) {
        this.$store.commit('Auth_SetNamespace', value || 'All');
      },
    },
  },
  watch: {
    namespace() {
      if (!this.injectionDialog) return;
      if (this.namespace && this.namespace !== 'All') {
        this.selectInjectionNamespace(this.namespace);
      }
    },
    namespaceDetails() {
      if (this.injectionTargetNamespace && this.getNamespaceDetail(this.injectionTargetNamespace)) return;
      this.injectionTargetNamespace = '';
      this.hydrateInjectionForm();
    },
  },
};
</script>

<style scoped>
.modal-backdrop {
  align-items: stretch;
  padding: 24px;
}

.modal-card.namespace-injection-modal {
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  margin: auto;
  max-height: min(760px, calc(100vh - 48px));
  max-width: 1080px;
  overflow: hidden;
  padding: 0;
}

.namespace-injection-modal .modal-header {
  border-bottom: 1px solid var(--pw-border);
  flex: 0 0 auto;
  margin: 0;
  padding: 18px 20px;
}

.namespace-injection-modal .modal-header h2 {
  font-size: 1.28rem;
  letter-spacing: 0;
}

.namespace-injection-form {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 14px;
  min-height: 0;
  overflow: hidden;
  padding: 16px 20px 0;
}

.namespace-injection-panel-title span,
.namespace-injection-badge {
  align-items: center;
  border-radius: 999px;
  display: inline-flex;
  line-height: 1.2;
}

.namespace-injection-layout {
  align-items: stretch;
  display: grid;
  flex: 1 1 auto;
  gap: 14px;
  grid-template-columns: minmax(320px, 0.9fr) minmax(420px, 1.1fr);
  min-height: 0;
  overflow: auto;
  padding: 1px 2px 2px;
}

.namespace-injection-panel {
  background: #fff;
  border: 1px solid rgba(216, 210, 198, 0.9);
  border-radius: 10px;
  box-shadow: 0 8px 22px rgba(25, 23, 20, 0.06);
  min-width: 0;
  padding: 14px;
}

.namespace-injection-panel-title {
  align-items: flex-start;
  display: flex;
  gap: 12px;
  justify-content: space-between;
  margin-bottom: 12px;
}

.namespace-injection-panel-title h3 {
  font-size: 0.98rem;
  letter-spacing: 0;
  line-height: 1.25;
  margin: 0;
}

.namespace-injection-panel-title span {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  color: var(--pw-muted);
  flex: 0 1 auto;
  font-size: 0.72rem;
  font-weight: 800;
  justify-content: center;
  max-width: 52%;
  min-height: 25px;
  padding: 4px 8px;
  text-align: right;
}

.namespace-injection-search {
  align-items: center;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: var(--pw-muted);
  display: flex;
  font-size: 0.82rem;
  font-weight: 800;
  gap: 8px;
  margin: 0 0 12px;
  padding: 8px 10px;
}

.namespace-injection-search {
  align-items: stretch;
  display: grid;
  gap: 6px;
}

.namespace-injection-search span {
  font-size: 0.78rem;
  font-weight: 800;
}

.namespace-injection-search input {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  min-height: 38px;
  padding: 0 10px;
}

.namespace-injection-list {
  display: grid;
  gap: 8px;
  max-height: min(42vh, 390px);
  overflow: auto;
  padding: 1px 3px 1px 1px;
}

.namespace-injection-row {
  align-items: start;
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  color: var(--pw-text);
  display: grid;
  gap: 8px;
  grid-template-columns: 1fr;
  min-height: 46px;
  padding: 9px 10px;
  text-align: left;
  transition: border-color 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;
}

.namespace-injection-row:hover {
  border-color: rgba(31, 41, 51, 0.34);
  box-shadow: 0 8px 18px rgba(25, 23, 20, 0.08);
}

.namespace-injection-row.selected {
  background: #eef1fe;
  border-color: var(--pw-accent);
  box-shadow: 0 0 0 2px rgba(58, 91, 217, 0.13);
}

.namespace-injection-row.system {
  background: #fff8eb;
}

.namespace-injection-name {
  font-weight: 800;
  line-height: 1.25;
  min-width: 0;
  overflow-wrap: anywhere;
}

.namespace-injection-row-main,
.namespace-injection-row-badges {
  align-items: flex-start;
  display: flex;
  gap: 7px;
  min-width: 0;
}

.namespace-injection-row-main {
  flex-wrap: wrap;
}

.namespace-injection-row-badges {
  flex-wrap: wrap;
  justify-content: flex-start;
}

.namespace-injection-revision {
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 7px;
  color: #1d4ed8;
  font-size: 0.7rem;
  font-weight: 800;
  line-height: 1.25;
  max-width: 100%;
  overflow-wrap: anywhere;
  padding: 3px 7px;
}

.namespace-injection-badge {
  background: #f3f4f6;
  border: 1px solid transparent;
  color: #4b5563;
  font-size: 0.72rem;
  font-weight: 900;
  justify-content: center;
  max-width: 100%;
  min-height: 24px;
  padding: 4px 8px;
  text-align: center;
}

.namespace-injection-badge.enabled {
  background: #ecfdf3;
  border-color: #bbf7d0;
  color: #166534;
}

.namespace-injection-badge.revision {
  background: #eff6ff;
  border-color: #bfdbfe;
  color: #1d4ed8;
}

.namespace-injection-badge.system {
  background: #fff7ed;
  border-color: #fed7aa;
  color: #9a3412;
}

.namespace-injection-badge.pending {
  background: #eef2ff;
  border-color: #c7d2fe;
  color: #3730a3;
}

.namespace-injection-editor {
  align-content: start;
  display: grid;
  gap: 12px;
}

.namespace-injection-explain {
  color: var(--pw-muted);
  line-height: 1.45;
  margin: -2px 0 0;
}

.namespace-current-state,
.namespace-warning,
.namespace-success,
.namespace-system-risk,
.namespace-apply-reason,
.namespace-editor-empty,
.namespace-injection-list .empty-state {
  border-radius: 8px;
}

.api-error-alert {
  align-items: center;
  display: flex;
  gap: 12px;
  justify-content: space-between;
  margin: 0 0 16px;
}

.api-error-alert .secondary-button {
  flex: 0 0 auto;
}

.namespace-current-state {
  align-items: center;
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  color: var(--pw-muted);
  display: flex;
  gap: 8px;
  justify-content: space-between;
  padding: 10px 12px;
}

.namespace-current-state strong {
  color: var(--pw-primary);
}

.namespace-injection-options {
  display: grid;
  gap: 9px;
}

.namespace-injection-option {
  align-items: flex-start;
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  cursor: pointer;
  display: grid;
  gap: 10px;
  grid-template-columns: auto minmax(0, 1fr);
  min-height: 58px;
  padding: 11px 12px;
  transition: border-color 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;
}

.namespace-injection-option:hover {
  border-color: rgba(31, 41, 51, 0.34);
}

.namespace-injection-option.selected {
  background: #eef1fe;
  border-color: rgba(58, 91, 217, 0.68);
  box-shadow: 0 0 0 2px rgba(58, 91, 217, 0.12);
}

.namespace-injection-option.recommended.selected {
  background: #ecfdf3;
  border-color: #86efac;
  box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.12);
}

.namespace-injection-option.warning.selected {
  background: #fff7ed;
  border-color: #fdba74;
  box-shadow: 0 0 0 2px rgba(249, 115, 22, 0.12);
}

.namespace-injection-option.disabled {
  background: #f8fafc;
  color: #94a3b8;
  cursor: not-allowed;
}

.namespace-injection-option input {
  margin-top: 4px;
}

.namespace-injection-option span {
  display: grid;
  gap: 3px;
  min-width: 0;
}

.namespace-injection-option strong,
.namespace-injection-option small {
  min-width: 0;
}

.namespace-injection-option small {
  color: var(--pw-muted);
  line-height: 1.35;
}

.namespace-injection-option.disabled small {
  color: #94a3b8;
}

.namespace-injection-editor .field {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  padding: 10px 12px;
}

.namespace-injection-editor .field select {
  border-radius: 8px;
  width: 100%;
}

.namespace-unsaved-change {
  background: #eef2ff;
  border: 1px solid #c7d2fe;
  border-radius: 8px;
  color: #3730a3;
  font-size: 0.82rem;
  font-weight: 900;
  padding: 9px 11px;
}

.namespace-warning {
  background: #fff7ed;
  border: 1px solid #fed7aa;
  color: #9a3412;
}

.namespace-success {
  background: #ecfdf3;
  border: 1px solid #bbf7d0;
  color: #166534;
}

.namespace-system-risk {
  background: #fff7ed;
  border: 1px solid #fed7aa;
  color: #9a3412;
  display: grid;
  gap: 4px;
}

.namespace-apply-reason {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 700;
  padding: 9px 10px;
}

.namespace-injection-list .empty-state {
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
  color: #64748b;
  font-weight: 700;
  padding: 18px;
  text-align: center;
}

.namespace-empty-state,
.namespace-editor-empty {
  display: grid;
  gap: 8px;
}

.namespace-empty-state p,
.namespace-editor-empty p {
  color: #64748b;
  margin: 0;
}

.namespace-empty-state .secondary-button {
  border-radius: 8px;
  justify-self: center;
}

.namespace-injection-footer {
  align-items: start;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), #fff 30%),
    var(--pw-surface);
  border-top: 1px solid var(--pw-border);
  display: grid;
  flex: 0 0 auto;
  gap: 14px;
  grid-template-columns: minmax(0, 1fr) auto;
  margin: 0 -20px;
  padding: 12px 20px 16px;
}

.namespace-injection-footer-copy {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.namespace-injection-modal .modal-actions {
  background: transparent;
  border-top: 0;
  justify-content: flex-end;
  margin: 0;
  padding: 0;
  position: static;
}

.namespace-injection-modal .modal-actions .primary-button,
.namespace-injection-modal .modal-actions .secondary-button {
  border-radius: 8px;
  min-width: 112px;
}

.namespace-confirm-backdrop {
  align-items: center;
  z-index: 1200;
}

.namespace-confirm-dialog {
  border-radius: 12px;
  max-width: min(680px, calc(100vw - 48px));
  overflow: hidden;
  padding: 0;
}

.namespace-confirm-dialog .modal-header {
  border-bottom: 1px solid var(--pw-border);
  margin: 0;
  padding: 18px 20px;
}

.namespace-confirm-dialog .modal-header h2 {
  font-size: 1.25rem;
  letter-spacing: 0;
}

.namespace-confirm-body {
  display: grid;
  gap: 12px;
  padding: 18px 20px;
}

.namespace-confirm-body p {
  color: var(--pw-muted);
  line-height: 1.45;
  margin: 0;
}

.namespace-confirm-summary {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  display: grid;
  gap: 8px 12px;
  grid-template-columns: max-content minmax(0, 1fr);
  padding: 12px;
}

.namespace-confirm-summary span {
  color: var(--pw-muted);
  font-weight: 800;
}

.namespace-confirm-summary strong {
  color: var(--pw-primary);
  min-width: 0;
  overflow-wrap: anywhere;
}

.namespace-confirm-list {
  display: grid;
  gap: 8px;
}

.namespace-confirm-row {
  align-items: flex-start;
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 8px;
  cursor: pointer;
  display: grid;
  gap: 10px;
  grid-template-columns: auto minmax(0, 1fr);
  padding: 11px 12px;
}

.namespace-confirm-row input {
  margin-top: 4px;
}

.namespace-confirm-row span {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.namespace-confirm-row strong {
  color: var(--pw-primary);
  overflow-wrap: anywhere;
}

.namespace-confirm-row small {
  align-items: center;
  color: var(--pw-muted);
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  line-height: 1.35;
}

.namespace-confirm-row code {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  color: #334155;
  overflow-wrap: anywhere;
  padding: 4px 7px;
}

.namespace-confirm-actions {
  border-top: 1px solid var(--pw-border);
  justify-content: flex-end;
  margin: 0;
  padding: 14px 20px 18px;
}

.management-backdrop {
  align-items: center;
  padding: 24px;
}

.management-dialog {
  max-width: 520px;
  padding: 22px;
}

.management-dialog.compact {
  max-width: 440px;
}

.management-dialog .modal-header {
  margin-bottom: 14px;
}

.management-dialog .modal-actions {
  margin-top: 16px;
}

.management-dialog .language-dialog-actions {
  justify-content: center;
}

.language-dialog-actions .primary-button,
.language-dialog-actions .secondary-button {
  min-width: 104px;
}

.language-choice-list {
  display: grid;
  gap: 10px;
}

.language-choice {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  color: var(--pw-primary);
  display: grid;
  gap: 4px;
  min-height: 74px;
  padding: 14px;
  text-align: left;
}

.language-choice.selected {
  background: #eef1fe;
  border-color: var(--pw-accent);
  box-shadow: 0 0 0 2px rgba(58, 91, 217, 0.14);
}

.language-choice span,
.dialog-copy {
  color: var(--pw-muted);
  line-height: 1.45;
}

.dialog-copy {
  margin: 0;
}

.about-product-mark {
  align-items: center;
  background: var(--pw-accent);
  border-radius: 14px;
  color: #fff;
  display: flex;
  font-size: 1.5rem;
  font-weight: 900;
  height: 56px;
  justify-content: center;
  width: 56px;
}

.compact-about-dialog {
  max-width: 420px;
  padding: 18px;
  position: relative;
}

.about-close-button {
  background: #fff;
  border: 1px solid var(--pw-border);
  border-radius: 999px;
  color: var(--pw-primary);
  font-weight: 900;
  min-height: 38px;
  padding: 0 18px;
  position: absolute;
  right: 18px;
  top: 18px;
}

.about-summary {
  align-items: center;
  display: flex;
  gap: 14px;
  padding-right: 96px;
}

.about-summary h2 {
  font-size: 1.65rem;
  letter-spacing: 0;
  line-height: 1.08;
  margin: 0;
}

.about-summary .eyebrow {
  margin-bottom: 4px;
}

.about-version-grid {
  background: var(--pw-surface-muted);
  border: 1px solid var(--pw-border);
  border-radius: 10px;
  display: grid;
  gap: 0;
  margin: 16px 0 0;
  overflow: hidden;
}

.about-version-grid div {
  align-items: center;
  display: grid;
  gap: 12px;
  grid-template-columns: 96px minmax(0, 1fr);
  min-height: 42px;
  padding: 10px 12px;
}

.about-version-grid div + div {
  border-top: 1px solid var(--pw-border);
}

.about-version-grid dt {
  color: var(--pw-muted);
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.about-version-grid dd {
  color: var(--pw-primary);
  font-weight: 800;
  margin: 0;
  min-width: 0;
  overflow-wrap: anywhere;
}

@media (max-width: 860px) {
  .modal-backdrop {
    padding: 14px;
  }

  .modal-card.namespace-injection-modal {
    max-height: calc(100vh - 28px);
    max-width: 100%;
  }

  .namespace-injection-layout {
    grid-template-columns: 1fr;
  }

  .namespace-injection-list {
    max-height: 260px;
  }

  .namespace-injection-footer {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .modal-backdrop {
    align-items: flex-end;
    padding: 0;
  }

  .modal-card.namespace-injection-modal {
    border-radius: 12px 12px 0 0;
    max-height: 94vh;
  }

  .management-backdrop {
    align-items: center;
    padding: 14px;
  }

  .about-close-button {
    right: 14px;
    top: 14px;
  }

  .about-summary {
    padding-right: 86px;
  }

  .about-version-grid div {
    gap: 4px;
    grid-template-columns: 1fr;
  }

  .namespace-injection-modal .modal-header {
    align-items: flex-start;
    gap: 12px;
    padding: 16px;
  }

  .namespace-injection-form {
    padding: 14px 16px 0;
  }

  .namespace-injection-panel {
    padding: 12px;
  }

  .namespace-injection-panel-title {
    display: grid;
  }

  .namespace-injection-panel-title span {
    justify-self: start;
    max-width: 100%;
    text-align: left;
  }

  .namespace-injection-row {
    grid-template-columns: 1fr;
  }

  .namespace-injection-row-badges {
    justify-content: flex-start;
  }

  .namespace-injection-badge {
    justify-self: start;
    max-width: 100%;
  }

  .namespace-injection-footer {
    margin: 0 -16px;
    padding: 12px 16px 16px;
  }

  .namespace-injection-modal .modal-actions {
    display: grid;
    gap: 8px;
    grid-template-columns: 1fr;
  }

  .namespace-injection-modal .modal-actions .primary-button,
  .namespace-injection-modal .modal-actions .secondary-button {
    width: 100%;
  }

  .namespace-confirm-backdrop {
    align-items: flex-end;
    padding: 0;
  }

  .namespace-confirm-dialog {
    border-radius: 12px 12px 0 0;
    max-height: 92vh;
    max-width: 100%;
  }

  .namespace-confirm-body {
    max-height: calc(92vh - 150px);
    overflow: auto;
    padding: 16px;
  }

  .namespace-confirm-summary {
    grid-template-columns: 1fr;
  }

  .namespace-confirm-actions {
    display: grid;
    gap: 8px;
    grid-template-columns: 1fr;
    padding: 12px 16px 16px;
  }

  .namespace-confirm-actions .primary-button,
  .namespace-confirm-actions .secondary-button {
    width: 100%;
  }
}
</style>
