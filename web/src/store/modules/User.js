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
  user: {
    username: '',
    name: '',
    email: '',
    uid: ''
  },
  users: []
}

// getters
const getters = {
  User_GetStatus: (state) => {
    return state.status;
  },
  User_GetErrorHandle: (state) => {
    return state.error_handle;
  },
  User_GetMeta: (state) => {
    return state.meta;
  },
  User_GetItem: (state) => {
    return state.user;
  },
  User_GetItems: (state) => {
    return state.users;
  },
}

export default {
    state: state,
    getters: getters,
    actions: actions.User,
    mutations: mutations.User
}
