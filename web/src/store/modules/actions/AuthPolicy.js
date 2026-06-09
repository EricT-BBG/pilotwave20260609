import axios from 'axios';
import { buildApiUrl } from '../../../lib/runtime';

const getErrorMessage = (err) => {
  const data = err.response?.data;
  return data?.error || data || '';
}

const filterMatchLabels = (labels = []) => {
  return (labels || []).filter((item) => item?.key && item?.value);
}

const AuthPolicy_GetItems = ({ commit }, payload) => {
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
    axios.get(buildApiUrl('/security/authpolicies'), data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('AuthPolicy_GetItems', {
            authPolicys: res.data.results || [],
            meta: res.data.meta || { page:1, limit: 0, total: 0 }
          });

          resolve(res.data.results);
        }
      }, err => {
        console.log('err', err.response);
        commit('AuthPolicy_GetItems', {
          authPolicys: [],
          meta: { page:1, limit: 0, total: 0 },
        });

        resolve([]);
      });
  });
}

const AuthPolicy_GetMenuItems = ({ commit }, payload) => {
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
    axios.get(buildApiUrl('/security/authpolicies'), data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('AuthPolicy_GetMenuItems', {
            authPolicys: res.data.results || [],
          });

          resolve(res.data.results);
        }
      }, err => {
        console.log('err', err.response);
        commit('AuthPolicy_GetMenuItems', {
          authPolicys: [],
        });

        resolve([]);
      });
  });
}

const AuthPolicy_GetItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  return new Promise((resolve) => {
    axios.get(buildApiUrl('/security/authpolicy/' + payload.namespace + '/' + payload.name))
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('AuthPolicy_GetItem', {
            authPolicy: res.data || '',
          });

          resolve(res.data);
        }
      }, err => {
        console.log('err', err.response);
        commit('AuthPolicy_GetItem', {
          authPolicy: '',
        });

        resolve('');
      });
  });
}

const AuthPolicy_NewItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    name: payload.name,
    namespace: payload.namespace,
    action: payload.action,
    rules: payload.rules,
    selectorMatchLabels: filterMatchLabels(payload.labels)
  }

  axios.post(buildApiUrl('/security/authpolicies'), data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('AuthPolicy_SetStatus', {
          status: 'create_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;

      commit('AuthPolicy_SetStatus', {
        status: 'create_error',
        error_handle: errMsg
      });
    });
}

const AuthPolicy_UpdateItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    action: payload.action,
    rules: payload.rules,
    selectorMatchLabels: filterMatchLabels(payload.labels),
    resourceversion: payload.resourceVersion || '',
  }

  axios.put(buildApiUrl('/security/authpolicy/' + payload.namespace + '/' + payload.name), data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('AuthPolicy_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
      }
    }, err => {
      const errMsg = getErrorMessage(err);
      
      commit('AuthPolicy_SetStatus', {
        status: err.response?.status === 409 ? 'update_conflict' : 'update_error',
        error_handle: errMsg
      });
    });
}

const AuthPolicy_DelItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;
  
  axios.delete(buildApiUrl('/security/authpolicy/' + payload.namespace + '/' + payload.name))
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('AuthPolicy_SetStatus', {
          status: 'delete_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;
      commit('AuthPolicy_SetStatus', {
        status: 'delete_error',
        error_handle: errMsg
      });
    });
}

export default {
  AuthPolicy_GetItems,
  AuthPolicy_GetMenuItems,
  AuthPolicy_GetItem,
  AuthPolicy_NewItem,
  AuthPolicy_UpdateItem,
  AuthPolicy_DelItem,
}
