import { v4 as uuidv4 } from 'uuid';

const isTlsProtocol = (protocol) => ['HTTPS', 'TLS'].includes(String(protocol).toUpperCase());

const Gateway_ResetStatus = (state) => {
  state.status = '';
  state.error_handle = '';
}

const Gateway_SetStatus = (state, payload) => {
  state.status = payload.status;
  state.error_handle = payload.error_handle;
}

const Gateway_GetItem = (state, payload) => {
  let servers = [];
  const gateway = payload.gateway || {};
  for (let i in gateway.servers || []) {
    let server = {
      hosts: gateway.servers[i].hosts,
      ports: gateway.servers[i].ports,
    } 

    servers.push(server);
  }

  state.servers = servers;
  state.gateway = gateway;
  state.selectorMatchLabels = gateway.selectormatchlabels || {};
  state.resourceVersion = gateway.resourceversion || '';
}

const Gateway_GetItems = (state, payload) => {
  let gateways = [];
  if (payload.gateways) {
    for (let index in payload.gateways) {
      let servers = payload.gateways[index].servers;
      let hosts = [];
      let hostNames = [];
      let ports = [];
      for (let i in servers) {
        let h = servers[i].hosts.join(', ');
        hosts.push(h);
        hostNames = hostNames.concat(servers[i].hosts);

        let portItems = servers[i].ports;
        for (let j in portItems) {
          ports.push(portItems[j].port + ' (' + portItems[j].protocol + ')')
        }
      }

      gateways.push({
        id: uuidv4(), // For istio api
        name: payload.gateways[index].name,
        hosts: hosts.join(' ; '),
        hostNames: hostNames,
        hostsCount: hosts.length,
        ports: ports.join(', '),
        namespace: payload.gateways[index].namespace,
        createdAt: payload.gateways[index].createdAt,
        resourceversion: payload.gateways[index].resourceversion || '',
        text: payload.gateways[index].name,
        value: payload.gateways[index].name + ',' + payload.gateways[index].namespace
      })
    }
  }

  state.gateways = gateways;
  state.meta = payload.meta;
}

const Gateway_FilterItems = (state, payload) => {
  let gateways = [];
  if (payload != '') {
    gateways = state.gatewayMenu.map(ele => {
      if (ele.name.indexOf(payload) >= 0) {
        return ele;
      }
    }).filter(notUndefined => notUndefined !== undefined);
    
    state.gateways = gateways;
    state.meta =  { page: 1, limit: 0, total: gateways.length }
  } else {
    state.gateways = state.gatewayMenu;
    state.meta =  { page: 1, limit: 0, total: state.gatewayMenu.length }
  }
}

const Gateway_GetMenuItems = (state, payload) => {
  let gateways = [];
  if (payload.gateways) {
    for (let index in payload.gateways) {
      let servers = payload.gateways[index].servers;
      let hosts = [];
      let hostNames = [];
      let ports = [];
      for (let i in servers) {
        let h = servers[i].hosts.join(', ');
        hosts.push(h);
        hostNames = hostNames.concat(servers[i].hosts);

        let portItems = servers[i].ports;
        for (let j in portItems) {
          ports.push(portItems[j].port + ' (' + portItems[j].protocol + ')')
        }
      }

      gateways.push({
        id: uuidv4(), // For istio api
        name: payload.gateways[index].name,
        hosts: hosts.join(' ; '),
        hostNames: hostNames,
        hostsCount: hosts.length,
        ports: ports.join(', '),
        namespace: payload.gateways[index].namespace,
        createdAt: payload.gateways[index].createdAt,
        text: payload.gateways[index].name,
        value: payload.gateways[index].name + ',' + payload.gateways[index].namespace
      })
    }
  }

  state.gatewayMenu = gateways;
}

const Gateway_GetTLSCertificates = (state, payload) => {
  state.tlsCertificates = payload.certificates || [];
}

const Gateway_GetBlackWhiteList = (state, payload) => {
  let lists = [];
  if (payload.lists) {
    for (let index in payload.lists) {
      lists.push({
        id: payload.lists[index].id,
        domain: payload.lists[index].domain,
        description: payload.lists[index].description,
        category: payload.lists[index].category,
        createdAt: payload.lists[index].createdAt,
      })
    }
  }

  state.bwlist = lists;
  state.meta = payload.meta;
}

const Gateway_GetMappings = (state, payload) => {
  let mappings = [];
  if (payload.mappings) {
    for (let index in payload.mappings) {
      mappings.push(payload.mappings[index]);
      if (payload.routers.length) {
        for (let i in payload.routers) {
          let router = payload.routers[i]
          if (payload.mappings[index].name === router.name && payload.mappings[index].namespace === router.namespace) {
            mappings[index] = router; // replace mapping data
            break;
          }
        }
      }
    }
  }

  state.mappings = mappings;
  state.mappingResourceVersions = payload.resourceVersions || {};
}

const Gateway_UpdateHosts = (state, payload) => {
  state.servers[payload.serverIndex].hosts = payload.hosts;
}

const Gateway_AddServers = (state) => {
  state.servers.push({
    hosts: [],
    ports: [{
      protocol: 'HTTP',
      port: 80,
      cert: '',
      pkey: '',
      cacert: '',
      credentialname: '',
      mode: ''
    }]
  });
}

const Gateway_RemoveServer = (state, payload) => {
  state.servers.splice(payload.serverIndex, 1);
}

const Gateway_AddPorts = (state, payload) => {
  state.servers[payload.serverIndex].ports.push({
    protocol: 'HTTP',
    port: 80,
    cert: '',
    pkey: '',
    cacert: '',
    credentialname: '',
    mode: ''
  })
}

const Gateway_RemovePort = (state, payload) => {
  state.servers[payload.serverIndex].ports.splice(payload.index, 1);
}

const Gateway_UpdatePort = (state, payload) => {
  let value = payload.value;
  if (payload.key === 'port') value = Number(payload.value);

  state.servers[payload.serverIndex].ports[payload.index][payload.key] = value;

  if (payload.key === 'cert' || payload.key === 'pkey' || payload.key === 'cacert') {
    state.servers[payload.serverIndex].ports[payload.index].name = '';
    if (!state.servers[payload.serverIndex].ports[payload.index].mode) {
      state.servers[payload.serverIndex].ports[payload.index].mode = 'SIMPLE';
    }
    // state.servers[payload.serverIndex].ports[payload.index].credentialname = '';
  }

  if (payload.key === 'protocol' && !isTlsProtocol(payload.value)) {
    state.servers[payload.serverIndex].ports[payload.index].cert = '';
    state.servers[payload.serverIndex].ports[payload.index].pkey = '';
    state.servers[payload.serverIndex].ports[payload.index].cacert = '';
    state.servers[payload.serverIndex].ports[payload.index].name = '';
    state.servers[payload.serverIndex].ports[payload.index].mode = '';
    state.servers[payload.serverIndex].ports[payload.index].credentialname = '';
  }
}

const Gateway_ResetCert = (state, payload) => {
    state.servers[payload.serverIndex].ports[payload.index].cert = '';
    state.servers[payload.serverIndex].ports[payload.index].pkey = '';
    state.servers[payload.serverIndex].ports[payload.index].cacert = '';
    state.servers[payload.serverIndex].ports[payload.index].name = '';
    state.servers[payload.serverIndex].ports[payload.index].mode = 'SIMPLE';
    // state.servers[payload.serverIndex].ports[payload.index].credentialname = '';
}

const Gateway_ResetData = (state) => {
  state.gateway = '';
  state.selectorMatchLabels = {};
  state.resourceVersion = '';
  state.tlsCertificates = [];
  state.servers = [
    {
      hosts: [],
      ports: [{
        protocol: 'HTTP',
        port: 80,
        cert: '',
        pkey: '',
        cacert: '',
        name: '',
        credentialname: '',
        mode: ''
      }]
    }
  ]
}

export default {
  Gateway_ResetStatus,
  Gateway_SetStatus,
  Gateway_GetItem,
  Gateway_GetItems,
  Gateway_GetMenuItems,
  Gateway_GetTLSCertificates,
  Gateway_GetBlackWhiteList,
  Gateway_GetMappings,
  Gateway_UpdateHosts,
  Gateway_AddServers,
  Gateway_RemoveServer,
  Gateway_AddPorts,
  Gateway_RemovePort,
  Gateway_UpdatePort,
  Gateway_ResetCert,
  Gateway_ResetData,
  Gateway_FilterItems
}
