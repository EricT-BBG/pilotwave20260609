export const SUPPORTED_LOCALES = ['en', 'tw'];
export const DEFAULT_LOCALE = 'en';

export function normalizeLocale(locale) {
  const value = String(locale || '').trim().toLowerCase();
  if (!value) return '';
  if (value === 'tw' || value.startsWith('zh')) return 'tw';
  if (value.startsWith('en')) return 'en';
  return '';
}

export function detectBrowserLocale(navigatorLike = globalThis.navigator) {
  const candidates = [
    ...(Array.isArray(navigatorLike?.languages) ? navigatorLike.languages : []),
    navigatorLike?.language,
    navigatorLike?.userLanguage,
  ];

  for (const candidate of candidates) {
    const locale = normalizeLocale(candidate);
    if (SUPPORTED_LOCALES.includes(locale)) return locale;
  }

  return DEFAULT_LOCALE;
}

export function resolveLocale({
  storage = globalThis.sessionStorage,
  navigatorLike = globalThis.navigator,
} = {}) {
  try {
    const storedLocale = normalizeLocale(storage?.getItem('locale'));
    if (SUPPORTED_LOCALES.includes(storedLocale)) return storedLocale;
  } catch {
    return detectBrowserLocale(navigatorLike);
  }

  return detectBrowserLocale(navigatorLike);
}
