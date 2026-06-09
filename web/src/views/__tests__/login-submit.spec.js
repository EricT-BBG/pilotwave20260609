import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, expect, it } from 'vitest';

const srcRoot = resolve(import.meta.dirname, '../..');

describe('login submit behavior', () => {
  it('submits when Enter is pressed in the password field', () => {
    const source = readFileSync(resolve(srcRoot, 'views/Landingpage.vue'), 'utf8');
    const passwordInput = source.match(/data-testid="login-password"[\s\S]*?\/>/)?.[0] || '';

    expect(passwordInput).toContain('@keydown.enter.prevent="submit"');
  });
});
