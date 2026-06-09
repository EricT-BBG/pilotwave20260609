import axios from 'axios';
import { buildApiUrl } from '../../../lib/runtime';

const User_GetItems = ({ commit }, payload) => {
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
  if (payload.search) {
    data.params.search = payload.search;
  }
  
  return new Promise((resolve) => {
    axios.get(buildApiUrl('/users'), data)
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('User_GetItems', {
            users: res.data.users || [],
            meta: res.data.meta || { page: 1, limit: 0, total: 0 }
          });

          resolve(res.data.users);
        }
      }, err => {
        console.log('err', err.response);
        commit('User_GetItems', {
          users: [],
          meta: { page: 1, limit: 0, total: 0 },
        });

        resolve([]);
      });
  });
}

const User_GetItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;
  
  return new Promise((resolve) => {
    axios.get(buildApiUrl('/user/' + payload.id))
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('User_GetItem', {
            user: res.data || '',
          });

          resolve(res.data);
        }
      }, err => {
        console.log('err', err.response);
        commit('User_GetItem', {
          user: '',
        });

        resolve('');
      });
  });
}

const User_NewItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    name: payload.name,
    username: payload.username,
    password: payload.password,
    email: payload.email,
    permissions: payload.permissions
  }

  axios.post(buildApiUrl('/users'), data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('User_SetStatus', {
          status: 'create_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;

      commit('User_SetStatus', {
        status: 'create_error',
        error_handle: errMsg
      });
    });
}

const User_UpdateItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    name: payload.name,
    email: payload.email,
    permissions: payload.permissions
  }

  axios.put(buildApiUrl('/user/' + payload.id), data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('User_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;

      commit('User_SetStatus', {
        status: 'update_error',
        error_handle: errMsg
      });
    });
}

const User_UpdatePwd = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  let data = {
    password: payload.password,
  }

  axios.put(buildApiUrl('/user/' + payload.id + '/resetpassword'), data)
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('User_SetStatus', {
          status: 'update_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;

      commit('User_SetStatus', {
        status: 'update_error',
        error_handle: errMsg
      });
    });
}

const User_DelItem = ({ commit }, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;
  
  axios.delete(buildApiUrl('/user/' + payload.id))
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        commit('User_SetStatus', {
          status: 'delete_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;
      commit('User_SetStatus', {
        status: 'delete_error',
        error_handle: errMsg
      });
    });
}

export default {
  User_GetItems,
  User_GetItem,
  User_NewItem,
  User_UpdateItem,
  User_UpdatePwd,
  User_DelItem,
}
