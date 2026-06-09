import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { readdirSync, statSync } from 'node:fs';
import { resolve } from 'node:path';
import enMessages from '../../plugins/translate/en/lang';
import twMessages from '../../plugins/translate/tw/lang';

const srcRoot = resolve(import.meta.dirname, '../..');

const visibleEnglishLiterals = [
  ['views/router/Router.vue', 'subtitle="Istio virtual service"'],
  ['views/router/Router.vue', 'empty-title="No routers have been created"'],
  ['views/gateway/Gateway.vue', 'subtitle="Istio gateway service"'],
  ['views/gateway/Gateway.vue', 'empty-title="No gateways have been created"'],
  ['views/policy/Authpolicy.vue', 'subtitle="Istio Authorization Policy"'],
  ['views/user/User.vue', 'subtitle="Pilotwave user account service"'],
  ['components/Template.vue', '<p class="eyebrow">Language</p>'],
  ['components/Template.vue', '<p class="eyebrow">Sign Out</p>'],
  ['components/Template.vue', '<dt>Version</dt>'],
  ['views/Home.vue', '<p class="eyebrow">Istio virtual service overview</p>'],
  ['views/Home.vue', 'No routers have been created'],
  ['views/Home.vue', 'Create the first router'],
  ['views/Home.vue', 'No router to select'],
  ['views/Home.vue', '第一個 Router'],
  ['views/Home.vue', '<p class="eyebrow">Ready For Setup</p>'],
  ['components/gateway/Dashboard.vue', '<p class="dashboard-title">TLS ready ports</p>'],
  ['components/gateway/Cytoscape.vue', '<span class="card-kicker">Router</span>'],
  ['views/gateway/GatewayDetail.vue', 'Loading gateway detail...'],
  ['views/router/RouterDetail.vue', 'Loading router detail...'],
  ['views/policy/NewAuthpolicy.vue', '← Back'],
  ['components/policy/EditSetting.vue', 'No labels configured.'],
  ['components/auth/LabelItem.vue', 'Delete</button>'],
  ['components/policy/RuleItem.vue', 'No when conditions configured.'],
  ['views/user/User.vue', '>Prev</button>'],
  ['views/user/User.vue', '>Next</button>'],
  ['components/ResourceListPage.vue', '>Close</button>'],
  ['components/ResourceListPage.vue', "'Selected'"],
  ['components/ResourceListPage.vue', "'Select'"],
  ['components/ResourceListPage.vue', "'Ready For Setup'"],
  ['components/ResourceListPage.vue', "'No Matches'"],
  ['components/ResourceListPage.vue', "'Istio is unavailable'"],
  ['components/ResourceListPage.vue', 'No ${this.title} configured yet'],
  ['components/router/GatewaySetting.vue', '<option value="All">All</option>'],
  ['components/router/GatewaySetting.vue', 'Router gateway mapping changed in Kubernetes. Reload before submitting again.'],
  ['components/router/RouterSetting.vue', 'Router rules changed in Kubernetes. Reload before submitting again.'],
  ['components/router/Cytoscape.vue', "$t('Gateway.Router')"],
  ['components/router/GatewaySetting.vue', 'No gateways associated.'],
  ['components/router/GatewaySetting.vue', '>Total:'],
  ['components/gateway/RouterSetting.vue', '<option value="All">All</option>'],
  ['components/gateway/RouterSetting.vue', 'Gateway VirtualService mapping changed in Kubernetes. Reload before submitting again.'],
  ['components/gateway/RouterSetting.vue', 'No routers associated.'],
  ['components/gateway/RouterSetting.vue', '>Total:'],
];

describe('visible UI copy localization', () => {
  it('does not leave known customer-visible English literals outside i18n', () => {
    const remaining = visibleEnglishLiterals.filter(([file, literal]) => {
      const source = readFileSync(resolve(srcRoot, file), 'utf8');
      return source.includes(literal);
    });

    expect(remaining).toEqual([]);
  });

  it('defines static translation keys used by Vue templates', () => {
    const files = collectVueFiles(srcRoot);
    const usedKeys = files.flatMap((file) => {
      const source = readFileSync(file, 'utf8');
      return Array.from(source.matchAll(/\$t\(['"]([^'"]+)['"]\)/g), (match) => match[1]);
    });

    const missingKeys = [...new Set(usedKeys)].filter((key) => !enMessages[key] || !twMessages[key]);

    expect(missingKeys).toEqual([]);
  });

  it('labels the Istio route resource as VirtualService in primary UI copy', () => {
    expect(twMessages['System.ServiceGateway']).toBe('Gateway 管理');
    expect(enMessages['System.ServiceRouter']).toBe('VirtualService Management');
    expect(enMessages['Router.New']).toBe('New VirtualService');
    expect(enMessages['Router.Remove']).toBe('Remove VirtualService');
    expect(enMessages['Router.EmptyTitle']).toBe('No VirtualServices yet');
    expect(enMessages['Gateway.Router']).toBe('VirtualService');

    expect(twMessages['System.ServiceRouter']).toBe('VirtualService 管理');
    expect(twMessages['Router.New']).toBe('新增 VirtualService');
    expect(twMessages['Router.Remove']).toBe('移除 VirtualService');
    expect(twMessages['Router.EmptyTitle']).toBe('尚未建立 VirtualService');
    expect(twMessages['Gateway.Router']).toBe('VirtualService');
  });
});

function collectVueFiles(dir) {
  return readdirSync(dir).flatMap((entry) => {
    const fullPath = resolve(dir, entry);
    const stat = statSync(fullPath);
    if (stat.isDirectory()) return collectVueFiles(fullPath);
    return fullPath.endsWith('.vue') ? [fullPath] : [];
  });
}
