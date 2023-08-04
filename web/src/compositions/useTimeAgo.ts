import TimeAgo from 'javascript-time-ago';
import en from 'javascript-time-ago/locale/en.json';

import { getUserLanguage } from '~/utils/locale';

TimeAgo.addDefaultLocale(en);

const addedLocales = ['en'];

export default () => new TimeAgo(getUserLanguage());
export async function loadTimeAgoLocale(locale: string) {
  if (!addedLocales.includes(locale)) {
    const { default: timeAgoLocale } = await import(`~/assets/timeAgoLocales/${locale}.js`);
    TimeAgo.addLocale(timeAgoLocale);
    addedLocales.push(locale);
  }
}
