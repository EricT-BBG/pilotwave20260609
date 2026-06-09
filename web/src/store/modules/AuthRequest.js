import actions from './actions';
import mutations from './mutations';

// initial state
const state = {
  error_handle: null,
  status: null,
  meta: {
    page: 1,
    limit: 0,
    total: 0
  },
  authRequest: '',
  resourceVersion: '',
  authRequests: [],
  authMenuRequests: [],
  jwtRules: [{
    issuer: '',
    jwksUri: '',
    audiences: [],
  }],
  labels: [{
    key: '',
    value: ''
  }]
}

// getters
const getters = {
  AuthRequest_GetStatus: (state) => {
    return state.status;
  },
  AuthRequest_GetErrorHandle: (state) => {
    return state.error_handle;
  },
  AuthRequest_GetMeta: (state) => {
    return state.meta;
  },
  AuthRequest_GetItem: (state) => {
    return state.authRequest;
  },
  AuthRequest_GetMenuItem: (state) => {
    return state.authMenuRequests;
  },
  AuthRequest_GetItems: (state) => {
    return state.authRequests;
  },
  AuthRequest_GetJwtRules: (state) => {
    return state.jwtRules;
  },
  AuthRequest_GetLabels: (state) => {
    return state.labels;
  },
  AuthRequest_GetResourceVersion: (state) => {
    return state.resourceVersion;
  },
}

export default {
    state: state,
    getters: getters,
    actions: actions.AuthRequest,
    mutations: mutations.AuthRequest
}
