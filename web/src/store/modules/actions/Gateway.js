import axios from 'axios';

let api_url=process.env.VUE_APP_API_URL;
const api_version=process.env.VUE_APP_API_VERSION;

if (process.env.VUE_APP_API_URL === '//') {
  api_url = window.location.protocol + '//' +window.location.host;
}

const getErrorMessage = (err) => {
  const data = err.response?.data;
  return data?.error || data || '';
}

const isTlsProtocol = (protocol) => ['HTTPS', 'TLS'].includes(String(protocol || '').toUpperCase());

const encodeGatewayTLSMaterial = (servers) => {
  for (let index in servers) {
    let server = servers[index];
    for (let i in server.ports) {
      const port = server.ports[i];
      if (!isTlsProtocol(port.protocol)) continue;
      if (!port.mode) port.mode = 'SIMPLE';
      if (port.cert) port.cert = window.btoa(port.cert);
      if (port.pkey) port.pkey = window.btoa(port.pkey);
      if (port.cacert) port.cacert = window.btoa(port.cacert);
    }
  }
}

const findGatewayTLSValidationError = (servers) => {
  for (let index in servers) {
    let server = servers[index];
    for (let i in server.ports) {
      const port = server.ports[i];
      if (!isTlsProtocol(port.protocol)) continue;
      const mode = String(port.mode || 'SIMPLE').toUpperCase();
      const hasExistingSecret = Boolean(port.credentialname && !port.cert && !port.pkey && !port.cacert);
      if (hasExistingSecret) continue;
      if (!port.cert) return `Port ${port.port} requires a certificate or existing Secret.`;
      if (!port.pkey) return `Port ${port.port} requires a private key.`;
      if (mode === 'MUTUAL' && !port.cacert) return `Port ${port.port} requires a CA bundle for mTLS.`;
    }
  }
  return '';
}

const Gateway_GetItems = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;
  
  let data = {
    params: {
      page: payload.page || 1,
      limit: payload.limit || 20
    }
  }
  
  if (payload.namespace && payload.namespace !== 'All') {
    data.params.namespace = payload.namespace;
  }

  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/gateways', data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Gateway_GetItems', {
            gateways: res.data.gateways || [],
            meta: res.data.meta || { page: 1, limit: 0, total: 0 }
          });

          resolve(res.data.gateways);
        }
      }, err => {
        console.log('err', err.response);
        commit('Gateway_GetItems', {
          gateways: [],
          meta: { page: 1, limit: 0, total: 0 },
        });

        resolve([]);
      });
  });
}

const Gateway_GetMenuItems = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    params: {
      page: payload.page || 1,
      limit: payload.limit || 20
    }
  }
  
  if (payload.namespace && payload.namespace !== 'All') {
    data.params.namespace = payload.namespace;
  }

  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/gateways', data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Gateway_GetMenuItems', {
            gateways: res.data.gateways || [],
          });

          resolve(res.data.gateways);
        }
      }, err => {
        console.log('err', err.response);
        commit('Gateway_GetMenuItems', {
          gateways: [],
        });

        resolve([]);
      });
  });
}

const Gateway_GetItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/gateway/' + payload.namespace + '/' + payload.name)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Gateway_GetItem', {
            gateway: res.data || '',
          });

          resolve(res.data);
        }
      }, err => {
        console.log('err', err.response);
        commit('Gateway_GetItem', {
          gateway: '',
        });

        resolve('');
      });
  });
}

const Gateway_GetTLSCertificates = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/gateway/' + payload.namespace + '/' + payload.name + '/tls-certificates')
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          const certificates = res.data.certificates || [];
          commit('Gateway_GetTLSCertificates', {
            certificates,
          });

          resolve(certificates);
        }
      }, err => {
        console.log('err', err.response);
        commit('Gateway_GetTLSCertificates', {
          certificates: [],
        });

        resolve([]);
      });
  });
}

const Gateway_CheckTLSSecretExists = (_context, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  return axios.get(api_url + api_version + '/gateway/tls-secret/exists', {
    params: {
      credentialname: payload.credentialname || '',
    }
  }).then(res => res.data || {
    exists: false,
  });
}

const Gateway_NewItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    name: payload.name,
    namespace: payload.namespace,
    servers: [],
  }
  if (payload.selectorMatchLabels) {
    data.selectormatchlabels = payload.selectorMatchLabels;
  }

  let servers = JSON.parse(JSON.stringify(payload.servers));
  const validationError = findGatewayTLSValidationError(servers);
  if (validationError) {
    commit('Gateway_SetStatus', {
      status: 'create_error',
      error_handle: validationError
    });
    return;
  }
  encodeGatewayTLSMaterial(servers);

  data.servers = servers;

  axios.post(api_url + api_version  + '/gateways', data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Gateway_SetStatus', {
          status: 'create_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;

      commit('Gateway_SetStatus', {
        status: 'create_error',
        error_handle: errMsg
      });
    });
}

const Gateway_UpdateItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    servers: [],
    selectormatchlabels: payload.selectorMatchLabels || {},
    resourceversion: payload.resourceVersion || '',
  }

  let servers = JSON.parse(JSON.stringify(payload.servers));
  const validationError = findGatewayTLSValidationError(servers);
  if (validationError) {
    commit('Gateway_SetStatus', {
      status: 'update_error',
      error_handle: validationError
    });
    return;
  }
  encodeGatewayTLSMaterial(servers);

  data.servers = servers;
  
  axios.put(api_url + api_version  + '/gateway/' + payload.namespace + '/' + payload.name, data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Gateway_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
      }
    }, err => {
      const errMsg = getErrorMessage(err);

      commit('Gateway_SetStatus', {
        status: err.response?.status === 409 ? 'update_conflict' : 'update_error',
        error_handle: errMsg
      });
    });
}

const Gateway_DelItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  axios.delete(api_url + api_version + '/gateway/' + payload.namespace + '/' + payload.name)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Gateway_SetStatus', {
          status: 'delete_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;
      commit('Gateway_SetStatus', {
        status: 'delete_error',
        error_handle: errMsg
      });
    });
}

const Gateway_GetBlackWhiteList = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    params: {
      page: payload.page || 1,
      limit: payload.limit || 20
    }
  }

  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/gateway/' + payload.id + '/bwlists', data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Gateway_GetBlackWhiteList', {
            lists: res.data.lists || [],
            meta: res.data.meta || { page:1, limit: 0, total: 0 }
          });

          resolve(res.data.lists);
        }
      }, err => {
        console.log('err', err.response);
        commit('Gateway_GetBlackWhiteList', {
          lists: [],
          meta: { page:1, limit: 0, total: 0 },
        });

        resolve([]);
      });
  });
}

const Gateway_NewBlackWhiteList = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    category: payload.category,
    domain: payload.domain,
    description: payload.description,
  }

  axios.post(api_url + api_version  + '/gateway/' + payload.id + '/bwlists', data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Gateway_SetStatus', {
          status: 'create_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;

      commit('Gateway_SetStatus', {
        status: 'create_error',
        error_handle: errMsg
      });
    });
}

const Gateway_DelBlackWhiteList = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  axios.delete(api_url + api_version + '/bwlist/' + payload.id)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Gateway_SetStatus', {
          status: 'delete_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;
      commit('Gateway_SetStatus', {
        status: 'delete_error',
        error_handle: errMsg
      });
    });
}

const Gateway_GetMappings = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/gateway/' + payload.namespace + '/' + payload.name + '/routers')
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Gateway_GetMappings', {
            mappings: res.data.routers || [],
            routers: payload.routers || [],
            resourceVersions: res.data.resourceversions || {},
          });

          resolve(res.data.routers);
        }
      }, err => {
        console.log('err', err.response);
        commit('Gateway_GetMappings', {
          mappings: [],
          routers: [],
          resourceVersions: {},
        });

        resolve('');
      });
  });
}

const Gateway_MappingRouters = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    routers: payload.routers || [],
    resourceversions: payload.resourceVersions || {},
  }

  axios.put(api_url + api_version  + '/gateway/' + payload.namespace + '/' + payload.name + '/routers', data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Gateway_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
      }
    }, err => {
      const errMsg = getErrorMessage(err);

      commit('Gateway_SetStatus', {
        status: err.response?.status === 409 ? 'update_conflict' : 'update_error',
        error_handle: errMsg
      });
    });
}

export default {
  Gateway_GetMenuItems,
  Gateway_GetItems,
  Gateway_GetItem,
  Gateway_GetTLSCertificates,
  Gateway_CheckTLSSecretExists,
  Gateway_NewItem,
  Gateway_UpdateItem,
  Gateway_DelItem,
  Gateway_GetBlackWhiteList,
  Gateway_NewBlackWhiteList,
  Gateway_DelBlackWhiteList,
  Gateway_GetMappings,
  Gateway_MappingRouters,
}
