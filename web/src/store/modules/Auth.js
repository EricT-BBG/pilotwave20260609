import actions from './actions';
import mutations from './mutations';
import { normalizeLocale, resolveLocale, SUPPORTED_LOCALES } from '../../lib/locale';

// initial state
const state = {
  error_handle: null,
  status: null,
  language: resolveLocale(),
  user: '',
  namespace: 'All',
  namespaces: [],
  namespaceDetails: [],
  clusterCapabilities: {
    istio: {
      installed: true,
      disabled: false,
      missingCRDs: [],
      availableCRDs: [],
      defaultInjectionAvailable: false,
      revisionInjectionAvailable: false,
      revisions: [],
      revisionTags: [],
      revisionDetectionMessage: '',
      message: '',
    },
  },
}

// getters
const getters = {
  Auth_GetStatus: (state) => {
    return state.status;
  },
  Auth_GetErrorHandle: (state) => {
    return state.error_handle;
  },
  Auth_GetNamespace: (state) => {
    return state.namespace;
  },
  Auth_GetNamespaces: (state) => {
    return state.namespaces;
  },
  Auth_GetNamespaceDetails: (state) => {
    return state.namespaceDetails;
  },
  Auth_GetClusterCapabilities: (state) => {
    return state.clusterCapabilities;
  },
  Auth_GetIstioCapabilities: (state) => {
    return state.clusterCapabilities?.istio || {
      installed: true,
      disabled: false,
      missingCRDs: [],
      availableCRDs: [],
      defaultInjectionAvailable: false,
      revisionInjectionAvailable: false,
      revisions: [],
      revisionTags: [],
      revisionDetectionMessage: '',
      message: '',
    };
  },
  Auth_GetLanguage: (state) => {
    let lang = normalizeLocale(sessionStorage.getItem('locale'));
    if (SUPPORTED_LOCALES.includes(lang)) state.language = lang;

    return state.language;
  },
  Auth_GetUserInfo: (state) => {
    if (state.user)
      return state.user;

    let user = sessionStorage.getItem('member');
    if (user) user = JSON.parse(user);

    return user;
  }
}

export default {
  state: state,
  getters: getters,
  actions: actions.Auth,
  mutations: mutations.Auth
}
