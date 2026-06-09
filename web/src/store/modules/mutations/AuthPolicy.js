import { v4 as uuidv4 } from 'uuid';

const AuthPolicy_ResetStatus = (state) => {
  state.status = '';
  state.error_handle = '';
}

const AuthPolicy_SetStatus = (state, payload) => {
  state.status = payload.status;
  state.error_handle = payload.error_handle;
}

const AuthPolicy_GetItem = (state, payload) => {
  state.rules = payload.authPolicy.rules;
  state.labels = payload.authPolicy.selectorMatchLabels;
  state.authPolicy = payload.authPolicy;
  state.resourceVersion = payload.authPolicy.resourceversion || payload.authPolicy.resourceVersion || '';
}

const AuthPolicy_GetItems = (state, payload) => {
  let authPolicys = [];
  if (payload.authPolicys) {
    for (let index in payload.authPolicys) {
      let labels = [];
      let labelItems = payload.authPolicys[index].selectorMatchLabels;
      for (let i in labelItems) {
        labels.push(labelItems[i].key + ' (' + labelItems[i].value + ')')
      }

      authPolicys.push({
        id: uuidv4(), // For istio api
        name: payload.authPolicys[index].name,
        ruleCount: payload.authPolicys[index].rules.length,
        labels: labels.join(', '),
        namespace: payload.authPolicys[index].namespace,
        createdAt: payload.authPolicys[index].createdAt,
      })
    }
  }

  state.authPolicys = authPolicys;
  state.meta = payload.meta;
}

const AuthPolicy_FilterItems = (state, payload) => {
  let authPolicys = [];
  if (payload != '') {
    authPolicys = state.authMenuPolicys.map(ele => {
      if (ele.name.indexOf(payload) >= 0) {
        return ele;
      }
    }).filter(notUndefined => notUndefined !== undefined);
    
    state.authPolicys = authPolicys;
    state.meta =  { page: 1, limit: 0, total: authPolicys.length }
  } else {
    state.authPolicys = state.authMenuPolicys;
    state.meta =  { page: 1, limit: 0, total: state.authMenuPolicys.length }
  }
}

const AuthPolicy_GetMenuItems = (state, payload) => {
  let authPolicys = [];
  if (payload.authPolicys) {
    for (let index in payload.authPolicys) {
      let labels = [];
      let labelItems = payload.authPolicys[index].selectorMatchLabels;
      for (let i in labelItems) {
        labels.push(labelItems[i].key + ' (' + labelItems[i].value + ')')
      }

      authPolicys.push({
        id: uuidv4(), // For istio api
        name: payload.authPolicys[index].name,
        ruleCount: payload.authPolicys[index].rules.length,
        labels: labels.join(', '),
        namespace: payload.authPolicys[index].namespace,
        createdAt: payload.authPolicys[index].createdAt,
      })
    }
  }

  state.authMenuPolicys = authPolicys;
}

const AuthPolicy_AddRules = (state) => {
  state.rules.push({
    from: [
      {
        source: {
          principals: [],
          notPrincipals: [],
          requestPrincipals: [],
          notRequestPrincipals: [],
          namespaces: [],
          notNamespaces: [],
          ipBlocks: [],
          notIpBlocks: [],
          remoteIpBlocks: [],
          notRemoteIpBlocks: []
        }
      }
    ],
    to: [
      {
        operation: {
          hosts: [],
          notHosts: [],
          ports: [],
          notPorts: [],
          methods: [],
          notMethods: [],
          paths: [],
          notPaths: []
        }
      }
    ],
    when: []
  });
}

const AuthPolicy_RemoveRule = (state, payload) => {
  state.rules.splice(payload.ruleIndex, 1);
}

const AuthPolicy_AddLabels = (state) => {
  state.labels.push({
    key: '',
    value: ''
  });
}

const AuthPolicy_RemoveLabel = (state, payload) => {
  state.labels.splice(payload.index, 1);
}

const AuthPolicy_UpdateLabel = (state, payload) => {
  let value = payload.value;
  state.labels[payload.index][payload.key] = value;
}

const AuthPolicy_ResetData = (state) => {
  state.rules = [{
    from: [
      {
        source: {
          principals: [],
          notPrincipals: [],
          requestPrincipals: [],
          notRequestPrincipals: [],
          namespaces: [],
          notNamespaces: [],
          ipBlocks: [],
          notIpBlocks: [],
          remoteIpBlocks: [],
          notRemoteIpBlocks: []
        }
      }
    ],
    to: [
      {
        operation: {
          hosts: [],
          notHosts: [],
          ports: [],
          notPorts: [],
          methods: [],
          notMethods: [],
          paths: [],
          notPaths: []
        }
      }
    ],
    when: []
  }]
  state.labels = []
}

const AuthPolicy_AddWhen = (state, payload) => {
  state.rules[payload.ruleIndex].when.push({
    key: '',
    values: [],
    notValues: []
  });
}

const AuthPolicy_RemoveWhen = (state, payload) => {
  state.rules[payload.ruleIndex].when.splice(payload.index, 1);
}

const AuthPolicy_UpdateWhen = (state, payload) => {
  let value = payload.value;
  state.rules[payload.ruleIndex].when[payload.index][payload.key] = value;
}

const AuthPolicy_AddTo = (state, payload) => {
  state.rules[payload.ruleIndex].to.push({
    operation: {
      hosts: [],
      notHosts: [],
      methods: [],
      notMethods: [],
      paths: [],
      notPaths: []
    }
  });
}

const AuthPolicy_RemoveTo = (state, payload) => {
  state.rules[payload.ruleIndex].to.splice(payload.index, 1);
}

const AuthPolicy_UpdateTo = (state, payload) => {
  let value = payload.value;
  state.rules[payload.ruleIndex].to[payload.index].operation[payload.key] = value;
}

const AuthPolicy_AddFrom = (state, payload) => {
  state.rules[payload.ruleIndex].from.push({
    source: {
      principals: [],
      notPrincipals: [],
      requestPrincipals: [],
      notRequestPrincipals: [],
      namespaces: [],
      notNamespaces: [],
      ipBlocks: [],
      notIpBlocks: [],
      remoteIpBlocks: [],
      notRemoteIpBlocks: [],
    }
  });
}

const AuthPolicy_RemoveFrom = (state, payload) => {
  state.rules[payload.ruleIndex].from.splice(payload.index, 1);
}

const AuthPolicy_UpdateFrom = (state, payload) => {
  let value = payload.value;
  state.rules[payload.ruleIndex].from[payload.index].source[payload.key] = value;
}

export default {
  AuthPolicy_ResetStatus,
  AuthPolicy_SetStatus,
  AuthPolicy_GetItem,
  AuthPolicy_GetItems,
  AuthPolicy_GetMenuItems,
  AuthPolicy_AddRules,
  AuthPolicy_RemoveRule,
  AuthPolicy_AddLabels,
  AuthPolicy_RemoveLabel,
  AuthPolicy_UpdateLabel,
  AuthPolicy_ResetData,
  AuthPolicy_AddWhen,
  AuthPolicy_RemoveWhen,
  AuthPolicy_UpdateWhen,
  AuthPolicy_AddTo,
  AuthPolicy_RemoveTo,
  AuthPolicy_UpdateTo,
  AuthPolicy_AddFrom,
  AuthPolicy_RemoveFrom,
  AuthPolicy_UpdateFrom,
  AuthPolicy_FilterItems
}
