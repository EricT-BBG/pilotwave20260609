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
  authPolicy: '',
  resourceVersion: '',
  authPolicys: [],
  authMenuPolicys: [],
  rules: [{
    from: [
      {
        source: {
          principals: [],
          notPrincipals: [],
          requestPrincipals: [],
          notRequestPrincipals: [],
          namespaces: [],
          notNamespaces: [],
          ipBlocks: [],
          notIpBlocks: [],
          remoteIpBlocks: [],
          notRemoteIpBlocks: []
        }
      }
    ],
    to: [
      {
        operation: {
          hosts: [],
          notHosts: [],
          ports:[],
          notPorts: [],
          methods: [],
          notMethods: [],
          paths: [],
          notPaths: []
        }
      }
    ],
    when: [
      // {
      //  key: '',
      //  values: [],
      //  notValues: []
      // }
    ]
  }],
  labels: [
    // {
    //  key: '',
    //  value: ''
    // }
  ]
}

// getters
const getters = {
  AuthPolicy_GetStatus: (state) => {
    return state.status;
  },
  AuthPolicy_GetErrorHandle: (state) => {
    return state.error_handle;
  },
  AuthPolicy_GetMeta: (state) => {
    return state.meta;
  },
  AuthPolicy_GetItem: (state) => {
    return state.authPolicy;
  },
  AuthPolicy_GetItems: (state) => {
    return state.authPolicys;
  },
  AuthPolicy_GetMenuItems: (state) => {
    return state.authMenuPolicys;
  },
  AuthPolicy_GetRules: (state) => {
    return state.rules;
  },
  AuthPolicy_GetLabels: (state) => {
    return state.labels;
  },
  AuthPolicy_GetResourceVersion: (state) => {
    return state.resourceVersion;
  },
}

export default {
    state: state,
    getters: getters,
    actions: actions.AuthPolicy,
    mutations: mutations.AuthPolicy
}
