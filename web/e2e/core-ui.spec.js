import { test, expect } from '@playwright/test';

const apiBase = '/api/v1';
const adminUser = {
  uid: 'e2e-admin',
  username: 'admin',
  name: 'Pilotwave Admin',
  email: 'admin@example.test',
  permissions: ['admin'],
  token: 'e2e-admin-token',
};
const viewerUser = {
  uid: 'e2e-viewer',
  username: 'viewer',
  name: 'Pilotwave Viewer',
  email: 'viewer@example.test',
  permissions: [],
  token: 'e2e-viewer-token',
};
const routerItem = {
  id: 'router-1',
  name: 'checkout-router',
  namespace: 'pilotwave-e2e',
  hosts: ['checkout.pilotwave.local'],
  protocol: 'http',
};

function metricAtHour(value, hour = 23) {
  const date = new Date();
  date.setHours(hour, 0, 0, 0);
  const metrics = Array.from({ length: 24 }, () => null);
  metrics[hour] = {
    timestamp: Math.floor(date.getTime() / 1000),
    value,
  };
  return metrics;
}

async function mockCoreApi(page) {
  await page.route(`**${apiBase}/auth/signin`, async (route) => {
    const request = route.request();
    const payload = request.postDataJSON();
    const user = payload.username === 'viewer' ? viewerUser : adminUser;

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(user),
    });
  });

  await page.route(`**${apiBase}/namespaces`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        namespaces: ['default', 'pilotwave-e2e'],
        items: [
          {
            name: 'default',
            systemNamespace: true,
            istioInjection: { mode: 'disabled', status: 'disabled' },
          },
          {
            name: 'pilotwave-e2e',
            labels: { 'istio-injection': 'enabled' },
            istioInjection: { mode: 'enabled', status: 'enabled' },
          },
        ],
      }),
    });
  });

  await page.route(`**${apiBase}/cluster/capabilities`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        istio: {
          installed: true,
          disabled: false,
          revisions: ['test-revision'],
          revisionTags: [],
        },
      }),
    });
  });

  await page.route(`**${apiBase}/routers**`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        routers: [routerItem],
        meta: { page: 1, limit: 20, total: 1 },
      }),
    });
  });

  await page.route(`**${apiBase}/grafana`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: 'monitoring-1',
        configured: true,
        provider: 'prometheus',
        host: 'prometheus.monitoring.svc.cluster.local',
        port: '9090',
        datasourceId: '',
        isTls: false,
        skipTlsVerify: false,
      }),
    });
  });

  await page.route(`**${apiBase}/router/*/*/successrate**`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        metrics: metricAtHour(99.4),
        successRate: 99.4,
        totalSuccessReqests: 497,
        totalReqests: 500,
      }),
    });
  });

  await page.route(`**${apiBase}/router/*/*/latency**`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        metrics: metricAtHour(123.4),
      }),
    });
  });

  await page.route(`**${apiBase}/router/*/*/ops**`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        metrics: metricAtHour(42),
      }),
    });
  });

  await page.route(`**${apiBase}/monitoring/test`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ ok: true, message: 'Monitoring source is reachable.' }),
    });
  });

  await page.route(`**${apiBase}/grafanas`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ status: 'update_success' }),
    });
  });
}

async function signIn(page, account = 'admin') {
  await page.goto('/');
  await page.getByTestId('login-account').fill(account);
  await page.getByTestId('login-password').fill('admin');
  await Promise.all([
    page.waitForURL('**/dashboard'),
    page.getByTestId('login-submit').click(),
  ]);
}

test.describe('core auth session and monitoring UI', () => {
  test.beforeEach(async ({ page }) => {
    await mockCoreApi(page);
  });

  test('redirects protected routes to login until a session signs in and signs out cleanly', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByRole('heading', { name: /Sign in to your account|登入你的個人帳號/ })).toBeVisible();

    await signIn(page);
    await expect(page).toHaveURL(/\/dashboard$/);
    await expect(page.getByRole('heading', { name: /Dashboard|儀表板/ })).toBeVisible();
    await expect(page.getByText(/Monitoring Source|監控資料來源/)).toBeVisible();
    await expect(page.evaluate(() => sessionStorage.getItem('accessToken'))).resolves.toBe(adminUser.token);

    await page.getByTestId('topbar-logout-open').click();
    await expect(page.getByTestId('logout-confirm-dialog')).toBeVisible();
    await page.getByTestId('logout-confirm-dialog').getByRole('button', { name: /Signout|登出/ }).click();

    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByRole('heading', { name: /Sign in to your account|登入你的個人帳號/ })).toBeVisible();
    await expect(page.evaluate(() => sessionStorage.getItem('accessToken'))).resolves.toBeNull();
  });

  test('keeps non-admin sessions away from user management', async ({ page }) => {
    await signIn(page, 'viewer');

    await expect(page.getByText(/Monitoring Source|監控資料來源/)).toHaveCount(0);
    await page.goto('/users');

    await expect(page).toHaveURL(/\/dashboard$/);
    await expect(page.getByRole('heading', { name: /Dashboard|儀表板/ })).toBeVisible();
    await expect(page.getByText(/Account Management|帳號管理/)).toHaveCount(0);
  });

  test('shows dashboard router metrics and validates the monitoring source dialog', async ({ page }) => {
    await signIn(page);

    await expect(page.getByRole('heading', { name: /Dashboard|儀表板/ })).toBeVisible();
    await expect(page.getByLabel(/Select Router|選擇路由/)).toHaveValue('pilotwave-e2e/checkout-router');
    await expect(page.getByText('99.4%')).toHaveCount(2);
    await expect(page.getByText(/500 .*Request times today|500 .*今日連線呼叫次數/)).toBeVisible();
    await expect(page.getByText(/123.4/)).toBeVisible();
    await expect(page.getByText(/42/)).toBeVisible();

    await page.getByText(/Monitoring Source|監控資料來源/).click();
    await expect(page.getByRole('heading', { name: /Monitoring Source|監控資料來源/ })).toBeVisible();
    await expect(page.getByText('http://prometheus.monitoring.svc.cluster.local:9090')).toBeVisible();

    await page.getByRole('button', { name: /Test connection|測試連線/ }).click();
    await expect(page.getByText('Monitoring source is reachable.')).toBeVisible();
  });
});
