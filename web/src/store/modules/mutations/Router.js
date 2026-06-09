import moment from 'moment';
import { v4 as uuidv4 } from 'uuid';

const formatNumber = (value, precision) => {
  return Math.round(Math.round(value * Math.pow(10, (precision || 0) + 1)) / 10) / Math.pow(10, (precision || 0));
}

const Router_ResetStatus = (state) => {
  state.status = '';
  state.error_handle = '';
}

const Router_SetStatus = (state, payload) => {
  state.status = payload.status;
  state.error_handle = payload.error_handle;
}

const Router_GetItem = (state, payload) => {
  state.router = payload.router;
}

const Router_GetItems = (state, payload) => {
  let routers = [];
  if (payload.routers) {
    for (let index in payload.routers) {
      routers.push({
        id: uuidv4(), // For istio api
        name: payload.routers[index].name,
        protocol: payload.routers[index].protocol,
        namespace: payload.routers[index].namespace,
        httpCount: payload.routers[index].httpCount,
        createdAt: payload.routers[index].createdAt,
        resourceversion: payload.routers[index].resourceversion || '',
        text: payload.routers[index].name + ' (' + payload.routers[index].protocol  + ')',
        value: payload.routers[index].name + ',' + payload.routers[index].namespace
      })
    }
  }

  state.routers = routers;
  if (payload.meta) {
    state.meta = payload.meta;
  }
}

const Router_GetMenuItems = (state, payload) => {
  let routers = [];
  if (payload.routers) {
    for (let index in payload.routers) {
      routers.push({
        id: uuidv4(), // For istio api
        name: payload.routers[index].name,
        protocol: payload.routers[index].protocol,
        namespace: payload.routers[index].namespace,
        httpCount: payload.routers[index].httpCount,
        createdAt: payload.routers[index].createdAt,
        resourceversion: payload.routers[index].resourceversion || '',
        text: payload.routers[index].name + ' (' + payload.routers[index].protocol  + ')',
        value: payload.routers[index].name + ',' + payload.routers[index].namespace
      })
    }
  }

  state.routerMenu = routers;
}

const Router_FilterItems = (state, payload) => {
  let routers = [];
  if (payload != '') {
    routers = state.routerMenu.map(ele => {
      if (ele.name.indexOf(payload) >= 0) {
        return ele;
      }
    }).filter(notUndefined => notUndefined !== undefined);
    
    state.routers = routers;
    state.meta =  { page: 1, limit: 0, total: routers.length }
  } else {
    state.routers = state.routerMenu;
    state.meta =  { page: 1, limit: 0, total: state.routerMenu.length }
  }
}

const Router_GetMappings = (state, payload) => {
  let mappings = [];
  if (payload.mappings) {
    for (let index in payload.mappings) {
      mappings.push(payload.mappings[index]);
      if (payload.gateways.length) {
        for (let i in payload.gateways) {
          let gateway = payload.gateways[i]
          if (payload.mappings[index].name === gateway.name && payload.mappings[index].namespace === gateway.namespace) {
            mappings[index] = gateway; // replace mapping data
            break;
          }
        }
      }
    }
  }

  state.mappings = mappings;
  state.mappingResourceVersion = payload.resourceVersion || '';
}

const Router_GetRule = (state, payload) => {
  state.httpItems = payload.httpItems;
  state.ruleResourceVersion = payload.resourceVersion || '';
}

const Router_AddHttp = (state) => {
  state.httpItems.push({
    prefixs: [''],
    headers: [],
    rewrite: '',
    fixedDelay: '',
    timeout: '',
    destinations: [{
      host: '',
      port: 80,
      weight: 0,
      subset: ''
    }] 
  });
}

const Router_RemoveHttp = (state, payload) => {
  state.httpItems.splice(payload.httpIndex, 1);
}

const Router_AddPrefix = (state, payload) => {
  state.httpItems[payload.httpIndex].prefixs.push('');
}

const Router_RemovePrefix = (state, payload) => {
  state.httpItems[payload.httpIndex].prefixs.splice(payload.index, 1);
}

const Router_UpdatePrefix = (state, payload) => {
  state.httpItems[payload.httpIndex].prefixs[payload.index] = payload.value;
}

const Router_AddHeader = (state, payload) => {
  state.httpItems[payload.httpIndex].headers.push({
    key: payload.key,
    value: payload.value
  });
}

const Router_RemoveHeader = (state, payload) => {
  state.httpItems[payload.httpIndex].headers.splice(payload.index, 1);
}

const Router_UpdateHeader = (state, payload) => {
  state.httpItems[payload.httpIndex].headers[payload.index][payload.key] = payload.value;
}

const Router_UpdateHttp = (state, payload) => {
  state.httpItems[payload.httpIndex][payload.key] = payload.value;

  if (payload.key === 'fixedDelay' || payload.key === 'timeout') state.httpItems[payload.httpIndex][payload.key] = Number(payload.value);
}

const Router_AddDestination = (state, payload) => {
  state.httpItems[payload.httpIndex].destinations.push({
    host:  '',
    port:  80,
    weight:  0,
    subset:  '',
  });
}

const Router_RemoveDestination = (state, payload) => {
  state.httpItems[payload.httpIndex].destinations.splice(payload.index, 1);
}

const Router_UpdateDestination = (state, payload) => {
  state.httpItems[payload.httpIndex].destinations[payload.index][payload.key] = payload.value;

  if (payload.key === 'port' || payload.key === 'weight') state.httpItems[payload.httpIndex].destinations[payload.index][payload.key] = Number(payload.value);
}


const Router_GetSuccessRate = (state, payload) => {
  let metrics = [];
  for (let i=0; i<=23; i++) {
    if (payload.metrics[i]) {
      if (moment.unix(payload.metrics[i].timestamp).format('H') == i) {
        metrics.push(formatNumber(payload.metrics[i].value, 1));
        continue;
      }
    }

    metrics.push(0);
  }

  let failRequest = 0;
  if (payload.totalReqests > 0) {
    failRequest = payload.totalReqests - payload.totalSuccessReqests;
  }

  state.successRate = metrics;
  state.successAvg = formatNumber(payload.successRate, 1);
  state.successRequest = payload.totalSuccessReqests || 0;
  state.failRequest = failRequest;
  state.totalReqest = payload.totalReqests || 0;
}

const Router_GetHourSuccessRate = (state, payload) => {
  let failRequest = 0;
  if (payload.totalReqests > 0) {
    failRequest = payload.totalReqests - payload.totalSuccessReqests;
    if (failRequest < 0) failRequest = 0;
  }

  state.successHourAvg = formatNumber(payload.successRate, 1);
  state.successHourRequest = payload.totalSuccessReqests || 0;
  state.failHourRequest = failRequest;
  state.totalHourReqest = payload.totalReqests || 0;
}

const Router_GetLatency = (state, payload) => {
  let metrics = [];
  for (let i=0; i<=23; i++) {
    if (payload.metrics[i]) {
      if (moment.unix(payload.metrics[i].timestamp).format('H') == i) {
        metrics.push(formatNumber(payload.metrics[i].value, 1));
        continue;
      }
    }

    metrics.push(0);
  }

  state.latency= metrics;
}

const Router_GetOPS = (state, payload) => {
  let metrics = [];
  for (let i=0; i<=23; i++) {
    if (payload.metrics[i]) {
      if (moment.unix(payload.metrics[i].timestamp).format('H') == i) {
        metrics.push(formatNumber(payload.metrics[i].value, 1));
        continue;
      }
    }

    metrics.push(0);
  }

  state.ops= metrics;
}

const Router_GetGrafana = (state, payload) => {
  state.grafana = {
    configured: false,
    provider: 'grafana',
    host: '',
    port: '',
    token: '',
    datasourceId: '1',
    isTls: false,
    skipTlsVerify: false,
    ...(payload.grafana || {}),
  };
}

export default {
  Router_ResetStatus,
  Router_SetStatus,
  Router_GetItem,
  Router_GetItems,
  Router_GetMenuItems,
  Router_GetMappings,
  Router_GetRule,
  Router_AddHttp,
  Router_RemoveHttp,
  Router_AddPrefix,
  Router_RemovePrefix,
  Router_UpdatePrefix,
  Router_AddHeader,
  Router_RemoveHeader,
  Router_UpdateHeader,
  Router_UpdateHttp,
  Router_AddDestination,
  Router_RemoveDestination,
  Router_UpdateDestination,
  Router_GetSuccessRate,
  Router_GetHourSuccessRate,
  Router_GetLatency,
  Router_GetOPS,
  Router_GetGrafana,
  Router_FilterItems
}
