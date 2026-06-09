import { createI18n } from 'vue-i18n';
import en from './translate/en/lang';
import tw from './translate/tw/lang';
import { resolveLocale } from '../lib/locale';

export default createI18n({
  legacy: true,
  locale: resolveLocale(),
  fallbackLocale: 'en',
  messages: { en, tw }
});
