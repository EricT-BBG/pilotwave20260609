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
  router: '',
  routers: [],
  routerMenu: [],
  mappings: [],
  httpItems: [{
    prefixs: [''],
    headers: [],
    rewrite: '',
    fixedDelay: '',
    timeout: '',
    destinations: [{
      host: '',
      port: '',
      weight: '',
      subset: ''
    }] 
  }],
  ruleResourceVersion: '',
  mappingResourceVersion: '',
  successRate: [],
  successAvg: 0,
  successRequest: 0,
  failRequest: 0,
  totalReqest: 0,
  successHourAvg: 0,
  successHourRequest: 0,
  failHourRequest: 0,
  totalHourReqest: 0,
  latency: [],
  ops: [],
  protocols: [
    {
      name: 'HTTP',
      value: 'http'
    },
    {
      name: 'HTTP2',
      value: 'http2'
    },
    {
      name: 'HTTPS',
      value: 'https'
    },
    {
      name: 'gRPC',
      value: 'grpc'
    },
    {
      name: 'TCP',
      value: 'tcp'
    },
    {
      name: 'UDP',
      value: 'udp'
    },
    {
      name: 'TLS/SSL',
      value: 'tls'
    },
  ],
  grafana: {
    configured: false,
    provider: 'grafana',
    host: '',
    port: '',
    token: '',
    datasourceId: '1',
    isTls: false,
    skipTlsVerify: false
  }
}

// getters
const getters = {
  Router_GetStatus: (state) => {
    return state.status;
  },
  Router_GetErrorHandle: (state) => {
    return state.error_handle;
  },
  Router_GetMeta: (state) => {
    return state.meta;
  },
  Router_GetItem: (state) => {
    return state.router;
  },
  Router_GetProtocols: (state) => {
    return state.protocols;
  },
  Router_GetItems: (state) => {
    return state.routers;
  },
  Router_GetMenuItems: (state) => {
    return state.routerMenu;
  },
  Router_GetHttpItems: (state) => {
    return state.httpItems;
  },
  Router_GetRuleResourceVersion: (state) => {
    return state.ruleResourceVersion;
  },
  Router_GetMappings: (state) => {
    return state.mappings;
  },
  Router_GetMappingResourceVersion: (state) => {
    return state.mappingResourceVersion;
  },
  Router_GetSuccessRate: (state) => {
    return state.successRate;
  },
  Router_GetSuccessAvg: (state) => {
    return state.successAvg;
  },
  Router_GetSuccessRequest: (state) => {
    return state.successRequest;
  },
  Router_GetFailRequest: (state) => {
    return state.failRequest;
  },
  Router_GetTotalRequest: (state) => {
    return state.totalReqest;
  },
  Router_GetSuccessHourAvg: (state) => {
    return state.successHourAvg;
  },
  Router_GetSuccessHourRequest: (state) => {
    return state.successHourRequest;
  },
  Router_GetFailHourRequest: (state) => {
    return state.failHourRequest;
  },
  Router_GetTotalHourRequest: (state) => {
    return state.totalHourReqest;
  },
  Router_GetLatency: (state) => {
    return state.latency;
  },
  Router_GetOPS: (state) => {
    return state.ops;
  },
  Router_GetGrafana: (state) => {
    return state.grafana;
  },
}

export default {
    state: state,
    getters: getters,
    actions: actions.Router,
    mutations: mutations.Router
}
