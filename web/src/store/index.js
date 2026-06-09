import { createStore } from 'vuex';
import Auth from './modules/Auth';
import User from './modules/User';
import Gateway from './modules/Gateway';
import Router from './modules/Router';
import AuthRequest from './modules/AuthRequest';
import AuthPolicy from './modules/AuthPolicy';

export default createStore({
  modules: {
    Auth,
    User,
    Gateway,
    Router,
    AuthRequest,
    AuthPolicy
  }
});
