import actions from './actions';
import mutations from './mutations';

// initial state
const state = {
  error_handle: null,
  status: null,
  protocols: [
    {
      name: 'HTTP',
      value: 'HTTP'
    },
    {
      name: 'HTTP2',
      value: 'HTTP2'
    },
    {
      name: 'HTTPS',
      value: 'HTTPS'
    },
    {
      name: 'gRPC',
      value: 'GRPC'
    },
    {
      name: 'TCP',
      value: 'TCP'
    },
    {
      name: 'UDP',
      value: 'UDP'
    },
    {
      name: 'TLS/SSL',
      value: 'TLS'
    },
  ],
  bwlist: [],
  meta: {
    page: 1,
    limit: 0,
    total: 0
  },
  gateway: '',
  selectorMatchLabels: {},
  resourceVersion: '',
  gateways: [],
  gatewayMenu: [],
  tlsCertificates: [],
  mappings: [],
  mappingResourceVersions: {},
  servers: [
    {
      hosts: [],
      ports: [{
        protocol: 'HTTP',
        port: 80,
        cert: '',
        pkey: '',
        cacert: '',
        name: '',
        credentialname: '',
        mode: ''
      }]
    }
  ]
}

// getters
const getters = {
  Gateway_GetStatus: (state) => {
    return state.status;
  },
  Gateway_GetErrorHandle: (state) => {
    return state.error_handle;
  },
  Gateway_GetBlackWhiteList: (state) => {
    return state.bwlist;
  },
  Gateway_GetMeta: (state) => {
    return state.meta;
  },
  Gateway_GetItem: (state) => {
    return state.gateway;
  },
  Gateway_GetItems: (state) => {
    return state.gateways;
  },
  Gateway_GetMenuItems: (state) => {
    return state.gatewayMenu;
  },
  Gateway_GetTLSCertificates: (state) => {
    return state.tlsCertificates;
  },
  Gateway_GetProtocols: (state) => {
    return state.protocols;
  },
  Gateway_GetMappings: (state) => {
    return state.mappings;
  },
  Gateway_GetMappingResourceVersions: (state) => {
    return state.mappingResourceVersions;
  },
  Gateway_GetServers: (state) => {
    return state.servers;
  },
  Gateway_GetSelectorMatchLabels: (state) => {
    return state.selectorMatchLabels;
  },
  Gateway_GetResourceVersion: (state) => {
    return state.resourceVersion;
  },
}

export default {
    state: state,
    getters: getters,
    actions: actions.Gateway,
    mutations: mutations.Gateway
}
