import { normalizeLocale, SUPPORTED_LOCALES } from '../../../lib/locale';

const Auth_ResetStatus = (state) => {
  state.status = '';
  state.error_handle = '';
}

const Auth_SetStatus = (state, payload) => {
  state.status = payload.status;
  state.error_handle = payload.error_handle;

  let user = sessionStorage.getItem('member');
  if (user) state.user = JSON.parse(user);
}

const Auth_SetLanguage = (state, payload) => {
  const lang = normalizeLocale(payload.lang);
  if (!SUPPORTED_LOCALES.includes(lang))
    return;

  state.language = lang;
  sessionStorage.setItem('locale', lang);
}

const getNamespaceName = (item) => {
  if (typeof item === 'string') return item;
  return item?.name || item?.namespace || item?.metadata?.name || '';
}

const getNamespaceLabels = (item) => {
  return item?.labels || item?.metadata?.labels || {};
}

const getNamespaceInjection = (item) => {
  const labels = getNamespaceLabels(item);
  const injection = item?.istioInjection || item?.istio_injection || item?.injection || {};
  const revision = injection.revision || item?.revision || labels['istio.io/rev'] || '';
  let mode = injection.mode || item?.mode || '';

  if (!mode && typeof injection === 'string') mode = injection;
  if (!mode && labels['istio-injection'] === 'enabled') mode = 'enabled';
  if (!mode && labels['istio-injection'] === 'disabled') mode = 'disabled';
  if (!mode && revision) mode = 'revision';
  if (!mode) mode = 'disabled';

  const status = injection.status || item?.status || (mode === 'revision' && revision ? `revision:${revision}` : mode);

  return {
    mode,
    revision,
    status,
  };
}

const normalizeNamespace = (item) => {
  const name = getNamespaceName(item);
  const labels = getNamespaceLabels(item);
  const normalized = {
    name,
    istioInjection: getNamespaceInjection(item),
    systemNamespace: item?.systemNamespace === true || item?.system_namespace === true,
    raw: item,
  };

  if (Object.keys(labels).length) normalized.labels = labels;

  return normalized;
}

const Auth_GetNamespaces = (state, payload) => {
  let namespaces = [];
  let namespaceDetails = [];
  if (payload.namespaces.length) {
    for (let i in payload.namespaces) {
      const namespace = normalizeNamespace(payload.namespaces[i]);
      if (!namespace.name) continue;
      namespaces.push(namespace.name);
      namespaceDetails.push(namespace);
    }
    namespaces.push('All');
  }
  state.namespaces = namespaces;
  state.namespaceDetails = namespaceDetails;
}

const normalizeClusterCapabilities = (payload = {}) => {
  const istio = payload.istio || {};

  return {
    ...payload,
    istio: {
      installed: istio.installed !== false,
      disabled: istio.disabled === true,
      missingCRDs: Array.isArray(istio.missingCRDs) ? istio.missingCRDs : [],
      availableCRDs: Array.isArray(istio.availableCRDs) ? istio.availableCRDs : [],
      defaultInjectionAvailable: istio.defaultInjectionAvailable === true,
      revisionInjectionAvailable: istio.revisionInjectionAvailable === true,
      revisions: Array.isArray(istio.revisions) ? istio.revisions : [],
      revisionTags: Array.isArray(istio.revisionTags) ? istio.revisionTags : [],
      revisionDetectionMessage: istio.revisionDetectionMessage || '',
      message: istio.message || '',
    },
  };
}

const Auth_SetClusterCapabilities = (state, payload) => {
  state.clusterCapabilities = normalizeClusterCapabilities(payload);
}

const Auth_SetNamespace = (state, payload) => {
  state.namespace = payload || 'All';
}

export default {
  Auth_SetNamespace,
  Auth_GetNamespaces,
  Auth_SetClusterCapabilities,
  Auth_ResetStatus,
  Auth_SetStatus,
  Auth_SetLanguage
}
