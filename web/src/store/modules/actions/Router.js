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

const inFlightMetricRequests = new Map();

const metricRequestKey = (url, data) => {
  return `${url}?${JSON.stringify(data?.params || {})}`;
}

const getMetricResponse = (url, data) => {
  const key = metricRequestKey(url, data);
  if (inFlightMetricRequests.has(key)) {
    return inFlightMetricRequests.get(key);
  }

  const request = axios.get(url, data)
    .finally(() => {
      inFlightMetricRequests.delete(key);
    });
  inFlightMetricRequests.set(key, request);
  return request;
}

const Router_GetItems = ({ commit }, payload) => {
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
    axios.get(api_url + api_version + '/routers', data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetItems', {
            routers: res.data.routers || [],
            meta: res.data.meta || { page: 1, limit: 0, total: 0 }
          });

          resolve(res.data.routers);
        }
      }, err => {
        console.log('err', err.response);
        commit('Router_GetItems', {
          routers: [],
          meta: { page: 1, limit: 0, total: 0 },
        });

        resolve([]);
      });
  });
}

const Router_GetMenuItems = ({ commit }, payload) => {
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
    axios.get(api_url + api_version + '/routers', data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetMenuItems', {
            routers: res.data.routers || [],
          });

          resolve(res.data.routers);
        }
      }, err => {
        console.log('err', err.response);
        commit('Router_GetMenuItems', {
          routers: [],
        });

        resolve([]);
      });
  });
}

const Router_GetItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetItem', {
            router: res.data || '',
          });

          resolve(res.data);
        }
      }, err => {
        console.log('err', err.response);
        commit('Router_GetItem', {
          router: '',
        });

        resolve('');
      });
  });
}

const Router_NewItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    name: payload.name,
    protocol: payload.protocol,
    namespace: payload.namespace,
    hosts: payload.hosts,
  }

  return axios.post(api_url + api_version  + '/routers', data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        const gateways = payload.gateways || [];
        if (!gateways.length) {
          commit('Router_SetStatus', {
            status: 'create_success',
            error_handle: ''
          });
          return res.data;
        }

        return axios.put(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name + '/gateways', {
          gateways,
          resourceversion: ''
        }).then(mappingRes => {
          if (mappingRes.status >= 200 && mappingRes.status < 300) {
            commit('Router_SetStatus', {
              status: 'create_success',
              error_handle: ''
            });
          }
          return mappingRes.data;
        });
      }
    }, err => {
      const errMsg = getErrorMessage(err);

      commit('Router_SetStatus', {
        status: 'create_error',
        error_handle: errMsg
      });
    }).catch(err => {
      const errMsg = getErrorMessage(err);

      commit('Router_SetStatus', {
        status: 'create_error',
        error_handle: errMsg
      });
    });
}

const Router_UpdateItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    protocol: payload.protocol,
    hosts: payload.hosts,
    resourceversion: payload.resourceVersion || '',
  }
  
  axios.put(api_url + api_version  + '/router/' + payload.namespace + '/' + payload.name, data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Router_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
      }
    }, err => {
      const errMsg = getErrorMessage(err);

      commit('Router_SetStatus', {
        status: err.response?.status === 409 ? 'update_conflict' : 'update_error',
        error_handle: errMsg
      });
    });
}

const Router_DelItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;
  
  axios.delete(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Router_SetStatus', {
          status: 'delete_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;
      commit('Router_SetStatus', {
        status: 'delete_error',
        error_handle: errMsg
      });
    });
}

const Router_GetMappings = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name + '/gateways')
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetMappings', {
            mappings: res.data.gateways || [],
            gateways: payload.gateways || [],
            resourceVersion: res.data.resourceversion || '',
          });

          resolve(res.data.gateways);
        }
      }, err => {
        console.log('err', err.response);
        commit('Router_GetMappings', {
          mappings: [],
          gateways: [],
          resourceVersion: '',
        });

        resolve([]);
      });
  });
}

const Router_MappingGateways = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    gateways: payload.gateways || [],
    resourceversion: payload.resourceVersion || '',
  }
  
  axios.put(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name + '/gateways', data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Router_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
      }
    }, err => {
      const errMsg = getErrorMessage(err);

      commit('Router_SetStatus', {
        status: err.response?.status === 409 ? 'update_conflict' : 'update_error',
        error_handle: errMsg
      });
    });
}

const Router_DelMapping = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    params: {
      gatewayId: payload.gatewayId,
    }
  }

  axios.delete(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name + '/gateways', data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Router_SetStatus', {
          status: 'delete_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;
      commit('Router_SetStatus', {
        status: 'delete_error',
        error_handle: errMsg
      });
    });
}

const Router_GetRule = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;
  
  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name + '/rules')
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetRule', {
            httpItems: res.data.https || [],
            resourceVersion: res.data.resourceversion || '',
          });

          resolve(res.data.https);
        }
      }, err => {
        console.log('err', err.response);
        commit('Router_GetRule', {
          httpItems: [],
          resourceVersion: '',
        });

        resolve([]);
      });
  });
}

const Router_UpdateRules = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    name: payload.name,
    namespace: payload.namespace,
    https: payload.httpItems,
    resourceversion: payload.resourceVersion || '',
  }
  
  axios.put(api_url + api_version  + '/router/' + payload.namespace + '/' + payload.name + '/rules', data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Router_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
      }
    }, err => {
      const errMsg = getErrorMessage(err);

      commit('Router_SetStatus', {
        status: err.response?.status === 409 ? 'update_conflict' : 'update_error',
        error_handle: errMsg
      });
    });
}

const Router_GetSuccessRate = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    params: {
      startTime: payload.startTime,
      endTime: payload.endTime,
      interval: payload.interval
    }
  }
  
  return new Promise((resolve) => {
    getMetricResponse(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name + '/successrate', data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetSuccessRate', {
            metrics: res.data.metrics || [],
            successRate: res.data.successRate || 0,
            totalSuccessReqests: res.data.totalSuccessReqests || 0,
            totalReqests: res.data.totalReqests || 0
          });

          resolve({
            ok: true,
            metrics: res.data.metrics || [],
            successRate: res.data.successRate || 0,
            totalSuccessReqests: res.data.totalSuccessReqests || 0,
            totalReqests: res.data.totalReqests || 0
          });
        }
      }, err => {
        console.log('err', err.response);
        const message = err.response?.data?.error || 'Unable to load success rate metrics.';
        commit('Router_GetSuccessRate', {
          metrics: [],
          successRate: 0,
          totalSuccessReqests: 0,
          totalReqests: 0
        });

        resolve({ ok: false, message });
      });
  });
}

const Router_GetHourSuccessRate = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    params: {
      startTime: payload.startTime,
      endTime: payload.endTime,
      interval: payload.interval
    }
  }
  
  return new Promise((resolve) => {
    getMetricResponse(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name + '/successrate', data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetHourSuccessRate', {
            successRate: res.data.successRate || 0,
            totalSuccessReqests: res.data.totalSuccessReqests || 0,
            totalReqests: res.data.totalReqests || 0
          });

          resolve({
            ok: true,
            successRate: res.data.successRate || 0,
            totalSuccessReqests: res.data.totalSuccessReqests || 0,
            totalReqests: res.data.totalReqests || 0
          });
        }
      }, err => {
        console.log('err', err.response);
        const message = err.response?.data?.error || 'Unable to load hourly success rate metrics.';
        commit('Router_GetHourSuccessRate', {
          successRate: 0,
          totalSuccessReqests: 0,
          totalReqests: 0
        });

        resolve({ ok: false, message });
      });
  });
}

const Router_GetLatency = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    params: {
      percentage: payload.percentage,
      startTime: payload.startTime,
      endTime: payload.endTime,
      interval: payload.interval
    }
  }
  
  return new Promise((resolve) => {
    getMetricResponse(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name + '/latency', data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetLatency', {
            metrics: res.data.metrics || [],
          });

          resolve({
            ok: true,
            metrics: res.data.metrics || [],
          });
        }
      }, err => {
        console.log('err', err.response);
        const message = err.response?.data?.error || 'Unable to load latency metrics.';
        commit('Router_GetLatency', {
          metrics: [],
        });

        resolve({ ok: false, message });
      });
  });
}

const Router_GetOPS = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    params: {
      startTime: payload.startTime,
      endTime: payload.endTime,
      interval: payload.interval
    }
  }
  
  return new Promise((resolve) => {
    getMetricResponse(api_url + api_version + '/router/' + payload.namespace + '/' + payload.name + '/ops', data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetOPS', {
            metrics: res.data.metrics || [],
          });

          resolve({
            ok: true,
            metrics: res.data.metrics || [],
          });
        }
      }, err => {
        console.log('err', err.response);
        const message = err.response?.data?.error || 'Unable to load OPS metrics.';
        commit('Router_GetOPS', {
          metrics: [],
        });

        resolve({ ok: false, message });
      });
  });
}

const Router_GetGrafana = ({ commit }) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;
  
  return new Promise((resolve) => {
    axios.get(api_url + api_version + '/grafana')
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Router_GetGrafana', {
            grafana: res.data || '',
          });

          resolve({
            grafana: res.data || '',
          });
        }
      }, err => {
        console.log('err', err.response);
        commit('Router_GetGrafana', {
          grafana: '',
        });

        resolve({
            grafana: '',
        });
      });
  });
}

const Router_UpdateGrafana = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    id: payload.id,
    provider: payload.provider || 'grafana',
    host: payload.host,
    port: payload.port,
    token: payload.token,
    datasourceId: payload.datasourceId || '1',
    isTls: payload.isTls,
    skipTlsVerify: payload.skipTlsVerify,
  }

  return new Promise((resolve) => {
    axios.post(api_url + api_version  + '/grafanas', data)
      .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('Router_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
        resolve(res.data);
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;

      commit('Router_SetStatus', {
        status: 'update_error',
        error_handle: errMsg
      });
      resolve('');
    });
  });
}

const Router_TestGrafana = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  const data = {
    provider: payload.provider || 'grafana',
    host: payload.host,
    port: payload.port,
    token: payload.token,
    datasourceId: payload.datasourceId || '1',
    isTls: payload.isTls,
    skipTlsVerify: payload.skipTlsVerify,
  };

  return new Promise((resolve) => {
    axios.post(api_url + api_version + '/monitoring/test', data, { timeout: 8000 })
      .then(res => {
        resolve(res.data || { ok: false, message: 'Empty response from monitoring source test.' });
      }, err => {
        let errMsg = 'Monitoring source test failed.';
        if (err.response?.data?.error) errMsg = err.response.data.error;
        commit('Router_SetStatus', {
          status: 'monitoring_test_error',
          error_handle: errMsg
        });
        resolve({ ok: false, message: errMsg });
      });
  });
}

export default {
  Router_GetItems,
  Router_GetItem,
  Router_GetMenuItems,
  Router_NewItem,
  Router_UpdateItem,
  Router_DelItem,
  Router_GetMappings,
  Router_MappingGateways,
  Router_DelMapping,
  Router_GetRule,
  Router_UpdateRules,
  Router_GetSuccessRate,
  Router_GetHourSuccessRate,
  Router_GetLatency,
  Router_GetOPS,
  Router_GetGrafana,
  Router_UpdateGrafana,
  Router_TestGrafana
}
