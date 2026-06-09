import { describe, expect, it } from 'vitest';

import { detectBrowserLocale, resolveLocale } from '../locale';

function storageWith(value) {
  return {
    getItem: () => value
  };
}

describe('locale resolution', () => {
  it('uses a supported stored locale before browser language', () => {
    expect(resolveLocale({
      storage: storageWith('en'),
      navigatorLike: { language: 'zh-TW', languages: ['zh-TW'] }
    })).toBe('en');
  });

  it('detects Traditional Chinese browser language for the login default', () => {
    expect(detectBrowserLocale({ language: 'zh-TW', languages: ['zh-TW', 'en-US'] })).toBe('tw');
    expect(detectBrowserLocale({ language: 'zh-Hant-TW', languages: ['zh-Hant-TW'] })).toBe('tw');
  });

  it('maps Simplified Chinese browser language to the available Chinese locale', () => {
    expect(resolveLocale({
      storage: storageWith(null),
      navigatorLike: { language: 'zh-CN', languages: ['zh-CN'] }
    })).toBe('tw');
  });

  it('falls back to English for unsupported browser languages', () => {
    expect(resolveLocale({
      storage: storageWith(null),
      navigatorLike: { language: 'ja-JP', languages: ['ja-JP'] }
    })).toBe('en');
  });
});
