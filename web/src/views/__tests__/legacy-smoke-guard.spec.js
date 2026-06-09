import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

const currentDir = dirname(fileURLToPath(import.meta.url));
const srcRoot = resolve(currentDir, '../..');

const smokePathFiles = [
  'App.vue',
  'components/Navigation.vue',
  'components/ResourceListPage.vue',
  'components/Template.vue',
  'views/Home.vue',
  'views/Landingpage.vue',
  'views/auth/Requestauth.vue',
  'views/gateway/Gateway.vue',
  'views/policy/Authpolicy.vue',
  'views/router/Router.vue',
  'views/user/User.vue'
];

const legacyPatterns = [
  {
    name: 'Vuetify v-* component tags',
    pattern: /<\/?v-[a-z]/i
  },
  {
    name: 'Vuelidate validationMixin',
    pattern: /\bvalidationMixin\b/
  },
  {
    name: 'Vuelidate package imports',
    pattern: /from\s+['"]vuelidate(?:\/lib\/validators)?['"]/
  },
  {
    name: 'Vuelidate $v proxy',
    pattern: /\$v\b/
  }
];

describe('Vue 3 smoke path legacy guard', () => {
  it.each(smokePathFiles)('%s does not reintroduce compatibility APIs', (relativePath) => {
    const source = readFileSync(resolve(srcRoot, relativePath), 'utf8');
    const violations = legacyPatterns
      .filter(({ pattern }) => pattern.test(source))
      .map(({ name }) => name);

    expect(violations).toEqual([]);
  });
});
