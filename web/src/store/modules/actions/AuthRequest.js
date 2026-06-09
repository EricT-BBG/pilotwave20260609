import axios from 'axios';
import { buildApiUrl } from '../../../lib/runtime';

const getErrorMessage = (err) => {
  const data = err.response?.data;
  return data?.error || data || '';
}

const filterMatchLabels = (labels = []) => {
  return (labels || []).filter((item) => item?.key && item?.value);
}

const AuthRequest_GetItems = ({ commit }, payload) => {
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
    axios.get(buildApiUrl('/security/requestauths'), data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('AuthRequest_GetItems', {
            authRequests: res.data.results || [],
            meta: res.data.meta || { page:1, limit: 0, total: 0 }
          });

          resolve(res.data.results);
        }
      }, err => {
        console.log('err', err.response);
        commit('AuthRequest_GetItems', {
          authRequests: [],
          meta: { page:1, limit: 0, total: 0 },
        });

        resolve([]);
      });
  });
}

const AuthRequest_GetMenuItems = ({ commit }, payload) => {
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
    axios.get(buildApiUrl('/security/requestauths'), data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('AuthRequest_GetMenuItems', {
            authRequests: res.data.results || [],
            meta: res.data.meta || { page:1, limit: 0, total: 0 }
          });

          resolve(res.data.results);
        }
      }, err => {
        console.log('err', err.response);
        commit('AuthRequest_GetMenuItems', {
          authRequests: [],
          meta: { page:1, limit: 0, total: 0 },
        });

        resolve([]);
      });
  });
}

const AuthRequest_GetItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;
  
  return new Promise((resolve) => {
    axios.get(buildApiUrl('/security/requestauth/' + payload.namespace + '/' + payload.name))
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('AuthRequest_GetItem', {
            authRequest: res.data || '',
          });

          resolve(res.data);
        }
      }, err => {
        console.log('err', err.response);
        commit('AuthRequest_GetItem', {
          authRequest: '',
        });

        resolve('');
      });
  });
}

const AuthRequest_NewItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    name: payload.name,
    namespace: payload.namespace,
    jwtRules: payload.rules,
    selectorMatchLabels: filterMatchLabels(payload.labels)
  }

  axios.post(buildApiUrl('/security/requestauths'), data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('AuthRequest_SetStatus', {
          status: 'create_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;

      commit('AuthRequest_SetStatus', {
        status: 'create_error',
        error_handle: errMsg
      });
    });
}

const AuthRequest_UpdateItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    jwtRules: payload.rules,
    selectorMatchLabels: filterMatchLabels(payload.labels),
    resourceversion: payload.resourceVersion || '',
  }

  axios.put(buildApiUrl('/security/requestauth/' + payload.namespace + '/' + payload.name), data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('AuthRequest_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
      }
    }, err => {
      const errMsg = getErrorMessage(err);

      commit('AuthRequest_SetStatus', {
        status: err.response?.status === 409 ? 'update_conflict' : 'update_error',
        error_handle: errMsg
      });
    });
}

const AuthRequest_DelItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;
  
  axios.delete(buildApiUrl('/security/requestauth/' + payload.namespace + '/' + payload.name))
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('AuthRequest_SetStatus', {
          status: 'delete_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;
      commit('AuthRequest_SetStatus', {
        status: 'delete_error',
        error_handle: errMsg
      });
    });
}

export default {
  AuthRequest_GetItems,
  AuthRequest_GetMenuItems,
  AuthRequest_GetItem,
  AuthRequest_NewItem,
  AuthRequest_UpdateItem,
  AuthRequest_DelItem,
}
