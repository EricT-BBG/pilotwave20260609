import { v4 as uuidv4 } from 'uuid';

const AuthRequest_ResetStatus = (state) => {
  state.status = '';
  state.error_handle = '';
}

const AuthRequest_SetStatus = (state, payload) => {
  state.status = payload.status;
  state.error_handle = payload.error_handle;
}

const AuthRequest_GetItem = (state, payload) => {
  state.jwtRules = payload.authRequest.jwtRules;
  state.labels = payload.authRequest.selectorMatchLabels;
  state.authRequest = payload.authRequest;
  state.resourceVersion = payload.authRequest.resourceversion || payload.authRequest.resourceVersion || '';
}

const AuthRequest_GetItems = (state, payload) => {
  let authRequests = [];
  if (payload.authRequests) {
    for (let index in payload.authRequests) {
      let labels = [];
      let labelItems = payload.authRequests[index].selectorMatchLabels;
      for (let i in labelItems) {
        labels.push(labelItems[i].key + ' (' + labelItems[i].value + ')')
      }

      authRequests.push({
        id: uuidv4(), // For istio api
        name: payload.authRequests[index].name,
        ruleCount: payload.authRequests[index].jwtRules.length,
        labels: labels.join(', '),
        namespace: payload.authRequests[index].namespace,
        createdAt: payload.authRequests[index].createdAt,
      })
    }
  }

  state.authRequests = authRequests;
  state.meta = payload.meta;
}

const AuthRequest_FilterItems = (state, payload) => {
  let authRequests = [];
  if (payload != '') {
    authRequests = state.authMenuRequests.map(ele => {
      if (ele.name.indexOf(payload) >= 0) {
        return ele;
      }
    }).filter(notUndefined => notUndefined !== undefined);
    
    state.authRequests = authRequests;
    state.meta =  { page: 1, limit: 0, total: authRequests.length }
  } else {
    state.authRequests = state.authMenuRequests;
    state.meta =  { page: 1, limit: 0, total: state.authMenuRequests.length }
  }
}

const AuthRequest_GetMenuItems = (state, payload) => {
  let authRequests = [];
  if (payload.authRequests) {
    for (let index in payload.authRequests) {
      let labels = [];
      let labelItems = payload.authRequests[index].selectorMatchLabels;
      for (let i in labelItems) {
        labels.push(labelItems[i].key + ' (' + labelItems[i].value + ')')
      }

      authRequests.push({
        id: uuidv4(), // For istio api
        name: payload.authRequests[index].name,
        ruleCount: payload.authRequests[index].jwtRules.length,
        labels: labels.join(', '),
        namespace: payload.authRequests[index].namespace,
        createdAt: payload.authRequests[index].createdAt,
      })
    }
  }

  state.authMenuRequests = authRequests;
}

const AuthRequest_AddRules = (state) => {
  state.jwtRules.push({
    issuer: '',
    jwksUri: '',
    audiences: [],
  });
}

const AuthRequest_RemoveRule = (state, payload) => {
  state.jwtRules.splice(payload.ruleIndex, 1);
}

const AuthRequest_AddLabels = (state) => {
  state.labels.push({
    key: '',
    value: ''
  });
}

const AuthRequest_RemoveLabel = (state, payload) => {
  state.labels.splice(payload.index, 1);
}

const AuthRequest_ResetData = (state) => {
  state.jwtRules = [{
    issuer: '',
    jwksUri: '',
    audiences: [],
  }]
  state.labels = [{
    key: '',
    value: ''
  }]
}

const AuthRequest_UpdateRule = (state, payload) => {
  let value = payload.value;
  state.jwtRules[payload.ruleIndex][payload.key] = value;
}

const AuthRequest_UpdateLabel = (state, payload) => {
  let value = payload.value;
  state.labels[payload.index][payload.key] = value;
}

export default {
  AuthRequest_ResetStatus,
  AuthRequest_SetStatus,
  AuthRequest_GetItem,
  AuthRequest_GetItems,
  AuthRequest_GetMenuItems,
  AuthRequest_AddRules,
  AuthRequest_RemoveRule,
  AuthRequest_AddLabels,
  AuthRequest_RemoveLabel,
  AuthRequest_ResetData,
  AuthRequest_UpdateRule,
  AuthRequest_UpdateLabel,
  AuthRequest_FilterItems
}
