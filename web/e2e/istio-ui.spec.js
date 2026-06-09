import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { test, expect } from '@playwright/test';

const context = process.env.ISTIO_CONTEXT || 'colima-legacy-1-18';
const namespace = process.env.ISTIO_E2E_NAMESPACE || 'pilotwave-e2e-istio';
const injectionNamespace = process.env.ISTIO_E2E_INJECTION_NAMESPACE || 'pilotwave-e2e-injection';
const testRevision = process.env.ISTIO_E2E_TEST_REVISION || 'test-revision';
const gatewayName = 'pw-e2e-gw';
const tlsGatewayName = 'pw-e2e-tls-gw';
const listGatewayAlphaName = 'pw-e2e-gw-alpha';
const listGatewayCharlieName = 'pw-e2e-gw-charlie';
const routerName = 'pw-e2e-router';
const authPolicyName = 'pw-e2e-authz';
const requestAuthName = 'pw-e2e-requestauth';
const host = 'pw-e2e.pilotwave.local';
const tlsHost = 'pw-e2e-tls.pilotwave.local';
const updatedGatewayHost = 'pw-e2e-updated.pilotwave.local';
const updatedRouterHost = 'pw-e2e-router-updated.pilotwave.local';
const listGatewayAlphaHost = 'alpha.pw-e2e.pilotwave.local';
const listGatewayCharlieHost = 'charlie.pw-e2e.pilotwave.local';
const preservedLabelKey = 'pilotwave.io/e2e-preserve';
const preservedLabelValue = 'keep-label';
const preservedAnnotationKey = 'pilotwave.io/e2e-note';
const preservedAnnotationValue = 'keep-annotation';
const preservedRouterHeaderValue = 'keep-header';
const preservedDestinationRuleExportTo = '.';
const preservedDestinationRuleMaxConnections = 7;
const tlsSecretNamespace = process.env.ISTIO_E2E_TLS_SECRET_NAMESPACE || 'istio-system';
const userManagedTLSSecretName = 'pw-e2e-user-managed-tls';
const apiBase = '/api/v1';
const sensitiveNamespaces = new Set(['default', 'kube-system', 'kube-public', 'kube-node-lease', 'istio-system']);
const injectionLabelKeys = ['istio-injection', 'istio.io/rev'];
let createdNamespaces = new Set();
let originalInjectionLabels = {};

function fixture(name) {
  return readFileSync(new URL(`./fixtures/${name}`, import.meta.url), 'utf8').trim();
}

const tlsCertA = fixture('tls-cert-a.pem');
const tlsKeyA = fixture('tls-key-a.pem');
const tlsCertB = fixture('tls-cert-b.pem');
const tlsKeyB = fixture('tls-key-b.pem');
const userManagedTLSCert = fixture('tls-user-managed-cert.pem');
const userManagedTLSKey = fixture('tls-user-managed-key.pem');

function kubectl(args, options = {}) {
  const output = execFileSync('kubectl', ['--context', context, ...args], {
    encoding: 'utf8',
    stdio: options.stdio || (options.input ? ['pipe', 'pipe', 'pipe'] : ['ignore', 'pipe', 'pipe']),
    input: options.input
  });
  return typeof output === 'string' ? output.trim() : '';
}

function namespaceExists(name) {
  try {
    kubectl(['get', 'namespace', name], { stdio: 'pipe' });
    return true;
  } catch {
    return false;
  }
}

function namespacePhase(name) {
  try {
    return kubectl(['get', 'namespace', name, '-o', 'jsonpath={.status.phase}'], { stdio: 'pipe' });
  } catch {
    return '';
  }
}

function waitForNamespaceDeleted(name) {
  for (let i = 0; i < 120; i += 1) {
    if (!namespaceExists(name)) return;
    execFileSync('sleep', ['1']);
  }
  throw new Error(`Timed out waiting for namespace deletion: ${name}`);
}

function ensureSafeNamespace(name, purpose) {
  if (sensitiveNamespaces.has(name)) {
    throw new Error(`${purpose} namespace must not be a sensitive namespace: ${name}`);
  }
}

function ensureNamespace(name) {
  ensureSafeNamespace(name, 'E2E');
  if (namespaceExists(name)) {
    if (namespacePhase(name) !== 'Terminating') return;
    waitForNamespaceDeleted(name);
  }

  kubectl(['create', 'namespace', name], { stdio: 'ignore' });
  createdNamespaces.add(name);
}

function getNamespaceLabel(name, key) {
  try {
    const value = kubectl(['get', 'namespace', name, '-o', `go-template={{ index .metadata.labels "${key}" }}`]);
    return value === '<no value>' ? '' : value;
  } catch {
    return '';
  }
}

function saveNamespaceLabels(name) {
  originalInjectionLabels[name] = {};
  for (const key of injectionLabelKeys) {
    originalInjectionLabels[name][key] = getNamespaceLabel(name, key);
  }
}

function removeNamespaceLabel(name, key) {
  try {
    kubectl(['label', 'namespace', name, `${key}-`, '--overwrite'], { stdio: 'ignore' });
  } catch {
    // Removing an absent namespace label is safe to ignore during setup/cleanup.
  }
}

function restoreNamespaceLabels(name) {
  if (createdNamespaces.has(name)) {
    kubectl(['delete', 'namespace', name, '--ignore-not-found=true', '--wait=false'], { stdio: 'ignore' });
    return;
  }

  const labels = originalInjectionLabels[name] || {};
  for (const key of injectionLabelKeys) {
    const value = labels[key];
    if (value) {
      kubectl(['label', 'namespace', name, `${key}=${value}`, '--overwrite'], { stdio: 'ignore' });
    } else {
      removeNamespaceLabel(name, key);
    }
  }
}

function prepareNamespaces() {
  ensureNamespace(namespace);
  ensureNamespace(injectionNamespace);
  saveNamespaceLabels(injectionNamespace);
  for (const key of injectionLabelKeys) {
    removeNamespaceLabel(injectionNamespace, key);
  }
}

function cleanup() {
  const deletes = [
    ['-n', namespace, 'delete', 'gateway', gatewayName, '--ignore-not-found=true'],
    ['-n', namespace, 'delete', 'gateway', tlsGatewayName, '--ignore-not-found=true'],
    ['-n', namespace, 'delete', 'gateway', listGatewayAlphaName, '--ignore-not-found=true'],
    ['-n', namespace, 'delete', 'gateway', listGatewayCharlieName, '--ignore-not-found=true'],
    ['-n', namespace, 'delete', 'virtualservice', routerName, '--ignore-not-found=true'],
    ['-n', namespace, 'delete', 'destinationrule', routerName, '--ignore-not-found=true'],
    ['-n', namespace, 'delete', 'authorizationpolicy', authPolicyName, '--ignore-not-found=true'],
    ['-n', namespace, 'delete', 'requestauthentication', requestAuthName, '--ignore-not-found=true'],
    ['-n', tlsSecretNamespace, 'delete', 'secret', userManagedTLSSecretName, '--ignore-not-found=true']
  ];

  for (const args of deletes) {
    kubectl(args, { stdio: 'ignore' });
  }

  deleteManagedTLSSecretsForGateway(gatewayName);
  deleteManagedTLSSecretsForGateway(tlsGatewayName);
}

function encodePEM(value) {
  return Buffer.from(value, 'utf8').toString('base64');
}

function tlsSecretPrefixForGateway(name) {
  return `pilotwave-${name}-${tlsSecretNamespace}-port-`;
}

function managedTLSSecretsForGateway(name) {
  try {
    const output = kubectl(['-n', tlsSecretNamespace, 'get', 'secrets', '-o', 'jsonpath={.items[*].metadata.name}']);
    return output
      .split(/\s+/)
      .filter((item) => item.startsWith(tlsSecretPrefixForGateway(name)))
      .sort();
  } catch {
    return [];
  }
}

function deleteManagedTLSSecretsForGateway(name) {
  for (const secretName of managedTLSSecretsForGateway(name)) {
    kubectl(['-n', tlsSecretNamespace, 'delete', 'secret', secretName, '--ignore-not-found=true'], { stdio: 'ignore' });
  }
}

function secretExists(name) {
  try {
    kubectl(['-n', tlsSecretNamespace, 'get', 'secret', name], { stdio: 'pipe' });
    return true;
  } catch {
    return false;
  }
}

function applyTLSSecret(name, cert, key) {
  const secret = {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name,
      namespace: tlsSecretNamespace
    },
    type: 'kubernetes.io/tls',
    stringData: {
      'tls.crt': cert,
      'tls.key': key
    }
  };

  kubectl(['apply', '-f', '-'], {
    input: JSON.stringify(secret),
    stdio: ['pipe', 'pipe', 'pipe']
  });
}

function gatewayResourceVersion(name) {
  return kubectl([
    '-n', namespace,
    'get', 'gateway', name,
    '-o', 'jsonpath={.metadata.resourceVersion}'
  ]);
}

function gatewayCredentialName(name) {
  return kubectl([
    '-n', namespace,
    'get', 'gateway', name,
    '-o', 'jsonpath={.spec.servers[0].tls.credentialName}'
  ]);
}

function resourceVersion(kind, name) {
  return kubectl([
    '-n', namespace,
    'get', kind, name,
    '-o', 'jsonpath={.metadata.resourceVersion}'
  ]);
}

async function expectOnlyManagedTLSSecretForGateway(name) {
  await expect.poll(() => managedTLSSecretsForGateway(name)).toHaveLength(1);
  return managedTLSSecretsForGateway(name)[0];
}

function patchIstioResource(kind, name, patch) {
  kubectl([
    '-n', namespace,
    'patch', kind, name,
    '--type=merge',
    '-p', JSON.stringify(patch)
  ], { stdio: 'ignore' });
}

function getMetadataValue(kind, name, field, key) {
  const value = kubectl([
    '-n', namespace,
    'get', kind, name,
    '-o', `go-template={{ index .metadata.${field} "${key}" }}`
  ]);
  return value === '<no value>' ? '' : value;
}

function getIstioResource(kind, name) {
  return JSON.parse(kubectl([
    '-n', namespace,
    'get', kind, name,
    '-o', 'json'
  ]));
}

function decorateGatewayForPreservation() {
  patchIstioResource('gateway', gatewayName, {
    metadata: {
      labels: {
        [preservedLabelKey]: preservedLabelValue
      },
      annotations: {
        [preservedAnnotationKey]: preservedAnnotationValue
      }
    }
  });
}

function decorateRouterForPreservation() {
  patchIstioResource('virtualservice', routerName, {
    metadata: {
      labels: {
        [preservedLabelKey]: preservedLabelValue
      },
      annotations: {
        [preservedAnnotationKey]: preservedAnnotationValue
      }
    },
    spec: {
      gateways: ['mesh', `${namespace}/${gatewayName}`],
      http: [{
        name: 'preserve-primary',
        headers: {
          request: {
            set: {
              'x-pilotwave-e2e': preservedRouterHeaderValue
            }
          }
        },
        route: [{
          destination: {
            host: 'localhost'
          }
        }],
        timeout: '5s'
      }]
    }
  });
}

function decorateDestinationRuleForPreservation() {
  patchIstioResource('destinationrule', routerName, {
    metadata: {
      labels: {
        [preservedLabelKey]: preservedLabelValue
      },
      annotations: {
        [preservedAnnotationKey]: preservedAnnotationValue
      }
    },
    spec: {
      exportTo: [preservedDestinationRuleExportTo],
      trafficPolicy: {
        connectionPool: {
          tcp: {
            maxConnections: preservedDestinationRuleMaxConnections
          }
        },
        loadBalancer: {
          simple: 'LEAST_CONN'
        }
      }
    }
  });
}

function decorateSecurityResourceForPreservation(kind, name) {
  patchIstioResource(kind, name, {
    metadata: {
      labels: {
        [preservedLabelKey]: preservedLabelValue
      },
      annotations: {
        [preservedAnnotationKey]: preservedAnnotationValue
      }
    }
  });
}

async function expectPreservedMetadata(kind, name) {
  await expect.poll(() => getMetadataValue(kind, name, 'labels', preservedLabelKey)).toBe(preservedLabelValue);
  await expect.poll(() => getMetadataValue(kind, name, 'annotations', preservedAnnotationKey)).toBe(preservedAnnotationValue);
}

async function expectPreservedDestinationRuleSpec() {
  await expect.poll(() => getIstioResource('destinationrule', routerName).spec.exportTo || [])
    .toContain(preservedDestinationRuleExportTo);

  await expect.poll(() => {
    const item = getIstioResource('destinationrule', routerName);
    return item.spec.trafficPolicy?.connectionPool?.tcp?.maxConnections || 0;
  }).toBe(preservedDestinationRuleMaxConnections);

  await expect.poll(() => {
    const item = getIstioResource('destinationrule', routerName);
    return item.spec.trafficPolicy?.loadBalancer?.simple || '';
  }).toBe('LEAST_CONN');
}

async function expectPreservedRouterSpec() {
  await expect.poll(() => getIstioResource('virtualservice', routerName).spec.gateways || [])
    .toContain(`${namespace}/${gatewayName}`);

  await expect.poll(() => {
    const item = getIstioResource('virtualservice', routerName);
    return item.spec.http?.[0]?.headers?.request?.set?.['x-pilotwave-e2e'] || '';
  }).toBe(preservedRouterHeaderValue);

  await expect.poll(() => getIstioResource('virtualservice', routerName).spec.http?.[0]?.timeout || '')
    .toBe('5s');
}

async function expectNamespaceOptions(page, selector, expectedNamespaces) {
  const options = page.locator(selector).locator('option');
  await expect.poll(async () => options.count()).toBeGreaterThanOrEqual(expectedNamespaces.length);

  const actual = await options.evaluateAll((nodes) => nodes.map((node) => node.value || node.textContent.trim()));
  for (const expected of expectedNamespaces) {
    expect(actual).toContain(expected);
  }
}

async function openNamespaceMenu(page) {
  const trigger = page.getByTestId('namespace-menu-open');
  if (await page.getByTestId('namespace-menu').isVisible().catch(() => false)) return;
  await trigger.click();
  await expect(page.getByTestId('namespace-menu')).toBeVisible();
}

function namespaceMenuOption(page, name) {
  return page
    .getByTestId('namespace-menu')
    .locator('.namespace-menu-option')
    .filter({ hasText: name });
}

async function expectNamespaceMenuOptions(page, expectedNamespaces) {
  await openNamespaceMenu(page);
  const namespaceLabels = page.getByTestId('namespace-menu').locator('.namespace-menu-option-main strong');
  await expect.poll(async () => namespaceLabels.count()).toBeGreaterThanOrEqual(expectedNamespaces.length + 1);

  const actual = await namespaceLabels.evaluateAll((nodes) => nodes.map((node) => node.textContent.trim()));
  for (const expected of expectedNamespaces) {
    expect(actual).toContain(expected);
  }
}

function serverHostsInput(page, serverIndex) {
  return page.locator(`[data-testid="gateway-server-hosts"][data-server-index="${serverIndex}"]`);
}

function gatewayPortProtocol(page, serverIndex, portIndex) {
  return page.locator(`[data-testid="gateway-port-protocol"][data-server-index="${serverIndex}"][data-port-index="${portIndex}"]`);
}

function gatewayPortNumber(page, serverIndex, portIndex) {
  return page.locator(`[data-testid="gateway-port-number"][data-server-index="${serverIndex}"][data-port-index="${portIndex}"]`);
}

function resourceSortHeader(page, columnKey) {
  return page.getByTestId(`resource-sort-${columnKey}`).locator('xpath=ancestor::th');
}

async function login(page) {
  await page.goto('/');
  await page.getByTestId('login-account').fill(process.env.E2E_USERNAME || 'admin');
  await page.getByTestId('login-password').fill(process.env.E2E_PASSWORD || 'admin');
  await Promise.all([
    page.waitForURL('**/dashboard'),
    page.getByTestId('login-submit').click()
  ]);
  await expect(page).toHaveURL(/\/dashboard$/);
}

async function authHeaders(page) {
  const token = await page.evaluate(() => sessionStorage.getItem('accessToken'));
  return {
    Authentication: token,
    'Content-Type': 'application/json',
    Accept: 'application/json'
  };
}

function tlsGatewayPayload({ cert, pkey, credentialname, resourceversion } = {}) {
  const port = {
    protocol: 'HTTPS',
    port: 443
  };

  if (credentialname) {
    port.credentialname = credentialname;
  } else {
    port.cert = encodePEM(cert);
    port.pkey = encodePEM(pkey);
  }

  return {
    name: tlsGatewayName,
    namespace,
    servers: [{
      hosts: [tlsHost],
      ports: [port]
    }],
    selectormatchlabels: {
      istio: 'ingressgateway'
    },
    ...(resourceversion ? { resourceversion } : {})
  };
}

async function createTLSGatewayByAPI(page, cert, pkey) {
  const response = await page.request.post(`${apiBase}/gateways`, {
    headers: await authHeaders(page),
    data: tlsGatewayPayload({ cert, pkey })
  });
  const body = await response.json();
  expect(response.status(), JSON.stringify(body)).toBe(200);
  expect(body.status).toBe('create_success');
}

async function createGatewayByAPI(page, { name, servers }) {
  const response = await page.request.post(`${apiBase}/gateways`, {
    headers: await authHeaders(page),
    data: {
      name,
      namespace,
      servers,
    }
  });
  await expectSuccessResponse(response, 'create_success');
}

async function expectSuccessResponse(response, expectedStatus) {
  const body = await response.json();
  expect(response.status(), JSON.stringify(body)).toBe(200);
  expect(body.status).toBe(expectedStatus);
}

async function updateTLSGatewayByAPI(page, options) {
  const response = await page.request.put(`${apiBase}/gateway/${namespace}/${tlsGatewayName}`, {
    headers: await authHeaders(page),
    data: tlsGatewayPayload({
      ...options,
      resourceversion: gatewayResourceVersion(tlsGatewayName)
    })
  });
  const body = await response.json();
  expect(response.status(), JSON.stringify(body)).toBe(200);
  expect(body.status).toBe('update_success');
}

async function openTLSGatewayUpdateDialogFromCertificateDetail(page) {
  await page.goto('/tls-certificates');

  const row = page.getByTestId('tls-certificate-row').filter({ hasText: tlsHost }).first();
  await expect(row).toContainText(tlsGatewayName);
  await row.click();

  await expect(page).toHaveURL(/\/tls-certificates\/.+/);
  await expect(page.getByTestId('tls-certificate-detail')).toContainText(tlsGatewayName);
  await expect(page.getByTestId('tls-certificate-detail')).toContainText(tlsHost);

  await page.getByRole('button', { name: /Renew \/ Update|更新 \/ 換發/ }).click();
  await expect(page).toHaveURL(new RegExp(`/gateway/${tlsGatewayName}.*tab=setting`));

  await page.getByRole('button', { name: /Configure TLS|設定 TLS/ }).click();
  const dialog = page.getByRole('dialog', { name: /Configure TLS Credential|設定 TLS 憑證/ });
  await expect(dialog).toBeVisible();
  await expect(page.getByTestId('tls-current-certificate')).toContainText(tlsHost);
  await expect(page.getByTestId('tls-current-certificate')).toContainText(await gatewayCredentialName(tlsGatewayName));

  return dialog;
}

async function createRouterByAPI(page) {
  const response = await page.request.post(`${apiBase}/routers`, {
    headers: await authHeaders(page),
    data: {
      name: routerName,
      namespace,
      hosts: [host],
      protocol: 'http'
    }
  });
  await expectSuccessResponse(response, 'create_success');
}

async function updateRouterRulesByAPI(page, subset) {
  const response = await page.request.put(`${apiBase}/router/${namespace}/${routerName}/rules`, {
    headers: await authHeaders(page),
    data: {
      https: [{
        prefixs: ['/e2e'],
        destinations: [{
          host: routerName,
          port: 80,
          weight: 100,
          subset
        }]
      }]
    }
  });
  await expectSuccessResponse(response, 'update_success');
}

async function createAuthorizationPolicyByAPI(page) {
  const response = await page.request.post(`${apiBase}/security/authpolicies`, {
    headers: await authHeaders(page),
    data: {
      name: authPolicyName,
      namespace,
      action: 'allow',
      rules: [{
        from: [{
          source: {
            ipBlocks: ['0.0.0.0/0']
          }
        }],
        to: [{
          operation: {
            paths: ['/e2e']
          }
        }]
      }]
    }
  });
  await expectSuccessResponse(response, 'create_success');
}

async function updateAuthorizationPolicyByAPI(page) {
  const response = await page.request.put(`${apiBase}/security/authpolicy/${namespace}/${authPolicyName}`, {
    headers: await authHeaders(page),
    data: {
      action: 'deny',
      resourceversion: resourceVersion('authorizationpolicy', authPolicyName),
      rules: [{
        from: [{
          source: {
            namespaces: [namespace]
          }
        }],
        to: [{
          operation: {
            paths: ['/e2e-updated']
          }
        }]
      }]
    }
  });
  await expectSuccessResponse(response, 'update_success');
}

async function createRequestAuthenticationByAPI(page) {
  const response = await page.request.post(`${apiBase}/security/requestauths`, {
    headers: await authHeaders(page),
    data: {
      name: requestAuthName,
      namespace,
      jwtRules: [{
        issuer: 'testing@pilotwave',
        jwksUri: 'https://raw.githubusercontent.com/istio/istio/release-1.7/security/tools/jwt/samples/jwks.json'
      }]
    }
  });
  await expectSuccessResponse(response, 'create_success');
}

async function updateRequestAuthenticationByAPI(page) {
  const response = await page.request.put(`${apiBase}/security/requestauth/${namespace}/${requestAuthName}`, {
    headers: await authHeaders(page),
    data: {
      resourceversion: resourceVersion('requestauthentication', requestAuthName),
      jwtRules: [{
        issuer: 'testing-updated@pilotwave',
        jwksUri: 'https://raw.githubusercontent.com/istio/istio/release-1.7/security/tools/jwt/samples/jwks.json',
        audiences: ['pilotwave-e2e']
      }]
    }
  });
  await expectSuccessResponse(response, 'update_success');
}

async function verifyNamespaceRefresh(page) {
  const namespaces = kubectl(['get', 'namespace', '-o', 'jsonpath={.items[*].metadata.name}'])
    .split(/\s+/)
    .filter(Boolean);

  await page.goto('/gateways');
  await expectNamespaceMenuOptions(page, namespaces);

  const refreshResponse = page.waitForResponse((response) => (
    response.url().includes(`${apiBase}/namespaces`) &&
    response.request().method() === 'GET'
  ));
  await page.getByTestId('namespace-refresh').click();
  expect((await refreshResponse).status()).toBe(200);
  await expectNamespaceMenuOptions(page, namespaces);

  await page.goto('/new/gateway');
  await expectNamespaceOptions(page, '[data-testid="gateway-namespace"]', namespaces);

  await page.goto('/new/router');
  await expectNamespaceOptions(page, '[data-testid="router-namespace"]', namespaces);
}

function namespaceInjectionLabels() {
  return {
    injection: getNamespaceLabel(injectionNamespace, 'istio-injection'),
    revision: getNamespaceLabel(injectionNamespace, 'istio.io/rev')
  };
}

async function setNamespaceInjectionMode(page, mode, revision = '') {
  await page.goto('/dashboard');
  await page.getByTestId('nav-namespace-injection-open').click();
  await page.getByTestId(`namespace-injection-row-${injectionNamespace}`).click();
  await page.locator(`input[name="namespace-injection-mode"][value="${mode}"]`).check();
  if (mode === 'revision') {
    const revisionInput = page.getByTestId('namespace-injection-revision');
    await expect(revisionInput).toBeEnabled();
    await revisionInput.selectOption(revision);
    await expect(revisionInput).toHaveValue(revision);
  }
  const saveButton = page.getByTestId('namespace-injection-save');
  await expect(saveButton).toBeEnabled();
  await saveButton.click();
  await expect(page.getByTestId('namespace-injection-confirm-dialog')).toBeVisible();
  const confirmButton = page.getByTestId('namespace-injection-confirm-apply');
  await expect(confirmButton).toBeDisabled();
  const confirmationItems = page.getByTestId('namespace-injection-confirm-list').locator('input[type="checkbox"]');
  const confirmationCount = await confirmationItems.count();
  expect(confirmationCount).toBeGreaterThan(0);
  for (let index = 0; index < confirmationCount; index += 1) {
    await confirmationItems.nth(index).check();
  }
  await expect(confirmButton).toBeEnabled();
  const responsePromise = page.waitForResponse((response) => (
    response.url().includes(`/api/v1/namespace/${injectionNamespace}/istio-injection`) &&
    response.request().method() === 'PATCH'
  ));
  await confirmButton.click();
  const response = await responsePromise;
  expect(response.status()).toBe(200);
}

function expectedInjectionLabels(expectedMode) {
  if (expectedMode === 'enabled') {
    return { injection: 'enabled', revision: '' };
  }

  if (expectedMode === 'disabled') {
    return { injection: 'disabled', revision: '' };
  }

  if (expectedMode === 'revision') {
    return { injection: '', revision: testRevision };
  }

  throw new Error(`Unsupported expected injection mode: ${expectedMode}`);
}

async function expectInjectionLabels(expectedMode) {
  await expect.poll(() => namespaceInjectionLabels()).toEqual(expectedInjectionLabels(expectedMode));
}

async function createGateway(page) {
  await page.goto('/new/gateway');
  await page.getByTestId('gateway-name').fill(gatewayName);
  await page.getByTestId('gateway-namespace').selectOption(namespace);
  await serverHostsInput(page, 0).fill(host);
  await gatewayPortProtocol(page, 0, 0).selectOption('HTTP');
  await gatewayPortNumber(page, 0, 0).fill('80');
  await Promise.all([
    page.waitForURL('**/gateways'),
    page.getByTestId('gateway-submit').click()
  ]);

  const gatewayHost = kubectl([
    '-n', namespace,
    'get', 'gateway', gatewayName,
    '-o', 'jsonpath={.spec.servers[0].hosts[0]}'
  ]);
  expect(gatewayHost).toBe(host);
}

async function createMultiPortGatewayThroughUI(page) {
  await page.goto('/new/gateway');
  await page.getByTestId('gateway-name').fill(gatewayName);
  await page.getByTestId('gateway-namespace').selectOption(namespace);

  await expect(page.getByTestId('gateway-server-remove-0')).toBeDisabled();
  await expect(page.getByTestId('gateway-port-remove-0-0')).toHaveCount(0);

  await serverHostsInput(page, 0).fill(host);
  await gatewayPortProtocol(page, 0, 0).selectOption('HTTP');
  await gatewayPortNumber(page, 0, 0).fill('80');

  await page.getByTestId('gateway-server-add-port-0').click();
  await expect(gatewayPortNumber(page, 0, 1)).toBeVisible();
  await expect(page.getByTestId('gateway-port-remove-0-0')).toBeVisible();
  await expect(page.getByTestId('gateway-port-remove-0-1')).toBeVisible();

  await gatewayPortProtocol(page, 0, 1).selectOption('HTTP2');
  await gatewayPortNumber(page, 0, 1).fill('8080');
  await page.getByTestId('gateway-port-remove-0-1').click();
  await expect(gatewayPortNumber(page, 0, 1)).toHaveCount(0);
  await expect(page.getByTestId('gateway-port-remove-0-0')).toHaveCount(0);

  await page.getByTestId('gateway-server-add-port-0').click();
  await gatewayPortProtocol(page, 0, 1).selectOption('HTTP2');
  await gatewayPortNumber(page, 0, 1).fill('8080');

  await page.getByTestId('gateway-add-server').click();
  await expect(page.getByTestId('gateway-server-remove-0')).toBeEnabled();
  await expect(page.getByTestId('gateway-server-remove-1')).toBeEnabled();

  await serverHostsInput(page, 1).fill('delete-me.pw-e2e.pilotwave.local');
  await gatewayPortNumber(page, 1, 0).fill('9090');
  await page.getByTestId('gateway-server-remove-1').click();
  await expect(serverHostsInput(page, 1)).toHaveCount(0);
  await expect(page.getByTestId('gateway-server-remove-0')).toBeDisabled();

  await Promise.all([
    page.waitForURL('**/gateways'),
    page.getByTestId('gateway-submit').click()
  ]);

  await expect.poll(() => {
    const item = getIstioResource('gateway', gatewayName);
    return item.spec.servers.map((server) => ({
      hosts: server.hosts,
      port: `${server.port.number}:${server.port.protocol}`,
    }));
  }).toEqual([
    {
      hosts: [host],
      port: '80:HTTP',
    },
    {
      hosts: [host],
      port: '8080:HTTP2',
    }
  ]);
}

async function updateGateway(page) {
  await page.goto(`/gateway/${gatewayName}?name=${gatewayName}&namespace=${namespace}`);
  await page.getByTestId('gateway-tab-setting').click();
  await serverHostsInput(page, 0).fill(updatedGatewayHost);
  await page.getByTestId('gateway-update-submit').click();

  await expect.poll(() => kubectl([
    '-n', namespace,
    'get', 'gateway', gatewayName,
    '-o', 'jsonpath={.spec.servers[0].hosts[0]}'
  ])).toBe(updatedGatewayHost);
}

async function updateGatewayServerDeletionRules(page) {
  await page.goto(`/gateway/${gatewayName}?name=${gatewayName}&namespace=${namespace}`);
  await page.getByTestId('gateway-tab-setting').click();

  await expect(serverHostsInput(page, 0)).toHaveValue(host);
  await expect(gatewayPortNumber(page, 0, 0)).toHaveValue('80');
  await expect(gatewayPortNumber(page, 0, 1)).toHaveValue('8080');
  await expect(page.getByTestId('gateway-server-remove-0')).toBeDisabled();
  await page.getByTestId('gateway-add-server').click();
  await serverHostsInput(page, 1).fill('temporary.pw-e2e.pilotwave.local');
  await gatewayPortNumber(page, 1, 0).fill('9090');
  await expect(page.getByTestId('gateway-server-remove-0')).toBeEnabled();
  await expect(page.getByTestId('gateway-server-remove-1')).toBeEnabled();

  await page.getByTestId('gateway-server-remove-0').click();
  await expect(serverHostsInput(page, 1)).toHaveCount(0);
  await expect(serverHostsInput(page, 0)).toHaveValue('temporary.pw-e2e.pilotwave.local');
  await expect(page.getByTestId('gateway-server-remove-0')).toBeDisabled();

  const updateResponse = page.waitForResponse((response) => (
    response.url().includes(`${apiBase}/gateway/${namespace}/${gatewayName}`) &&
    response.request().method() === 'PUT'
  ));
  await page.getByTestId('gateway-update-submit').click();
  expect((await updateResponse).status()).toBe(200);
  await expect.poll(() => {
    const item = getIstioResource('gateway', gatewayName);
    return item.spec.servers.map((server) => ({
      host: server.hosts[0],
      port: server.port.number,
    }));
  }).toEqual([{
    host: 'temporary.pw-e2e.pilotwave.local',
    port: 9090,
  }]);
}

async function verifyGatewayUpdateConflict(page) {
  const response = await page.request.put(`${apiBase}/gateway/${namespace}/${gatewayName}`, {
    headers: await authHeaders(page),
    data: {
      servers: [{
        hosts: ['stale-gateway.pilotwave.local'],
        ports: [{ port: 80, protocol: 'HTTP' }]
      }],
      selectormatchlabels: {},
      resourceversion: 'stale-resource-version'
    }
  });

  const body = await response.json();
  expect(response.status(), JSON.stringify(body)).toBe(409);
  expect(body.error).toContain('Gateway was modified');
}

async function createRouter(page) {
  await page.goto('/new/router');
  await page.getByTestId('router-name').fill(routerName);
  await page.getByTestId('router-namespace').selectOption(namespace);
  await page.getByTestId('router-hosts').fill(host);
  await page.getByTestId('router-protocol').selectOption('http');
  await Promise.all([
    page.waitForURL('**/routers?page=1'),
    page.getByTestId('router-submit').click()
  ]);

  const virtualServiceHost = kubectl([
    '-n', namespace,
    'get', 'virtualservice', routerName,
    '-o', 'jsonpath={.spec.hosts[0]}'
  ]);
  expect(virtualServiceHost).toBe(host);
}

async function updateRouter(page) {
  await page.goto(`/router/${routerName}?name=${routerName}&namespace=${namespace}`);
  await page.getByTestId('router-tab-setting').click();
  await page.getByTestId('router-edit-hosts').fill(updatedRouterHost);
  await page.getByTestId('router-update-submit').click();

  await expect.poll(() => kubectl([
    '-n', namespace,
    'get', 'virtualservice', routerName,
    '-o', 'jsonpath={.spec.hosts[0]}'
  ])).toBe(updatedRouterHost);
}

async function verifyRouterUpdateConflict(page) {
  const response = await page.request.put(`${apiBase}/router/${namespace}/${routerName}`, {
    headers: await authHeaders(page),
    data: {
      protocol: 'http',
      hosts: ['stale-router.pilotwave.local'],
      resourceversion: 'stale-resource-version'
    }
  });

  const body = await response.json();
  expect(response.status(), JSON.stringify(body)).toBe(409);
  expect(body.error).toContain('Router was modified');
}

async function createAuthorizationPolicy(page) {
  await page.goto('/new/authpolicy');
  await page.getByTestId('authpolicy-name').fill(authPolicyName);
  await page.getByTestId('authpolicy-namespace').selectOption(namespace);
  await page.getByTestId('authpolicy-action').selectOption('audit');
  await page.getByTestId('authpolicy-from-ipBlocks').fill('0.0.0.0/0');
  await page.getByTestId('authpolicy-to-paths').fill('/e2e');
  await Promise.all([
    page.waitForURL('**/authpolicies'),
    page.getByTestId('authpolicy-submit').click()
  ]);

  const action = kubectl([
    '-n', namespace,
    'get', 'authorizationpolicy', authPolicyName,
    '-o', 'jsonpath={.spec.action}'
  ]);
  expect(action).toBe('AUDIT');
}

async function updateAuthorizationPolicy(page) {
  await page.goto(`/authpolicy/${authPolicyName}?name=${authPolicyName}&namespace=${namespace}`);
  await page.getByTestId('authpolicy-edit-action').selectOption('deny');
  await page.getByTestId('authpolicy-to-paths').fill('/e2e-updated');
  await page.getByTestId('authpolicy-update-submit').click();

  await expect.poll(() => kubectl([
    '-n', namespace,
    'get', 'authorizationpolicy', authPolicyName,
    '-o', 'jsonpath={.spec.action}'
  ])).toBe('DENY');

  await expect.poll(() => kubectl([
    '-n', namespace,
    'get', 'authorizationpolicy', authPolicyName,
    '-o', 'jsonpath={.spec.rules[0].to[0].operation.paths[0]}'
  ])).toBe('/e2e-updated');
}

async function createRequestAuthentication(page) {
  await page.goto('/new/requestauth');
  await page.getByTestId('requestauth-name').fill(requestAuthName);
  await page.getByTestId('requestauth-namespace').selectOption(namespace);
  await page.getByTestId('requestauth-rule-issuer').fill('testing@pilotwave');
  await page.getByTestId('requestauth-rule-jwks-uri').fill('https://raw.githubusercontent.com/istio/istio/release-1.7/security/tools/jwt/samples/jwks.json');
  await Promise.all([
    page.waitForURL('**/requestauths'),
    page.getByTestId('requestauth-submit').click()
  ]);

  const name = kubectl([
    '-n', namespace,
    'get', 'requestauthentication', requestAuthName,
    '-o', 'jsonpath={.metadata.name}'
  ]);
  expect(name).toBe(requestAuthName);
}

async function updateRequestAuthentication(page) {
  await page.goto(`/requestauth/${requestAuthName}?name=${requestAuthName}&namespace=${namespace}`);
  await page.getByTestId('requestauth-rule-issuer').fill('testing-updated@pilotwave');
  await page.getByTestId('requestauth-rule-audiences').fill('pilotwave-e2e');
  await page.getByTestId('requestauth-update-submit').click();

  await expect.poll(() => kubectl([
    '-n', namespace,
    'get', 'requestauthentication', requestAuthName,
    '-o', 'jsonpath={.spec.jwtRules[0].issuer}'
  ])).toBe('testing-updated@pilotwave');

  await expect.poll(() => kubectl([
    '-n', namespace,
    'get', 'requestauthentication', requestAuthName,
    '-o', 'jsonpath={.spec.jwtRules[0].audiences[0]}'
  ])).toBe('pilotwave-e2e');
}

async function deleteResourceFromList(page, listUrl, resourceName, resourceKind) {
  await page.goto(listUrl);
  await page.getByTestId('resource-search').fill(resourceName);
  await page.getByTestId(`resource-select-${namespace}-${resourceName}`).click();
  await page.getByTestId('resource-delete-open').click();
  await page.getByTestId('resource-delete-confirm').click();
  await expect.poll(() => {
    try {
      kubectl(['-n', namespace, 'get', resourceKind, resourceName], { stdio: 'pipe' });
      return 'exists';
    } catch {
      return 'deleted';
    }
  }).toBe('deleted');
}

async function expectGatewayListOrder(page, expectedNames) {
  const rows = page.locator('tbody .resource-row');
  await expect.poll(async () => rows.count()).toBeGreaterThanOrEqual(expectedNames.length);
  for (let index = 0; index < expectedNames.length; index += 1) {
    await expect(rows.nth(index).locator('.index-col')).toHaveText(String(index + 1));
    await expect(rows.nth(index).locator('td').nth(2)).toContainText(expectedNames[index]);
  }
}

test.beforeAll(() => {
  prepareNamespaces();
});

test.beforeEach(() => {
  cleanup();
});

test.afterEach(() => {
  cleanup();
});

test.afterAll(() => {
  cleanup();
  restoreNamespaceLabels(injectionNamespace);
  if (createdNamespaces.has(namespace)) {
    kubectl(['delete', 'namespace', namespace, '--ignore-not-found=true', '--wait=false'], { stdio: 'ignore' });
  }
});

test('web UI creates, updates, and deletes Istio resources through the UI', async ({ page }) => {
  await login(page);
  await verifyNamespaceRefresh(page);

  await createGateway(page);
  await createRouter(page);
  await createAuthorizationPolicy(page);
  await createRequestAuthentication(page);
  decorateGatewayForPreservation();
  decorateRouterForPreservation();

  await verifyGatewayUpdateConflict(page);
  await verifyRouterUpdateConflict(page);

  await updateGateway(page);
  await expectPreservedMetadata('gateway', gatewayName);

  await updateRouter(page);
  await expectPreservedMetadata('virtualservice', routerName);
  await expectPreservedRouterSpec();

  await updateAuthorizationPolicy(page);
  await updateRequestAuthentication(page);

  await deleteResourceFromList(page, '/requestauths', requestAuthName, 'requestauthentication');
  await deleteResourceFromList(page, '/authpolicies', authPolicyName, 'authorizationpolicy');
  await deleteResourceFromList(page, '/routers', routerName, 'virtualservice');
  await deleteResourceFromList(page, '/gateways', gatewayName, 'gateway');
});

test('web UI manages Gateway server and port add/remove rules', async ({ page }) => {
  await login(page);

  await createMultiPortGatewayThroughUI(page);
  await updateGatewayServerDeletionRules(page);
});

test('web UI sorts Gateway list rows while keeping visible row numbers sequential', async ({ page }) => {
  await login(page);
  await createGatewayByAPI(page, {
    name: listGatewayCharlieName,
    servers: [{
      hosts: [listGatewayCharlieHost],
      ports: [{ port: 8080, protocol: 'HTTP' }],
    }]
  });
  await createGatewayByAPI(page, {
    name: listGatewayAlphaName,
    servers: [{
      hosts: [listGatewayAlphaHost],
      ports: [{ port: 80, protocol: 'HTTP' }],
    }]
  });

  await page.goto('/gateways');
  await page.getByTestId('resource-search').fill('pw-e2e-gw-');
  await expect.poll(async () => page.locator('tbody .resource-row').count()).toBeGreaterThanOrEqual(2);

  await page.getByTestId('resource-sort-name').click();
  await expect(resourceSortHeader(page, 'name')).toHaveAttribute('aria-sort', 'ascending');
  await expectGatewayListOrder(page, [listGatewayAlphaName, listGatewayCharlieName]);

  await page.getByTestId('resource-sort-name').click();
  await expect(resourceSortHeader(page, 'name')).toHaveAttribute('aria-sort', 'descending');
  await expectGatewayListOrder(page, [listGatewayCharlieName, listGatewayAlphaName]);
});

test('API replaces Pilotwave-managed Gateway TLS certificates without deleting user-managed secrets', async ({ page }) => {
  await login(page);

  await createTLSGatewayByAPI(page, tlsCertA, tlsKeyA);
  const firstManagedSecret = await expectOnlyManagedTLSSecretForGateway(tlsGatewayName);
  await expect.poll(() => gatewayCredentialName(tlsGatewayName)).toBe(firstManagedSecret);

  await updateTLSGatewayByAPI(page, { cert: tlsCertB, pkey: tlsKeyB });
  const secondManagedSecret = await expectOnlyManagedTLSSecretForGateway(tlsGatewayName);
  expect(secondManagedSecret).not.toBe(firstManagedSecret);
  await expect.poll(() => secretExists(firstManagedSecret)).toBe(false);
  await expect.poll(() => gatewayCredentialName(tlsGatewayName)).toBe(secondManagedSecret);

  applyTLSSecret(userManagedTLSSecretName, userManagedTLSCert, userManagedTLSKey);
  await updateTLSGatewayByAPI(page, { credentialname: userManagedTLSSecretName });
  await expect.poll(() => secretExists(userManagedTLSSecretName)).toBe(true);
  await expect.poll(() => managedTLSSecretsForGateway(tlsGatewayName)).toHaveLength(0);
  await expect.poll(() => gatewayCredentialName(tlsGatewayName)).toBe(userManagedTLSSecretName);

  await updateTLSGatewayByAPI(page, { cert: tlsCertA, pkey: tlsKeyA });
  const finalManagedSecret = await expectOnlyManagedTLSSecretForGateway(tlsGatewayName);
  await expect.poll(() => secretExists(userManagedTLSSecretName)).toBe(true);
  await expect.poll(() => gatewayCredentialName(tlsGatewayName)).toBe(finalManagedSecret);
});

test('web UI opens TLS certificate detail and dismisses the update dialog without changing TLS settings', async ({ page }) => {
  await login(page);

  await createTLSGatewayByAPI(page, tlsCertA, tlsKeyA);
  const credentialName = await expectOnlyManagedTLSSecretForGateway(tlsGatewayName);

  let dialog = await openTLSGatewayUpdateDialogFromCertificateDetail(page);
  await dialog.locator('footer').getByRole('button', { name: /Cancel|取消/ }).click();
  await expect(dialog).toBeHidden();
  await expect.poll(() => gatewayCredentialName(tlsGatewayName)).toBe(credentialName);

  dialog = await openTLSGatewayUpdateDialogFromCertificateDetail(page);
  await page.keyboard.press('Escape');
  await expect(dialog).toBeHidden();
  await expect.poll(() => gatewayCredentialName(tlsGatewayName)).toBe(credentialName);
});

test('API updates preserve custom fields outside Pilotwave-managed Istio fragments', async ({ page }) => {
  await login(page);

  await createRouterByAPI(page);
  await updateRouterRulesByAPI(page, 'v1');
  decorateDestinationRuleForPreservation();

  await updateRouterRulesByAPI(page, 'v2');
  await expectPreservedMetadata('destinationrule', routerName);
  await expectPreservedDestinationRuleSpec();
  await expect.poll(() => getIstioResource('destinationrule', routerName).spec.subsets.map((item) => item.name))
    .toContain('v2');

  await createAuthorizationPolicyByAPI(page);
  decorateSecurityResourceForPreservation('authorizationpolicy', authPolicyName);
  await updateAuthorizationPolicyByAPI(page);
  await expectPreservedMetadata('authorizationpolicy', authPolicyName);
  await expect.poll(() => getIstioResource('authorizationpolicy', authPolicyName).spec.action || '')
    .toBe('DENY');

  await createRequestAuthenticationByAPI(page);
  decorateSecurityResourceForPreservation('requestauthentication', requestAuthName);
  await updateRequestAuthenticationByAPI(page);
  await expectPreservedMetadata('requestauthentication', requestAuthName);
  await expect.poll(() => getIstioResource('requestauthentication', requestAuthName).spec.jwtRules?.[0]?.issuer || '')
    .toBe('testing-updated@pilotwave');
});

test('web UI manages safe namespace Istio injection modes', async ({ page }) => {
  await login(page);
  await verifyNamespaceRefresh(page);

  await setNamespaceInjectionMode(page, 'enabled');
  await expectInjectionLabels('enabled');

  await setNamespaceInjectionMode(page, 'disabled');
  await expectInjectionLabels('disabled');

  await setNamespaceInjectionMode(page, 'revision', testRevision);
  await expectInjectionLabels('revision');
});

test('namespace injection editor preserves unsaved row drafts', async ({ page }) => {
  await login(page);
  await verifyNamespaceRefresh(page);

  await page.goto('/dashboard');
  await page.getByTestId('nav-namespace-injection-open').click();
  await page.getByTestId(`namespace-injection-row-${injectionNamespace}`).click();
  await page.locator('input[name="namespace-injection-mode"][value="enabled"]').check();
  await expect(page.locator('input[name="namespace-injection-mode"][value="enabled"]')).toBeChecked();

  await page.getByTestId(`namespace-injection-row-${namespace}`).click();
  await expect(page.locator('input[name="namespace-injection-mode"][value="enabled"]')).not.toBeChecked();

  await page.getByTestId(`namespace-injection-row-${injectionNamespace}`).click();
  await expect(page.locator('input[name="namespace-injection-mode"][value="enabled"]')).toBeChecked();

  await page.locator('input[name="namespace-injection-mode"][value="disabled"]').check();
  await page.getByTestId(`namespace-injection-row-${injectionNamespace}`).click();
  await expect(page.locator('input[name="namespace-injection-mode"][value="disabled"]')).toBeChecked();
});

test('namespace menu shows injection status, reloads, and persists selected namespace', async ({ page }) => {
  await login(page);
  await setNamespaceInjectionMode(page, 'enabled');

  await page.goto('/gateways');
  await openNamespaceMenu(page);

  const injectionOption = namespaceMenuOption(page, injectionNamespace);
  await expect(injectionOption).toBeVisible();
  await expect(injectionOption.locator('.namespace-status-badge')).toHaveText(/Injected|已注入/);
  await injectionOption.click();

  await expect(page.getByTestId('namespace-menu-open')).toContainText(injectionNamespace);
  await expect(page.getByTestId('namespace-menu-open').locator('.namespace-status-badge')).toHaveText(/Injected|已注入/);

  await openNamespaceMenu(page);
  await expect(namespaceMenuOption(page, injectionNamespace)).toHaveClass(/selected/);

  const refreshResponse = page.waitForResponse((response) => (
    response.url().includes(`${apiBase}/namespaces`) &&
    response.request().method() === 'GET'
  ));
  await page.getByTestId('namespace-refresh').click();
  expect((await refreshResponse).status()).toBe(200);
  await expect(page.getByTestId('namespace-menu-open')).toContainText(injectionNamespace);
});

test('global API unavailable alert appears when an API request receives no response', async ({ page }) => {
  await login(page);
  await page.route(`**${apiBase}/namespaces`, (route) => route.abort('failed'));

  await page.goto('/gateways');
  await openNamespaceMenu(page);
  await page.getByTestId('namespace-refresh').click();

  const alert = page.getByTestId('api-error-alert');
  await expect(alert).toBeVisible();
  await expect(alert).toContainText(/API service is not responding|API 服務沒有回應/);
});
