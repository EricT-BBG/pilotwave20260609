import axios from 'axios';
import { buildApiUrl } from '../../../lib/runtime';

const Auth_Signin = ({ commit }, payload) => {
  let data = {
    username: payload.account.trim(),
    password: payload.password.trim()
  }

  axios.post(buildApiUrl('/auth/signin'), data, {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    }
  })
    .then(res => {
      if (res.status >= 200 && res.status < 300) {
        sessionStorage.clear();
        sessionStorage.setItem('member', JSON.stringify(res.data));
        sessionStorage.setItem('accessToken', res.data.token);

        commit('Auth_SetStatus', {
          status: 'signin_success',
          error_handle: ''
        });
      }
    }, err => {
      let errMsg = '';
      if (err.response) errMsg = err.response.data;

      commit('Auth_SetStatus', {
        status: 'signin_error',
        error_handle: errMsg
      });
    });
}

const Auth_GetNamespaces = ({ commit }) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  return new Promise((resolve) => {
    axios.get(buildApiUrl('/namespaces'))
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          const namespaces = res.data.namespaces || [];
          const namespaceItems = res.data.items || namespaces;
          commit('Auth_GetNamespaces', {
            namespaces: namespaceItems,
          });

          resolve(namespaces);
        }
      }, err => {
        console.log('err', err.response);
        commit('Auth_GetNamespaces', {
          namespaces: [],
        });

        resolve([]);
      });
  });
}

const Auth_GetClusterCapabilities = ({ commit }) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  return new Promise((resolve) => {
    axios.get(buildApiUrl('/cluster/capabilities'))
      .then(res => {
        if (res.status >= 200 && res.status < 300) {
          commit('Auth_SetClusterCapabilities', res.data || {});
          resolve(res.data || {});
        }
      }, err => {
        console.log('err', err.response);
        const fallback = {
          istio: {
            installed: true,
            disabled: false,
            missingCRDs: [],
            availableCRDs: [],
            message: '',
          },
        };
        commit('Auth_SetClusterCapabilities', fallback);
        resolve(fallback);
      });
  });
}

const Auth_UpdateNamespaceInjection = (context, payload) => {
  const accessToken = sessionStorage.getItem('accessToken');
  axios.defaults.headers.common['Authentication'] = accessToken;

  const data = {
    mode: payload.mode,
    revision: payload.mode === 'revision' ? (payload.revision || '').trim() : '',
    allowSystemNamespace: Boolean(payload.allowSystemNamespace),
  };

  return axios.patch(buildApiUrl('/namespace/' + encodeURIComponent(payload.name) + '/istio-injection'), data, {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    }
  }).then(res => res.data);
}

export default {
  Auth_Signin,
  Auth_GetNamespaces,
  Auth_GetClusterCapabilities,
  Auth_UpdateNamespaceInjection
}
